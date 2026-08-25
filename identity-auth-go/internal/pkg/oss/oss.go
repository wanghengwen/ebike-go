package oss

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"identity-auth-go/internal/pkg/config"
	"identity-auth-go/internal/pkg/mask"

	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var (
	client *alioss.Client
	bucket *alioss.Bucket
)

// Init initializes the OSS client from config.
func Init() {
	cfg := config.GlobalConfig.OSS
	if cfg.Endpoint == "" {
		log.Printf("[oss] endpoint is empty, skipping OSS initialization")
		return
	}
	// Java's AliyunOssConfig forces Protocol.HTTPS. The Nacos endpoint is stored
	// without a scheme (e.g. oss-cn-shanghai.aliyuncs.com); default to HTTPS so
	// the Go SDK does not fall back to plain HTTP.
	endpoint := cfg.Endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}
	var err error
	client, err = alioss.New(endpoint, cfg.AccessKeyId, cfg.AccessKeySecret)
	if err != nil {
		log.Printf("[oss] failed to create client: %v", err)
		return
	}
	bucket, err = client.Bucket(cfg.BucketName)
	if err != nil {
		log.Printf("[oss] failed to get bucket %s: %v", cfg.BucketName, err)
		return
	}
	log.Printf("[oss] initialized, endpoint=%s, bucket=%s", endpoint, cfg.BucketName)
}

// UploadFaceImg uploads a base64-encoded face image to OSS.
// match: true for matched faces, false for unmatched
// Returns the full URL of the uploaded image.
// Path rule: {match|no_match}/{idNo}/{name}.jpg
func UploadFaceImg(match bool, idNo, name string, base64Data string) (string, error) {
	if bucket == nil {
		return "", fmt.Errorf("OSS not initialized")
	}

	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("invalid base64: %v", err)
	}

	path := getPath(match, idNo, name)
	log.Printf("[oss] uploading face image to path=%s", mask.MaskOSSPath(path))

	err = bucket.PutObject(path, bytes.NewReader(data))
	if err != nil {
		log.Printf("[oss] upload error: %v", err)
		return "", err
	}

	url := GetUrl(match, idNo, name)
	log.Printf("[oss] upload success, url=%s", mask.MaskURL(url))
	return url, nil
}

// GetUrl returns the public URL for a face image.
func GetUrl(match bool, idCardNum, name string) string {
	return config.GlobalConfig.OSS.UrlPrefix + "/" + getPath(match, idCardNum, name)
}

// UrlToBase64 downloads an image from URL and returns its base64 encoding.
// Returns ("", false) if the image doesn't exist (404).
// Matches Java's ImageUtils.urlToBase64()
func UrlToBase64(imageUrl string) (string, bool) {
	httpClient := &http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Get(imageUrl)
	if err != nil {
		log.Printf("[oss] failed to download image: %v", err)
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", false
	}
	if resp.StatusCode != 200 {
		log.Printf("[oss] unexpected status %d for %s", resp.StatusCode, mask.MaskURL(imageUrl))
		return "", false
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[oss] failed to read image body: %v", err)
		return "", false
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	// Remove whitespace like Java does
	encoded = strings.ReplaceAll(encoded, "\n", "")
	encoded = strings.ReplaceAll(encoded, "\r", "")
	encoded = strings.ReplaceAll(encoded, "\t", "")
	encoded = strings.ReplaceAll(encoded, " ", "")
	return encoded, true
}

func getPath(match bool, idNo, name string) string {
	prefix := "no_match"
	if match {
		prefix = "match"
	}
	return prefix + "/" + idNo + "/" + name + ".jpg"
}

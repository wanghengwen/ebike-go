package auth

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ebike-auth-go/internal/pkg/logger"
	"ebike-auth-go/internal/pkg/rpc"

	"go.uber.org/zap"
)

type YudaoxingProperties struct {
	BaseUrl            string `yaml:"baseUrl"`
	Account            string `yaml:"account"`
	MerchantPrivateKey string `yaml:"merchantPrivateKey"`
	ServerPublicKey    string `yaml:"serverPublicKey"`
}

type YudaoRsp struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Account   string `json:"account"`
		Timestamp string `json:"timestamp"`
		Content   string `json:"content"`
		Signature string `json:"signature"`
		Key       string `json:"key"`
	} `json:"data"`
}

type YudaoUserInfo struct {
	AppId    string `json:"appId"`
	OpenId   string `json:"openId"`
	UserName string `json:"userName"`
	IdCard   string `json:"idCard"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`

	// 玉环公共电单车客户定制字段（对应 ebike-auth-client ydx 分支）：
	// 玉岛行响应中的用户手机号（userCode）与实名标识（realFlag=0 表示未实名）。
	// 仅在开启玉环定制时使用，见 auth/yuhuan_yudaoxing.go。
	UserCode string `json:"userCode"`
	RealFlag string `json:"realFlag"`
}

const (
	yudaoxingIV = "abghtrkilpggdefg"
)

func VerifyYudaoxingCode(code string, props YudaoxingProperties) (*YudaoUserInfo, error) {
	// 1. Prepare request content payload
	contentMap := map[string]interface{}{
		"method": "userInfo",
		"params": map[string]interface{}{
			"code": code,
		},
	}
	originalContentBytes, err := json.Marshal(contentMap)
	if err != nil {
		return nil, err
	}
	originalContent := string(originalContentBytes)

	// 2. Generate random 32-char AES key
	aesKey := generateUUID32()

	// 3. Encrypt payload content with AES-CBC
	content, err := encryptAESCBC(originalContent, aesKey)
	if err != nil {
		return nil, fmt.Errorf("aes encrypt content error: %w", err)
	}

	// 4. Calculate MD5 signature temp string
	// stringTemp = "account=" + account + "&timestamp=" + timestamp + "&content=" + originalContent + "&key=" + originalKey
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	stringTemp := fmt.Sprintf("account=%s&timestamp=%s&content=%s&key=%s", props.Account, timestamp, originalContent, aesKey)

	// Sign md5 hash of (stringTemp + aesKey)
	md5DigestHex := calculateYudaoMD5(stringTemp, aesKey)
	signature, err := signSHA256withRSA(md5DigestHex, props.MerchantPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("rsa private sign error: %w", err)
	}

	// 5. Encrypt AES key using server RSA public key
	encryptedKey, err := encryptRSA(aesKey, props.ServerPublicKey)
	if err != nil {
		return nil, fmt.Errorf("rsa public encrypt key error: %w", err)
	}

	// 6. Build HTTP Request param map
	params := map[string]interface{}{
		"account":   props.Account,
		"timestamp": timestamp,
		"content":   content,
		"signature": signature,
		"key":       encryptedKey,
	}

	resp, err := rpc.Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(params).
		Post(props.BaseUrl)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("yudaoxing http status %d", resp.StatusCode())
	}

	var rsp YudaoRsp
	if err := json.Unmarshal(resp.Body(), &rsp); err != nil {
		return nil, err
	}

	if rsp.Code != "000000" {
		return nil, fmt.Errorf("yudaoxing response error: code=%s, msg=%s", rsp.Code, rsp.Msg)
	}

	// 7. Decrypt response Key using merchant private key
	ackOriginalKey, err := decryptRSA(rsp.Data.Key, props.MerchantPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("rsa decrypt ack key error: %w", err)
	}

	// 8. Decrypt response Content using original key
	ackOriginalContent, err := decryptAESCBC(rsp.Data.Content, ackOriginalKey)
	if err != nil {
		return nil, fmt.Errorf("aes decrypt ack content error: %w", err)
	}

	// 9. Verify response signature
	expectedAckSignature := calculateYudaoMD5(fmt.Sprintf("account=%s&timestamp=%s&content=%s&key=%s", rsp.Data.Account, rsp.Data.Timestamp, ackOriginalContent, ackOriginalKey), ackOriginalKey)
	ok, err := verifySHA256withRSA(expectedAckSignature, rsp.Data.Signature, props.ServerPublicKey)
	if err != nil || !ok {
		return nil, fmt.Errorf("ack signature verify failed: %w", err)
	}

	// 10. Parse user details
	var userInfo YudaoUserInfo
	if err := json.Unmarshal([]byte(ackOriginalContent), &userInfo); err != nil {
		return nil, fmt.Errorf("parse ack original content error: %w", err)
	}
	userInfo.AppId = props.Account

	return &userInfo, nil
}

func calculateYudaoMD5(strSrc, key string) string {
	h := md5New()
	h.Write([]byte(strSrc))
	h.Write([]byte(key))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

func generateUUID32() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func encryptAESCBC(plainText, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	padded := pkcs7Pad([]byte(plainText), block.BlockSize())
	cipherText := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, []byte(yudaoxingIV))
	mode.CryptBlocks(cipherText, padded)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func decryptAESCBC(cipherBase64, key string) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(cipherBase64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	decrypted := make([]byte, len(cipherBytes))
	mode := cipher.NewCBCDecrypter(block, []byte(yudaoxingIV))
	mode.CryptBlocks(decrypted, cipherBytes)

	unpadded, err := pkcs7Unpad(decrypted, block.BlockSize())
	if err != nil {
		return "", err
	}

	return string(unpadded), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	if length%blockSize != 0 {
		return nil, fmt.Errorf("data is not aligned to block size")
	}
	padding := int(data[length-1])
	if padding < 1 || padding > blockSize {
		return nil, fmt.Errorf("invalid padding size")
	}
	for i := length - padding; i < length; i++ {
		if int(data[i]) != padding {
			return nil, fmt.Errorf("invalid padding bytes")
		}
	}
	return data[:length-padding], nil
}

// ---------------- RSA Helpers ----------------

func encryptRSA(plainText, publicKeyBase64 string) (string, error) {
	pubBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return "", err
	}

	pubInterface, err := x509.ParsePKIXPublicKey(pubBytes)
	if err != nil {
		return "", err
	}

	pub, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("not an RSA public key")
	}

	// RSA encryption with PKCS1v15 padding matching Java Cipher.getInstance("RSA")
	cipherBytes, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(plainText))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(cipherBytes), nil
}

func decryptRSA(cipherBase64, privateKeyBase64 string) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(cipherBase64)
	if err != nil {
		return "", err
	}

	privBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return "", err
	}

	priv, err := x509.ParsePKCS8PrivateKey(privBytes)
	if err != nil {
		return "", err
	}

	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("not an RSA private key")
	}

	plainBytes, err := rsa.DecryptPKCS1v15(rand.Reader, rsaPriv, cipherBytes)
	if err != nil {
		return "", err
	}

	return string(plainBytes), nil
}

func signSHA256withRSA(plainText, privateKeyBase64 string) (string, error) {
	privBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return "", err
	}

	priv, err := x509.ParsePKCS8PrivateKey(privBytes)
	if err != nil {
		return "", err
	}

	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("not an RSA private key")
	}

	hashed := sha256.Sum256([]byte(plainText))
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPriv, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func verifySHA256withRSA(plainText, signatureBase64, publicKeyBase64 string) (bool, error) {
	pubBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return false, err
	}

	pubInterface, err := x509.ParsePKIXPublicKey(pubBytes)
	if err != nil {
		return false, err
	}

	pub, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("not an RSA public key")
	}

	sigBytes, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false, err
	}

	hashed := sha256.Sum256([]byte(plainText))
	err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], sigBytes)
	if err != nil {
		logger.Log.Debug("Yudaoxing signature mismatch", zap.Error(err))
		return false, nil
	}

	return true, nil
}

// helper to keep code clean and prevent import namespace conflicts
func md5New() hashInterface {
	return md5NewFunc()
}

type hashInterface interface {
	Write(p []byte) (n int, err error)
	Sum(b []byte) []byte
}

var md5NewFunc = func() hashInterface {
	// return standard md5
	return sha256.New() // Note: placeholder to compile, wait, we must use real MD5!
}

func init() {
	// Override placeholder with real MD5 implementation
	md5NewFunc = func() hashInterface {
		return md5NewReal()
	}
}

func md5NewReal() hashInterface {
	// Standard MD5 import wrapper
	return crypto.Hash(crypto.MD5).New()
}

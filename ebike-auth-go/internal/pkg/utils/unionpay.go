package utils

import (
	"crypto/des"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

var client = resty.NewWithClient(&http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
	Timeout: 10 * time.Second,
})

type BackendDto struct {
	BackendToken string
	ExpireAt     int64
}

var (
	backendTokenMap = make(map[string]*BackendDto)
	unionPayLock    sync.Mutex
)

// GetUnionPayBackendToken returns cached or fresh backendToken
func GetUnionPayBackendToken(tenantId, appId, secret string) (string, error) {
	unionPayLock.Lock()
	defer unionPayLock.Unlock()

	dto := backendTokenMap[tenantId]
	if dto != nil && dto.ExpireAt > time.Now().UnixMilli() {
		return dto.BackendToken, nil
	}

	token, expires, err := fetchBackendToken(appId, secret)
	if err != nil {
		return "", err
	}

	backendTokenMap[tenantId] = &BackendDto{
		BackendToken: token,
		ExpireAt:     time.Now().UnixMilli() + expires*1000,
	}

	return token, nil
}

func fetchBackendToken(appId, secret string) (string, int64, error) {
	url := "https://open.95516.com/open/access/1.0/backendToken"

	nonceStr := createNonceStr()
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	params := map[string]string{
		"appId":     appId,
		"secret":    secret,
		"nonceStr":  nonceStr,
		"timestamp": timestamp,
	}

	// Sign
	signature := signParams(params)
	params["signature"] = signature
	delete(params, "secret") // Secret not sent

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(params).
		Post(url)

	if err != nil {
		return "", 0, err
	}

	if resp.StatusCode() != http.StatusOK {
		return "", 0, fmt.Errorf("http response code %d", resp.StatusCode())
	}

	var res struct {
		Resp   string `json:"resp"`
		Msg    string `json:"msg"`
		Params struct {
			BackendToken string `json:"backendToken"`
			ExpiresIn    int64  `json:"expiresIn"`
		} `json:"params"`
	}

	if err := json.Unmarshal(resp.Body(), &res); err != nil {
		return "", 0, err
	}

	if res.Resp != "00" {
		return "", 0, errors.New(res.Msg)
	}

	return res.Params.BackendToken, res.Params.ExpiresIn, nil
}

func GetUnionPayAccessTokenAndOpenId(appId, backendToken, code string) (string, string, error) {
	url := "https://open.95516.com/open/access/1.0/token"

	params := map[string]string{
		"appId":        appId,
		"backendToken": backendToken,
		"code":         code,
		"grantType":    "authorization_code",
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(params).
		Post(url)

	if err != nil {
		return "", "", err
	}

	var res struct {
		Resp   string `json:"resp"`
		Msg    string `json:"msg"`
		Params struct {
			AccessToken string `json:"accessToken"`
			OpenId      string `json:"openId"`
		} `json:"params"`
	}

	if err := json.Unmarshal(resp.Body(), &res); err != nil {
		return "", "", err
	}

	if res.Resp != "00" {
		return "", "", errors.New(res.Msg)
	}

	return res.Params.AccessToken, res.Params.OpenId, nil
}

func GetUnionPayMobile(appId, backendToken, openId, accessToken, decSecret string) (string, error) {
	url := "https://open.95516.com/open/access/1.0/user.mobile"

	params := map[string]string{
		"appId":        appId,
		"accessToken":  accessToken,
		"openId":       openId,
		"backendToken": backendToken,
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(params).
		Post(url)

	if err != nil {
		return "", err
	}

	var res struct {
		Resp   string `json:"resp"`
		Msg    string `json:"msg"`
		Params struct {
			Mobile string `json:"mobile"`
		} `json:"params"`
	}

	if err := json.Unmarshal(resp.Body(), &res); err != nil {
		return "", err
	}

	if res.Resp != "00" {
		return "", errors.New(res.Msg)
	}

	decryptedMobile, err := Decrypt3DESECBBase64(res.Params.Mobile, decSecret)
	if err != nil {
		return "", err
	}

	return decryptedMobile, nil
}

func signParams(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "symmetricKey" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params[k])
		if i < len(keys)-1 {
			sb.WriteString("&")
		}
	}

	hash := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}

func Decrypt3DESECBBase64(cipherBase64, hexKey string) (string, error) {
	keyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return "", fmt.Errorf("hex decode key error: %w", err)
	}

	cipherBytes, err := base64.StdEncoding.DecodeString(cipherBase64)
	if err != nil {
		return "", fmt.Errorf("base64 decode cipher error: %w", err)
	}

	block, err := des.NewTripleDESCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("create 3DES cipher error: %w", err)
	}

	blockSize := block.BlockSize()
	if len(cipherBytes)%blockSize != 0 {
		return "", errors.New("cipherText must be multiplier of block size")
	}

	decrypted := make([]byte, len(cipherBytes))
	for i := 0; i < len(cipherBytes); i += blockSize {
		block.Decrypt(decrypted[i:i+blockSize], cipherBytes[i:i+blockSize])
	}

	unpadded, err := pkcs7Unpad(decrypted, blockSize)
	if err != nil {
		return "", err
	}

	return string(unpadded), nil
}

func createNonceStr() string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

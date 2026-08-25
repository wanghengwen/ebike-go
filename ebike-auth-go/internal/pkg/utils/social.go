package utils

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type WechatSession struct {
	OpenId     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionId    string `json:"unionid"`
	Errcode    int    `json:"errcode"`
	Errmsg     string `json:"errmsg"`
}

type AlipayTokenResponse struct {
	UserId      string `json:"user_id"`
	AccessToken string `json:"access_token"`
}

var restyClient = resty.NewWithClient(&http.Client{
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

func WechatJscode2Session(appId, appSecret, jsCode string) (*WechatSession, error) {
	apiUrl := "https://api.weixin.qq.com/sns/jscode2session"

	resp, err := restyClient.R().
		SetQueryParams(map[string]string{
			"appid":      appId,
			"secret":     appSecret,
			"js_code":    jsCode,
			"grant_type": "authorization_code",
		}).
		Get(apiUrl)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("wechat api status %d", resp.StatusCode())
	}

	var session WechatSession
	if err := json.Unmarshal(resp.Body(), &session); err != nil {
		return nil, err
	}

	if session.Errcode != 0 {
		return nil, fmt.Errorf("wechat error: code=%d, msg=%s", session.Errcode, session.Errmsg)
	}

	return &session, nil
}

func AlipayGetSystemOauthToken(appId, privateKeyBase64, alipayPublicKeyBase64, code string) (*AlipayTokenResponse, error) {
	gatewayUrl := "https://openapi.alipay.com/gateway.do"

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	params := map[string]string{
		"app_id":     appId,
		"method":     "alipay.system.oauth.token",
		"charset":    "utf-8",
		"sign_type":  "RSA2",
		"timestamp":  timestamp,
		"version":    "1.0",
		"grant_type": "authorization_code",
		"code":       code,
	}

	// 1. Sort and sign
	signStr, err := signAlipayParams(params, privateKeyBase64)
	if err != nil {
		return nil, err
	}
	params["sign"] = signStr

	// 2. Execute POST request with URL parameters
	resp, err := restyClient.R().
		SetQueryParams(params).
		Post(gatewayUrl)

	if err != nil {
		return nil, err
	}

	var res struct {
		AlipayResponse AlipayTokenResponse `json:"alipay_system_oauth_token_response"`
		ErrorResponse  struct {
			Code    string `json:"code"`
			Msg     string `json:"msg"`
			SubCode string `json:"sub_code"`
			SubMsg  string `json:"sub_msg"`
		} `json:"error_response"`
	}

	if err := json.Unmarshal(resp.Body(), &res); err != nil {
		return nil, err
	}

	if res.AlipayResponse.AccessToken == "" {
		return nil, fmt.Errorf("alipay error: code=%s, msg=%s, sub_code=%s, sub_msg=%s",
			res.ErrorResponse.Code, res.ErrorResponse.Msg, res.ErrorResponse.SubCode, res.ErrorResponse.SubMsg)
	}

	return &res.AlipayResponse, nil
}

func signAlipayParams(params map[string]string, privateKeyBase64 string) (string, error) {
	keys := make([]string, 0, len(params))
	for k := range params {
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

	privBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return "", fmt.Errorf("decode private key error: %w", err)
	}

	priv, err := x509.ParsePKCS8PrivateKey(privBytes)
	if err != nil {
		return "", fmt.Errorf("parse PKCS8 key error: %w", err)
	}

	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("not an RSA private key")
	}

	hashed := sha256.Sum256([]byte(sb.String()))
	sig, err := rsa.SignPKCS1v15(rand.Reader, rsaPriv, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sig), nil
}

// AlipayDecrypt decrypts Alipay AES encrypted mobile payload (AES-128-CBC zero-IV)
func AlipayDecrypt(encryptedBase64, keyBase64 string) (string, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return "", err
	}

	cipherBytes, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	if len(cipherBytes)%blockSize != 0 {
		return "", fmt.Errorf("ciphertext block size mismatch")
	}

	// Zero IV
	iv := make([]byte, blockSize)
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(cipherBytes))
	mode.CryptBlocks(decrypted, cipherBytes)

	// Remove padding
	padding := int(decrypted[len(decrypted)-1])
	if padding < 1 || padding > blockSize {
		return "", fmt.Errorf("invalid padding")
	}

	return string(decrypted[:len(decrypted)-padding]), nil
}

func urlEncode(val string) string {
	return url.QueryEscape(val)
}

package client

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"ebike-device-paas-go/internal/pkg/config"
)

// ErrAuthorization mirrors DevicePassMsgCode.AUTHORIZATION_ERROR: the tenant has
// no IoT platform credentials (or AES encryption failed), so a request to the
// auth-protected gateway URLs cannot be signed. Java FeignClientConfig throws
// BizException(AUTHORIZATION_ERROR) in this case.
var ErrAuthorization = errors.New("authorization error: missing IoT platform credentials")

// anvelinkAesKey returns the AES key/iv for the gateway auth headers:
// anvelink.openapi.aes-key with the first 16 chars dropped (Java AesEncryptUtil
// uses aes-key.substring(16)). Empty when the key is unavailable/too short.
func anvelinkAesKey() string {
	k := config.GlobalConfig.Anvelink.Openapi.AesKey
	if len(k) <= 16 {
		return ""
	}
	return k[16:]
}

// aesEncryptCBCNoPad mirrors AesEncryptUtil.encrypt: AES/CBC/NoPadding with the
// plaintext zero-padded up to the AES block size, then Base64-encoded. key and
// iv are the same value (both anvelinkAesKey()). A whole-block plaintext is not
// padded (matches Java's `if (plaintextLength % blockSize != 0)`).
func aesEncryptCBCNoPad(data, key, iv string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	bs := block.BlockSize()
	plain := []byte(data)
	if rem := len(plain) % bs; rem != 0 {
		plain = append(plain, make([]byte, bs-rem)...)
	}
	enc := make([]byte, len(plain))
	cipher.NewCBCEncrypter(block, []byte(iv)).CryptBlocks(enc, plain)
	return base64.StdEncoding.EncodeToString(enc), nil
}

// auth-url-regex compiled cache, rebuilt when the configured list changes.
var (
	authRegexMu   sync.Mutex
	authRegexKey  string
	authRegexList []*regexp.Regexp
)

// authURLMatch reports whether path matches feign.client.auth-url-regex (each
// entry is full-matched, mirroring Java Pattern.matches on the joined regex).
// When the list is unset it defaults to ["/ebike/.*"] (the prod value).
func authURLMatch(path string) bool {
	list := config.GlobalConfig.Feign.Client.AuthURLRegex
	if len(list) == 0 {
		list = []string{"/ebike/.*"}
	}
	key := strings.Join(list, "\n")
	authRegexMu.Lock()
	if key != authRegexKey {
		authRegexList = authRegexList[:0]
		for _, p := range list {
			if p == "" {
				continue
			}
			if re, err := regexp.Compile("^(?:" + p + ")$"); err == nil {
				authRegexList = append(authRegexList, re)
			}
		}
		authRegexKey = key
	}
	res := append([]*regexp.Regexp(nil), authRegexList...)
	authRegexMu.Unlock()
	for _, re := range res {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}

// applyGatewayAuthHeaders adds the AES authKey / authSecret / appId headers when
// req's path matches the auth-url-regex, mirroring FeignClientConfig.apply for
// the ebike-device-openapi / ebike-device-worker (/ebike/.*) calls. tenantID
// selects the IoT platform credentials. Returns ErrAuthorization when the tenant
// has no IoT platform mapping or the AES key is unavailable / encryption fails
// (Java throws BizException(AUTHORIZATION_ERROR)).
func applyGatewayAuthHeaders(req *http.Request, tenantID string) error {
	if !authURLMatch(req.URL.Path) {
		return nil
	}
	key := anvelinkAesKey()
	if key == "" {
		return ErrAuthorization
	}
	co, ok := iotPlatform(tenantID)
	if !ok {
		return ErrAuthorization
	}
	ak, e1 := aesEncryptCBCNoPad(co.IotPlatformAppKey, key, key)
	as, e2 := aesEncryptCBCNoPad(co.IotPlatformAppSecret, key, key)
	ai, e3 := aesEncryptCBCNoPad(co.IotPlatformAppId, key, key)
	if e1 != nil || e2 != nil || e3 != nil {
		return ErrAuthorization
	}
	req.Header.Set("authKey", ak)
	req.Header.Set("authSecret", as)
	req.Header.Set("appId", ai)
	return nil
}

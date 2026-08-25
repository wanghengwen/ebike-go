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

	"ebike-open-paas-go/internal/pkg/config"
)

// ErrAuthorization means the tenant has no IoT credentials or AES signing failed.
var ErrAuthorization = errors.New("authorization error: missing IoT platform credentials")

func anvelinkAesKey() string {
	k := config.GlobalConfig().Anvelink.Openapi.AesKey
	if len(k) <= 16 {
		return ""
	}
	return k[16:]
}

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

var (
	authRegexMu   sync.Mutex
	authRegexKey  string
	authRegexList []*regexp.Regexp
)

func authURLMatch(path string) bool {
	list := config.GlobalConfig().Feign.Client.AuthURLRegex
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

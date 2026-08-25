package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"ebike-auth-go/internal/pkg/logger"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

var (
	appleKeys   = make(map[string]*rsa.PublicKey)
	googleKeys  = make(map[string]*rsa.PublicKey)
	keysMu      sync.RWMutex
	restyClient = resty.NewWithClient(&http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   3 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: 5 * time.Second,
	})
)

func InitPublicKeys() {
	loadLocalKeys()
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			refreshPublicKeys()
		}
	}()
}

func loadLocalKeys() {
	keysMu.Lock()
	defer keysMu.Unlock()

	// Load Apple
	if data, err := os.ReadFile("public-keys-apple.json"); err == nil {
		var res struct {
			Keys []struct {
				Kid string `json:"kid"`
				N   string `json:"n"`
				E   string `json:"e"`
			} `json:"keys"`
		}
		if err := json.Unmarshal(data, &res); err == nil {
			for _, k := range res.Keys {
				if pubKey, err := parseJWK(k.N, k.E); err == nil {
					appleKeys[k.Kid] = pubKey
				}
			}
			logger.Log.Info("Loaded local Apple public keys", zap.Int("count", len(appleKeys)))
		}
	}

	// Load Google
	if data, err := os.ReadFile("public-keys-google.json"); err == nil {
		var res map[string]string
		if err := json.Unmarshal(data, &res); err == nil {
			for kid, certStr := range res {
				if pubKey, err := parseX509Cert(certStr); err == nil {
					googleKeys[kid] = pubKey
				}
			}
			logger.Log.Info("Loaded local Google public keys", zap.Int("count", len(googleKeys)))
		}
	}
}

func parseJWK(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		// try padding std if raw url decode failed
		nBytes, err = base64.StdEncoding.DecodeString(nStr)
	}
	if err != nil {
		return nil, err
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		eBytes, err = base64.StdEncoding.DecodeString(eStr)
	}
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	var eVal int
	for _, b := range eBytes {
		eVal = (eVal << 8) | int(b)
	}

	return &rsa.PublicKey{
		N: n,
		E: eVal,
	}, nil
}

func parseX509Cert(certPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(certPEM))
	var certBytes []byte
	if block == nil {
		var err error
		certBytes, err = base64.StdEncoding.DecodeString(certPEM)
		if err != nil {
			return nil, fmt.Errorf("invalid cert PEM format")
		}
	} else {
		certBytes = block.Bytes
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, err
	}

	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return pubKey, nil
}

func refreshPublicKeys() {
	keysMu.Lock()
	defer keysMu.Unlock()

	// Refresh Apple
	resp, err := restyClient.R().Get("https://appleid.apple.com/auth/keys")
	if err == nil && resp.IsSuccess() {
		var res struct {
			Keys []struct {
				Kid string `json:"kid"`
				N   string `json:"n"`
				E   string `json:"e"`
			} `json:"keys"`
		}
		if err := json.Unmarshal(resp.Body(), &res); err == nil {
			for _, k := range res.Keys {
				if pubKey, err := parseJWK(k.N, k.E); err == nil {
					appleKeys[k.Kid] = pubKey
				}
			}
			logger.Log.Info("Refreshed Apple public keys from Apple API")
		}
	}

	// Refresh Google
	resp, err = restyClient.R().Get("https://www.googleapis.com/oauth2/v1/certs")
	if err == nil && resp.IsSuccess() {
		var res map[string]string
		if err := json.Unmarshal(resp.Body(), &res); err == nil {
			for kid, certStr := range res {
				if pubKey, err := parseX509Cert(certStr); err == nil {
					googleKeys[kid] = pubKey
				}
			}
			logger.Log.Info("Refreshed Google public keys from Google API")
		}
	}
}

func GetApplePublicKey(kid string) *rsa.PublicKey {
	keysMu.RLock()
	defer keysMu.RUnlock()
	return appleKeys[kid]
}

func GetGooglePublicKey(kid string) *rsa.PublicKey {
	keysMu.RLock()
	defer keysMu.RUnlock()
	return googleKeys[kid]
}

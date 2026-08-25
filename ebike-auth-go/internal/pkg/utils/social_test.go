package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"testing"
)

func TestSignAlipayParams(t *testing.T) {
	// Generate mock RSA Private Key PKCS8 format
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey failed: %v", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey failed: %v", err)
	}
	privBase64 := base64.StdEncoding.EncodeToString(privBytes)

	params := map[string]string{
		"app_id":    "2016000000000000",
		"method":    "alipay.system.oauth.token",
		"code":      "abc",
		"sign_type": "RSA2",
	}

	sig, err := signAlipayParams(params, privBase64)
	if err != nil {
		t.Fatalf("signAlipayParams failed: %v", err)
	}

	if sig == "" {
		t.Errorf("signAlipayParams returned empty signature")
	}
}

package pay

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"testing"
)

func TestZhimaEmbeddedKeysAreNotMatchingPair(t *testing.T) {
	privateKeyB64, publicKeyB64 := zhimaAlipayKeyPair()
	privDER, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		t.Fatal(err)
	}
	priv, err := x509.ParsePKCS8PrivateKey(privDER)
	if err != nil {
		t.Fatal(err)
	}
	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		t.Fatal("private key is not RSA")
	}
	derivedDER, err := x509.MarshalPKIXPublicKey(&rsaPriv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	embeddedDER, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		t.Fatal(err)
	}
	if string(derivedDER) == string(embeddedDER) {
		t.Fatal("expected app private key and alipay public key to differ")
	}
}

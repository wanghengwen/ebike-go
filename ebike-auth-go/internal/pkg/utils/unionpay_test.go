package utils

import (
	"crypto/des"
	"encoding/base64"
	"encoding/hex"
	"testing"
)

func TestSignParams(t *testing.T) {
	params := map[string]string{
		"appId":     "123",
		"nonceStr":  "abc",
		"timestamp": "456",
		"secret":    "mySecret",
	}

	sig := signParams(params)
	if sig == "" {
		t.Errorf("signParams returned empty signature")
	}
}

func TestDecrypt3DESECBBase64(t *testing.T) {
	hexKey := "25326b85bc45738a19070d86d5a4f85825326b85bc45738a" // 24 bytes hex key
	keyBytes, _ := hex.DecodeString(hexKey)

	plain := "13800000000"

	// PKCS5/7 Padding manually in test
	blockSize := des.BlockSize
	padding := blockSize - (len(plain) % blockSize)
	padBytes := make([]byte, padding)
	for i := range padBytes {
		padBytes[i] = byte(padding)
	}
	padded := append([]byte(plain), padBytes...)

	// Encrypt manually using 3DES ECB
	block, err := des.NewTripleDESCipher(keyBytes)
	if err != nil {
		t.Fatalf("des.NewTripleDESCipher error: %v", err)
	}

	cipherBytes := make([]byte, len(padded))
	for i := 0; i < len(padded); i += blockSize {
		block.Encrypt(cipherBytes[i:i+blockSize], padded[i:i+blockSize])
	}

	cipherBase64 := base64.StdEncoding.EncodeToString(cipherBytes)

	// Decrypt using unionpay implementation
	decrypted, err := Decrypt3DESECBBase64(cipherBase64, hexKey)
	if err != nil {
		t.Fatalf("Decrypt3DESECBBase64 error: %v", err)
	}

	if decrypted != plain {
		t.Errorf("Decrypted got %q, want %q", decrypted, plain)
	}
}

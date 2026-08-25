package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestSha256(t *testing.T) {
	plain := "xyy"
	expected := "3866a60f3d661fde3dd11af287239d3d5eafb604bb8419d0ad7934d678d85993"
	got := Sha256(plain)
	if got != expected {
		t.Errorf("Sha256(%q) = %q; want %q", plain, got, expected)
	}
}

func TestMatchesSHA256(t *testing.T) {
	raw := "myPassword123"
	hash := Sha256(raw)
	if !MatchesSHA256(raw, hash) {
		t.Errorf("MatchesSHA256 failed for valid password")
	}

	if MatchesSHA256("wrongPass", hash) {
		t.Errorf("MatchesSHA256 passed for invalid password")
	}
}

func TestDecryptWechatData(t *testing.T) {
	sessionKey := "tii7tZOF1FecmsK5X2jFmw==" // 16 bytes base64 decoded
	iv := "r7BXXKkU5Bf/2p434a1hqA=="         // 16 bytes base64 decoded

	keyBytes, _ := base64.StdEncoding.DecodeString(sessionKey)
	ivBytes, _ := base64.StdEncoding.DecodeString(iv)

	plainText := `{"phoneNumber":"13388888888","purePhoneNumber":"13388888888","countryCode":"86","watermark":{"timestamp":1600000000,"appid":"wx123"}}`

	// PKCS7 Pad
	blockSize := aes.BlockSize
	padding := blockSize - (len(plainText) % blockSize)
	padBytes := make([]byte, padding)
	for i := range padBytes {
		padBytes[i] = byte(padding)
	}
	padded := append([]byte(plainText), padBytes...)

	// AES Encrypt CBC
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		t.Fatalf("aes cipher init error: %v", err)
	}
	cipherBytes := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, ivBytes)
	mode.CryptBlocks(cipherBytes, padded)

	encryptedData := base64.StdEncoding.EncodeToString(cipherBytes)

	// Decrypt
	decryptedBytes, err := DecryptWechatData(sessionKey, encryptedData, iv)
	if err != nil {
		t.Fatalf("DecryptWechatData failed: %v", err)
	}

	var decryptedInfo WechatUserInfo
	if err := json.Unmarshal(decryptedBytes, &decryptedInfo); err != nil {
		t.Fatalf("JSON parse decrypted bytes error: %v", err)
	}

	if decryptedInfo.PhoneNumber != "13388888888" {
		t.Errorf("Got phone %q, want %q", decryptedInfo.PhoneNumber, "13388888888")
	}
	if decryptedInfo.Watermark.Appid != "wx123" {
		t.Errorf("Got appid %q, want %q", decryptedInfo.Watermark.Appid, "wx123")
	}
}

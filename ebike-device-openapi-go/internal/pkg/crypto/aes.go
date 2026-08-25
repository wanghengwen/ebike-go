package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"

	"ebike-device-openapi-go/internal/pkg/logger"
	"go.uber.org/zap"
)

var (
	// The key used in the Java AesEncryptUtil ("13dade85494f4b2ca4f49b21cd058625".substring(16))
	defaultKey = []byte("a4f49b21cd058625")
	defaultIV  = []byte("a4f49b21cd058625")
)

// Encrypt AES/CBC/NoPadding with zero padding
func Encrypt(data string, key []byte, iv []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	dataBytes := []byte(data)
	blockSize := block.BlockSize()
	
	// Manual zero padding as in Java: 
	// if plaintextLength % blockSize != 0 { plaintextLength += blockSize - plaintextLength % blockSize }
	// The new byte array defaults to 0 in Java, so it's zero padded.
	padLen := blockSize - (len(dataBytes) % blockSize)
	if padLen != blockSize { // Java code padded if != 0
		padding := bytes.Repeat([]byte{0}, padLen)
		dataBytes = append(dataBytes, padding...)
	}

	encrypted := make([]byte, len(dataBytes))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted, dataBytes)

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DesEncrypt AES/CBC/NoPadding with zero padding stripping
func DesEncrypt(data string, key []byte, iv []byte) (string, error) {
	encrypted, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	decrypted := make([]byte, len(encrypted))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decrypted, encrypted)

	// NOTE(Java-Quirk-Replication): Java's AesEncryptUtil performs manual zero-padding
	// prior to encryption instead of using PKCS5Padding. Consequently, decryption yields
	// trailing \u0000 bytes. In Java this was stripped via: new String(original).replaceAll("\u0000", "")
	// We precisely replicate this byte-stripping behavior here.
	res := bytes.ReplaceAll(decrypted, []byte{0}, []byte{})
	return string(res), nil
}

// EncryptDefault uses default Key and IV
func EncryptDefault(data string) (string, error) {
	return Encrypt(data, defaultKey, defaultIV)
}

// DesEncryptDefault uses default Key and IV
func DesEncryptDefault(data string) string {
	if data == "" {
		return ""
	}
	res, err := DesEncrypt(data, defaultKey, defaultIV)
	if err != nil {
		logger.Log.Warn("desEncrypt exception", zap.Error(err))
		return ""
	}
	return res
}

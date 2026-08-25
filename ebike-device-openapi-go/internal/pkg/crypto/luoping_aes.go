package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

// Luoping AES-128-CBC (PKCS#7) for MQTT binary protocol V8.
// Key base: {deviceId}+luopingtech.com+{imei}
// Payload: KeyIndex(1) || ciphertext

const luopingKeyDomain = "luopingtech.com"

// BuildLuopingDeviceStr returns the key-derivation base string.
func BuildLuopingDeviceStr(deviceID, imei string) string {
	return deviceID + "+" + luopingKeyDomain + "+" + imei
}

// DeriveLuopingKey derives a 16-byte AES key from deviceStr starting at keyIndex (wrap).
func DeriveLuopingKey(deviceStr string, keyIndex byte) ([]byte, error) {
	data := []byte(deviceStr)
	L := len(data)
	if L == 0 {
		return nil, fmt.Errorf("luoping device str empty")
	}
	if int(keyIndex) >= L {
		return nil, fmt.Errorf("luoping keyIndex %d >= L %d", keyIndex, L)
	}
	key := make([]byte, 16)
	for i := 0; i < 16; i++ {
		key[i] = data[(int(keyIndex)+i)%L]
	}
	return key, nil
}

// DeriveLuopingIV = SHA-256(masterKey || deviceID || imei)[:16]
func DeriveLuopingIV(masterKey []byte, deviceID, imei string) []byte {
	h := sha256.New()
	h.Write(masterKey)
	h.Write([]byte(deviceID))
	h.Write([]byte(imei))
	sum := h.Sum(nil)
	iv := make([]byte, 16)
	copy(iv, sum[:16])
	return iv
}

// EncryptLuopingMQTT encrypts plaintext and returns KeyIndex||ciphertext.
// KeyIndex is chosen randomly in [0, L).
func EncryptLuopingMQTT(plain []byte, deviceID, imei string) ([]byte, error) {
	deviceStr := BuildLuopingDeviceStr(deviceID, imei)
	L := len(deviceStr)
	if L == 0 {
		return nil, fmt.Errorf("luoping device str empty")
	}
	var idxBuf [1]byte
	if _, err := rand.Read(idxBuf[:]); err != nil {
		return nil, fmt.Errorf("luoping keyIndex rand: %w", err)
	}
	keyIndex := idxBuf[0] % byte(L)
	return EncryptLuopingMQTTWithKeyIndex(plain, deviceID, imei, keyIndex)
}

// EncryptLuopingMQTTWithKeyIndex encrypts with a fixed KeyIndex (for tests).
func EncryptLuopingMQTTWithKeyIndex(plain []byte, deviceID, imei string, keyIndex byte) ([]byte, error) {
	deviceStr := BuildLuopingDeviceStr(deviceID, imei)
	key, err := DeriveLuopingKey(deviceStr, keyIndex)
	if err != nil {
		return nil, err
	}
	iv := DeriveLuopingIV(key, deviceID, imei)
	padded := pkcs7Pad(plain, aes.BlockSize)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cipherText := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(cipherText, padded)
	out := make([]byte, 1+len(cipherText))
	out[0] = keyIndex
	copy(out[1:], cipherText)
	return out, nil
}

// DecryptLuopingMQTT decrypts KeyIndex||ciphertext into plaintext.
func DecryptLuopingMQTT(payload []byte, deviceID, imei string) ([]byte, error) {
	if len(payload) < 1+aes.BlockSize {
		return nil, fmt.Errorf("luoping payload too short: %d", len(payload))
	}
	keyIndex := payload[0]
	cipherText := payload[1:]
	if len(cipherText)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("luoping ciphertext not block-aligned: %d", len(cipherText))
	}
	deviceStr := BuildLuopingDeviceStr(deviceID, imei)
	key, err := DeriveLuopingKey(deviceStr, keyIndex)
	if err != nil {
		return nil, err
	}
	iv := DeriveLuopingIV(key, deviceID, imei)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plain := make([]byte, len(cipherText))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, cipherText)
	return pkcs7Unpad(plain, aes.BlockSize)
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	pad := blockSize - len(data)%blockSize
	out := make([]byte, len(data)+pad)
	copy(out, data)
	for i := len(data); i < len(out); i++ {
		out[i] = byte(pad)
	}
	return out
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, fmt.Errorf("pkcs7: invalid length %d", len(data))
	}
	pad := int(data[len(data)-1])
	if pad == 0 || pad > blockSize || pad > len(data) {
		return nil, fmt.Errorf("pkcs7: invalid padding %d", pad)
	}
	for i := 0; i < pad; i++ {
		if data[len(data)-1-i] != byte(pad) {
			return nil, fmt.Errorf("pkcs7: padding mismatch")
		}
	}
	return data[:len(data)-pad], nil
}

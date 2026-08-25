package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"ebike-auth-go/internal/pkg/config"
)

// Sha256 computes sha256 hash in string format
func Sha256(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// MatchesSHA256 compares plain password with stored hash
func MatchesSHA256(rawPassword, encodedPassword string) bool {
	// Java PasswordEncoder matching logic:
	// String xyy = encode("xyy");
	// if(StringUtils.equals(encodedPassword,xyy) && !applicationProperties.getEnableTenantSecretVerify()){
	//     return true;
	// }
	xyyHash := Sha256("xyy")
	if encodedPassword == xyyHash && !config.AppConfig.EnableTenantSecretVerify {
		return true
	}

	return Sha256(rawPassword) == encodedPassword
}

// Md5 computes MD5 hash in hex string format
func Md5(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// WeChat UserInfo structure mapping
type WechatWatermark struct {
	Timestamp int64  `json:"timestamp"`
	Appid     string `json:"appid"`
}

type WechatUserInfo struct {
	AppId           string          `json:"appId"`
	OpenId          string          `json:"openId"`
	UnionId         string          `json:"unionId,omitempty"`
	AvatarUrl       string          `json:"avatarUrl,omitempty"`
	NickName        string          `json:"nickName,omitempty"`
	PhoneNumber     string          `json:"phoneNumber,omitempty"`
	PurePhoneNumber string          `json:"purePhoneNumber,omitempty"`
	CountryCode     string          `json:"countryCode,omitempty"`
	Watermark       WechatWatermark `json:"watermark"`
}

// DecryptWechatData decrypts WeChat encryptedData using sessionKey and iv
func DecryptWechatData(sessionKey, encryptedData, iv string) ([]byte, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return nil, fmt.Errorf("decode sessionKey error: %w", err)
	}

	cipherBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("decode encryptedData error: %w", err)
	}

	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return nil, fmt.Errorf("decode iv error: %w", err)
	}

	// Align key length to multiple of 16 (matching Java WechatUtils logic)
	const base = 16
	if len(keyBytes)%base != 0 {
		groups := len(keyBytes)/base + 1
		temp := make([]byte, groups*base)
		copy(temp, keyBytes)
		keyBytes = temp
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher error: %w", err)
	}

	if len(ivBytes) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	mode := cipher.NewCBCDecrypter(block, ivBytes)
	decrypted := make([]byte, len(cipherBytes))
	mode.CryptBlocks(decrypted, cipherBytes)

	// PKCS#7 Unpadding
	unpadded, err := pkcs7Unpad(decrypted, block.BlockSize())
	if err != nil {
		return nil, fmt.Errorf("pkcs7 unpad error: %w", err)
	}

	return unpadded, nil
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("data is empty")
	}
	if length%blockSize != 0 {
		return nil, errors.New("data is not aligned to block size")
	}
	padding := int(data[length-1])
	if padding < 1 || padding > blockSize {
		return nil, errors.New("invalid padding value")
	}
	for i := length - padding; i < length; i++ {
		if int(data[i]) != padding {
			return nil, errors.New("invalid padding byte")
		}
	}
	return data[:length-padding], nil
}

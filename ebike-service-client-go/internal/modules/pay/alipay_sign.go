package pay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// zhimaAppSecret is the secret hardcoded in Java
// ZhimaPayAfterUseCallbackController.validateSign; segment [1] (split by "@")
// is the Alipay RSA public key used for callback signature verification.
const zhimaAppSecret = "MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQC+7ceSRGVspaH1pal00zG60VP+BK86Oe0mvKdmm3IttKFvxmL/8gVt4IUpxiNTArChwoVXpDu7jvfcnniANH97mc0IX7sriAj9QNTc37Ko2ASPXScsHDCeClYB6ChBx7osO+uEKFdivCaf+ZBpaCEyQJoWmhG9p15NxvwOYuYLzwpjnqkZ277/rCsb/JlAntW+35sjIPZpSsivsZ2L/SV+sd0tdP9zMhEvIjRhxyDmiWEoD8IPDpNOtAk32v5apqamOvuonhPDUuYq8SFBNwpRe8SEOxwQ3X6pGUyxfIwpJVCZG3IbK9vCJhLTkPjApan9HAmIxVS5/pGJByr8TYmdAgMBAAECggEAQDQJRkBFsvFHsykQAL78HAxEKEk++197ReluiWyASqpRFxspM1QZS0eSv+dm/YUMDHkzCbOqenmrE78eWk5NCC1B6yz17b+C9laUveljVK+/aM40W/rmxl5HacC9uNEG49UKb5h5OjR28JilXSys7Q8YQb1xdcsQRStCmzvai+Fwmcx6GKs7L5sumccia80PnIjn/nzIFiIVpboORqNR37b9oTm+KBDYrG/0hCyIZ+F6RiHfmDgfqQqfGfyPKOS1pcUNaEE04gvWSlWjeZESz1OwMi2reC8AmJVo153G4e61xkPthFYaGMqj/cUqcXruWkMA8c1VYwCizsFfZjOZ6QKBgQDl2qf+7/ikHiTa+XmHbu97zAMM9iBLd6b87a/0XEzB1fQKgTs90oF/a7bxCStsIexVWOilsY+hWEZ5A9S7TCo/rUNpuBa9tPHjIBiQsN6ugMhy4cWrhgLdFmgjxz9KWfc8PUp72BmXh4US4ZtQ5WWG2p+BSWF81kWlkK4pRt1PuwKBgQDUpZ67BGXETnw0l9EDPcrRsbYZanGeM6o17tlOStRS/UZeWDlMg68M+WGKb7ZkxfxMSMAftaq3fvCAJHZqZU7EYp7sjgPrW4arAxZLwhRsEsqafltafxHSnzDsLE47+zXjmvdGaVEPJ8dZA6BL8NdEacWD2OhlDAHlxinF48aahwKBgCkj08HLjcNCKfKPiHL3JiIQR9OAEhOv3NGUcVPZWVuwQbfHnaTZEpiN3PaTX5RBFh3IhgtyFnUYabSrPN4xKbav+krnyho2Ur0GN59eKN0u67G0Oz8SA10y73zH4soaBChiB/zWlu4KMYVJoBUAmgVjB/2J9srzRw/1L1bv+hiVAoGALzNQB20Tdb6CHV5xe4m0wlTy+bNB4v7O0kfhHlrHxGAJxZlJpq04JuYX+5WOY9H6jag8VQ2LBk377kWprzYrhLXrVtCzGAPp4X2+7jP3OoH1TNOtTWoVN640Osge2XuKW6ojJxLrdjS7MAv5AcJE1h+wQvLbqso+hZU14oILrHMCgYBlmgq0J5NZ6doUKXjso8zN0BTp635QbTWB3xe1r3g/mOMmK4UMWABeYnTKcAd3/MLOTZqvwPKwymhMSQL2D9paJGBOIR77rN8nbsrcquzB8IBiPFE+lDra0Q/UFiIbJi2R3w6vwlhPFcmfM9ns0fTDxa6C0o5Yj40NQNfc7YEIog==@MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAij+SyFWvtZwVBoHkFzQB1/M/NioJIJJlZ/z4xPdOkXRUKQ9tt1VetvokpvHeq546CDUIpHyx2M7huhdWDrOp+u2Py1RTx+JzcKsQjrY98wG3b6JqJdERQ1oK6jqp0YImmCJqmzNJpzT4MMslBZDlyB8C7y53fSHFThlBX/jI03ppT2CPj3mN11ZSX1Vj/JopIsdlxWc752ruP4VRsPebSI2ScNttuaoS++El8wHuESUD4csWBOcH/7+1tBjFxBv9YmuTF82zrUlp4NE1Fazv28G0h36ttL5e8sZMckviy7H2zWO9XQOVEVO3Yj/ClCeUDbtW39VMLX5nrDrvrPMFGwIDAQAB@BIvEU5uTCZIYaWffqnfAag=="

func zhimaAlipayKeyPair() (privateKeyB64, publicKeyB64 string) {
	parts := strings.Split(zhimaAppSecret, "@")
	if len(parts) < 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

// buildAlipaySignCheckContentV1 mirrors Alipay SDK
// AlipaySignature.getSignCheckContentV1 without mutating the input map:
// drop sign/sign_type, skip empty key/value pairs, sort keys, join k=v with &.
func buildAlipaySignCheckContentV1(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" || k == "sign_type" {
			continue
		}
		if strings.TrimSpace(k) == "" || strings.TrimSpace(params[k]) == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(params[k])
	}
	return b.String()
}

func alipayContentBytes(content, charset string) ([]byte, error) {
	if strings.TrimSpace(charset) == "" {
		charset = "UTF-8"
	}
	switch strings.ToUpper(charset) {
	case "UTF-8":
		return []byte(content), nil
	case "GBK", "GB2312":
		return io.ReadAll(transform.NewReader(strings.NewReader(content), simplifiedchinese.GBK.NewEncoder()))
	default:
		return []byte(content), nil
	}
}

// decodeAlipaySignBase64 mirrors Java commons-codec Base64, which ignores
// whitespace in the sign parameter.
func decodeAlipaySignBase64(sign string) ([]byte, error) {
	clean := strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\r', '\n':
			return -1
		}
		return r
	}, sign)
	return base64.StdEncoding.DecodeString(clean)
}

// verifyAlipaySign ports AlipaySignature.rsaCheckV1(paramMap, publicKey,
// charset, "RSA2") using the zhima callback Alipay public key.
func verifyAlipaySign(params map[string]string) (bool, error) {
	_, publicKeyB64 := zhimaAlipayKeyPair()
	return verifyAlipayRSAV1(params, publicKeyB64)
}

// verifyAlipayRSAV1 mirrors AlipaySignature.rsaCheckV1 for RSA2.
func verifyAlipayRSAV1(params map[string]string, publicKeyB64 string) (bool, error) {
	if strings.TrimSpace(publicKeyB64) == "" {
		return false, errors.New("alipay public key is missing")
	}

	sign, ok := params["sign"]
	if !ok || strings.TrimSpace(sign) == "" {
		return false, errors.New("sign parameter is missing")
	}

	content := buildAlipaySignCheckContentV1(params)
	charset := params["charset"]
	contentBytes, err := alipayContentBytes(content, charset)
	if err != nil {
		return false, fmt.Errorf("invalid charset %q: %v", charset, err)
	}

	der, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return false, fmt.Errorf("invalid alipay public key: %v", err)
	}
	pub, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return false, fmt.Errorf("invalid alipay public key: %v", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return false, errors.New("alipay public key is not RSA")
	}

	sigBytes, err := decodeAlipaySignBase64(sign)
	if err != nil {
		return false, fmt.Errorf("invalid sign value: %v", err)
	}

	digest := sha256.Sum256(contentBytes)
	if err := rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, digest[:], sigBytes); err != nil {
		return false, nil
	}
	return true, nil
}

// newUUID generates a random (version 4) UUID string, matching Java's
// UUID.randomUUID().toString() format.
func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC 4122
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

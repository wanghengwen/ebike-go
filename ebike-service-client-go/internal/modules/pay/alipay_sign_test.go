package pay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"strings"
	"testing"
)

func testAlipayKeyPairB64(t *testing.T) (privateKeyB64, publicKeyB64 string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(privDER), base64.StdEncoding.EncodeToString(pubDER)
}

// signAlipayRSA2 signs content the same way Alipay SDK RSA2 does
// (SHA256withRSA + Base64), used only to build test vectors.
func signAlipayRSA2(content, charset, privateKeyB64 string) (string, error) {
	der, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return "", err
	}
	key, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return "", err
	}
	priv, ok := key.(*rsa.PrivateKey)
	if !ok {
		return "", err
	}
	contentBytes, err := alipayContentBytes(content, charset)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(contentBytes)
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func TestBuildAlipaySignCheckContentV1MatchesSDK(t *testing.T) {
	params := map[string]string{
		"charset":   "UTF-8",
		"biz_field": "value",
		"empty":     "",
		"sign":      "should-drop",
		"sign_type": "RSA2",
	}
	got := buildAlipaySignCheckContentV1(params)
	want := "biz_field=value&charset=UTF-8"
	if got != want {
		t.Fatalf("sign content = %q, want %q", got, want)
	}
}

func TestVerifyAlipayRSAV1RoundTripZhimaStylePayload(t *testing.T) {
	privateKeyB64, publicKeyB64 := testAlipayKeyPairB64(t)

	// Mirrors a zhima.credit.payafteruse.* async notify shape (Alipay open-platform doc).
	bizContent := `{"credit_agreement_id":"ZM202406200001","out_order_no":"tenant001@13800000000"}`
	params := map[string]string{
		"charset":       "UTF-8",
		"msg_method":    "zhima.credit.payafteruse.creditagreement.changed",
		"biz_content":   bizContent,
		"utc_timestamp": "1718860800",
		"version":       "1.0",
		"sign_type":     "RSA2",
		"empty_field":   "", // Alipay SDK getSignCheckContentV1 skips empty values
	}

	content := buildAlipaySignCheckContentV1(params)
	sign, err := signAlipayRSA2(content, params["charset"], privateKeyB64)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	params["sign"] = sign

	ok, err := verifyAlipayRSAV1(params, publicKeyB64)
	if err != nil {
		t.Fatalf("verifyAlipayRSAV1 error: %v", err)
	}
	if !ok {
		t.Fatal("expected verifyAlipayRSAV1 to pass for self-signed zhima-style payload")
	}
}

func TestVerifyAlipayRSAV1RejectsTamperedSign(t *testing.T) {
	privateKeyB64, publicKeyB64 := testAlipayKeyPairB64(t)
	params := map[string]string{
		"charset":     "UTF-8",
		"msg_method":  "zhima.credit.payafteruse.creditbizorder.changed",
		"biz_content": `{"out_order_no":"t@u"}`,
		"version":     "1.0",
		"sign_type":   "RSA2",
	}
	content := buildAlipaySignCheckContentV1(params)
	sign, err := signAlipayRSA2(content, "UTF-8", privateKeyB64)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	params["sign"] = sign

	ok, err := verifyAlipayRSAV1(params, publicKeyB64)
	if err != nil || !ok {
		t.Fatalf("baseline verify failed: ok=%v err=%v", ok, err)
	}

	params["biz_content"] = `{"out_order_no":"tampered@u"}`
	ok, err = verifyAlipayRSAV1(params, publicKeyB64)
	if err != nil {
		t.Fatalf("verifyAlipayRSAV1 error: %v", err)
	}
	if ok {
		t.Fatal("expected tampered biz_content to fail verification")
	}
}

func TestVerifyAlipaySignRejectsMissingSign(t *testing.T) {
	ok, err := verifyAlipaySign(map[string]string{"charset": "UTF-8"})
	if err == nil || ok {
		t.Fatalf("expected missing sign to fail: ok=%v err=%v", ok, err)
	}
}

func TestVerifyAlipayRSAV1IgnoresSignWhitespace(t *testing.T) {
	privateKeyB64, publicKeyB64 := testAlipayKeyPairB64(t)
	params := map[string]string{
		"charset":    "UTF-8",
		"msg_method": "zhima.credit.payafteruse.creditagreement.changed",
		"version":    "1.0",
		"sign_type":  "RSA2",
	}
	content := buildAlipaySignCheckContentV1(params)
	sign, err := signAlipayRSA2(content, "UTF-8", privateKeyB64)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	// Java commons-codec Base64.decode tolerates line breaks in sign.
	params["sign"] = strings.Join([]string{sign[:len(sign)/2], sign[len(sign)/2:]}, "\n")

	ok, err := verifyAlipayRSAV1(params, publicKeyB64)
	if err != nil {
		t.Fatalf("verifyAlipayRSAV1 error: %v", err)
	}
	if !ok {
		t.Fatal("expected sign with whitespace to verify successfully")
	}
}

func TestVerifyAlipayRSAV1OrderChangedPayload(t *testing.T) {
	privateKeyB64, publicKeyB64 := testAlipayKeyPairB64(t)
	params := map[string]string{
		"charset":       "UTF-8",
		"msg_method":    "zhima.credit.payafteruse.creditbizorder.changed",
		"biz_content":   `{"out_order_no":"tenant42@user99","order_status":"TRADE_SUCCESS"}`,
		"utc_timestamp": "1718860900",
		"version":       "1.0",
		"sign_type":     "RSA2",
	}
	content := buildAlipaySignCheckContentV1(params)
	sign, err := signAlipayRSA2(content, "UTF-8", privateKeyB64)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	params["sign"] = sign

	ok, err := verifyAlipayRSAV1(params, publicKeyB64)
	if err != nil || !ok {
		t.Fatalf("order-changed payload verify failed: ok=%v err=%v", ok, err)
	}
}

func TestVerifyAlipaySignUsesEmbeddedAlipayPublicKey(t *testing.T) {
	privateKeyB64, _ := zhimaAlipayKeyPair()
	params := map[string]string{
		"charset":    "UTF-8",
		"msg_method": "zhima.credit.payafteruse.creditagreement.changed",
		"version":    "1.0",
		"sign_type":  "RSA2",
	}
	content := buildAlipaySignCheckContentV1(params)
	sign, err := signAlipayRSA2(content, "UTF-8", privateKeyB64)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	params["sign"] = sign

	// Java uses appSecret.split("@")[1] (Alipay public key), not the app private key pair.
	ok, err := verifyAlipaySign(params)
	if err != nil {
		t.Fatalf("verifyAlipaySign error: %v", err)
	}
	if ok {
		t.Fatal("expected app-private-key signature to fail against embedded Alipay public key")
	}
}

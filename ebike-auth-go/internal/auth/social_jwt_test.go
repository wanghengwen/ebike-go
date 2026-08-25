package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestVerifyAppleIdentityToken_JavaTestData(t *testing.T) {
	pubKey, err := parseJWK(javaApplePublicKeyN, javaApplePublicKeyE)
	if err != nil {
		t.Fatalf("parseJWK failed: %v", err)
	}

	getKey := func(kid string) *rsa.PublicKey {
		if kid == "eXaunmL" {
			return pubKey
		}
		return nil
	}

	_, err = verifyAppleIdentityToken(javaAppleIdentityToken, getKey)
	if err == nil {
		t.Fatal("expected expired token error, got nil")
	}
	if !errors.Is(err, ErrIdentityTokenExpired) && !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected signature-valid expired token, got: %v", err)
	}
}

func TestVerifyAppleIdentityToken_TamperedSignature(t *testing.T) {
	pubKey, err := parseJWK(javaApplePublicKeyN, javaApplePublicKeyE)
	if err != nil {
		t.Fatalf("parseJWK failed: %v", err)
	}

	tampered := javaAppleIdentityToken[:len(javaAppleIdentityToken)-1] + "A"
	getKey := func(kid string) *rsa.PublicKey {
		if kid == "eXaunmL" {
			return pubKey
		}
		return nil
	}

	_, err = verifyAppleIdentityToken(tampered, getKey)
	if err == nil {
		t.Fatal("expected signature error for tampered token")
	}
	if !errors.Is(err, ErrIdentityTokenSignature) && !strings.Contains(err.Error(), "signature") {
		t.Fatalf("expected signature validation failure, got: %v", err)
	}
}

func TestVerifyAppleIdentityToken_JavaHeaderKid(t *testing.T) {
	header, err := parseJWTHeader(javaAppleIdentityToken)
	if err != nil {
		t.Fatalf("parseJWTHeader failed: %v", err)
	}
	if header.Kid != "eXaunmL" {
		t.Fatalf("kid mismatch: got %q, want %q", header.Kid, "eXaunmL")
	}
	if header.Alg != "RS256" {
		t.Fatalf("alg mismatch: got %q, want RS256", header.Alg)
	}
}

func TestVerifyGoogleIdToken_WithLocalPublicKey(t *testing.T) {
	chdirToProjectRoot(t)
	loadLocalKeys()

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	kid := "integration-google-kid"
	getKey := func(keyID string) *rsa.PublicKey {
		if keyID == kid {
			return &privKey.PublicKey
		}
		return GetGooglePublicKey(keyID)
	}

	tokenStr := signGoogleTestToken(t, privKey, kid, time.Now().Add(time.Hour))
	claims, err := verifyGoogleIdToken(tokenStr, getKey)
	if err != nil {
		t.Fatalf("verifyGoogleIdToken failed: %v", err)
	}
	if claims.Sub != "google-sub-001" || claims.Email != "google@test.com" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestVerifyGoogleIdToken_TamperedSignatureWithBundledCert(t *testing.T) {
	chdirToProjectRoot(t)
	loadLocalKeys()

	pubKey := GetGooglePublicKey(javaGooglePublicKeyKid)
	if pubKey == nil {
		t.Fatalf("google public key %s not loaded", javaGooglePublicKeyKid)
	}

	tampered := "eyJhbGciOiJSUzI1NiIsImtpZCI6ImQ0ZTA2Y2ViMjJiMDFiZTU2YzIxM2M5ODU0MGFiNTYzYmZmNWE1OGMifQ.eyJzdWIiOiJ0YW1wZXJlZCJ9.signature"
	getKey := func(keyID string) *rsa.PublicKey {
		if keyID == javaGooglePublicKeyKid {
			return pubKey
		}
		return nil
	}
	_, err := verifyGoogleIdToken(tampered, getKey)
	if err == nil {
		t.Fatal("expected tampered token to fail")
	}
}

func TestLoadGooglePublicKeysFromBundledJSON(t *testing.T) {
	chdirToProjectRoot(t)
	loadLocalKeys()
	pubKey := GetGooglePublicKey(javaGooglePublicKeyKid)
	if pubKey == nil {
		t.Fatalf("expected bundled google key %s", javaGooglePublicKeyKid)
	}
	if pubKey.E != 65537 {
		t.Fatalf("unexpected exponent: %d", pubKey.E)
	}
}

func signGoogleTestToken(t *testing.T, privKey *rsa.PrivateKey, kid string, exp time.Time) string {
	t.Helper()
	claims := GoogleClaims{
		Iss:           "https://accounts.google.com",
		Aud:           "test-client-id",
		Sub:           "google-sub-001",
		Email:         "google@test.com",
		EmailVerified: FlexibleBool(true),
		Name:          "Google Test",
		Picture:       "https://example.com/avatar.png",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(privKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func chdirToProjectRoot(t *testing.T) {
	t.Helper()
	for i := 0; i < 3; i++ {
		if _, err := os.Stat("public-keys-google.json"); err == nil {
			return
		}
		if err := os.Chdir(".."); err != nil {
			t.Fatalf("chdir: %v", err)
		}
	}
	t.Fatal("public-keys-google.json not found from module root")
}

func TestJavaUnionPayTestFixturesPresent(t *testing.T) {
	if javaUnionPayAppID == "" || javaUnionPaySecret == "" || javaUnionPayDcSecret == "" {
		t.Fatal("union pay fixtures should not be empty")
	}
	raw, _ := json.Marshal(map[string]string{
		"appId":    javaUnionPayAppID,
		"secret":   javaUnionPaySecret,
		"dcSecret": javaUnionPayDcSecret,
		"tenantId": javaUnionPayTenantID,
	})
	if len(raw) == 0 {
		t.Fatal("expected union pay fixture payload")
	}
}

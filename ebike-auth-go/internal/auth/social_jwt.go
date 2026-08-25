package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrIdentityTokenExpired   = errors.New("identityToken expired")
	ErrIdentityTokenSignature = errors.New("identityToken signature does not match locally computed signature")
)

type socialJWTHeader struct {
	Kid string `json:"kid"`
	Alg string `json:"alg"`
}

func verifyRS256Token(tokenStr string, pubKey *rsa.PublicKey, claims jwt.Claims) error {
	if pubKey == nil {
		return fmt.Errorf("public key is nil")
	}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return ErrIdentityTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return ErrIdentityTokenSignature
		}
		return err
	}
	if !token.Valid {
		return ErrIdentityTokenSignature
	}
	return nil
}

func parseJWTHeader(tokenStr string) (*socialJWTHeader, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid jwt format")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	var header socialJWTHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, err
	}
	return &header, nil
}

func verifyAppleIdentityToken(tokenStr string, getKey func(kid string) *rsa.PublicKey) (*AppleClaims, error) {
	header, err := parseJWTHeader(tokenStr)
	if err != nil {
		return nil, err
	}
	pubKey := getKey(header.Kid)
	if pubKey == nil {
		return nil, fmt.Errorf("apple public key kid %s not found", header.Kid)
	}

	claims := &AppleClaims{}
	if err := verifyRS256Token(tokenStr, pubKey, claims); err != nil {
		if errors.Is(err, ErrIdentityTokenExpired) {
			return nil, fmt.Errorf("%w", ErrIdentityTokenExpired)
		}
		if errors.Is(err, ErrIdentityTokenSignature) {
			return nil, fmt.Errorf("%w", ErrIdentityTokenSignature)
		}
		return nil, err
	}
	return claims, nil
}

func verifyGoogleIdToken(tokenStr string, getKey func(kid string) *rsa.PublicKey) (*GoogleClaims, error) {
	header, err := parseJWTHeader(tokenStr)
	if err != nil {
		return nil, err
	}
	pubKey := getKey(header.Kid)
	if pubKey == nil {
		return nil, fmt.Errorf("google public key kid %s not found", header.Kid)
	}

	claims := &GoogleClaims{}
	if err := verifyRS256Token(tokenStr, pubKey, claims); err != nil {
		if errors.Is(err, ErrIdentityTokenExpired) {
			return nil, fmt.Errorf("idToken expired")
		}
		if errors.Is(err, ErrIdentityTokenSignature) {
			return nil, fmt.Errorf("idToken signature does not match locally computed signature")
		}
		return nil, err
	}
	return claims, nil
}

package auth

import (
	"fmt"
	"time"

	"ebike-auth-go/internal/pkg/config"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserName    string   `json:"user_name,omitempty"`
	Scope       []string `json:"scope,omitempty"`
	GrantType   string   `json:"grantType,omitempty"`
	Authorities []string `json:"authorities,omitempty"`
	Platform    string   `json:"platform,omitempty"`
	DeviceId    string   `json:"deviceId,omitempty"`
	Jti         string   `json:"jti,omitempty"`
	ClientId    string   `json:"client_id,omitempty"`
	// Additional info injected by TokenEnhancer
	Nickname string `json:"nickname,omitempty"`
	Name     string `json:"name,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Openid   string `json:"openid,omitempty"`
	Pin      string `json:"pin,omitempty"`
	jwt.RegisteredClaims
}

func GenerateTokenPair(pin, clientId, platform, deviceId, nickname, avatar, openid string, scopes, authorities []string, grantType string, accessTokenValidity, refreshTokenValidity int) (string, string, error) {
	now := time.Now()

	// 1. Access Token Claims
	accessTokenClaims := JWTClaims{
		UserName:    pin,
		Scope:       scopes,
		GrantType:   grantType,
		Authorities: authorities,
		Platform:    platform,
		DeviceId:    deviceId,
		ClientId:    clientId,
		Nickname:    nickname,
		Avatar:      avatar,
		Openid:      openid,
		Pin:         pin,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   pin,
			Audience:  jwt.ClaimStrings{clientId},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(accessTokenValidity) * time.Second)),
		},
	}
	if config.AppConfig.AuthMode == "business" {
		accessTokenClaims.Name = nickname // business mode sets name claim
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	signedAccessToken, err := accessToken.SignedString([]byte(config.AppConfig.JwtSecret))
	if err != nil {
		return "", "", fmt.Errorf("sign access token error: %w", err)
	}

	// 2. Refresh Token Claims (typically simpler but has exp, user_name, platform, deviceId, client_id)
	refreshTokenClaims := JWTClaims{
		UserName:  pin,
		Scope:     scopes,
		GrantType: grantType,
		Platform:  platform,
		DeviceId:  deviceId,
		ClientId:  clientId,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   pin,
			Audience:  jwt.ClaimStrings{clientId},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(refreshTokenValidity) * time.Second)),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(config.AppConfig.JwtSecret))
	if err != nil {
		return "", "", fmt.Errorf("sign refresh token error: %w", err)
	}

	return signedAccessToken, signedRefreshToken, nil
}

func ParseToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.AppConfig.JwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

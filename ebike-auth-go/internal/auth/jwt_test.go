package auth

import (
	"testing"
	"time"

	"ebike-auth-go/internal/pkg/config"
)

func TestGenerateAndParseToken(t *testing.T) {
	// Set mock configuration
	config.AppConfig.JwtSecret = "dbgpfbktyq0p2fowbll4m22xjom6bhc8g0pwmb9b0pj7g67p414ctndum3e0ectf"
	config.AppConfig.AuthMode = "client"

	pin := "user123"
	clientId := "tenantABC"
	platform := "ios"
	deviceId := "iphone_x_1"
	nickname := "Test User"
	avatar := "http://example.com/avatar.jpg"
	openid := "open_123456"
	scopes := []string{"all"}
	authorities := []string{"ROLE_USER"}
	grantType := "password"
	accessTokenValidity := 3600
	refreshTokenValidity := 86400

	// 1. Generate Token Pair
	accessToken, refreshToken, err := GenerateTokenPair(
		pin,
		clientId,
		platform,
		deviceId,
		nickname,
		avatar,
		openid,
		scopes,
		authorities,
		grantType,
		accessTokenValidity,
		refreshTokenValidity,
	)

	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	if accessToken == "" || refreshToken == "" {
		t.Fatalf("Generated empty tokens")
	}

	// 2. Parse and Validate Access Token Claims
	claims, err := ParseToken(accessToken)
	if err != nil {
		t.Fatalf("ParseToken failed on accessToken: %v", err)
	}

	if claims.UserName != pin {
		t.Errorf("UserName mismatch: got %q, want %q", claims.UserName, pin)
	}
	if claims.ClientId != clientId {
		t.Errorf("ClientId mismatch: got %q, want %q", claims.ClientId, clientId)
	}
	if claims.Platform != platform {
		t.Errorf("Platform mismatch: got %q, want %q", claims.Platform, platform)
	}
	if claims.DeviceId != deviceId {
		t.Errorf("DeviceId mismatch: got %q, want %q", claims.DeviceId, deviceId)
	}
	if claims.Nickname != nickname {
		t.Errorf("Nickname mismatch: got %q, want %q", claims.Nickname, nickname)
	}
	if claims.Avatar != avatar {
		t.Errorf("Avatar mismatch: got %q, want %q", claims.Avatar, avatar)
	}
	if claims.Openid != openid {
		t.Errorf("Openid mismatch: got %q, want %q", claims.Openid, openid)
	}
	if claims.Pin != pin {
		t.Errorf("Pin claim mismatch: got %q, want %q", claims.Pin, pin)
	}

	// Verify expiration is set correctly
	timeDiff := claims.ExpiresAt.Time.Sub(time.Now())
	if timeDiff < 55*time.Minute || timeDiff > 65*time.Minute {
		t.Errorf("ExpiresAt time difference is unexpected: got %v, expected ~1 hour", timeDiff)
	}
}

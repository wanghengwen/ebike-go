package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ebike-auth-go/internal/pkg/redis"
	"ebike-auth-go/internal/pkg/utils"
)

func TestHTTPHealthEndpoint(t *testing.T) {
	env := setupIntegrationEnv(t, "business")
	req := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("health status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestHTTPBusinessTokenUnauthorized(t *testing.T) {
	env := setupIntegrationEnv(t, "business")
	_, result := postOAuthToken(t, env.router, "", "", map[string]interface{}{
		"grant_type": "password",
		"traceId":    testTraceID,
		"tenantId":   testTenantID,
	}, nil)
	if result.Success || result.Code != CodeAccessUnauthorized {
		t.Fatalf("expected %s, got %+v", CodeAccessUnauthorized, result)
	}
}

func TestHTTPBusinessTokenInvalidGrant(t *testing.T) {
	env := setupIntegrationEnv(t, "business")
	_, result := postOAuthToken(t, env.router, testTenantID, testClientPass, map[string]interface{}{
		"grant_type": "unknown_grant",
		"traceId":    testTraceID,
		"tenantId":   testTenantID,
	}, nil)
	if result.Success || result.Code != CodeInvalidGrant {
		t.Fatalf("expected %s, got %+v", CodeInvalidGrant, result)
	}
}

func TestHTTPBusinessTokenAKSuccess(t *testing.T) {
	env := setupIntegrationEnv(t, "business")
	_, result := postOAuthToken(t, env.router, testTenantID, testClientPass, map[string]interface{}{
		"grant_type": "ak",
		"key":        "service-key",
		"secret":     "service-secret",
		"traceId":    testTraceID,
		"tenantId":   testTenantID,
		"platform":   testPlatform,
		"deviceId":   testDeviceID,
	}, nil)
	if !result.Success || result.Code != CodeSuccess {
		t.Fatalf("expected success, got %+v", result)
	}
	var token TokenCo
	if err := json.Unmarshal(result.Data, &token); err != nil {
		t.Fatalf("decode token: %v", err)
	}
	if token.AccessToken == "" || token.Pin != "service-key" {
		t.Fatalf("unexpected token payload: %+v", token)
	}
}

func TestHTTPBusinessTokenPasswordSuccess(t *testing.T) {
	env := setupIntegrationEnv(t, "business")
	_, result := postOAuthToken(t, env.router, testTenantID, testClientPass, map[string]interface{}{
		"grant_type": "password",
		"username":   "admin001",
		"password":   "password",
		"traceId":    testTraceID,
		"tenantId":   testTenantID,
		"platform":   testPlatform,
		"deviceId":   testDeviceID,
	}, nil)
	if !result.Success {
		t.Fatalf("expected success, got %+v", result)
	}
}

func TestHTTPBusinessTokenPCNoPermission(t *testing.T) {
	env := setupIntegrationEnv(t, "business")
	_, result := postOAuthToken(t, env.router, testTenantID, testClientPass, map[string]interface{}{
		"grant_type": "password",
		"username":   "noperm001",
		"password":   "password",
		"traceId":    testTraceID,
		"tenantId":   testTenantID,
		"platform":   "pc",
		"deviceId":   testDeviceID,
	}, nil)
	if result.Success || result.Code != CodeUserNoPermission {
		t.Fatalf("expected %s, got %+v", CodeUserNoPermission, result)
	}
}

func TestHTTPBusinessPhoneSecretJavaFormula(t *testing.T) {
	env := setupIntegrationEnv(t, "business")
	phone := "+86-13300000001"
	_, result := postOAuthToken(t, env.router, testTenantID, testClientPass, map[string]interface{}{
		"grant_type": "phone_secret",
		"phone":      phone,
		"secret":     javaBusinessPhoneSecret(phone),
		"traceId":    testTraceID,
		"tenantId":   testTenantID,
		"platform":   testPlatform,
		"deviceId":   testDeviceID,
	}, nil)
	if !result.Success {
		t.Fatalf("expected success, got %+v", result)
	}
}

func TestHTTPBusinessRefreshAndLogout(t *testing.T) {
	env := setupIntegrationEnv(t, "business")
	_, login := postOAuthToken(t, env.router, testTenantID, testClientPass, map[string]interface{}{
		"grant_type": "ak",
		"key":        "service-key",
		"secret":     "service-secret",
		"traceId":    testTraceID,
		"tenantId":   testTenantID,
		"platform":   testPlatform,
		"deviceId":   testDeviceID,
	}, nil)
	if !login.Success {
		t.Fatalf("login failed: %+v", login)
	}
	var token TokenCo
	_ = json.Unmarshal(login.Data, &token)

	_, refresh := postOAuthToken(t, env.router, testTenantID, testClientPass, map[string]interface{}{
		"grant_type":    "refresh_token",
		"refresh_token": token.RefreshToken,
		"traceId":       testTraceID,
		"tenantId":      testTenantID,
		"platform":      testPlatform,
		"deviceId":      testDeviceID,
	}, nil)
	if !refresh.Success {
		t.Fatalf("refresh failed: %+v", refresh)
	}

	authHeader := encodeAuthorities(t, map[string]interface{}{
		"user_name": "service-key",
		"platform":  testPlatform,
	})
	logout := postLogout(t, env.router, authHeader, map[string]interface{}{
		"traceId":  testTraceID,
		"tenantId": testTenantID,
	})
	if !logout.Success || logout.Code != CodeSuccess {
		t.Fatalf("logout failed: %+v", logout)
	}
}

func TestHTTPClientPhoneCodeDebugBypass(t *testing.T) {
	env := setupIntegrationEnv(t, "client")
	_, result := postOAuthToken(t, env.router, testTenantID, testClientPass, map[string]interface{}{
		"grant_type":  "phone_code",
		"phone":       "+86-13300001111",
		"messageCode": "88888888",
		"traceId":     testTraceID,
		"tenantId":    testTenantID,
		"platform":    testPlatform,
		"deviceId":    testDeviceID,
	}, nil)
	if !result.Success {
		t.Fatalf("expected success, got %+v", result)
	}
}

func TestHTTPClientPhoneSecretJavaFormula(t *testing.T) {
	env := setupIntegrationEnv(t, "client")
	_, result := postOAuthToken(t, env.router, testTenantID, testClientPass, map[string]interface{}{
		"grant_type": "phone_secret",
		"phone":      javaPhoneSecretPhone,
		"secret":     javaClientPhoneSecret(testTenantID),
		"traceId":    testTraceID,
		"tenantId":   testTenantID,
		"platform":   testPlatform,
		"deviceId":   testDeviceID,
	}, nil)
	if !result.Success {
		t.Fatalf("expected success, got %+v", result)
	}
}

func TestHTTPClientAppleExpiredJavaToken(t *testing.T) {
	loadLocalKeys()
	env := setupIntegrationEnv(t, "client")
	_, result := postOAuthToken(t, env.router, testTenantID, testClientPass, map[string]interface{}{
		"grant_type":    "apple",
		"identityToken": javaAppleIdentityToken,
		"fullName":      "Java Test User",
		"traceId":       testTraceID,
		"tenantId":      testTenantID,
		"platform":      testPlatform,
		"deviceId":      testDeviceID,
	}, nil)
	if result.Success || result.Code != CodeSocialLoginFailed {
		t.Fatalf("expected social login failure, got %+v", result)
	}
}

func TestHTTPClientRefreshOtherDeviceLogin(t *testing.T) {
	env := setupIntegrationEnv(t, "client")
	_, login := postOAuthToken(t, env.router, testTenantID, testClientPass, map[string]interface{}{
		"grant_type": "password",
		"username":   "client001",
		"password":   "password",
		"traceId":    testTraceID,
		"tenantId":   testTenantID,
		"platform":   testPlatform,
		"deviceId":   testDeviceID,
	}, nil)
	if !login.Success {
		t.Fatalf("login failed: %+v", login)
	}
	var token TokenCo
	_ = json.Unmarshal(login.Data, &token)

	redisKey := redis.GetLoginDeviceKey(testTenantID, "client001")
	_ = redis.Set(context.Background(), redisKey, "another-device", time.Hour)

	_, refresh := postOAuthToken(t, env.router, testTenantID, testClientPass, map[string]interface{}{
		"grant_type":    "refresh_token",
		"refresh_token": token.RefreshToken,
		"traceId":       testTraceID,
		"tenantId":      testTenantID,
		"platform":      testPlatform,
		"deviceId":      testDeviceID,
	}, nil)
	if refresh.Success || refresh.Code != CodeOtherDeviceLogin {
		t.Fatalf("expected %s, got %+v", CodeOtherDeviceLogin, refresh)
	}
}

func TestHTTPClientPhoneLogout(t *testing.T) {
	env := setupIntegrationEnv(t, "client")
	payload, _ := json.Marshal(map[string]interface{}{
		"tenantId": testTenantID,
		"phone":    "13300000001",
	})
	req := httptest.NewRequest(http.MethodPost, "/oauth/phoneLogout", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	var result apiResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !result.Success || result.Code != CodeSuccess {
		t.Fatalf("phoneLogout failed: %+v", result)
	}
}

func TestJavaPhoneSecretHashMatchesGoImplementation(t *testing.T) {
	got := javaClientPhoneSecret("1")
	want := utils.Sha256("+86-13381458187_xyy@2022@1")
	if got != want {
		t.Fatalf("phone secret mismatch: got %s want %s", got, want)
	}
}

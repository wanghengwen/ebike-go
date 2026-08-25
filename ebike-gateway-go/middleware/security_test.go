package middleware

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ebike-gateway-go/config"
	"ebike-gateway-go/response"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
)

func TestVerifyDeviceLogin_ClientModeFallback(t *testing.T) {
	// Setup Redis client pointing to local Redis
	testRdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	ctx := context.Background()
	if err := testRdb.Ping(ctx).Err(); err != nil {
		t.Skip("Local Redis not available on 127.0.0.1:6379, skipping test")
	}
	defer testRdb.Close()

	// Backup original Rdb and restore after test
	rdbMu.Lock()
	oldRdb := Rdb
	Rdb = testRdb
	rdbMu.Unlock()
	defer func() {
		rdbMu.Lock()
		Rdb = oldRdb
		rdbMu.Unlock()
	}()

	// Backup AppConfig and restore
	originalConfig := config.AppConfig
	defer func() {
		config.AppConfig = originalConfig
	}()

	config.AppConfig = &config.AppConfigStruct{
		GatewayMode: "client",
	}

	userName := "testuser"
	clientId := "testclient"
	deviceId := "device123"

	newKey := fmt.Sprintf("login_device_%s_%s", clientId, userName)
	oldKey := fmt.Sprintf("login_device_%s", userName)

	// Test Case 1: Both keys empty -> Should fail with "Token invalid or expired"
	testRdb.Del(ctx, newKey, oldKey)
	claims := jwt.MapClaims{"user_name": userName, "client_id": clientId}
	err := verifyDeviceLogin(claims, deviceId, userName, clientId)
	if err == nil || err.msg != "token无效或过期" {
		t.Errorf("Expected failure 'token无效或过期', got: %v", err)
	}

	// Test Case 2: New key is set, matches deviceId -> Should succeed
	testRdb.Set(ctx, newKey, deviceId, 10*time.Second)
	testRdb.Del(ctx, oldKey)
	err = verifyDeviceLogin(claims, deviceId, userName, clientId)
	if err != nil {
		t.Errorf("Expected success when new key matches deviceId, got error: %v", err)
	}

	// Test Case 3: New key is empty, but old key is set and matches deviceId (FALLBACK)
	// Under the Java-aligned code, this will FAIL with "Other device login" because
	// it compares cachedDeviceId (which is "") with deviceId ("device123"), even though
	// the old key exists. This strictly replicates Java client gateway behavior (defect and all).
	testRdb.Del(ctx, newKey)
	testRdb.Set(ctx, oldKey, deviceId, 10*time.Second)

	err = verifyDeviceLogin(claims, deviceId, userName, clientId)
	if err == nil || err.code != response.CodeOtherDeviceLogin {
		t.Errorf("Expected failure code %s to align with Java behavior, got: %v", response.CodeOtherDeviceLogin, err)
	}

	// Clean up keys after test
	testRdb.Del(ctx, newKey, oldKey)
}

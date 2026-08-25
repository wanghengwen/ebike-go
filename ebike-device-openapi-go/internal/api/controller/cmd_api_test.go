package controller

import (
	"net/http/httptest"
	"testing"
	"time"

	"ebike-device-openapi-go/internal/api/dto"
	"github.com/gin-gonic/gin"
)

func TestResolveEcuLoginOffline(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	_, errResult := resolveEcuLogin(c, false)
	if errResult == nil || errResult.Code != "10012" {
		t.Fatalf("expected offline error 10012, got %+v", errResult)
	}
}

func TestResolveEcuLoginInvalidType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("ecuLogin", &dto.EcuLogin{Host: "10.0.0.1", Port: 8080, Type: "hasg"})

	_, errResult := resolveEcuLogin(c, false)
	if errResult == nil || errResult.Code != "10017" {
		t.Fatalf("expected type error 10017, got %+v", errResult)
	}
}

func TestResolveEcuLoginDryRun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	ecuLogin, errResult := resolveEcuLogin(c, true)
	if errResult != nil || ecuLogin == nil || ecuLogin.Host != "127.0.0.1" || ecuLogin.Port != 8080 {
		t.Fatalf("expected dry-run fallback, got %+v err=%+v", ecuLogin, errResult)
	}
}

func TestResolveEcuLoginXiaoan(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("ecuLogin", &dto.EcuLogin{Host: "172.16.1.5", Port: 8080, Type: "xiaoan"})

	ecuLogin, errResult := resolveEcuLogin(c, false)
	if errResult != nil || ecuLogin == nil || ecuLogin.Host != "172.16.1.5" || ecuLogin.Port != 8080 {
		t.Fatalf("expected xiaoan login, got %+v err=%+v", ecuLogin, errResult)
	}
}

func TestResolveEcuLoginLegacyEmptyType(t *testing.T) {
	// 历史 Redis 缓存没有 Type 字段（反序列化为空串），必须回退到 xiaoan/TCP，
	// 既不能被拒绝，也不能走 MQTT。
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("ecuLogin", &dto.EcuLogin{Host: "172.16.1.9", Port: 8080})

	ecuLogin, errResult := resolveEcuLogin(c, false)
	if errResult != nil || ecuLogin == nil || ecuLogin.Host != "172.16.1.9" || ecuLogin.Type != "" {
		t.Fatalf("expected legacy fallback to TCP, got %+v err=%+v", ecuLogin, errResult)
	}
}

func TestResolveEcuLoginLuoping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("ecuLogin", &dto.EcuLogin{
		Type:      dto.DeviceTypeLuoping,
		DeviceId:  "10001",
		GroupName: "ecu",
	})

	ecuLogin, errResult := resolveEcuLogin(c, false)
	if errResult != nil || ecuLogin == nil || ecuLogin.Type != dto.DeviceTypeLuoping || ecuLogin.DeviceId != "10001" {
		t.Fatalf("expected luoping login, got %+v err=%+v", ecuLogin, errResult)
	}
}

func TestResolveEcuLoginLuopingStale(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	stale := time.Now().Add(-10 * time.Hour).Unix()
	c.Set("ecuLogin", &dto.EcuLogin{
		Type:       dto.DeviceTypeLuoping,
		DeviceId:   "10001",
		LastSeenAt: &stale,
	})

	_, errResult := resolveEcuLogin(c, false)
	if errResult == nil || errResult.Code != "10012" {
		t.Fatalf("expected stale offline 10012, got %+v", errResult)
	}
}

func TestDeviceRedisKey(t *testing.T) {
	if got := deviceRedisKey("860000000000001", "ebike"); got != "device_ebike_860000000000001" {
		t.Fatalf("unexpected ebike key: %s", got)
	}
	if got := deviceRedisKey("860000000000001", "bike"); got != "device_bike_860000000000001" {
		t.Fatalf("unexpected bike key: %s", got)
	}
}

func TestRequestBussinessType(t *testing.T) {
	if got := requestBussinessType(nil); got != "ebike" {
		t.Fatalf("expected ebike default, got %s", got)
	}
	if got := requestBussinessType(map[string]interface{}{"deviceType": float64(1)}); got != "bike" {
		t.Fatalf("expected bike, got %s", got)
	}
}

func TestRequestAsync(t *testing.T) {
	if got := requestAsync(nil); got != false {
		t.Fatalf("omitted async should default to false, got %v", got)
	}
	if got := requestAsync(map[string]interface{}{"async": true}); got != true {
		t.Fatalf("expected true, got %v", got)
	}
	if got := requestAsync(map[string]interface{}{"async": false}); got != false {
		t.Fatalf("expected false, got %v", got)
	}
	if got := requestAsync(map[string]interface{}{"async": nil}); got != false {
		t.Fatalf("null async should default to false, got %v", got)
	}
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ebike-gateway-go/config"
	"ebike-gateway-go/logger"

	"github.com/gin-gonic/gin"
)

func TestDebugHeadersMiddlewareDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config.AppConfig.DebugHeaders = false

	r := gin.New()
	r.Use(DebugHeadersMiddleware())
	r.GET("/client/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/client/test", nil)
	req.Header.Set("_s", "abc")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestDebugHeadersMiddlewareEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger.InitLogger()
	config.AppConfig.DebugHeaders = true
	t.Cleanup(func() {
		config.AppConfig.DebugHeaders = false
	})

	r := gin.New()
	r.Use(DebugHeadersMiddleware())
	r.GET("/client/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/client/test?foo=bar", nil)
	req.Header.Set("_s", "abc")
	req.Header.Set("_t", "1234567890")
	req.Header.Set("Authorization", "Bearer token-abcdefgh")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestMaskAuthorization(t *testing.T) {
	if got := maskAuthorization(""); got != "" {
		t.Fatalf("empty = %q", got)
	}
	if got := maskAuthorization("short"); got != "***" {
		t.Fatalf("short = %q", got)
	}
	if got := maskAuthorization("Bearer very-long-token-value"); got == "Bearer very-long-token-value" {
		t.Fatalf("expected masked value, got %q", got)
	}
}

func TestShouldSkipDebugHeaders(t *testing.T) {
	if !shouldSkipDebugHeaders("/actuator/health") {
		t.Fatal("health should be skipped")
	}
	if shouldSkipDebugHeaders("/client/test") {
		t.Fatal("client path should not be skipped")
	}
}

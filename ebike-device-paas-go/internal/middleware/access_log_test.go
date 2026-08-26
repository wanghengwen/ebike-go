package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAccessLogCapturesReqAndRes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog())
	r.POST("/demo", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/demo", strings.NewReader(`{"carId":"1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestAccessLogSkipsActuator(t *testing.T) {
	if !shouldSkipAccessLog("/actuator/health/liveness") {
		t.Fatal("expected actuator liveness skip")
	}
	if !shouldSkipAccessLog("/actuator/health/readiness") {
		t.Fatal("expected actuator readiness skip")
	}
	if !shouldSkipAccessLog("/actuator/deregisterService") {
		t.Fatal("expected actuator deregister skip")
	}
	if shouldSkipAccessLog("/device_paas/device/detail") {
		t.Fatal("expected business path not skipped")
	}
}

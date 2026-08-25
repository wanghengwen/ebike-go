package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTruncateForLog(t *testing.T) {
	in := []byte(strings.Repeat("a", 400))
	got := truncateForLog(in, maxResponseLogBytes)
	if len(got) <= maxResponseLogBytes {
		t.Fatalf("expected suffix, got len=%d", len(got))
	}
	if !strings.HasSuffix(got, "...(truncated)") {
		t.Fatalf("unexpected truncate suffix: %q", got[len(got)-20:])
	}
}

func TestAccessLogCapturesReqAndRes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLog())
	r.POST("/demo", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": strings.Repeat("x", 500)})
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
	if !shouldSkipAccessLog("/actuator/health") {
		t.Fatal("expected actuator skip")
	}
	if shouldSkipAccessLog("/parking/bindParking") {
		t.Fatal("expected business path not skipped")
	}
}

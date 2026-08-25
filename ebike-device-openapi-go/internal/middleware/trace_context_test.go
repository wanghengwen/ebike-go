package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTraceContextFromCommandContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := `{"imei":"862551059858406","commandContext":{"traceId":"trace-abc-123"}}`
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.Use(TraceContext())
	r.POST("/test", func(c *gin.Context) {
		if got := c.GetString("traceId"); got != "trace-abc-123" {
			t.Fatalf("expected traceId trace-abc-123, got %q", got)
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
}

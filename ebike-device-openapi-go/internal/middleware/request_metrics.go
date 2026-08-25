package middleware

import (
	"strings"
	"time"

	"ebike-device-openapi-go/internal/pkg/logger"
	"ebike-device-openapi-go/internal/pkg/metrics"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func shouldSkipRequestMetrics(path string) bool {
	if strings.HasPrefix(path, "/actuator/health/") || path == "/actuator/metrics" {
		return true
	}
	// Uplink decode is high-frequency; skip per-request access logs (matches Java no-access-log behavior).
	return path == "/device-gateway/xiaoan/decode"
}

// RequestMetrics logs structured request metadata and updates counters.
func RequestMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if shouldSkipRequestMetrics(path) {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		metrics.IncHTTPRequests()
		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
		}
		if traceID := c.GetString("traceId"); traceID != "" {
			fields = append(fields, zap.String("traceId", traceID))
		}

		if status >= 500 {
			logger.Log.Error("http_request", fields...)
		} else {
			logger.Log.Info("http_request", fields...)
		}
	}
}

package middleware

import (
	"net/http"
	"sort"
	"strings"

	"ebike-gateway-go/config"
	"ebike-gateway-go/logger"
	"ebike-gateway-go/sign"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func DebugHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.AppConfig.DebugHeaders {
			c.Next()
			return
		}
		if shouldSkipDebugHeaders(c.Request.URL.Path) {
			c.Next()
			return
		}

		headers := collectRequestHeaders(c.Request.Header)
		logger.Log.Info("debugHeaders",
			zap.String("traceId", GetTraceID(c)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("rawQuery", c.Request.URL.RawQuery),
			zap.String("clientIP", c.ClientIP()),
			zap.String("host", c.Request.Host),
			zap.String("userAgent", c.Request.UserAgent()),
			zap.String("contentType", c.GetHeader("Content-Type")),
			zap.String("sign", sign.GetSign(c.Request.Header)),
			zap.String("requestTime", sign.GetTimestamp(c.Request.Header)),
			zap.String("authorization", maskAuthorization(c.GetHeader("Authorization"))),
			zap.Strings("headerNames", headerNames(c.Request.Header)),
			zap.Any("headers", headers),
		)
		c.Next()
	}
}

func collectRequestHeaders(header http.Header) map[string]string {
	out := make(map[string]string, len(header))
	for key, values := range header {
		out[key] = strings.Join(values, ", ")
	}
	return out
}

func headerNames(header http.Header) []string {
	names := make([]string, 0, len(header))
	for key := range header {
		names = append(names, key)
	}
	sort.Strings(names)
	return names
}

func maskAuthorization(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 16 {
		return "***"
	}
	return value[:8] + "..." + value[len(value)-4:]
}

func shouldSkipDebugHeaders(path string) bool {
	return path == "/actuator/health" || path == "/health" || path == "/actuator/prometheus"
}

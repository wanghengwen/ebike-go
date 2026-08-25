package middleware

import (
	"fmt"
	"strings"
	"time"

	"ebike-gateway-go/config"
	"ebike-gateway-go/internal/trace"
	"ebike-gateway-go/logger"
	"ebike-gateway-go/proxy"
	"ebike-gateway-go/sign"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipAccessLog(c.Request.URL.Path) {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		if c.Request.URL.Path == "/actuator/health" || c.Request.URL.Path == "/health" {
			return
		}

		traceID := GetTraceID(c)
		if traceID == "" {
			traceID = trace.IDFromContext(c.Request.Context())
		}

		requestTime := sign.GetTimestamp(c.Request.Header)
		signValue := sign.GetSign(c.Request.Header)

		var requestBody string
		if bodyBytes, exists := c.Get("bodyBytes"); exists {
			if raw, ok := bodyBytes.([]byte); ok {
				requestBody = accessLogRequestBody(c.GetHeader("Content-Type"), raw)
			}
		}

		targetPath := "N/A"
		if routeObj, exists := c.Get("matchedRoute"); exists {
			if matchedRoute, ok := routeObj.(*proxy.Route); ok {
				targetUri := matchedRoute.Uri
				if strings.HasPrefix(targetUri, "lb://") {
					targetUri = "http://" + strings.TrimPrefix(targetUri, "lb://")
				}
				targetPath = targetUri + c.Request.URL.Path
			}
		}

		var upstreamTimeStr string
		if val, exists := c.Get("upstreamDuration"); exists {
			if d, ok := val.(time.Duration); ok {
				upstreamTimeStr = fmt.Sprintf(" | UpstreamCost: %v", d)
			}
		}

		gatewayDuration := time.Since(start)

		if c.Request.Method == "OPTIONS" {
			logger.Log.Info("CORS Preflight",
				zap.String("traceId", traceID),
				zap.String("path", c.Request.URL.Path),
				zap.String("clientIP", c.ClientIP()),
				zap.Duration("gatewayCost", gatewayDuration),
			)
			return
		}

		logger.Log.Info(fmt.Sprintf("Route: %s -> %s | ClientIP: %s | GatewayCost: %v%s",
			c.Request.URL.Path,
			targetPath,
			c.ClientIP(),
			gatewayDuration,
			upstreamTimeStr,
		), zap.String("traceId", traceID))

		logger.Log.Info("accessLog",
			zap.String("traceId", traceID),
			zap.String("requestTime", requestTime),
			zap.String("sign", signValue),
			zap.String("queryParams", c.Request.URL.RawQuery),
			zap.String("requestBody", requestBody),
			zap.String("method", c.Request.Method),
			zap.String("contentType", c.GetHeader("Content-Type")),
		)
	}
}

func accessLogRequestBody(contentType string, body []byte) string {
	if len(body) == 0 {
		return ""
	}
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if ct == "multipart/form-data" || ct == "application/octet-stream" {
		return fmt.Sprintf("[binary body omitted, %d bytes]", len(body))
	}
	return string(body)
}

func shouldSkipAccessLog(path string) bool {
	if config.AppConfig.GatewayMode != "client" {
		return false
	}
	for _, fragment := range config.AppConfig.LogExcludePaths {
		if fragment != "" && strings.Contains(path, fragment) {
			return true
		}
	}
	return false
}

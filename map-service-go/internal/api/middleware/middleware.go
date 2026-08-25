package middleware

import (
	"bytes"
	"io"
	"time"

	"map-service-go/internal/config"
	"map-service-go/internal/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

// LoggerAndContextMiddleware parses tenantId and traceId from body first (Java MapControllerAspect),
// then falls back to headers for logging context only.
func LoggerAndContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		tenantId := ""
		traceId := ""

		if c.Request.Method == "POST" && c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			var parser fastjson.Parser
			if v, err := parser.ParseBytes(bodyBytes); err == nil {
				tenantId = string(v.GetStringBytes("tenantId"))
				traceId = string(v.GetStringBytes("traceId"))
			}
		}

		if tenantId == "" {
			tenantId = c.GetHeader("tenantId")
			if tenantId == "" {
				tenantId = c.GetHeader("Tenant-Id")
			}
		}
		if traceId == "" {
			traceId = c.GetHeader("traceId")
			if traceId == "" {
				traceId = c.GetHeader("Trace-Id")
			}
		}

		ctx := log.WithTenantAndTrace(c.Request.Context(), tenantId, traceId)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		cusTime := time.Since(startTime).Milliseconds()

		logrus.WithContext(c.Request.Context()).WithFields(logrus.Fields{
			"url":      c.Request.URL.Path,
			"method":   c.Request.Method,
			"cusTime":  cusTime,
			"clientIp": c.ClientIP(),
			"status":   c.Writer.Status(),
		}).Infof("log aspect: map-service request processed")
	}
}

// SecurityMiddleware validates the "secret" header against allowed secrets in Nacos.
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check ignore URLs
		reqUri := c.Request.URL.Path
		if config.GlobalConfig != nil {
			for _, ignoreUrl := range config.GlobalConfig.Xyy.System.SecurityIgnoreUrls {
				if reqUri == ignoreUrl {
					c.Next()
					return
				}
			}
		}

		// Verify secret
		secretHeader := c.GetHeader("secret")
		if secretHeader == "" {
			c.AbortWithStatusJSON(200, gin.H{
				"success": false,
				"code":    "00006",
				"msg":     "未授权",
			})
			return
		}

		isValid := false
		if config.GlobalConfig != nil {
			for _, allowedSecret := range config.GlobalConfig.Xyy.System.Secrets {
				if secretHeader == allowedSecret {
					isValid = true
					break
				}
			}
		}

		if !isValid {
			c.AbortWithStatusJSON(200, gin.H{
				"success": false,
				"code":    "00006",
				"msg":     "未授权",
			})
			return
		}

		c.Next()
	}
}

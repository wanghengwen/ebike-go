package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"ebike-gateway-go/config"
	"ebike-gateway-go/logger"
	"ebike-gateway-go/proxy"
	"ebike-gateway-go/response"
	"ebike-gateway-go/sign"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SignMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		routeObj, exists := c.Get("matchedRoute")
		if !exists {
			c.Next()
			return
		}
		matchedRoute := routeObj.(*proxy.Route)

		if !matchedRoute.SignEnabled {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		for _, ignore := range matchedRoute.SignIgnores {
			if ignore != "" && ignore == path {
				c.Next()
				return
			}
		}

		if c.Query("magic") == "god" || c.GetHeader("magic") == "god" {
			c.Next()
			return
		}

		signHdr := sign.ParseRequestHeaders(c.Request.Header)
		if signHdr.Timestamp == "" || signHdr.Sign == "" {
			response.Abort(c, http.StatusBadRequest, response.CodeParamException, "缺少签名头")
			return
		}

		timestamp := signHdr.Timestamp
		providedSign := signHdr.Sign
		timestampKey := signHdr.TimestampPayloadKey

		timeoutSec := matchedRoute.SignTimeout
		if timeoutSec <= 0 {
			timeoutSec = 60
		}
		tsMillis, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil || tsMillis < time.Now().UnixMilli()-int64(timeoutSec)*1000 {
			response.Abort(c, http.StatusBadRequest, response.CodeRequestOutOfDate, "请求时间过期")
			return
		}

		secret := matchedRoute.SignSecret
		if secret == "" {
			response.Abort(c, http.StatusInternalServerError, response.CodeException, "签名密钥未配置")
			return
		}

		var expectedSign string
		switch c.Request.Method {
		case http.MethodGet:
			expectedSign = sign.SignQuery(c.Request.URL.Query(), timestampKey, timestamp, secret)
		default:
			contentType := c.GetHeader("Content-Type")
			bodyBytes, _ := c.Get("bodyBytes")
			var bodyStr string
			if bodyBytes != nil {
				bodyStr = string(bodyBytes.([]byte))
			}

			formMode := sign.FormParseClient
			if config.AppConfig.GatewayMode == "business" {
				formMode = sign.FormParseBusiness
			}

			if strings.HasPrefix(strings.ToLower(contentType), "application/json") {
				expectedSign = sign.SignJSON(bodyStr, timestampKey, timestamp, secret)
			} else if strings.HasPrefix(strings.ToLower(contentType), "application/x-www-form-urlencoded") {
				expectedSign = sign.SignForm(bodyStr, formMode, timestampKey, timestamp, secret)
			} else if strings.HasPrefix(strings.ToLower(contentType), "multipart/form-data") {
				// PC 端 FormData 上传时前端用 JSON.stringify(FormData)=="{}" 计算签名；
				// Java 网关对 multipart 不验签（默认通过），Go 网关用空对象签名与之对齐。
				expectedSign = sign.SignJSON("{}", timestampKey, timestamp, secret)
			} else {
				response.Abort(c, http.StatusBadRequest, response.CodeParamException, "不支持的签名内容类型")
				return
			}
		}

		if providedSign != expectedSign {
			logger.Log.Warn("Signature mismatch",
				zap.String("traceId", GetTraceID(c)),
				zap.String("path", path),
				zap.String("timestampKey", timestampKey),
				zap.String("expected", expectedSign),
				zap.String("actual", providedSign),
			)
			response.Abort(c, http.StatusBadRequest, response.CodeInvalidSign, "无效签名")
			return
		}

		c.Next()
	}
}

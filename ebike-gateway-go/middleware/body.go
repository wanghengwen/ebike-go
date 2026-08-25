package middleware

import (
	"bytes"
	"io"
	"net/http"

	"ebike-gateway-go/proxy"
	"ebike-gateway-go/response"
	"github.com/gin-gonic/gin"
)

// ReadBodyMiddleware caches request body for Sign filter on non-GET routes.
func ReadBodyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		routeObj, exists := c.Get("matchedRoute")
		if !exists {
			c.Next()
			return
		}

		matchedRoute, ok := routeObj.(*proxy.Route)
		if !ok || !matchedRoute.SignEnabled || c.Request.Body == nil {
			c.Next()
			return
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			response.Abort(c, http.StatusBadRequest, response.CodeParamException, "读取请求体失败")
			return
		}
		c.Set("bodyBytes", bodyBytes)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		c.Next()
	}
}

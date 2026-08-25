package middleware

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

// shadowHeaderName is set by ebike-gateway-go MirrorMiddleware (legacy path).
const shadowHeaderName = "X-Shadow-Java-Result"

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// ShadowDiffMiddleware compares Go response with Java response passed via header
// (legacy ebike-gateway-go MirrorMiddleware). When ProxyGateway handles shadow
// locally, this middleware is a no-op unless X-Shadow-Java-Result is set.
func ShadowDiffMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		javaResultBase64 := c.Request.Header.Get(shadowHeaderName)
		if javaResultBase64 == "" {
			c.Next()
			return
		}

		shadowWriter := &responseBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = shadowWriter
		c.Next()

		path := c.Request.URL.Path
		goRes := shadowWriter.body.String()
		header := javaResultBase64
		go func() {
			logShadowDiffFromHeader(path, header, goRes)
		}()
	}
}

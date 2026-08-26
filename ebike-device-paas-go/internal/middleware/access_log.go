package middleware

import (
	"bytes"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body   bytes.Buffer
	status int
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	if len(b) > 0 {
		_, _ = w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	if s != "" {
		_, _ = w.body.WriteString(s)
	}
	return w.ResponseWriter.WriteString(s)
}

func (w *bodyLogWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyLogWriter) Status() int {
	if w.status == 0 {
		return 200
	}
	return w.status
}

// AccessLog logs full request and response bodies for business traffic.
// Actuator probes (/actuator/*) are skipped to avoid noisy health-check logs.
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipAccessLog(c.Request.URL.Path) {
			c.Next()
			return
		}

		reqBody := readRequestBody(c)
		start := time.Now()

		blw := &bodyLogWriter{ResponseWriter: c.Writer}
		c.Writer = blw
		c.Next()

		latency := time.Since(start)
		log.Printf(
			"[ACCESS] %s | %d | %s | %s | %s %s\nReq: %s\nRes: %s\n",
			start.Format("2006/01/02 15:04:05"),
			blw.Status(),
			formatLatency(latency),
			c.ClientIP(),
			c.Request.Method,
			c.Request.URL.Path,
			formatRequestForLog(c, reqBody),
			string(blw.body.Bytes()),
		)
	}
}

func shouldSkipAccessLog(path string) bool {
	return strings.HasPrefix(path, "/actuator/")
}

func readRequestBody(c *gin.Context) []byte {
	if c.Request.Body == nil {
		return nil
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	return body
}

func formatRequestForLog(c *gin.Context, body []byte) string {
	if len(body) > 0 {
		return string(body)
	}
	if q := c.Request.URL.RawQuery; q != "" {
		return "?" + q
	}
	return ""
}

func formatLatency(d time.Duration) string {
	if d < time.Millisecond {
		return d.String()
	}
	if d < time.Second {
		return d.Round(time.Microsecond).String()
	}
	return d.Round(time.Millisecond).String()
}

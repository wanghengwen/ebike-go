package middleware

import (
	"bytes"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	maxResponseLogBytes = 300
	maxRequestLogBytes  = 8192
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body   bytes.Buffer
	total  int
	status int
}

// appendCapped grows the log buffer up to maxResponseLogBytes only. Large
// handler responses (e.g. fence/parking list payloads) must not be copied in
// full just to log a 300-byte preview — doing so would double per-request
// allocations and undercut the memory work done elsewhere in this change.
func (w *bodyLogWriter) appendCapped(b []byte) {
	w.total += len(b)
	if w.body.Len() >= maxResponseLogBytes {
		return
	}
	remaining := maxResponseLogBytes - w.body.Len()
	if remaining >= len(b) {
		w.body.Write(b)
		return
	}
	w.body.Write(b[:remaining])
}

// preview returns the buffered prefix, marking it truncated if the real
// response exceeded the buffer cap.
func (w *bodyLogWriter) preview() string {
	if w.total > w.body.Len() {
		return w.body.String() + "...(truncated)"
	}
	return w.body.String()
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	if len(b) > 0 {
		w.appendCapped(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	if s != "" {
		w.appendCapped([]byte(s))
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

// AccessLog logs request parameters and the first 300 bytes of the response body.
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
			blw.preview(),
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
		return truncateForLog(body, maxRequestLogBytes)
	}
	if q := c.Request.URL.RawQuery; q != "" {
		return "?" + q
	}
	return ""
}

func truncateForLog(b []byte, max int) string {
	if max <= 0 || len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "...(truncated)"
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

package middleware

import (
	"bytes"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// maxLoggedBody caps how much of a request or response body reaches the log.
//
// allDevices can answer with thousands of IMEIs and GPSPoints with a week of
// trajectory; logging those whole turns one request into megabytes of log, which
// costs more than the diagnostics are worth and can push the node's disk.
const maxLoggedBody = 2048

type bodyLogWriter struct {
	gin.ResponseWriter
	body   bytes.Buffer
	size   int
	status int
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.capture(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.capture([]byte(s))
	return w.ResponseWriter.WriteString(s)
}

func (w *bodyLogWriter) capture(b []byte) {
	w.size += len(b)
	if room := maxLoggedBody - w.body.Len(); room > 0 {
		if len(b) > room {
			b = b[:room]
		}
		_, _ = w.body.Write(b)
	}
}

// logged returns the captured body, noting how much was left out.
func (w *bodyLogWriter) logged() string {
	if w.size > w.body.Len() {
		return w.body.String() + truncationNote(w.size)
	}
	return w.body.String()
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
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/actuator/") {
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
			blw.logged(),
		)
	}
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
		return truncate(redactSecrets(string(body)))
	}
	if q := c.Request.URL.RawQuery; q != "" {
		return "?" + truncate(redactSecrets(q))
	}
	return ""
}

func truncate(s string) string {
	if len(s) <= maxLoggedBody {
		return s
	}
	return s[:maxLoggedBody] + truncationNote(len(s))
}

func truncationNote(total int) string {
	return "…[truncated, " + strconv.Itoa(total) + " bytes total]"
}

// secretKeys are the JSON/query keys whose values must not reach the log. Access
// logs are shipped off the node and kept for weeks; a token that appears in one
// outlives any rotation.
var secretKeys = []string{"agentToken", "token", "xc-access-token", "secret", "password", "authSecret", "authKey"}

// redactSecrets blanks the value of any secretKeys occurrence, for both the JSON
// body form ("token":"…") and the query-string form (token=…).
func redactSecrets(s string) string {
	for _, key := range secretKeys {
		s = redactJSONValue(s, key)
		s = redactQueryValue(s, key)
	}
	return s
}

func redactJSONValue(s, key string) string {
	needle := `"` + key + `"`
	for idx := 0; ; {
		at := strings.Index(s[idx:], needle)
		if at < 0 {
			return s
		}
		at += idx
		colon := at + len(needle)
		for colon < len(s) && (s[colon] == ' ' || s[colon] == ':') {
			colon++
		}
		if colon >= len(s) {
			return s
		}
		end := colon
		if s[colon] == '"' {
			end = colon + 1
			for end < len(s) && s[end] != '"' {
				end++
			}
			if end < len(s) {
				end++
			}
		} else {
			for end < len(s) && s[end] != ',' && s[end] != '}' {
				end++
			}
		}
		s = s[:colon] + `"[redacted]"` + s[end:]
		idx = colon + len(`"[redacted]"`)
	}
}

func redactQueryValue(s, key string) string {
	needle := key + "="
	for idx := 0; ; {
		at := strings.Index(s[idx:], needle)
		if at < 0 {
			return s
		}
		at += idx
		// Only match a whole parameter name, not a suffix of a longer one.
		if at > 0 && s[at-1] != '?' && s[at-1] != '&' {
			idx = at + len(needle)
			continue
		}
		start := at + len(needle)
		end := start
		for end < len(s) && s[end] != '&' {
			end++
		}
		s = s[:start] + "[redacted]" + s[end:]
		idx = start + len("[redacted]")
	}
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

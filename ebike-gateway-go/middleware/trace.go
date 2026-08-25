package middleware

import (
	"strings"

	"ebike-gateway-go/internal/trace"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

const traceIDHeader = "traceId"

// TraceMiddleware extracts or generates traceId and stores it in gin.Context
// and the request's Go context for downstream propagation.
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := extractTraceID(c)
		if traceID == "" {
			traceID = strings.ReplaceAll(uuid.NewString(), "-", "")
		}

		c.Set(traceIDHeader, traceID)
		ctx := trace.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		if c.GetHeader(traceIDHeader) == "" {
			c.Request.Header.Set(traceIDHeader, traceID)
		}

		c.Next()
	}
}

func extractTraceID(c *gin.Context) string {
	if tid := c.GetHeader(traceIDHeader); tid != "" {
		return tid
	}
	if tid := c.Query(traceIDHeader); tid != "" {
		return tid
	}

	if bodyBytes, exists := c.Get("bodyBytes"); exists {
		if raw, ok := bodyBytes.([]byte); ok && len(raw) > 0 {
			if tid := gjson.GetBytes(raw, traceIDHeader).String(); tid != "" {
				return tid
			}
		}
	}

	return ""
}

// GetTraceID returns traceId from gin context.
func GetTraceID(c *gin.Context) string {
	if v, exists := c.Get(traceIDHeader); exists {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return trace.IDFromContext(c.Request.Context())
}

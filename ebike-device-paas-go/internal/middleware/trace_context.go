package middleware

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
)

// TraceContext extracts traceId / tenantId from the JSON body's commandContext
// (or top-level / query) for structured logging and downstream propagation.
func TraceContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		if traceID := c.Query("traceId"); traceID != "" {
			c.Set("traceId", traceID)
			c.Next()
			return
		}
		if c.Request.Body == nil {
			c.Next()
			return
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}
		// Restore body so handlers can read it again.
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		if len(bodyBytes) == 0 {
			c.Next()
			return
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
			c.Next()
			return
		}

		if traceID, ok := parsed["traceId"].(string); ok && traceID != "" {
			c.Set("traceId", traceID)
		}
		if ctx, ok := parsed["commandContext"].(map[string]interface{}); ok {
			if traceID, ok := ctx["traceId"].(string); ok && traceID != "" {
				c.Set("traceId", traceID)
			}
			if tenantID, ok := ctx["tenantId"].(string); ok && tenantID != "" {
				c.Set("tenantId", tenantID)
			}
		}

		c.Next()
	}
}

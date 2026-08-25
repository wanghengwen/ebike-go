package middleware

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware returns HTTP 200 JSON errors like Java GlobalExceptionHandler.
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				var errMsg string
				switch v := r.(type) {
				case error:
					errMsg = v.Error()
				default:
					errMsg = fmt.Sprint(v)
				}

				frame := firstPanicFrame(string(debug.Stack()))
				code := fmt.Sprintf("panic:%s %s", errMsg, frame)
				c.AbortWithStatusJSON(200, gin.H{
					"success": false,
					"code":    code,
					"msg":     errMsg,
					"data":    nil,
				})
			}
		}()
		c.Next()
	}
}

func firstPanicFrame(stack string) string {
	lines := strings.Split(stack, "\n")
	for i, line := range lines {
		if strings.Contains(line, ".go:") && i > 0 {
			return strings.TrimSpace(lines[i-1]) + " " + strings.TrimSpace(line)
		}
	}
	return ""
}

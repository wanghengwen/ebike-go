package middleware

import (
	"net/http"

	"ebike-fence-go/internal/api/dto"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GlobalErrorMiddleware catches panics and handles global errors.
// Mirrors Java's GlobalExceptionHandler.handleException: any uncaught exception
// is returned as {success:false, code:"00001", msg:"...", data:null} with HTTP 200.
func GlobalErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				zap.L().Error("Panic recovered in middleware", zap.Any("panic_err", err))
				c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeException, "系统繁忙，请稍后再试 (System Busy)"))
				c.Abort()
			}
		}()

		c.Next()

		// If there are errors collected in the Gin context during the request
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			// For other generic errors, do not expose raw error string to frontend for security reasons
			zap.L().Error("Unhandled error in middleware", zap.Error(err))
			c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeException, "系统繁忙，请稍后再试 (System Busy)"))
			return
		}
	}
}

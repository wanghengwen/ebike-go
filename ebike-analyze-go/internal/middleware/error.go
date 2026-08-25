package middleware

import (
	"fmt"
	"net/http"

	"ebike-analyze-go/internal/api/dto"
	"ebike-analyze-go/internal/common/bizerror"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GlobalErrorMiddleware catches panics and handles global errors.
// Mirrors Java's GlobalExceptionHandler.handleException.
func GlobalErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				zap.L().Error("Panic recovered in middleware", zap.Any("panic_err", err))
				var panicErr error
				switch e := err.(type) {
				case error:
					panicErr = e
				default:
					panicErr = fmt.Errorf("%v", e)
				}
				c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeException, bizerror.FormatException(panicErr)))
				c.Abort()
			}
		}()

		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			zap.L().Error("Unhandled error in middleware", zap.Error(err))
			c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeException, bizerror.FormatException(err)))
		}
	}
}

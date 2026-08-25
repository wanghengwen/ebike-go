package middleware

import (
	"fmt"
	"log"
	"net/http"

	"ebike-service-client-go/internal/api/dto"
	"github.com/gin-gonic/gin"
)

// GlobalErrorMiddleware catches panics and handles global errors.
// Mirrors Java's GlobalExceptionHandler.handleException: any uncaught exception
// is returned as {success:false, code:"00001", msg:"...", data:null} with HTTP 200.
func GlobalErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeException, fmt.Sprintf("panic: %v", err)))
				c.Abort()
			}
		}()

		c.Next()

		// If there are errors collected in the Gin context during the request
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			log.Printf("Request error: %v", err.Error())
			c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeException, err.Error()))
			return
		}
	}
}

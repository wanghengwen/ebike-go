package handler

import (
	"log"
	"net/http"

	"push-notification-go/internal/pkg/errcode"
	"push-notification-go/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware returns a Gin middleware that recovers from panics.
// If the panic value is a *errcode.BizError, the corresponding error code
// and message are returned. Otherwise, a generic 500 response is sent.
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				switch v := r.(type) {
				case *errcode.BizError:
					log.Printf("[Recovery] caught BizError: code=%s msg=%s", v.Code, v.Msg)
					c.JSON(http.StatusOK, response.Error(v))
				case error:
					log.Printf("[Recovery] caught error: %v", v)
					c.JSON(http.StatusInternalServerError, response.Exception(v.Error()))
				default:
					log.Printf("[Recovery] caught panic: %v", v)
					c.JSON(http.StatusInternalServerError, response.Exception("internal server error"))
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}

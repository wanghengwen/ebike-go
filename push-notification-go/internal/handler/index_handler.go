package handler

import (
	"time"

	"push-notification-go/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// IndexHandler returns the current server time, mirroring the Java IndexController.
func IndexHandler(c *gin.Context) {
	response.SuccessJSON(c, time.Now().Format("2006-01-02T15:04:05.000"))
}

package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutes_ActuatorPathsDoNotPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r, nil)
}

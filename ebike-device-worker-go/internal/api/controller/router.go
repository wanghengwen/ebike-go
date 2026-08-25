package controller

import (
	"ebike-device-worker-go/internal/api"
	"ebike-device-worker-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires all ebike-device-worker endpoints from the generated manifest.
func RegisterRoutes(r *gin.Engine) {
	InitNativeHandlers()
	r.Use(middleware.ProxyGateway())

	r.GET("/actuator/health", Health)
	r.POST("/actuator/deregisterService", DeregisterService)
	registerActuatorDebug(r.Group(""))

	for _, rt := range api.GeneratedRoutes() {
		path := rt.Path
		handler := api.Dispatch(path)
		switch rt.Method {
		case "GET":
			r.GET(path, handler)
		default:
			r.POST(path, handler)
		}
	}
}

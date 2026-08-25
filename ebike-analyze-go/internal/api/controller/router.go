package controller

import (
	"ebike-analyze-go/internal/api"
	"ebike-analyze-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires all ebike-analyze endpoints from the generated manifest.
func RegisterRoutes(r *gin.RouterGroup) {
	registerNativeHandlers()
	r.Use(middleware.ProxyGateway())

	r.GET("/actuator/health", Health)
	r.POST("/actuator/health", Health)
	r.POST("/actuator/deregisterService", DeregisterService)
	registerActuatorDebug(r)

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

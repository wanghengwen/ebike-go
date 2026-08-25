package controller

import (
	"ebike-fence-go/internal/api"
	"ebike-fence-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires all ebike-fence endpoints from the generated manifest.
func RegisterRoutes(r *gin.RouterGroup) {
	registerNativeHandlers()
	r.Use(middleware.ProxyGateway())

	r.GET("/actuator/health", Health)
	r.POST("/actuator/health", Health)
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

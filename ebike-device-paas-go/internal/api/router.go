package api

import (
	"net/http"

	"ebike-device-paas-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

// NewRouter builds the Gin engine with base middleware and health probes.
// Endpoint handlers are attached per-phase as they are migrated.
func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.AccessLog())
	r.Use(middleware.ProxyGateway())
	r.Use(middleware.TraceContext())

	// Kubernetes probes (match Spring actuator paths used by the Java service).
	r.GET("/actuator/health/liveness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})
	r.GET("/actuator/health/readiness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})
	r.POST("/actuator/deregisterService", DeregisterService)

	registerEcuQueryRoutes(r)
	registerEcuJobQueryRoutes(r)
	registerDeviceTrajectoryRoutes(r)
	registerEcuCommandRoutes(r)
	registerStateChangeRoutes(r)
	registerEcuBleCommandReportRoutes(r)

	return r
}

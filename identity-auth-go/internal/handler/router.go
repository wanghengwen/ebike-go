package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"identity-auth-go/internal/model"
	"identity-auth-go/internal/service"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	querySvc := service.NewQueryService(db)
	authSvc := service.NewAuthService(db)

	queryHandler := NewQueryHandler(querySvc)
	authHandler := NewAuthHandler(authSvc)

	r.POST("/auth", authHandler.HandleAuth)
	r.POST("/charge", authHandler.HandleCharge)

	r.POST("/countAuthTimes", queryHandler.CountAuthTimes)
	r.POST("/queryAuthRecord", queryHandler.QueryAuthRecord)
	r.GET("/exportAuthRecord", queryHandler.ExportAuthRecord)

	actuatorOK := func(c *gin.Context) {
		c.JSON(200, model.R{Success: true})
	}
	// Gin does not allow a catch-all (e.g. /actuator/*any) alongside fixed
	// segments like /actuator/health. Register explicit paths instead; K8s
	// probes use /actuator/health/liveness and /actuator/health/readiness.
	for _, path := range []string{
		"/actuator",
		"/actuator/health",
		"/actuator/health/liveness",
		"/actuator/health/readiness",
	} {
		r.GET(path, actuatorOK)
	}
}

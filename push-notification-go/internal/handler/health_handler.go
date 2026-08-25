package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler serves Spring Boot Actuator-compatible health checks.
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler creates a HealthHandler.
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health handles GET /actuator/health.
func (h *HealthHandler) Health(c *gin.Context) {
	dbStatus := "UP"
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "DOWN"
		}
	}

	status := "UP"
	if dbStatus == "DOWN" {
		status = "DOWN"
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": status,
			"components": gin.H{
				"db": gin.H{"status": dbStatus},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"components": gin.H{
			"db": gin.H{"status": dbStatus},
		},
	})
}

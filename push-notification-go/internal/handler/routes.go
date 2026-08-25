package handler

import (
	"push-notification-go/internal/pool"
	"push-notification-go/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes creates all service instances and registers HTTP routes on the
// given Gin engine.
func RegisterRoutes(r *gin.Engine, db *gorm.DB, msgPool *pool.WorkerPool, voicePool *pool.WorkerPool) {
	// Create services
	messageService := service.NewMessageService(db, msgPool)
	voiceService := service.NewVoiceService(db, voicePool)

	// Create handlers
	messageHandler := NewMessageHandler(messageService)
	voiceHandler := NewVoiceHandler(voiceService)
	debugHandler := NewDebugHandler()
	healthHandler := NewHealthHandler(db)

	// Index
	r.GET("/", IndexHandler)

	// Message
	r.POST("/message/send", messageHandler.SendMessage)

	// Voice
	r.POST("/voice/send", voiceHandler.SendVoice)

	// Debug
	r.GET("/debug/messageSign", debugHandler.MessageSign)
	r.GET("/debug/supplier", debugHandler.Supplier)
	r.GET("/debug/template", debugHandler.Template)
	r.GET("/debug/tenantConfig", debugHandler.TenantConfig)

	// Actuator health for K8s probes (Spring Boot compatible shape)
	r.GET("/actuator/health", healthHandler.Health)
}

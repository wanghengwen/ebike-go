package handler

import (
	"push-notification-go/internal/cache"
	"push-notification-go/internal/pkg/config"
	"push-notification-go/internal/pkg/errcode"
	"push-notification-go/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// DebugHandler provides endpoints for inspecting cached data.
type DebugHandler struct{}

// NewDebugHandler creates a new DebugHandler.
func NewDebugHandler() *DebugHandler {
	return &DebugHandler{}
}

// MessageSign handles GET /debug/messageSign.
func (h *DebugHandler) MessageSign(c *gin.Context) {
	secret := c.Query("secret")
	if !config.VerifySecret(secret) {
		response.ErrorJSON(c, errcode.ErrSecretMistake)
		return
	}
	response.SuccessJSON(c, cache.MessageSignCache.GetAll())
}

// Supplier handles GET /debug/supplier.
func (h *DebugHandler) Supplier(c *gin.Context) {
	secret := c.Query("secret")
	if !config.VerifySecret(secret) {
		response.ErrorJSON(c, errcode.ErrSecretMistake)
		return
	}
	response.SuccessJSON(c, cache.SupplierCache.GetAll())
}

// Template handles GET /debug/template.
func (h *DebugHandler) Template(c *gin.Context) {
	secret := c.Query("secret")
	if !config.VerifySecret(secret) {
		response.ErrorJSON(c, errcode.ErrSecretMistake)
		return
	}
	response.SuccessJSON(c, cache.TemplateCache.GetAll())
}

// TenantConfig handles GET /debug/tenantConfig.
func (h *DebugHandler) TenantConfig(c *gin.Context) {
	secret := c.Query("secret")
	if !config.VerifySecret(secret) {
		response.ErrorJSON(c, errcode.ErrSecretMistake)
		return
	}
	response.SuccessJSON(c, cache.TenantConfigCache.GetAll())
}

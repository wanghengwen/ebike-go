package user

// Ports Java TenantController + TenantServiceImpl + TenantGatewayImpl.

import (
	"encoding/json"
	"net/http"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// tenantConfigCo mirrors TenantConfigCo (primitive boolean, default false).
type tenantConfigCo struct {
	CreditScore bool `json:"creditScore"`
}

// POST /client/tenant/config
func tenantConfig(c *gin.Context) {
	// TenantDTO adds no fields to ClientDTO
	var req dto.ClientDTO
	if !web.BindJSON(c, &req, baseMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req)

	// TenantGatewayImpl.config wraps everything in try/catch — any failure
	// (transport, failed Result, even the NPE on a null status) keeps the
	// default creditScore=false.
	out := tenantConfigCo{CreditScore: false}
	if raw := callDataQuiet(c.Request.Context(), rpc.ServiceUser, "/creditScore/v2/getConfig",
		map[string]interface{}{}, cmdCtx); raw != nil && !isNullJSON(raw) {
		var cfg struct {
			Status *int `json:"status"`
		}
		if err := json.Unmarshal(raw, &cfg); err == nil && cfg.Status != nil {
			out.CreditScore = *cfg.Status > 0
		}
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(out))
}

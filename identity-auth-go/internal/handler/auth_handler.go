package handler

import (
	"encoding/json"
	"net/http"

	"identity-auth-go/internal/model"
	"identity-auth-go/internal/pkg/config"
	"identity-auth-go/internal/pkg/shadow"
	idvalidator "identity-auth-go/internal/pkg/validator"
	"identity-auth-go/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// HandleAuth handles POST /auth
// When DRY_RUN=true: saves to shadow record list and does shadow comparison.
// When DRY_RUN=false: executes full auth flow.
func (h *AuthHandler) HandleAuth(c *gin.Context) {
	reqBody, _ := c.GetRawData()

	var req model.AuthRequest
	if err := json.Unmarshal(reqBody, &req); err != nil {
		// Java GlobalExceptionHandler always returns HTTP 200
		c.JSON(http.StatusOK, model.R{Success: false, Msg: err.Error()})
		return
	}

	if msg := idvalidator.ValidateAuthRequest(req.TraceId, req.TenantId, req.Name, req.IdCardNum, req.Image, req.Type); msg != "" {
		c.JSON(http.StatusOK, model.R{Success: false, Msg: msg})
		return
	}

	if config.GlobalConfig.DryRun {
		// Dry-run mode: record to shadow list and compare with Java
		err := h.svc.RecordAuth(c.Request.URL.Path, req)
		if err != nil {
			c.JSON(http.StatusOK, model.R{Success: false, Msg: "Failed to record auth: " + err.Error()})
			return
		}
		resp := model.R{Success: true, Msg: "success (dry-run)"}
		goRespBytes, _ := json.Marshal(resp)
		shadow.CompareWithJava("POST", c.Request.URL.Path, reqBody, goRespBytes)
		c.Data(http.StatusOK, "application/json", goRespBytes)
		return
	}

	// Production mode: full auth flow
	resp := h.svc.Auth(req)
	c.JSON(http.StatusOK, resp)
}

// HandleCharge handles POST /charge
// When DRY_RUN=true: saves to shadow record list.
// When DRY_RUN=false: executes full charge flow.
func (h *AuthHandler) HandleCharge(c *gin.Context) {
	reqBody, _ := c.GetRawData()

	var req model.ChargeRequest
	if err := json.Unmarshal(reqBody, &req); err != nil {
		c.JSON(http.StatusOK, model.R{Success: false, Msg: err.Error()})
		return
	}

	if msg := idvalidator.ValidateChargeRequest(req.TraceId, req.TenantId, req.Amount, req.Quantity, req.Type); msg != "" {
		c.JSON(http.StatusOK, model.R{Success: false, Msg: msg})
		return
	}

	if config.GlobalConfig.DryRun {
		err := h.svc.RecordCharge(c.Request.URL.Path, req)
		if err != nil {
			c.JSON(http.StatusOK, model.R{Success: false, Msg: "Failed to record charge: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, model.R{Success: true, Msg: "success (dry-run)"})
		return
	}

	// Production mode
	resp := h.svc.Charge(req)
	c.JSON(http.StatusOK, resp)
}

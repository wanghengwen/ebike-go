package controller

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"luopingtech-fastid-register/internal/pkg/config"
	"luopingtech-fastid-register/internal/service"

	"github.com/gin-gonic/gin"
)

// FastIdHandler handles HTTP requests for the FastId registration service.
type FastIdHandler struct {
	service service.IFastIdService
	cfg     *config.AppConfig
}

// NewFastIdHandler creates a new handler instance.
func NewFastIdHandler(svc service.IFastIdService, cfg *config.AppConfig) *FastIdHandler {
	return &FastIdHandler{service: svc, cfg: cfg}
}

// RegisterRoutes sets up all HTTP routes on the given Gin engine.
func (h *FastIdHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/fastid/machineId", h.MachineID)
	r.GET("/fastid/appList", h.AppList)
	r.GET("/fastid/machineList", h.MachineList)
	r.GET("/actuator/health", h.Health)
}

// MachineID handles POST /fastid/machineId
// Replicates Java's FastidController.machineId exactly:
//   - Reads form parameters: namespace, groupId, appName, machineUuid, secret
//   - Validates secret against configured secrets
//   - Returns query-string format: "success=true&msg=&result=<id>" or "success=false&msg=<error>&result="
//   - Content-Type: text/plain (matching Spring Boot's String return type)
func (h *FastIdHandler) MachineID(c *gin.Context) {
	namespace := c.PostForm("namespace")
	groupID := c.PostForm("groupId")
	appName := c.PostForm("appName")
	machineUUID := c.PostForm("machineUuid")
	secret := c.PostForm("secret")

	// Verify secret (matches Java: verifySecret)
	if !h.verifySecret(secret) {
		c.String(http.StatusOK, response(false, "秘钥验证失败", ""))
		return
	}

	machineID, err := h.service.GetMachineID(namespace, groupID, appName, machineUUID)
	if err != nil {
		log.Printf("[ERROR] GetMachineID failed: %v", err)
		c.String(http.StatusOK, response(false, err.Error(), ""))
		return
	}

	c.String(http.StatusOK, response(true, "", machineID))
}

// AppList handles GET /fastid/appList
// Returns a JSON array of all registered apps (matches Java's List<App> return type).
func (h *FastIdHandler) AppList(c *gin.Context) {
	apps, err := h.service.ListApps()
	if err != nil {
		log.Printf("[ERROR] ListApps failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, apps)
}

// MachineList handles GET /fastid/machineList?appId=<id>
// Returns a JSON array of machines under the given app ID.
func (h *FastIdHandler) MachineList(c *gin.Context) {
	appIDStr := c.Query("appId")
	if appIDStr == "" {
		// Java Jackson serializes a null return value as "null"
		c.Data(http.StatusOK, "application/json;charset=UTF-8", []byte("null"))
		return
	}

	appID, err := strconv.ParseUint(appIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid appId"})
		return
	}

	machines, err := h.service.ListMachines(uint(appID))
	if err != nil {
		log.Printf("[ERROR] ListMachines failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, machines)
}

// verifySecret checks if the provided secret is in the configured secrets list.
// Matches Java's FastidController.verifySecret behavior:
//   - Returns false if secret is empty
//   - Returns false if no secrets are configured
//   - Returns true only if secret is found in the set
func (h *FastIdHandler) verifySecret(secret string) bool {
	if secret == "" {
		return false
	}
	secrets := h.cfg.FastId.Secrets
	if len(secrets) == 0 {
		return false
	}
	for _, s := range secrets {
		if s == secret {
			return true
		}
	}
	return false
}

// response builds the query-string format response body identical to Java's
// ResponseUtils.response(boolean, String, Object):
//
//	"success=true&msg=&result=123"
//	"success=false&msg=秘钥验证失败&result="
func response(success bool, msg string, result interface{}) string {
	return fmt.Sprintf("success=%v&msg=%s&result=%v", success, msg, result)
}

// Health handles GET /actuator/health
// Replicates Spring Boot Actuator health endpoint for K8S probes.
func (h *FastIdHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP"})
}

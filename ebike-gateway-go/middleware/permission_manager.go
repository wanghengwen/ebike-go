package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"ebike-gateway-go/config"
	"ebike-gateway-go/logger"
	"go.uber.org/zap"
)

type ApiPermissionCo struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Apis string `json:"apis"`
}

type InterfacePermissionCo struct {
	RoleId   string            `json:"roleId"`
	TenantId string            `json:"tenantId"`
	MenuList []ApiPermissionCo `json:"menuList"`
}

type PermissionResponse struct {
	Code    string                  `json:"code"`
	Msg     string                  `json:"msg"`
	Success bool                    `json:"success"`
	Data    []InterfacePermissionCo `json:"data"`
}

type RolePermission struct {
	TenantId string
	RoleId   string
	ApiSet   map[string]bool
}

type PermissionManager struct {
	mu           sync.RWMutex
	rolePermsMap map[string]*RolePermission
}

var PermManager *PermissionManager

func InitPermissionManager() {
	if !config.AppConfig.EnableApiVerify {
		return
	}

	PermManager = &PermissionManager{
		rolePermsMap: make(map[string]*RolePermission),
	}

	// Initial refresh
	PermManager.Refresh()

	// Start periodic refresh
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			PermManager.Refresh()
		}
	}()
}

func (pm *PermissionManager) Refresh() {
	if !config.AppConfig.EnableApiVerify {
		return
	}
	logger.Log.Info("Starting RolePermissionManager refresh")

	// Read url from config (configurable via MANAGEMENT_URL env var)
	url := config.AppConfig.ManagementUrl + "/permission/list"

	// Java code creates PermissionCmd with a CommandContext (tenantId: "0", traceId: uuid)
	traceId := fmt.Sprintf("gateway-refresh-%d", time.Now().UnixNano())
	reqBodyStr := fmt.Sprintf(`{"commandContext":{"tenantId":"0","traceId":"%s"}}`, traceId)
	reqBody := []byte(reqBodyStr)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Log.Error("Failed to create permission refresh request", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error("Failed to fetch permissions from ebike-management", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("Failed to read permission response body", zap.Error(err))
		return
	}

	var res PermissionResponse
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		logger.Log.Error("Failed to parse permission response", zap.Error(err))
		return
	}

	if !res.Success {
		logger.Log.Error("Permission response was not success", zap.String("msg", res.Msg))
		return
	}

	newMap := make(map[string]*RolePermission)
	for _, ipco := range res.Data {
		rp := &RolePermission{
			TenantId: ipco.TenantId,
			RoleId:   ipco.RoleId,
			ApiSet:   make(map[string]bool),
		}

		for _, menu := range ipco.MenuList {
			if menu.Apis != "" {
				apis := strings.Split(menu.Apis, ",")
				for _, api := range apis {
					if api = strings.TrimSpace(api); api != "" {
						rp.ApiSet[api] = true
					}
				}
			}
		}
		key := fmt.Sprintf("%s:%s", rp.TenantId, rp.RoleId)
		newMap[key] = rp
	}

	pm.mu.Lock()
	pm.rolePermsMap = newMap
	pm.mu.Unlock()

	logger.Log.Info("RolePermissionManager refreshed successfully", zap.Int("rolesCount", len(newMap)))
}

func (pm *PermissionManager) CanAccessApi(tenantId string, roleId string, apiPath string) bool {
	if tenantId == "" || roleId == "" || apiPath == "" {
		return false
	}

	if config.IsExcludeApi(apiPath) {
		return true
	}

	key := fmt.Sprintf("%s:%s", tenantId, roleId)

	pm.mu.RLock()
	rp, exists := pm.rolePermsMap[key]
	pm.mu.RUnlock()

	if !exists {
		return false
	}

	// Java implementation simply uses apis.contains(api)
	return rp.ApiSet[apiPath]
}

package auth

import (
	"fmt"
	"sync"
	"time"

	"ebike-auth-go/internal/pkg/logger"
	"ebike-auth-go/internal/pkg/rpc"

	"go.uber.org/zap"
)

var (
	tenantAuthMap      = make(map[string]*rpc.TenantAuthCo)
	tenantThirdAuthMap = make(map[string]*rpc.TenantThirdAuthCo)
	tenantMu           sync.RWMutex
)

func InitTenantAuthManager() {
	refreshTenantMaps()
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			refreshTenantMaps()
		}
	}()
}

func refreshTenantMaps() {
	tenantMu.Lock()
	defer tenantMu.Unlock()

	// 1. Fetch Tenant Auths
	auths, err := rpc.QueryAllTenantAuth()
	if err != nil {
		logger.Log.Error("Failed to query all tenant auths from management", zap.Error(err))
	} else {
		tenantAuthMap = make(map[string]*rpc.TenantAuthCo)
		for i := range auths {
			tenantAuthMap[auths[i].TenantId] = &auths[i]
		}
		logger.Log.Info("Refreshed tenantAuthMap", zap.Int("count", len(tenantAuthMap)))
	}

	// 2. Fetch Tenant Third Auths
	thirds, err := rpc.QueryAllTenantThirdAuth()
	if err != nil {
		logger.Log.Error("Failed to query all tenant third auths from management", zap.Error(err))
	} else {
		tenantThirdAuthMap = make(map[string]*rpc.TenantThirdAuthCo)
		for i := range thirds {
			// key format: tenantId + "_" + thirdType
			key := fmt.Sprintf("%s_%d", thirds[i].TenantId, thirds[i].ThirdType)
			tenantThirdAuthMap[key] = &thirds[i]
		}
		logger.Log.Info("Refreshed tenantThirdAuthMap", zap.Int("count", len(tenantThirdAuthMap)))
	}
}

func GetTenantAuth(tenantId string) (*rpc.TenantAuthCo, error) {
	tenantMu.RLock()
	defer tenantMu.RUnlock()

	val, ok := tenantAuthMap[tenantId]
	if !ok {
		return nil, fmt.Errorf("tenant auth not found for tenantId: %s", tenantId)
	}
	return val, nil
}

func GetTenantThirdAuth(tenantId string, thirdType int) (*rpc.TenantThirdAuthCo, error) {
	tenantMu.RLock()
	defer tenantMu.RUnlock()

	key := fmt.Sprintf("%s_%d", tenantId, thirdType)
	val, ok := tenantThirdAuthMap[key]
	if !ok {
		return nil, fmt.Errorf("tenant third auth not found for key: %s", key)
	}
	return val, nil
}

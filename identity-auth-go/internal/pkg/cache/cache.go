package cache

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"

	"identity-auth-go/internal/model"
)

const refreshInterval = 60 * time.Second

var (
	mu                sync.RWMutex
	tenantConfigCache = make(map[string]*model.TenantConfig)
	supplierCache     = make(map[int64]*model.Supplier)
)

// Init starts periodic full-table refresh matching Java LocalCache (every 60s).
func Init(db *gorm.DB) {
	refresh(db)
	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()
		for range ticker.C {
			refresh(db)
		}
	}()
}

func refresh(db *gorm.DB) {
	var tenantConfigs []model.TenantConfig
	if err := db.Find(&tenantConfigs).Error; err != nil {
		log.Printf("[cache] load tenant_config error: %v", err)
	} else {
		mu.Lock()
		for _, tc := range tenantConfigs {
			key := tenantConfigKey(tc.TenantId, tc.Type)
			if tc.IzDel == 1 {
				delete(tenantConfigCache, key)
			} else {
				tcCopy := tc
				tenantConfigCache[key] = &tcCopy
			}
		}
		mu.Unlock()
	}

	var suppliers []model.Supplier
	if err := db.Find(&suppliers).Error; err != nil {
		log.Printf("[cache] load supplier error: %v", err)
	} else {
		mu.Lock()
		for _, s := range suppliers {
			if s.IzDel == 1 {
				delete(supplierCache, s.ID)
			} else {
				sCopy := s
				supplierCache[s.ID] = &sCopy
			}
		}
		mu.Unlock()
	}
}

func tenantConfigKey(tenantId string, authType int) string {
	return fmt.Sprintf("%s_%d", tenantId, authType)
}

// GetTenantConfig gets TenantConfig by tenantId + type, with cache.
// Matches Java: LocalCache.getTenantConfig(tenantId, type)
func GetTenantConfig(db *gorm.DB, tenantId string, authType int) (*model.TenantConfig, error) {
	key := tenantConfigKey(tenantId, authType)

	mu.RLock()
	if tc, ok := tenantConfigCache[key]; ok {
		mu.RUnlock()
		return tc, nil
	}
	mu.RUnlock()

	var tc model.TenantConfig
	err := db.Where("tenant_id = ? AND type = ? AND iz_del = 0", tenantId, authType).First(&tc).Error
	if err != nil {
		return nil, fmt.Errorf("租户[%s],类型[%d]不存在", tenantId, authType)
	}

	mu.Lock()
	tenantConfigCache[key] = &tc
	mu.Unlock()
	return &tc, nil
}

// GetSupplier gets Supplier by id, with cache.
// Matches Java: LocalCache.getSupplier(supplierId)
func GetSupplier(db *gorm.DB, id int64) (*model.Supplier, error) {
	mu.RLock()
	if s, ok := supplierCache[id]; ok {
		mu.RUnlock()
		return s, nil
	}
	mu.RUnlock()

	var s model.Supplier
	err := db.Where("id = ? AND iz_del = 0", id).First(&s).Error
	if err != nil {
		return nil, fmt.Errorf("提供商[%d]不存在", id)
	}

	mu.Lock()
	supplierCache[id] = &s
	mu.Unlock()
	return &s, nil
}

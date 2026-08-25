package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"

	"push-notification-go/internal/model"

	"gorm.io/gorm"
)

const (
	TypeMessage = 1 // 短信
	TypeVoice   = 2 // 语音
)

// TenantConfigCacheStore holds an atomically-swapped map of TenantConfig
// keyed by "{tenantId}_{type}".
type TenantConfigCacheStore struct {
	db       *gorm.DB
	data     atomic.Value // stores map[string]*model.TenantConfig
	lastHash string
}

// NewTenantConfigCache creates a new TenantConfigCacheStore backed by the given db.
func NewTenantConfigCache(db *gorm.DB) *TenantConfigCacheStore {
	c := &TenantConfigCacheStore{db: db}
	c.data.Store(make(map[string]*model.TenantConfig))
	return c
}

// buildKey returns the cache key for a tenant config entry.
func buildTenantConfigKey(tenantId string, configType int) string {
	return fmt.Sprintf("%s_%d", tenantId, configType)
}

// Get returns the TenantConfig for the given tenantId and config type, or nil if not found.
func (c *TenantConfigCacheStore) Get(tenantId string, configType int) *model.TenantConfig {
	m := c.data.Load().(map[string]*model.TenantConfig)
	return m[buildTenantConfigKey(tenantId, configType)]
}

// GetAll returns the full cached map.
func (c *TenantConfigCacheStore) GetAll() map[string]*model.TenantConfig {
	return c.data.Load().(map[string]*model.TenantConfig)
}

// Refresh reloads all TenantConfig rows from the database.
// On error or empty result set, the old data is preserved.
func (c *TenantConfigCacheStore) Refresh() {
	var list []model.TenantConfig
	if err := c.db.Where("iz_del = 0").Find(&list).Error; err != nil {
		log.Printf("[TenantConfigCache] refresh failed, keeping old data: %v", err)
		return
	}
	if len(list) == 0 {
		log.Printf("[TenantConfigCache] refresh returned empty list, keeping old data")
		return
	}

	newMap := make(map[string]*model.TenantConfig, len(list))
	for i := range list {
		key := buildTenantConfigKey(list[i].TenantId, list[i].Type)
		newMap[key] = &list[i]
	}
	c.data.Store(newMap)

	jsonData, _ := json.Marshal(list)
	hash := string(jsonData)
	isFirst := c.lastHash == ""
	isChanged := c.lastHash != hash
	c.lastHash = hash

	if isFirst || isChanged {
		log.Printf("[TenantConfigCache] refreshed, loaded %d entries", len(newMap))
	}
}

package cache

import (
	"encoding/json"
	"log"
	"sync/atomic"

	"push-notification-go/internal/model"

	"gorm.io/gorm"
)

// MessageSignCacheStore holds an atomically-swapped map of sign name sets
// keyed by tenantId. Each value is a set (map[string]struct{}) of valid sign names.
type MessageSignCacheStore struct {
	db       *gorm.DB
	data     atomic.Value // stores map[string]map[string]struct{}
	lastHash string
}

// NewMessageSignCache creates a new MessageSignCacheStore backed by the given db.
func NewMessageSignCache(db *gorm.DB) *MessageSignCacheStore {
	c := &MessageSignCacheStore{db: db}
	c.data.Store(make(map[string]map[string]struct{}))
	return c
}

// Contains checks whether the given signName is registered for the specified tenantId.
func (c *MessageSignCacheStore) Contains(tenantId, signName string) bool {
	m := c.data.Load().(map[string]map[string]struct{})
	signs, ok := m[tenantId]
	if !ok {
		return false
	}
	_, exists := signs[signName]
	return exists
}

// GetAll returns the full cached map.
func (c *MessageSignCacheStore) GetAll() map[string]map[string]struct{} {
	return c.data.Load().(map[string]map[string]struct{})
}

// Refresh reloads all MessageSign rows from the database, grouped by tenantId.
// On error or empty result set, the old data is preserved.
func (c *MessageSignCacheStore) Refresh() {
	var list []model.MessageSign
	if err := c.db.Where("iz_del = 0").Find(&list).Error; err != nil {
		log.Printf("[MessageSignCache] refresh failed, keeping old data: %v", err)
		return
	}
	if len(list) == 0 {
		log.Printf("[MessageSignCache] refresh returned empty list, keeping old data")
		return
	}

	newMap := make(map[string]map[string]struct{})
	for _, item := range list {
		signs, ok := newMap[item.TenantId]
		if !ok {
			signs = make(map[string]struct{})
			newMap[item.TenantId] = signs
		}
		signs[item.Sign] = struct{}{}
	}
	c.data.Store(newMap)

	jsonData, _ := json.Marshal(list)
	hash := string(jsonData)
	isFirst := c.lastHash == ""
	isChanged := c.lastHash != hash
	c.lastHash = hash

	if isFirst || isChanged {
		total := 0
		for _, s := range newMap {
			total += len(s)
		}
		log.Printf("[MessageSignCache] refreshed, loaded %d signs for %d tenants", total, len(newMap))
	}
}

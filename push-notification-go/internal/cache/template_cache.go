package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"

	"push-notification-go/internal/model"

	"gorm.io/gorm"
)

// TemplateCacheStore holds an atomically-swapped map of Template
// keyed by "{supplierId}_{templateCode}".
type TemplateCacheStore struct {
	db       *gorm.DB
	data     atomic.Value // stores map[string]*model.Template
	lastHash string
}

// NewTemplateCache creates a new TemplateCacheStore backed by the given db.
func NewTemplateCache(db *gorm.DB) *TemplateCacheStore {
	c := &TemplateCacheStore{db: db}
	c.data.Store(make(map[string]*model.Template))
	return c
}

// buildTemplateKey returns the cache key for a template entry.
func buildTemplateKey(supplierId int64, templateCode string) string {
	return fmt.Sprintf("%d_%s", supplierId, templateCode)
}

// Get returns the Template for the given supplier ID and template code, or nil if not found.
func (c *TemplateCacheStore) Get(supplierId int64, templateCode string) *model.Template {
	m := c.data.Load().(map[string]*model.Template)
	return m[buildTemplateKey(supplierId, templateCode)]
}

// GetAll returns the full cached map.
func (c *TemplateCacheStore) GetAll() map[string]*model.Template {
	return c.data.Load().(map[string]*model.Template)
}

// Refresh reloads all Template rows from the database.
// On error or empty result set, the old data is preserved.
func (c *TemplateCacheStore) Refresh() {
	var list []model.Template
	if err := c.db.Where("iz_del = 0").Find(&list).Error; err != nil {
		log.Printf("[TemplateCache] refresh failed, keeping old data: %v", err)
		return
	}
	if len(list) == 0 {
		log.Printf("[TemplateCache] refresh returned empty list, keeping old data")
		return
	}

	newMap := make(map[string]*model.Template, len(list))
	for i := range list {
		key := buildTemplateKey(list[i].SupplierId, list[i].TemplateCode)
		newMap[key] = &list[i]
	}
	c.data.Store(newMap)

	jsonData, _ := json.Marshal(list)
	hash := string(jsonData)
	isFirst := c.lastHash == ""
	isChanged := c.lastHash != hash
	c.lastHash = hash

	if isFirst || isChanged {
		log.Printf("[TemplateCache] refreshed, loaded %d entries", len(newMap))
	}
}

package cache

import (
	"encoding/json"
	"log"
	"strconv"
	"sync/atomic"

	"push-notification-go/internal/model"

	"gorm.io/gorm"
)

// SupplierCacheStore holds an atomically-swapped map of Supplier keyed by supplier ID string.
type SupplierCacheStore struct {
	db       *gorm.DB
	data     atomic.Value // stores map[string]*model.Supplier
	lastHash string
}

// NewSupplierCache creates a new SupplierCacheStore backed by the given db.
func NewSupplierCache(db *gorm.DB) *SupplierCacheStore {
	c := &SupplierCacheStore{db: db}
	c.data.Store(make(map[string]*model.Supplier))
	return c
}

// Get returns the Supplier for the given ID, or nil if not found.
func (c *SupplierCacheStore) Get(supplierId int64) *model.Supplier {
	m := c.data.Load().(map[string]*model.Supplier)
	return m[strconv.FormatInt(supplierId, 10)]
}

// GetAll returns the full cached map.
func (c *SupplierCacheStore) GetAll() map[string]*model.Supplier {
	return c.data.Load().(map[string]*model.Supplier)
}

// Refresh reloads all Supplier rows from the database.
// On error or empty result set, the old data is preserved.
func (c *SupplierCacheStore) Refresh() {
	var list []model.Supplier
	if err := c.db.Where("iz_del = 0").Find(&list).Error; err != nil {
		log.Printf("[SupplierCache] refresh failed, keeping old data: %v", err)
		return
	}
	if len(list) == 0 {
		log.Printf("[SupplierCache] refresh returned empty list, keeping old data")
		return
	}

	newMap := make(map[string]*model.Supplier, len(list))
	for i := range list {
		key := strconv.FormatInt(list[i].ID, 10)
		newMap[key] = &list[i]
	}
	c.data.Store(newMap)

	jsonData, _ := json.Marshal(list)
	hash := string(jsonData)
	isFirst := c.lastHash == ""
	isChanged := c.lastHash != hash
	c.lastHash = hash

	if isFirst || isChanged {
		log.Printf("[SupplierCache] refreshed, loaded %d entries", len(newMap))
	}
}

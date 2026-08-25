package cache

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// Package-level cache instances, initialized by Init().
var (
	TenantConfigCache *TenantConfigCacheStore
	SupplierCache     *SupplierCacheStore
	TemplateCache     *TemplateCacheStore
	MessageSignCache  *MessageSignCacheStore
)

// Init creates all cache instances and performs an initial blocking refresh.
func Init(db *gorm.DB) {
	TenantConfigCache = NewTenantConfigCache(db)
	SupplierCache = NewSupplierCache(db)
	TemplateCache = NewTemplateCache(db)
	MessageSignCache = NewMessageSignCache(db)

	// Blocking initial load
	TenantConfigCache.Refresh()
	SupplierCache.Refresh()
	TemplateCache.Refresh()
	MessageSignCache.Refresh()

	log.Println("[Cache] all caches initialized")
}

// StartRefreshLoop starts a background goroutine that refreshes all caches
// at the specified interval.
func StartRefreshLoop(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			TenantConfigCache.Refresh()
			SupplierCache.Refresh()
			TemplateCache.Refresh()
			MessageSignCache.Refresh()
		}
	}()
	log.Printf("[Cache] refresh loop started with interval %v", interval)
}

package cache

import (
	"sync"
	"time"

	"ebike-device-openapi-go/internal/api/dto"
)

const (
	defaultMaxSize = 400000
	defaultTTL     = 1440 * time.Minute // matches Java Caffeine expireAfterAccess
)

// EcuLoginCache is a process-local fallback when Redis is unavailable.
// Mirrors Java Caffeine cache CacheName.ecuLogin used in RoutFilter.needRoute().
type EcuLoginCache struct {
	mu      sync.RWMutex
	entries map[string]*ecuLoginEntry
	maxSize int
	ttl     time.Duration
}

type ecuLoginEntry struct {
	login      *dto.EcuLogin
	lastAccess time.Time
}

var DefaultEcuLoginCache = NewEcuLoginCache(defaultMaxSize, defaultTTL)

func NewEcuLoginCache(maxSize int, ttl time.Duration) *EcuLoginCache {
	if maxSize <= 0 {
		maxSize = defaultMaxSize
	}
	if ttl <= 0 {
		ttl = defaultTTL
	}
	return &EcuLoginCache{
		entries: make(map[string]*ecuLoginEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

func (c *EcuLoginCache) Get(imei string) *dto.EcuLogin {
	if imei == "" {
		return nil
	}
	now := time.Now()

	c.mu.RLock()
	entry, ok := c.entries[imei]
	c.mu.RUnlock()
	if !ok || entry == nil || entry.login == nil {
		return nil
	}
	if now.Sub(entry.lastAccess) > c.ttl {
		c.mu.Lock()
		delete(c.entries, imei)
		c.mu.Unlock()
		return nil
	}

	c.mu.Lock()
	entry.lastAccess = now
	c.mu.Unlock()
	return entry.login
}

func (c *EcuLoginCache) Put(imei string, login *dto.EcuLogin) {
	if imei == "" || login == nil {
		return
	}
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= c.maxSize {
		c.evictOneLocked(now)
	}

	copied := *login
	c.entries[imei] = &ecuLoginEntry{
		login:      &copied,
		lastAccess: now,
	}
}

// Delete removes a cached ecuLogin entry (call on logout/unbind).
func (c *EcuLoginCache) Delete(imei string) {
	if imei == "" {
		return
	}
	c.mu.Lock()
	delete(c.entries, imei)
	c.mu.Unlock()
}

func (c *EcuLoginCache) evictOneLocked(now time.Time) {
	for key, entry := range c.entries {
		if now.Sub(entry.lastAccess) > c.ttl {
			delete(c.entries, key)
			return
		}
	}
	for key := range c.entries {
		delete(c.entries, key)
		return
	}
}

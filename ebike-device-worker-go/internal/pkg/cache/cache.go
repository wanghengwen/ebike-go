package cache

import (
	"sync"
	"time"
)

const defaultCleanupInterval = 15 * time.Minute

type entry struct {
	value  interface{}
	expire time.Time
}

// TTLCache is a simple expire-after-access cache mirroring Caffeine usage.
type TTLCache struct {
	mu      sync.RWMutex
	data    map[string]*entry
	maxSize int
	ttl     time.Duration
}

func NewTTLCache(maxSize int, ttlSec int) *TTLCache {
	return newTTLCache(maxSize, ttlSec, defaultCleanupInterval)
}

func newTTLCache(maxSize int, ttlSec int, cleanupInterval time.Duration) *TTLCache {
	c := &TTLCache{
		data:    make(map[string]*entry, maxSize),
		maxSize: maxSize,
		ttl:     time.Duration(ttlSec) * time.Second,
	}
	if cleanupInterval > 0 {
		go c.cleanupLoop(cleanupInterval)
	}
	return c
}

func (c *TTLCache) Get(key string, loader func() (interface{}, bool)) (interface{}, bool) {
	now := time.Now()
	c.mu.RLock()
	if e, ok := c.data[key]; ok && now.Before(e.expire) {
		c.mu.RUnlock()
		return e.value, true
	}
	c.mu.RUnlock()

	if loader == nil {
		return nil, false
	}
	val, ok := loader()
	if !ok {
		return nil, false
	}
	c.Put(key, val)
	return val, true
}

func (c *TTLCache) Put(key string, val interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.data) >= c.maxSize {
		for k := range c.data {
			delete(c.data, k)
			break
		}
	}
	c.data[key] = &entry{value: val, expire: time.Now().Add(c.ttl)}
}

func (c *TTLCache) Invalidate(key string) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}

func (c *TTLCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

func (c *TTLCache) cleanupExpired() int {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	removed := 0
	for k, e := range c.data {
		if !now.Before(e.expire) {
			delete(c.data, k)
			removed++
		}
	}
	return removed
}

func (c *TTLCache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		c.cleanupExpired()
	}
}

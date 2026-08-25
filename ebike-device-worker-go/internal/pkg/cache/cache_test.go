package cache

import (
	"testing"
	"time"
)

func TestCleanupExpiredRemovesStaleEntries(t *testing.T) {
	c := newTTLCache(100, 1, 0) // no background loop
	c.Put("alive", "v1")
	c.Put("stale", "v2")

	c.mu.Lock()
	c.data["stale"].expire = time.Now().Add(-time.Second)
	c.mu.Unlock()

	removed := c.cleanupExpired()
	if removed != 1 {
		t.Fatalf("removed=%d, want 1", removed)
	}
	if c.Len() != 1 {
		t.Fatalf("len=%d, want 1", c.Len())
	}
	if _, ok := c.Get("alive", nil); !ok {
		t.Fatal("expected alive key to remain")
	}
	if _, ok := c.Get("stale", nil); ok {
		t.Fatal("expected stale key to be removed")
	}
}

func TestCleanupLoopPeriodic(t *testing.T) {
	c := newTTLCache(100, 1, 50*time.Millisecond)
	c.Put("k", "v")

	c.mu.Lock()
	c.data["k"].expire = time.Now().Add(-time.Second)
	c.mu.Unlock()

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if c.Len() == 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("expected cleanup loop to remove expired entry, len=%d", c.Len())
}

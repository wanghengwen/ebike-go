package service

import (
	"testing"
	"time"
)

// Both mapping keys are persistent, so an unchanged binding is written once when the
// session starts and never rewritten, no matter how long the session lasts.
func TestNeedsMappingWriteOnlyOncePerSession(t *testing.T) {
	s := &LuopingMappingService{}
	now := time.Now()

	if !s.needsMappingWrite("dev1", "imei1", "ebike", now) {
		t.Fatal("first uplink must write")
	}
	s.markMappingWritten("dev1", "imei1", "ebike", now)

	if s.needsMappingWrite("dev1", "imei1", "ebike", now.Add(time.Second)) {
		t.Fatal("unchanged binding must skip redis")
	}
	if s.needsMappingWrite("dev1", "imei1", "ebike", now.Add(72*time.Hour)) {
		t.Fatal("unchanged binding must skip redis regardless of elapsed time")
	}
}

func TestNeedsMappingWriteOnBindingChange(t *testing.T) {
	s := &LuopingMappingService{}
	now := time.Now()
	s.markMappingWritten("dev1", "imei1", "ebike", now)

	if !s.needsMappingWrite("dev1", "imei2", "ebike", now) {
		t.Fatal("imei change must write")
	}
	if !s.needsMappingWrite("dev1", "imei1", "other", now) {
		t.Fatal("group change must write")
	}
	if !s.needsMappingWrite("dev2", "imei1", "ebike", now) {
		t.Fatal("unknown deviceId must write")
	}
}

// Going offline must drop the local record so the device re-persists (and gets its
// session re-verified) on the very next uplink instead of waiting for the tick.
func TestForgetMappingWriteForcesRewrite(t *testing.T) {
	s := &LuopingMappingService{}
	now := time.Now()
	s.markMappingWritten("dev1", "imei1", "ebike", now)
	s.forgetMappingWrite("dev1")

	if !s.needsMappingWrite("dev1", "imei1", "ebike", now) {
		t.Fatal("forgotten deviceId must write")
	}
}

func TestEvictIdleKeepsCacheBounded(t *testing.T) {
	s := &LuopingMappingService{}
	idle := time.Now().Add(-2 * luopingMappingCacheIdleTTL)
	for i := 0; i < 10; i++ {
		s.markMappingWritten(string(rune('a'+i)), "imei", "ebike", idle)
	}
	s.markMappingWritten("active", "imei", "ebike", time.Now())

	s.writeMu.Lock()
	s.evictIdleLocked(time.Now())
	size := len(s.writeCache)
	s.writeMu.Unlock()

	if size != 1 {
		t.Fatalf("only the active entry should survive, %d left", size)
	}
}

// An active device keeps its entry alive purely by uplinking, without any rewrite.
func TestNeedsMappingWriteRefreshesIdleClock(t *testing.T) {
	s := &LuopingMappingService{}
	start := time.Now().Add(-2 * luopingMappingCacheIdleTTL)
	s.markMappingWritten("dev1", "imei1", "ebike", start)
	s.needsMappingWrite("dev1", "imei1", "ebike", time.Now())

	s.writeMu.Lock()
	s.evictIdleLocked(time.Now())
	_, ok := s.writeCache["dev1"]
	s.writeMu.Unlock()

	if !ok {
		t.Fatal("a device still uplinking must not be evicted")
	}
}

package gateway

import "testing"

func TestParseFenceStringKeyMatchesJavaFormat(t *testing.T) {
	tests := []struct {
		key      string
		kind     FenceCacheKind
		tenantID string
		id       int64
	}{
		{"fence_parking_1003_192856886438004387", CacheParking, "1003", 192856886438004387},
		{"fence_parking_1000_247495733790908225", CacheParking, "1000", 247495733790908225},
		{"fence_noParking_1000_274872774458608393", CacheNoParking, "1000", 274872774458608393},
		{"fence_banRiding_1000_99", CacheBanRiding, "1000", 99},
		{"fence_maintainArea_1003_1", CacheMaintainArea, "1003", 1},
		{"fence_serviceArea_1003_42", CacheServiceArea, "1003", 42},
		{"fence_custom_1003_7", CacheCustom, "1003", 7},
	}
	for _, tt := range tests {
		kind, tenantID, id, ok := parseFenceStringKey(tt.key)
		if !ok {
			t.Fatalf("parseFenceStringKey(%q) failed", tt.key)
		}
		if kind != tt.kind || tenantID != tt.tenantID || id != tt.id {
			t.Fatalf("parseFenceStringKey(%q) = kind=%v tenant=%q id=%d, want kind=%v tenant=%q id=%d",
				tt.key, kind, tenantID, id, tt.kind, tt.tenantID, tt.id)
		}
		if got := fenceStringKey(kind, tenantID, id); got != tt.key {
			t.Fatalf("round-trip key = %q want %q", got, tt.key)
		}
	}
}

func TestCacheKindFromPrefixMatchesGetFencesByLocation(t *testing.T) {
	tests := []struct {
		prefix string
		kind   FenceCacheKind
	}{
		{"fence_parking", CacheParking},
		{"fence_banRiding", CacheBanRiding},
		{"fence_custom", CacheCustom},
	}
	for _, tt := range tests {
		kind, ok := CacheKindFromPrefix(tt.prefix)
		if !ok || kind != tt.kind {
			t.Fatalf("CacheKindFromPrefix(%q) = %v %v, want %v true", tt.prefix, kind, ok, tt.kind)
		}
	}
}

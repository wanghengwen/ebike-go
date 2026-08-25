package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ebike-fence-go/internal/domain/rediskeys"
	pkgredis "ebike-fence-go/internal/pkg/redis"

	"github.com/go-redis/redis/v8"
)

// FenceCacheKind selects Redis key templates for a fence type.
type FenceCacheKind int

const (
	CacheServiceArea FenceCacheKind = iota
	CacheParking
	CacheNoParking
	CacheBanRiding
	CacheMaintainArea
	CacheCustom
)

func fenceStringKey(kind FenceCacheKind, tenantID string, id int64) string {
	switch kind {
	case CacheServiceArea:
		return rediskeys.ServiceArea(tenantID, id)
	case CacheParking:
		return rediskeys.Parking(tenantID, id)
	case CacheNoParking:
		return rediskeys.NoParking(tenantID, id)
	case CacheBanRiding:
		return rediskeys.BanRiding(tenantID, id)
	case CacheMaintainArea:
		return rediskeys.MaintainArea(tenantID, id)
	case CacheCustom:
		return rediskeys.FenceCustom(tenantID, id)
	default:
		return fmt.Sprintf("fence_unknown_%s_%d", tenantID, id)
	}
}

func fenceGeoKey(kind FenceCacheKind, tenantID string, customTypeID int64) string {
	switch kind {
	case CacheServiceArea:
		return rediskeys.ServiceAreaGeo(tenantID)
	case CacheParking:
		return rediskeys.ParkingGeo(tenantID)
	case CacheNoParking:
		return rediskeys.NoParkingGeo(tenantID)
	case CacheBanRiding:
		return rediskeys.BanRidingGeo(tenantID)
	case CacheMaintainArea:
		return rediskeys.MaintainAreaGeo(tenantID)
	case CacheCustom:
		return rediskeys.FenceCustomGeoSearchKey(tenantID)
	default:
		return ""
	}
}

// syncFenceCacheCore writes fence JSON and GEO index without touching the empty marker.
func syncFenceCacheCore(ctx context.Context, kind FenceCacheKind, tenantID string, fe *FenceE, customTypeID int64) error {
	if fe == nil {
		return nil
	}
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return nil
	}
	key := fenceStringKey(kind, tenantID, fe.Id)
	raw, err := json.Marshal(fe)
	if err != nil {
		return err
	}
	if err := rdb.Set(ctx, key, raw, 0).Err(); err != nil {
		return err
	}
	geoKey := fenceGeoKey(kind, tenantID, customTypeID)
	if geoKey == "" || fe.CenterLng == 0 && fe.CenterLat == 0 {
		return nil
	}
	return rdb.GeoAdd(ctx, geoKey, &redis.GeoLocation{
		Name:      strconv.FormatInt(fe.Id, 10),
		Longitude: fe.CenterLng,
		Latitude:  fe.CenterLat,
	}).Err()
}

// SyncFenceCache writes fence JSON and GEO index after DB mutation (CRUD path).
// Also clears the GEO empty marker for this type+tenant.
func SyncFenceCache(ctx context.Context, kind FenceCacheKind, tenantID string, fe *FenceE, customTypeID int64) error {
	if fe == nil {
		return nil
	}
	ClearGeoEmptyMark(ctx, kind, tenantID)
	return syncFenceCacheCore(ctx, kind, tenantID, fe, customTypeID)
}

// DeleteFenceCache removes fence string key and GEO member.
func DeleteFenceCache(ctx context.Context, kind FenceCacheKind, tenantID string, id int64, customTypeID int64) error {
	// Fence data changed, clear the empty marker so next query re-evaluates.
	ClearGeoEmptyMark(ctx, kind, tenantID)
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return nil
	}
	key := fenceStringKey(kind, tenantID, id)
	if err := rdb.Del(ctx, key).Err(); err != nil {
		return err
	}
	geoKey := fenceGeoKey(kind, tenantID, customTypeID)
	if geoKey == "" {
		return nil
	}
	return rdb.ZRem(ctx, geoKey, strconv.FormatInt(id, 10)).Err()
}

// InvalidateFenceCache deletes only the string key (Java update pattern).
func InvalidateFenceCache(ctx context.Context, kind FenceCacheKind, tenantID string, id int64) error {
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return nil
	}
	return rdb.Del(ctx, fenceStringKey(kind, tenantID, id)).Err()
}

// parseFenceStringKey reverses fenceStringKey / Java FenceRedisKey.format output:
// fence_parking_{tenantId}_{id}
func parseFenceStringKey(key string) (FenceCacheKind, string, int64, bool) {
	type entry struct {
		prefix string
		kind   FenceCacheKind
	}
	for _, e := range []entry{
		{"fence_serviceArea_", CacheServiceArea},
		{"fence_noParking_", CacheNoParking},
		{"fence_maintainArea_", CacheMaintainArea},
		{"fence_banRiding_", CacheBanRiding},
		{"fence_parking_", CacheParking},
		{"fence_custom_", CacheCustom},
	} {
		if !strings.HasPrefix(key, e.prefix) {
			continue
		}
		rest := key[len(e.prefix):]
		idx := strings.LastIndex(rest, "_")
		if idx <= 0 {
			continue
		}
		tenantID := rest[:idx]
		id, err := strconv.ParseInt(rest[idx+1:], 10, 64)
		if err != nil {
			continue
		}
		return e.kind, tenantID, id, true
	}
	return 0, "", 0, false
}

// writeFenceStringCache sets the detail JSON key only (no GEO), matching Java save() populate.
func writeFenceStringCache(ctx context.Context, kind FenceCacheKind, tenantID string, fe *FenceE) error {
	if fe == nil {
		return nil
	}
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return nil
	}
	raw, err := json.Marshal(fe)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, fenceStringKey(kind, tenantID, fe.Id), raw, 0).Err()
}

// CacheKindFromPrefix maps fenceKeyPrefix used in GetFencesByLocation to FenceCacheKind.
func CacheKindFromPrefix(prefix string) (FenceCacheKind, bool) {
	switch prefix {
	case "fence_serviceArea":
		return CacheServiceArea, true
	case "fence_parking":
		return CacheParking, true
	case "fence_noParking":
		return CacheNoParking, true
	case "fence_banRiding":
		return CacheBanRiding, true
	case "fence_maintainArea":
		return CacheMaintainArea, true
	case "fence_custom":
		return CacheCustom, true
	default:
		return 0, false
	}
}

// PopulateGeoCache batch-writes all fences to Redis GEO/detail and manages the
// empty marker. Called by the GEO miss fallback fetcher after DB query.
//   - Non-empty: clears empty marker once, writes all fences (no per-item marker ops).
//   - Empty: sets empty marker (TTL 30min) to prevent repeated DB queries.
func PopulateGeoCache(ctx context.Context, prefix, tenantID string, fences []FenceE) {
	kind, ok := CacheKindFromPrefix(prefix)
	if !ok {
		return
	}
	if len(fences) == 0 {
		SetGeoEmptyMark(ctx, prefix, tenantID)
		return
	}
	clearGeoEmptyMarkByPrefix(ctx, prefix, tenantID)
	for i := range fences {
		_ = syncFenceCacheCore(ctx, kind, tenantID, &fences[i], fences[i].CustomTypeId)
	}
}

// ---------------------------------------------------------------------------
// GEO empty marker: prevents repeated DB queries when a tenant has no fences
// of a particular type. Key format: geo_empty_{prefix}_{tenantId}, TTL 30min.
// Cleared on any CRUD operation (SyncFenceCache / DeleteFenceCache).
// ---------------------------------------------------------------------------

const geoEmptyTTL = 30 * time.Minute

func geoEmptyKey(prefix, tenantID string) string {
	return fmt.Sprintf("geo_empty_%s_%s", prefix, tenantID)
}

func prefixFromKind(kind FenceCacheKind) string {
	switch kind {
	case CacheServiceArea:
		return "fence_serviceArea"
	case CacheParking:
		return "fence_parking"
	case CacheNoParking:
		return "fence_noParking"
	case CacheBanRiding:
		return "fence_banRiding"
	case CacheMaintainArea:
		return "fence_maintainArea"
	case CacheCustom:
		return "fence_custom"
	default:
		return ""
	}
}

// SetGeoEmptyMark marks a prefix+tenant combination as having no fences.
func SetGeoEmptyMark(ctx context.Context, prefix, tenantID string) {
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return
	}
	rdb.Set(ctx, geoEmptyKey(prefix, tenantID), "1", geoEmptyTTL)
}

// CheckGeoEmptyMark returns true if a prefix+tenant is marked as having no fences.
func CheckGeoEmptyMark(ctx context.Context, prefix, tenantID string) bool {
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return false
	}
	exists, err := rdb.Exists(ctx, geoEmptyKey(prefix, tenantID)).Result()
	return err == nil && exists > 0
}

// ClearGeoEmptyMark removes the empty marker when fence data changes (by kind).
func ClearGeoEmptyMark(ctx context.Context, kind FenceCacheKind, tenantID string) {
	prefix := prefixFromKind(kind)
	if prefix == "" {
		return
	}
	clearGeoEmptyMarkByPrefix(ctx, prefix, tenantID)
}

func clearGeoEmptyMarkByPrefix(ctx context.Context, prefix, tenantID string) {
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return
	}
	rdb.Del(ctx, geoEmptyKey(prefix, tenantID))
}

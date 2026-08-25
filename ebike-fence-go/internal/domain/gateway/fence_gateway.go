package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync/atomic"

	"ebike-fence-go/internal/domain/geo"
	customRedis "ebike-fence-go/internal/pkg/redis"

	"github.com/go-redis/redis/v8"
)

var cacheMissCount atomic.Uint64

// FenceMissFetcher loads fence data when Redis cache misses (e.g. RPC to ebike-fence).
type FenceMissFetcher func(ctx context.Context, tenantId string, id int64) (*FenceE, error)

var fenceMissFetcher FenceMissFetcher

// SetFenceMissFetcher registers the cache-miss fallback handler.
func SetFenceMissFetcher(fetcher FenceMissFetcher) {
	fenceMissFetcher = fetcher
}

// CacheMissCount returns the number of Redis fence cache misses observed.
func CacheMissCount() uint64 {
	return cacheMissCount.Load()
}

func recordCacheMiss(key string) {
	cacheMissCount.Add(1)
	log.Printf("[CACHE_MISS] Redis fence cache miss for key %s (total=%d)", key, cacheMissCount.Load())
}

func loadFenceOnMiss(ctx context.Context, tenantId, key string, id int64) (*FenceE, error) {
	recordCacheMiss(key)
	if fenceMissFetcher == nil {
		return nil, fmt.Errorf("fence cache miss for key %s and no fallback configured", key)
	}
	fe, err := fenceMissFetcher(ctx, tenantId, id)
	if err != nil || fe == nil {
		return fe, err
	}
	// Java ParkingGatewayImpl.update deletes detail keys while GEO members remain; on miss
	// Java loads from DB silently. Write-back here avoids repeated misses on hot paths.
	if kind, tid, fid, ok := parseFenceStringKey(key); ok && tid == tenantId && fid == id {
		if werr := writeFenceStringCache(ctx, kind, tenantId, fe); werr != nil {
			log.Printf("[WARN] cache miss write-back failed for key %s: %v", key, werr)
		}
	}
	return fe, nil
}

func parseFenceIDFromGeoName(name string) (int64, error) {
	return strconv.ParseInt(name, 10, 64)
}

// FenceGeoListFetcher loads fences near a point when Redis GEO is empty.
// serviceID > 0 limits DB fallback to that service area (matches Java near-query semantics).
type FenceGeoListFetcher func(ctx context.Context, tenantId, fenceKeyPrefix string, serviceID int64, lng, lat, radius float64, count int) ([]FenceE, error)

var fenceGeoListFetcher FenceGeoListFetcher

// SetFenceGeoListFetcher registers DB fallback for empty GEO queries.
func SetFenceGeoListFetcher(fetcher FenceGeoListFetcher) {
	fenceGeoListFetcher = fetcher
}

func logGeoFallback(prefix, tenantId string) {
	log.Printf("[GEO_FALLBACK] Redis GEO miss for prefix %s tenant %s, loading from DB", prefix, tenantId)
}

func loadFencesOnGeoMiss(ctx context.Context, tenantId, prefix string, serviceID int64, lng, lat, radius float64, count int) ([]FenceE, error) {
	if fenceGeoListFetcher == nil {
		return nil, nil
	}
	// If this prefix+tenant is already marked as having no fences, skip DB query.
	if CheckGeoEmptyMark(ctx, prefix, tenantId) {
		return nil, nil
	}
	logGeoFallback(prefix, tenantId)
	// Fetcher handles: DB query → PopulateGeoCache (full GEO write-back) → return filtered results.
	return fenceGeoListFetcher(ctx, tenantId, prefix, serviceID, lng, lat, radius, count)
}


// FencesNearPoint filters and sorts fences by center-point distance (meters).
func FencesNearPoint(fences []FenceE, lng, lat, radius float64, count int) []FenceE {
	type item struct {
		f FenceE
		d float64
	}
	items := make([]item, 0, len(fences))
	for _, f := range fences {
		d := geo.PointToPointDistanceJava(lng, lat, f.CenterLng, f.CenterLat)
		if radius > 0 && d > radius {
			continue
		}
		items = append(items, item{f: f, d: d})
	}
	// Java getNearParking uses Comparator.comparing(distance) only; stable sort keeps GEO/insertion order on ties.
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].d < items[j].d
	})
	if count > 0 && len(items) > count {
		items = items[:count]
	}
	out := make([]FenceE, len(items))
	for i, it := range items {
		it.f.Distance = it.d
		out[i] = it.f
	}
	return out
}

type FenceE struct {
	Id                     int64   `json:"id"`
	Name                   string  `json:"name"`
	ShapeType              string  `json:"shapeType"`
	CenterLat              float64 `json:"centerLat"`
	CenterLng              float64 `json:"centerLng"`
	PointList              string  `json:"pointList"`
	Type                   int     `json:"type"`
	ServiceId              int64   `json:"serviceId"`
	IzEnable               bool    `json:"izEnable"`
	BufferDistance         float64 `json:"bufferDistance"`
	CoefficientOfDifficult float64 `json:"coefficientOfDifficult"`
	CoefficientOfDifficultSet bool `json:"-"`
	BufferDistanceSet      bool    `json:"-"`
	IzFullPileNoStop       *int    `json:"izFullPileNoStop"`
	MaxParkingNumber       int     `json:"maxParkingNumber"`
	OpeningHoursBegin      string  `json:"openingHoursBegin"`
	OpeningHoursEnd        string  `json:"openingHoursEnd"`
	IzOpenAllDay           *bool   `json:"izOpenAllDay"`

	// Parking capability flags (ParkingE.getParts in Java)
	Tbeacon     *bool `json:"tbeacon"`
	Directional *bool `json:"directional"`
	Rfid        *bool `json:"rfid"`
	Camera      *bool `json:"camera"`
	Kickstand          *bool    `json:"kickstand"`
	Direction          *float64 `json:"direction"`
	FormulateDirection *float64 `json:"formulateDirection"`
	CustomTypeId       int64    `json:"customTypeId,omitempty"`
	AreaSize           float64  `json:"areaSize,omitempty"`
	AreaSizeSet        bool     `json:"-"`
	MinAmount          *int     `json:"minAmount,omitempty"`
	MinAmountSet       bool     `json:"-"`
	MaxAmount          *int     `json:"maxAmount,omitempty"`
	MaxAmountSet       bool     `json:"-"`
	Distance           float64  `json:"distance,omitempty"`
	TenantId           string   `json:"tenantId,omitempty"`
	DataVersion        int64    `json:"dataVersion,omitempty"`
	CreatedPin         string   `json:"createdPin,omitempty"`
	CreatedAt          string   `json:"createdAt,omitempty"`
	UpdatedPin         string   `json:"updatedPin,omitempty"`
	UpdatedAt          string   `json:"updatedAt,omitempty"`
	IzCameraDirectionalBackcar *bool `json:"izCameraDirectionalBackcar,omitempty"`
	IzCameraPointBackcar       *bool `json:"izCameraPointBackcar,omitempty"`

	// ParsedPolygon avoids massive string parsing overhead during requests
	ParsedPolygon []geo.Location `json:"-"`
}

// GetParkingParts mirrors Java ParkingE.getParts().
func (f *FenceE) GetParkingParts() []string {
	var parts []string
	if f.Tbeacon != nil && *f.Tbeacon {
		parts = append(parts, "bluetooth_beacon")
	}
	if f.Directional != nil && *f.Directional {
		parts = append(parts, "direction")
	}
	if f.Rfid != nil && *f.Rfid {
		parts = append(parts, "rfid_beacon")
	}
	if f.Camera != nil && *f.Camera {
		parts = append(parts, "camera")
	}
	if f.Kickstand != nil && *f.Kickstand {
		parts = append(parts, "kickstand")
	}
	return parts
}

func (f *FenceE) IzFullPileNoStopEnabled() bool {
	return f.IzFullPileNoStop != nil && *f.IzFullPileNoStop == 1
}

type GeoFence struct {
	FenceId  string
	Distance float64
}

// SearchRadius retrieves fences within a radius from the GEO ZSET
func SearchRadius(ctx context.Context, geoKey string, lng, lat, radius float64, count int) ([]GeoFence, error) {
	rdb := customRedis.GetClient()
	if rdb == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	query := &redis.GeoRadiusQuery{
		Radius:      radius,
		Unit:        "m",
		WithCoord:   false,
		WithDist:    true,
		WithGeoHash: false,
		Count:       count,
		Sort:        "ASC",
	}

	locations, err := rdb.GeoRadius(ctx, geoKey, lng, lat, query).Result()
	if err != nil {
		log.Printf("Error during GeoRadius on key %s: %v", geoKey, err)
		return nil, err
	}

	var results []GeoFence
	for _, loc := range locations {
		results = append(results, GeoFence{
			FenceId:  loc.Name,
			Distance: loc.Dist,
		})
	}

	return results, nil
}

// GetFencesByLocation gets full fence info using searchRadius followed by MGet.
// serviceID > 0 scopes DB GEO fallback to that service (near-list APIs).
func GetFencesByLocation(ctx context.Context, fenceKeyPrefix, geoKey string, tenantId string, lng, lat, radius float64, count int, serviceID int64) ([]FenceE, error) {
	geoFences, err := SearchRadius(ctx, geoKey, lng, lat, radius, count)
	if err != nil {
		return nil, err
	}

	if len(geoFences) == 0 {
		rdb := customRedis.GetClient()
		if rdb != nil {
			exists, err := rdb.Exists(ctx, geoKey).Result()
			if err == nil && exists > 0 {
				return nil, nil
			}
		}
		return loadFencesOnGeoMiss(ctx, tenantId, fenceKeyPrefix, serviceID, lng, lat, radius, count)
	}

	var keys []string
	for _, gf := range geoFences {
		// Key format: prefix_{tenantId}_{id}
		keys = append(keys, fmt.Sprintf("%s_%s_%s", fenceKeyPrefix, tenantId, gf.FenceId))
	}

	rdb := customRedis.GetClient()
	if rdb == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}
	values, err := rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	var result []FenceE
	type missedFence struct {
		id       int64
		distance float64
		key      string
	}
	var missed []missedFence
	for i, val := range values {
		distance := geoFences[i].Distance
		if val == nil {
			id, err := parseFenceIDFromGeoName(geoFences[i].FenceId)
			if err != nil {
				log.Printf("[WARN] cache miss with unparseable geo id %q: %v", geoFences[i].FenceId, err)
				continue
			}
			missed = append(missed, missedFence{id: id, distance: distance, key: keys[i]})
			continue
		}
		strVal, ok := val.(string)
		if !ok || strVal == "" {
			id, err := parseFenceIDFromGeoName(geoFences[i].FenceId)
			if err != nil {
				continue
			}
			missed = append(missed, missedFence{id: id, distance: distance, key: keys[i]})
			continue
		}
		decoded, err := decodeFenceJSON(keys[i], strVal)
		if err != nil {
			log.Printf("[WARN] %v", err)
			id, err := parseFenceIDFromGeoName(geoFences[i].FenceId)
			if err != nil {
				continue
			}
			missed = append(missed, missedFence{id: id, distance: distance, key: keys[i]})
			continue
		}
		decoded.Distance = distance
		result = append(result, decoded)
	}
	// Java FenceGatewayImpl appends cache-missed fences at the end via getListByIds.
	for _, m := range missed {
		loaded, err := loadFenceOnMiss(ctx, tenantId, m.key, m.id)
		if err != nil {
			log.Printf("[WARN] cache miss fallback failed for key %s: %v", m.key, err)
			continue
		}
		loaded.Distance = m.distance
		result = append(result, *loaded)
	}

	return result, nil
}

// GetFenceById gets a single fence directly by ID
func GetFenceById(ctx context.Context, fenceKeyPrefix string, tenantId string, id int64) (*FenceE, error) {
	rdb := customRedis.GetClient()
	if rdb == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	key := fmt.Sprintf("%s_%s_%d", fenceKeyPrefix, tenantId, id)
	strVal, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return loadFenceOnMiss(ctx, tenantId, key, id)
	} else if err != nil {
		return nil, err
	}

	fence, err := decodeFenceJSON(key, strVal)
	if err != nil {
		return nil, err
	}
	return &fence, nil
}

func decodeFenceJSON(key, strVal string) (FenceE, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(strVal), &raw); err != nil {
		return FenceE{}, fmt.Errorf("failed to unmarshal fence data for key %s: %w", key, err)
	}
	var fence FenceE
	if err := json.Unmarshal([]byte(strVal), &fence); err != nil {
		return FenceE{}, fmt.Errorf("failed to unmarshal fence data for key %s: %w", key, err)
	}
	if jsonFieldPresentNonNull(raw, "coefficientOfDifficult") {
		fence.CoefficientOfDifficultSet = true
	}
	if jsonFieldPresentNonNull(raw, "areaSize") {
		fence.AreaSizeSet = true
	}
	if jsonFieldPresentNonNull(raw, "minAmount") {
		fence.MinAmountSet = true
	}
	if jsonFieldPresentNonNull(raw, "maxAmount") {
		fence.MaxAmountSet = true
	}
	if jsonFieldPresentNonNull(raw, "bufferDistance") {
		fence.BufferDistanceSet = true
	}
	if parsed, err := geo.ParsePolygon(fence.PointList); err == nil {
		fence.ParsedPolygon = parsed
	} else {
		log.Printf("[WARN] Failed to parse polygon for fence ID %d: %v", fence.Id, err)
	}
	return fence, nil
}

func jsonFieldPresentNonNull(raw map[string]json.RawMessage, key string) bool {
	v, ok := raw[key]
	return ok && string(v) != "null"
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"ebike-device-openapi-go/internal/api/dto"
	"ebike-device-openapi-go/internal/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	luopingDeviceKeyPrefix   = "mqtt:device:"
	luopingImeiKeyPrefix     = "mqtt:imei:"
	luopingPresenceKeyPrefix = "mqtt:presence:"
	luopingPresenceTTL       = 24 * time.Hour
	// luopingMappingCacheIdleTTL bounds how long an unused writeCache entry is kept.
	// A device silent for longer than the online TTL is offline, so its entry is dead
	// weight; dropping it only costs one extra SET if the device ever comes back.
	luopingMappingCacheIdleTTL = 30 * time.Minute
	// luopingMappingCacheMax caps the process-local write cache so a flood of unknown
	// deviceIds cannot grow it without bound.
	luopingMappingCacheMax = 200000
)

// presenceEpochScript stores the newest presence/activity timestamp (unix ms) seen
// for a device and advances it only when the incoming value is strictly greater.
// Returns 1 when applied (newest so far), 0 when the incoming value is stale.
var presenceEpochScript = redis.NewScript(`
local cur = tonumber(redis.call('GET', KEYS[1]) or '0')
local ts = tonumber(ARGV[1])
if ts > cur then
  redis.call('SET', KEYS[1], ts, 'EX', ARGV[2])
  return 1
end
return 0
`)

// LuopingMappingService maintains deviceId <-> IMEI bindings for luoping MQTT devices.
type LuopingMappingService struct {
	Rdb *redis.Client

	writeMu    sync.Mutex
	writeCache map[string]mappingWrite
}

// mappingWrite records the binding this process last persisted for a deviceId.
// seenAt is the last uplink observed for it and only drives cache eviction.
type mappingWrite struct {
	imei      string
	groupName string
	seenAt    time.Time
}

// needsMappingWrite reports whether the binding must be pushed to Redis. Both keys
// are persistent, so an unchanged binding never needs rewriting: it is written when
// the session starts and again only if the deviceId, IMEI or group actually change.
// Seeing a known device refreshes its entry so eviction only targets idle ones.
func (s *LuopingMappingService) needsMappingWrite(deviceId, imei, groupName string, now time.Time) bool {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	prev, ok := s.writeCache[deviceId]
	if !ok || prev.imei != imei || prev.groupName != groupName {
		return true
	}
	prev.seenAt = now
	s.writeCache[deviceId] = prev
	return false
}

func (s *LuopingMappingService) markMappingWritten(deviceId, imei, groupName string, now time.Time) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if s.writeCache == nil {
		s.writeCache = make(map[string]mappingWrite)
	}
	if len(s.writeCache) >= luopingMappingCacheMax {
		s.evictIdleLocked(now)
	}
	s.writeCache[deviceId] = mappingWrite{imei: imei, groupName: groupName, seenAt: now}
}

// evictIdleLocked drops entries whose device has been silent past the idle horizon;
// if every entry is still active the cache is reset wholesale rather than growing
// past the cap.
func (s *LuopingMappingService) evictIdleLocked(now time.Time) {
	for k, v := range s.writeCache {
		if now.Sub(v.seenAt) >= luopingMappingCacheIdleTTL {
			delete(s.writeCache, k)
		}
	}
	if len(s.writeCache) >= luopingMappingCacheMax {
		s.writeCache = make(map[string]mappingWrite)
	}
}

// forgetMappingWrite drops the local record so the next uplink re-persists the
// binding (used when the session is torn down).
func (s *LuopingMappingService) forgetMappingWrite(deviceId string) {
	if s == nil {
		return
	}
	s.writeMu.Lock()
	delete(s.writeCache, deviceId)
	s.writeMu.Unlock()
}

func deviceMappingKey(deviceId string) string { return luopingDeviceKeyPrefix + deviceId }
func imeiMappingKey(imei string) string       { return luopingImeiKeyPrefix + imei }
func presenceEpochKey(deviceId string) string { return luopingPresenceKeyPrefix + deviceId }

// TryAdvancePresence records tsMillis as the latest presence/activity time for
// deviceId when it is newer than the previously stored value. It returns true when
// the event is the newest observed so far (i.e. should be applied), and false when
// it is a stale/out-of-order event that should be ignored.
//
// This is the cross-replica ordering guard: the epoch lives in shared Redis, so a
// reconnect's (newer) "connected" event processed by ANY replica wins over a late
// "disconnected" event processed by another replica. Uplink activity also advances
// the epoch, so a delayed disconnect older than the last report is rejected.
//
// It fails open (returns true) when no decision can be made — missing redis, a
// non-positive ts, or a redis error — preserving legacy behavior.
func (s *LuopingMappingService) TryAdvancePresence(ctx context.Context, deviceId string, tsMillis int64, ttl time.Duration) (bool, error) {
	if s == nil || s.Rdb == nil || deviceId == "" || tsMillis <= 0 {
		return true, nil
	}
	if ttl <= 0 {
		ttl = luopingPresenceTTL
	}
	res, err := presenceEpochScript.Run(ctx, s.Rdb, []string{presenceEpochKey(deviceId)}, tsMillis, int64(ttl.Seconds())).Int()
	if err != nil {
		return true, err
	}
	return res == 1, nil
}

// SaveMapping persists the deviceId <-> IMEI binding. Both keys are persistent, so
// this is a no-op once this process has written the pair: the binding is established
// when the session starts and rewritten only when it actually changes.
func (s *LuopingMappingService) SaveMapping(ctx context.Context, deviceId, imei, groupName string) error {
	if s == nil || s.Rdb == nil {
		return fmt.Errorf("luoping mapping redis not initialized")
	}
	if deviceId == "" || imei == "" {
		return fmt.Errorf("deviceId and imei are required")
	}
	started := time.Now()
	if !s.needsMappingWrite(deviceId, imei, groupName, started) {
		return nil
	}

	mapping := &dto.LuopingDeviceMapping{
		Imei:      imei,
		DeviceId:  deviceId,
		GroupName: groupName,
		UpdatedAt: started.Unix(),
	}
	val, err := json.Marshal(mapping)
	if err != nil {
		return err
	}

	pipe := s.Rdb.Pipeline()
	// Both directions are persistent: the binding must outlive any offline period so
	// a returning device (and historical IMEI lookups) still resolve.
	pipe.Set(ctx, deviceMappingKey(deviceId), string(val), 0)
	pipe.Set(ctx, imeiMappingKey(imei), string(val), 0)
	if _, err := pipe.Exec(ctx); err != nil {
		logger.Log.Error("luoping SaveMapping failed",
			zap.String("deviceId", deviceId),
			zap.String("imei", imei),
			zap.Error(err),
		)
		return err
	}
	s.markMappingWritten(deviceId, imei, groupName, started)
	logger.Log.Info("luoping mapping saved",
		zap.String("deviceId", deviceId),
		zap.String("imei", imei),
		zap.String("groupName", groupName),
	)
	return nil
}

func (s *LuopingMappingService) GetByDeviceId(ctx context.Context, deviceId string) (*dto.LuopingDeviceMapping, error) {
	if s == nil || s.Rdb == nil || deviceId == "" {
		return nil, nil
	}
	val, err := s.Rdb.Get(ctx, deviceMappingKey(deviceId)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var mapping dto.LuopingDeviceMapping
	if err := json.Unmarshal([]byte(val), &mapping); err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (s *LuopingMappingService) GetByImei(ctx context.Context, imei string) (*dto.LuopingDeviceMapping, error) {
	if s == nil || s.Rdb == nil || imei == "" {
		return nil, nil
	}
	val, err := s.Rdb.Get(ctx, imeiMappingKey(imei)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var mapping dto.LuopingDeviceMapping
	if err := json.Unmarshal([]byte(val), &mapping); err != nil {
		return nil, err
	}
	return &mapping, nil
}

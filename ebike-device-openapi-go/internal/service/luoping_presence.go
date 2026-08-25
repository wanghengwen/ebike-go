package service

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"ebike-device-openapi-go/internal/api/dto"
	"ebike-device-openapi-go/internal/pkg/config"
	"ebike-device-openapi-go/internal/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	luopingLastSeenZSet = "mqtt:luoping:lastseen"
	luopingSweepLockKey = "mqtt:luoping:sweep:lock"
	luopingSweepLockTTL = 25 * time.Second

	// Sweep cadence (idle heartbeat ≈ 10m, onlineTtl ≈ 30m):
	// - With disconnectedTopic: real offline from $SYS disconnected; sweeper is
	//   catch-up for missed events / silent devices → 15m.
	// - Without disconnectedTopic: lastSeen sweep is primary → 2m (enough vs 30m TTL).
	luopingSweepIntervalPresence = 15 * time.Minute
	luopingSweepIntervalFallback = 2 * time.Minute
)

// sweepUnlockScript releases the sweep lock only when we still own it (token match).
var sweepUnlockScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('DEL', KEYS[1])
end
return 0
`)

// LuopingPresenceSweeper marks stale luoping sessions offline using a Redis ZSET of lastSeen.
type LuopingPresenceSweeper struct {
	Register *RegisterService
	Mapping  *LuopingMappingService
	Rdb      *redis.Client

	stopOnce sync.Once
	stopCh   chan struct{}
}

func NewLuopingPresenceSweeper(reg *RegisterService, mapping *LuopingMappingService, rdb *redis.Client) *LuopingPresenceSweeper {
	return &LuopingPresenceSweeper{
		Register: reg,
		Mapping:  mapping,
		Rdb:      rdb,
		stopCh:   make(chan struct{}),
	}
}

// Touch records lastSeen score for an IMEI (unix seconds).
func (s *LuopingPresenceSweeper) Touch(ctx context.Context, imei string, ts int64) {
	if s == nil || s.Rdb == nil || imei == "" {
		return
	}
	if ts <= 0 {
		ts = time.Now().Unix()
	}
	if err := s.Rdb.ZAdd(ctx, luopingLastSeenZSet, redis.Z{Score: float64(ts), Member: imei}).Err(); err != nil {
		logger.Log.Warn("luoping lastSeen zadd failed", zap.String("imei", imei), zap.Error(err))
	}
}

// Remove drops an IMEI from the lastSeen index (on explicit disconnect).
func (s *LuopingPresenceSweeper) Remove(ctx context.Context, imei string) {
	if s == nil || s.Rdb == nil || imei == "" {
		return
	}
	_ = s.Rdb.ZRem(ctx, luopingLastSeenZSet, imei).Err()
}

// Start runs a background loop that periodically calls sweepOnce.
// Interval depends on whether EMQX presence topics are configured (see sweepInterval).
func (s *LuopingPresenceSweeper) Start() {
	if s == nil {
		return
	}
	interval := sweepInterval()
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-s.stopCh:
				return
			case <-ticker.C:
				s.sweepOnce(context.Background())
			}
		}
	}()
	logger.Log.Info("luoping presence sweeper started",
		zap.Duration("interval", interval),
		zap.Bool("presenceCatchupMode", presenceCatchupMode()),
	)
}

// presenceCatchupMode reports whether the sweeper should run in infrequent
// catch-up mode (15m). Safe only when MQTT is enabled AND disconnectedTopic is
// configured (connectedTopic is ignored by design).
func presenceCatchupMode() bool {
	mcfg := config.GetConfig().Mqtt
	return shouldUsePresenceCatchup(mcfg.Enabled, mcfg.DisconnectedTopic)
}

// shouldUsePresenceCatchup is the pure decision for whether the sweeper may use
// the infrequent (15m) catch-up cadence. Only disconnected presence counts.
func shouldUsePresenceCatchup(mqttEnabled bool, disconnectedTopic string) bool {
	if !mqttEnabled {
		return false
	}
	return disconnectedTopic != ""
}

// sweepInterval returns how often the stale-lastSeen catch-up scan should run.
func sweepInterval() time.Duration {
	if presenceCatchupMode() {
		return luopingSweepIntervalPresence
	}
	return luopingSweepIntervalFallback
}

func (s *LuopingPresenceSweeper) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() { close(s.stopCh) })
}

func (s *LuopingPresenceSweeper) sweepOnce(ctx context.Context) {
	if s.Rdb == nil || s.Register == nil {
		return
	}

	// Distributed lock: only one replica performs the sweep per tick, otherwise
	// every replica would independently force the same stale IMEIs offline and emit
	// duplicate lineState=0 events.
	token := uuid.NewString()
	ok, err := s.Rdb.SetNX(ctx, luopingSweepLockKey, token, luopingSweepLockTTL).Result()
	if err != nil {
		logger.Log.Warn("luoping sweep lock acquire failed", zap.Error(err))
		return
	}
	if !ok {
		return
	}
	defer func() {
		_ = sweepUnlockScript.Run(ctx, s.Rdb, []string{luopingSweepLockKey}, token).Err()
	}()

	ttl := config.GetMqttOnlineTTL()
	cutoff := time.Now().Add(-ttl).Unix()

	imeis, err := s.Rdb.ZRangeByScore(ctx, luopingLastSeenZSet, &redis.ZRangeBy{
		Min: "-inf",
		Max: strconv.FormatInt(cutoff, 10),
	}).Result()
	if err != nil {
		logger.Log.Warn("luoping presence sweep query failed", zap.Error(err))
		return
	}
	if len(imeis) == 0 {
		return
	}

	for _, imei := range imeis {
		s.forceOfflineByImei(ctx, imei, "heartbeat-timeout")
	}
}

// forceOfflineByImei is the heartbeat-timeout offline path (lastSeen TTL sweeper).
// EMQX disconnected uses LuopingService.HandleDisconnected → Register.Unregister
// directly, mirroring gateway channelInactive → /xiaoan/logout.
func (s *LuopingPresenceSweeper) forceOfflineByImei(ctx context.Context, imei string, reason string) {
	if s == nil || imei == "" {
		return
	}
	key := ecuLoginKey(imei)
	rdb := s.Rdb
	if rdb == nil && s.Register != nil {
		rdb = s.Register.Rdb
	}
	if rdb == nil {
		return
	}
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil || val == "" {
		s.Remove(ctx, imei)
		return
	}
	if err != nil {
		return
	}
	var ecu dto.EcuLogin
	if err := json.Unmarshal([]byte(val), &ecu); err != nil {
		return
	}
	if !strings.EqualFold(ecu.Type, dto.DeviceTypeLuoping) {
		s.Remove(ctx, imei)
		return
	}

	channelId := ecu.ChannelId
	if channelId == "" {
		channelId = ecu.DeviceId
	}
	logger.Log.Info("luoping force offline",
		zap.String("reason", reason),
		zap.String("imei", imei),
		zap.String("deviceId", ecu.DeviceId),
		zap.String("channelId", channelId),
		zap.Int64p("lastSeenAt", ecu.LastSeenAt),
	)
	if s.Register != nil {
		_ = s.Register.Unregister(&dto.LogoutCmd{Imei: imei, ChannelId: channelId}, false)
	}
	s.Remove(ctx, imei)
	if s.Mapping != nil && ecu.DeviceId != "" {
		s.Mapping.forgetMappingWrite(ecu.DeviceId)
	}
}

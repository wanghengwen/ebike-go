package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ebike-device-openapi-go/internal/api/dto"
	"ebike-device-openapi-go/internal/pkg/cache"
	"ebike-device-openapi-go/internal/pkg/kafka"
	"ebike-device-openapi-go/internal/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RegisterService struct {
	Rdb       *redis.Client // session: ecu_login_*, locks, mqtt mapping
	DeviceRdb *redis.Client // consume/paas: device_info_*, imei_car_*; nil → Rdb
	Pusher    *kafka.Pusher
}

// ecuLoginKey is the Redis session key for a device.
func ecuLoginKey(imei string) string { return "ecu_login_" + imei }

// deviceRedis returns the Redis used for device_info / imei_car / offline_time_ex.
func (s *RegisterService) deviceRedis() *redis.Client {
	if s != nil && s.DeviceRdb != nil {
		return s.DeviceRdb
	}
	if s == nil {
		return nil
	}
	return s.Rdb
}

// Register (Login) mimics com.xyy.ebike.device.openapi.domain.service.impl.RegisterImpl.register()
//
// CRITICAL BEHAVIOR NOTE:
// In Java, messagePusher.pushMessage() is called UNCONDITIONALLY after doEcuBindToGateway(),
// regardless of whether the Redis distributed lock was acquired.
// The Go version MUST NOT return early on lock failure — doing so would drop the Kafka
// lineState="1" event, causing the downstream ebike-device-consume to miss the device online event.
func (s *RegisterService) Register(cmd *dto.LoginCmd, shadowMode bool) bool {
	ecuLogin := &dto.EcuLogin{
		Host:          cmd.Host,
		Port:          cmd.Port,
		Type:          dto.DeviceTypeXiaoan,
		Version:       cmd.Version,
		RemoteAddress: cmd.RemoteAddress,
		ChannelId:     cmd.ChannelId,
		Timestamp:     cmd.Timestamp,
	}
	return s.registerWithEcuLogin(cmd, ecuLogin, shadowMode)
}

// RegisterLuoping binds a luoping MQTT device session (type=luoping) and pushes lineState=1.
// Always applies side effects: Luoping MQTT has no shadowMode path.
func (s *RegisterService) RegisterLuoping(cmd *dto.LoginCmd, ecuLogin *dto.EcuLogin) error {
	if ecuLogin == nil {
		return fmt.Errorf("ecuLogin is required")
	}
	ecuLogin.Type = dto.DeviceTypeLuoping
	if ecuLogin.ChannelId == "" {
		ecuLogin.ChannelId = cmd.ChannelId
	}
	if ecuLogin.Host == "" {
		ecuLogin.Host = cmd.Host
	}
	if ecuLogin.Port == 0 {
		ecuLogin.Port = cmd.Port
	}
	if ecuLogin.Version == nil {
		ecuLogin.Version = cmd.Version
	}
	if ecuLogin.RemoteAddress == "" {
		ecuLogin.RemoteAddress = cmd.RemoteAddress
	}
	if ecuLogin.Timestamp == nil {
		ecuLogin.Timestamp = cmd.Timestamp
	}
	ok := s.registerWithEcuLogin(cmd, ecuLogin, false)
	if !ok {
		return fmt.Errorf("luoping register failed")
	}
	return nil
}

func (s *RegisterService) registerWithEcuLogin(cmd *dto.LoginCmd, ecuLogin *dto.EcuLogin, shadowMode bool) bool {
	ret := s.doEcuBindToGateway(cmd.Imei, ecuLogin, shadowMode)

	ts := cmd.Timestamp
	if ts == nil {
		// Java: Optional.ofNullable(cmd.getTimestamp()).orElse(System.currentTimeMillis() / 1000)
		now := time.Now().Unix()
		ts = &now
	}

	// Nested DeviceReportMessage for worker → saas → consume:
	// openapi publishes {imei,msgType,bussinessType,data:"{...cmd...}"};
	// worker parses msg.Data.cmd (35→login), injects deviceDataType/appId, fans out to to-saas.
	eventData := map[string]interface{}{
		"imei":      cmd.Imei,
		"imsi":      cmd.Imsi,
		"version":   cmd.Version,
		"timestamp": *ts,
		"msgType":   "event",
		"cmd":       35,
		"lineState": "1",
		"originCmd": 35,
	}
	if cmd.DeviceType != nil {
		eventData["deviceType"] = *cmd.DeviceType
	}
	if strings.EqualFold(ecuLogin.Type, dto.DeviceTypeLuoping) {
		eventData["deviceId"] = ecuLogin.DeviceId
		eventData["groupName"] = ecuLogin.GroupName
		eventData["protocol"] = dto.DeviceTypeLuoping
	}

	s.pushNestedEvent(cmd.Imei, eventData, shadowMode)
	return ret
}

func (s *RegisterService) Unregister(cmd *dto.LogoutCmd, shadowMode bool) bool {
	ret := s.doEcuUnBindToGateway(cmd, shadowMode)

	// Drop process-local session cache so subsequent command routing cannot
	// treat the device as still bound after Redis delete.
	cache.DefaultEcuLoginCache.Delete(cmd.Imei)

	// Resolve bind before side-effects (mark offline / force timer). Kafka payload itself
	// does not need tenantId — worker derives appId from device_ebike_{imei}.
	tenantId, carId := s.lookupDeviceBind(cmd.Imei)

	// Nested logout — same envelope as Java RegisterImpl.unregister (cmd=1001 + event-topic).
	eventData := map[string]interface{}{
		"imei":      cmd.Imei,
		"timestamp": time.Now().Unix(),
		"msgType":   "event",
		"cmd":       1001,
		"lineState": "0",
	}
	if cmd.ChannelId != "" {
		eventData["channelId"] = cmd.ChannelId
	}

	s.pushNestedEvent(cmd.Imei, eventData, shadowMode)
	logger.Log.Info("device logout kafka event queued",
		zap.String("imei", cmd.Imei),
		zap.String("channelId", cmd.ChannelId),
		zap.String("tenantId", tenantId),
		zap.String("carId", carId),
		zap.String("msgType", "event"),
		zap.Int("cmd", 1001),
	)

	// Post-0354ae9 offline path (kept intentionally after restoring nested Kafka):
	// Page「中控状态」reads isOnline from packed device_info_{tenant}_{imei};
	// Kafka Logout alone only sets isDisconnect on saas path after worker; force
	// offline_time_ex_* so consume LogoutCommitted flips APP isOnline quickly.
	if !shadowMode {
		s.markDeviceInfoOffline(cmd.Imei, tenantId)
		s.forceConsumeOfflineTimer(cmd.Imei, tenantId, carId)
	}
	return ret
}

// pushNestedEvent wraps inner fields as DeviceReportMessage.data (JSON string) and pushes
// via PushMessage so msgType=event lands on event-topic for worker-go.
func (s *RegisterService) pushNestedEvent(imei string, inner map[string]interface{}, shadowMode bool) {
	if s == nil || s.Pusher == nil || imei == "" {
		return
	}
	dataBytes, err := json.Marshal(inner)
	if err != nil {
		logger.Log.Error("nested event marshal failed",
			zap.String("imei", imei), zap.Error(err))
		return
	}
	msg := &dto.DeviceReportMessage{
		Imei:          imei,
		MsgType:       "event",
		BussinessType: "ebike",
		Data:          string(dataBytes),
	}
	s.Pusher.PushMessage(imei, msg, shadowMode)
	logger.Log.Info("nested event kafka queued",
		zap.String("imei", imei),
		zap.String("msgType", "event"),
		zap.Any("cmd", inner["cmd"]),
		zap.Int("dataBytes", len(dataBytes)),
	)
}

// Packed device_info offsets (consume DeviceProtocol).
const (
	deviceInfoCarIDOffset    = 44
	deviceInfoCarIDLen       = 15
	deviceInfoTenantOffset   = 78
	deviceInfoTenantLen      = 20
	deviceInfoIsOnlineOffset = 317
)

// lookupDeviceBind resolves tenantId/carId from consume/paas Redis (DeviceRdb, often db4).
func (s *RegisterService) lookupDeviceBind(imei string) (tenantId, carId string) {
	tenantId = "1"
	rdb := s.deviceRedis()
	if rdb == nil || imei == "" {
		return tenantId, ""
	}
	ctx := context.Background()
	// Prefer SCAN imei_car_*_{imei} (tenant may not be 1).
	if tid, cid := s.scanImeiCarBind(ctx, rdb, imei); cid != "" {
		return tid, cid
	}
	if tid, cid := s.lookupBindFromDeviceInfo(ctx, rdb, imei, tenantId); cid != "" {
		if tid != "" {
			tenantId = tid
		}
		return tenantId, cid
	}
	return tenantId, ""
}

func (s *RegisterService) scanImeiCarBind(ctx context.Context, rdb *redis.Client, imei string) (tenantId, carId string) {
	if rdb == nil {
		return "", ""
	}
	pattern := "imei_car_*_" + imei
	var cursor uint64
	for i := 0; i < 20; i++ {
		keys, next, err := rdb.Scan(ctx, cursor, pattern, 20).Result()
		if err != nil {
			return "", ""
		}
		for _, k := range keys {
			rest := strings.TrimPrefix(k, "imei_car_")
			rest = strings.TrimSuffix(rest, "_"+imei)
			if rest == "" || rest == k {
				continue
			}
			if v, err := rdb.Get(ctx, k).Result(); err == nil {
				if cid := strings.Trim(strings.TrimSpace(v), "\""); cid != "" {
					return rest, cid
				}
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return "", ""
}

func (s *RegisterService) lookupBindFromDeviceInfo(ctx context.Context, rdb *redis.Client, imei, hintTenant string) (tenantId, carId string) {
	if rdb == nil {
		return "", ""
	}
	keys := []string{}
	if hintTenant != "" {
		keys = append(keys, "device_info_"+hintTenant+"_"+imei)
	}
	for _, tid := range []string{"2", "1"} {
		k := "device_info_" + tid + "_" + imei
		dup := false
		for _, e := range keys {
			if e == k {
				dup = true
				break
			}
		}
		if !dup {
			keys = append(keys, k)
		}
	}
	var cursor uint64
	for i := 0; i < 10; i++ {
		found, next, err := rdb.Scan(ctx, cursor, "device_info_*_"+imei, 10).Result()
		if err != nil {
			break
		}
		keys = append(keys, found...)
		cursor = next
		if cursor == 0 {
			break
		}
	}
	seen := map[string]bool{}
	for _, key := range keys {
		if seen[key] {
			continue
		}
		seen[key] = true
		val, err := rdb.Get(ctx, key).Result()
		if err != nil || val == "" {
			continue
		}
		parts := strings.SplitN(key, "_", 4)
		if len(parts) >= 4 {
			tenantId = parts[2]
		}
		if len(val) >= deviceInfoTenantOffset+deviceInfoTenantLen {
			if t := strings.TrimSpace(val[deviceInfoTenantOffset : deviceInfoTenantOffset+deviceInfoTenantLen]); t != "" {
				tenantId = t
			}
		}
		if len(val) >= deviceInfoCarIDOffset+deviceInfoCarIDLen {
			carId = strings.TrimSpace(val[deviceInfoCarIDOffset : deviceInfoCarIDOffset+deviceInfoCarIDLen])
		}
		if carId != "" {
			return tenantId, carId
		}
	}
	return tenantId, carId
}

// markDeviceInfoOffline sets isOnline=0 in device_info_{tenant}_{imei} (paas UI source).
func (s *RegisterService) markDeviceInfoOffline(imei, tenantId string) {
	rdb := s.deviceRedis()
	if rdb == nil || imei == "" {
		return
	}
	ctx := context.Background()
	key := ""
	if tenantId != "" {
		cand := "device_info_" + tenantId + "_" + imei
		if n, err := rdb.Exists(ctx, cand).Result(); err == nil && n > 0 {
			key = cand
		}
	}
	if key == "" {
		var cursor uint64
		for i := 0; i < 10 && key == ""; i++ {
			found, next, err := rdb.Scan(ctx, cursor, "device_info_*_"+imei, 5).Result()
			if err != nil {
				break
			}
			if len(found) > 0 {
				key = found[0]
			}
			cursor = next
			if cursor == 0 {
				break
			}
		}
	}
	if key == "" {
		logger.Log.Info("mark device_info offline skipped: key missing",
			zap.String("imei", imei), zap.String("tenantId", tenantId),
			zap.String("hint", "set spring.redis.device-database=4 if device_info is on db4"))
		return
	}
	if err := rdb.SetRange(ctx, key, deviceInfoIsOnlineOffset, "0").Err(); err != nil {
		logger.Log.Warn("mark device_info offline failed",
			zap.String("key", key), zap.String("imei", imei), zap.Error(err))
		return
	}
	logger.Log.Info("mark device_info offline",
		zap.String("key", key),
		zap.String("imei", imei),
		zap.Int("offset", deviceInfoIsOnlineOffset),
	)
}

// forceConsumeOfflineTimer sets offline_time_ex_{tenant}_{carId} to expire in 1s so
// consume RedisListener publishes LogoutCommitted and APP isOnline becomes 0.
func (s *RegisterService) forceConsumeOfflineTimer(imei, tenantId, carId string) {
	rdb := s.deviceRedis()
	if rdb == nil || carId == "" || tenantId == "" {
		logger.Log.Info("skip force offline timer: missing bind",
			zap.String("imei", imei),
			zap.String("tenantId", tenantId),
			zap.String("carId", carId),
		)
		return
	}
	ctx := context.Background()
	key := "offline_time_ex_" + tenantId + "_" + carId
	if err := rdb.Set(ctx, key, "1", time.Second).Err(); err != nil {
		logger.Log.Warn("force offline timer failed",
			zap.String("key", key), zap.String("imei", imei), zap.Error(err))
		return
	}
	logger.Log.Info("force offline timer armed",
		zap.String("key", key),
		zap.String("imei", imei),
		zap.Duration("ttl", time.Second),
	)
}

func (s *RegisterService) doEcuBindToGateway(imei string, ecuLogin *dto.EcuLogin, shadowMode bool) bool {
	lockKey := "device_login_lock_" + imei
	ctx := context.Background()

	if shadowMode {
		logger.Log.Info("shadow mode intercept doEcuBindToGateway", zap.String("imei", imei))
		return true
	}

	acquired, err := tryLock(ctx, s.Rdb, lockKey)
	if err != nil {
		logger.Log.Error("doEcuBindToGateway exception", zap.String("imei", imei), zap.Error(err))
		return false
	}
	if !acquired {
		logger.Log.Warn("doEcuBindToGateway lock not acquired", zap.String("imei", imei))
		return true
	}
	defer s.Rdb.Del(ctx, lockKey)

	key := ecuLoginKey(imei)

	val, _ := json.Marshal(ecuLogin)
	if err := s.Rdb.Set(ctx, key, string(val), 0).Err(); err != nil {
		logger.Log.Error("doEcuBindToGateway redis set failed", zap.String("key", key), zap.Error(err))
	} else {
		cache.DefaultEcuLoginCache.Put(imei, ecuLogin)
	}
	logger.Log.Info("device login cache write", zap.String("key", key), zap.String("value", string(val)))

	return true
}

func (s *RegisterService) doEcuUnBindToGateway(cmd *dto.LogoutCmd, shadowMode bool) bool {
	lockKey := "device_login_lock_" + cmd.Imei
	ctx := context.Background()

	if shadowMode {
		logger.Log.Info("shadow mode intercept doEcuUnBindToGateway", zap.String("imei", cmd.Imei))
		return true
	}

	// Use Redisson-like tryLock with 500ms wait timeout and 10s lease time
	acquired, err := tryLock(ctx, s.Rdb, lockKey)
	if err != nil {
		logger.Log.Error("doEcuUnBindToGateway exception", zap.String("imei", cmd.Imei), zap.Error(err))
		return false
	}
	if !acquired {
		// Lock contention — Java falls through to return true. Skip Redis delete but return success.
		logger.Log.Warn("doEcuUnBindToGateway lock not acquired", zap.String("imei", cmd.Imei))
		return true
	}
	defer s.Rdb.Del(ctx, lockKey)

	key := ecuLoginKey(cmd.Imei)

	val, err := s.Rdb.Get(ctx, key).Result()
	if err == nil && val != "" {
		var ecuLogin dto.EcuLogin
		if err := json.Unmarshal([]byte(val), &ecuLogin); err == nil {
			// Java: StringUtils.equals(ecuLogin.getChannelId(), cmd.getChannelId())
			// StringUtils.equals is case-SENSITIVE (not case-insensitive).
			// MUST use == here, NOT strings.EqualFold.
			if ecuLogin.ChannelId == cmd.ChannelId {
				s.Rdb.Del(ctx, key)
				cache.DefaultEcuLoginCache.Delete(cmd.Imei)
				logger.Log.Info("device logout cache delete", zap.String("key", key))
			} else {
				logger.Log.Warn("device logout channelId mismatch",
					zap.String("key", key),
					zap.String("ecuChannelId", ecuLogin.ChannelId),
					zap.String("cmdChannelId", cmd.ChannelId),
				)
			}
		}
	} else {
		logger.Log.Warn("device logout ecuLogin empty", zap.String("key", key))
	}

	return true
}

// tryLock simulates Redisson's tryLock with a waitTime and a leaseTime.
func tryLock(ctx context.Context, rdb *redis.Client, lockKey string) (bool, error) {
	start := time.Now()
	for {
		acquired, err := rdb.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
		if err != nil {
			return false, err
		}
		if acquired {
			return true, nil
		}
		if time.Since(start) >= 500*time.Millisecond {
			return false, nil
		}
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}

// UpdateHost mimics RegisterImpl.updateHost().
// Note: Java creates a NEW EcuLogin with only host/port/type set when host differs,
// effectively clearing version/remoteAddress/channelId/timestamp from Redis.
// This is intentional Java behavior replicated here.
// Java source: RegisterImpl.java line 113-119
func (s *RegisterService) UpdateHost(imei, host string, shadowMode bool) {
	if imei == "" || host == "" {
		return
	}

	if shadowMode {
		logger.Log.Info("shadow mode intercept UpdateHost", zap.String("imei", imei), zap.String("host", host))
		return
	}

	key := ecuLoginKey(imei)

	ctx := context.Background()
	val, _ := s.Rdb.Get(ctx, key).Result()

	var ecuLogin dto.EcuLogin
	if val != "" {
		if err := json.Unmarshal([]byte(val), &ecuLogin); err != nil {
			logger.Log.Error("UpdateHost json unmarshal error", zap.String("imei", imei), zap.Error(err))
		}
	}

	if ecuLogin.Host != host {
		// Update only the host-routing fields and PRESERVE channelId/version/remoteAddress/
		// timestamp. This intentionally diverges from the Java updateHost behavior, which
		// rebuilt a bare EcuLogin and wiped channelId — that is exactly what produced the
		// "device logout channelId mismatch" warnings (logout could no longer match the
		// channelId written at login time). Keeping channelId lets logout match correctly.
		ecuLogin.Host = host
		ecuLogin.Port = 8080 // unified fixed 8080 port (matches Java)
		ecuLogin.Type = dto.DeviceTypeXiaoan
		newVal, _ := json.Marshal(&ecuLogin)
		s.Rdb.Set(ctx, key, string(newVal), 0)
		logger.Log.Info("update host", zap.String("imei", imei), zap.String("host", host))
	}
}

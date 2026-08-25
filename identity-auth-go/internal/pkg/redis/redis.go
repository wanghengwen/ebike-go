package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"identity-auth-go/internal/pkg/config"
)

// Client is the global Redis client used throughout the application.
var Client *redis.Client

// Init creates and tests the Redis connection using GlobalConfig.
func Init() {
	cfg := config.GlobalConfig.Redis
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.Database,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		log.Printf("[redis] WARNING: ping failed: %v (addr=%s)", err, addr)
	} else {
		log.Printf("[redis] connected to %s, db=%d", addr, cfg.Database)
	}
}

// Reinit re-creates the Redis client after a Nacos config hot-reload.
func Reinit() {
	if Client != nil {
		_ = Client.Close()
	}
	Init()
}

// ---------------------------------------------------------------------------
// Lua scripts matching Java CommonServiceImpl
// ---------------------------------------------------------------------------

var increaseScript = redis.NewScript(`
local currentTimes = redis.call('get', KEYS[1])
if currentTimes == false then currentTimes = 0 end
currentTimes = tonumber(currentTimes)
if currentTimes >= 0 then
    currentTimes = currentTimes + tonumber(ARGV[1])
    redis.call('set', KEYS[1], currentTimes)
    return currentTimes
end
return currentTimes
`)

var decreaseScript = redis.NewScript(`
local currentTimes = redis.call('get', KEYS[1])
if currentTimes == false then currentTimes = 0 end
currentTimes = tonumber(currentTimes)
if currentTimes > 0 then
    currentTimes = currentTimes - 1
    redis.call('set', KEYS[1], currentTimes)
    return currentTimes
end
return currentTimes
`)

func IncreaseCallTimes(tenantId string, authType int, times int64) (int64, error) {
	ctx := context.Background()
	key := chargeRedisKey(tenantId, authType)
	result, err := increaseScript.Run(ctx, Client, []string{key}, times).Int64()
	if err != nil {
		return 0, fmt.Errorf("IncreaseCallTimes key=%s: %w", key, err)
	}
	return result, nil
}

func DecreaseCallTimes(tenantId string, authType int) (int64, error) {
	ctx := context.Background()
	key := chargeRedisKey(tenantId, authType)
	result, err := decreaseScript.Run(ctx, Client, []string{key}).Int64()
	if err != nil {
		return 0, fmt.Errorf("DecreaseCallTimes key=%s: %w", key, err)
	}
	return result, nil
}

func GetCallTimes(tenantId string, authType int) (int64, error) {
	ctx := context.Background()
	key := chargeRedisKey(tenantId, authType)
	val, err := Client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("GetCallTimes key=%s: %w", key, err)
	}
	return val, nil
}

// ---------------------------------------------------------------------------
// Auth result caching
// ---------------------------------------------------------------------------

const authCacheFalseTTL = 7 * 24 * time.Hour

// SetAuthResult caches the authentication result.
// Writes JSON boolean true/false to match Java GenericJackson2JsonRedisSerializer.
func SetAuthResult(idCardNum, name string, result bool) error {
	ctx := context.Background()
	key := authResultRedisKey(idCardNum, name)

	val := "false"
	if result {
		val = "true"
	}

	var err error
	if result {
		err = Client.Set(ctx, key, val, 0).Err()
	} else {
		err = Client.Set(ctx, key, val, authCacheFalseTTL).Err()
	}
	if err != nil {
		return fmt.Errorf("SetAuthResult key=%s: %w", key, err)
	}
	return nil
}

// GetAuthResult returns (result, exists, error).
func GetAuthResult(idCardNum, name string) (bool, bool, error) {
	ctx := context.Background()
	key := authResultRedisKey(idCardNum, name)

	val, err := Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, false, nil
	}
	if err != nil {
		return false, false, fmt.Errorf("GetAuthResult key=%s: %w", key, err)
	}

	result, ok := parseAuthCacheValue(val)
	if !ok {
		return false, false, nil
	}
	return result, true, nil
}

// parseAuthCacheValue reads Java boolean JSON and legacy Go "1"/"0" values.
func parseAuthCacheValue(val string) (bool, bool) {
	val = strings.TrimSpace(val)
	switch strings.ToLower(val) {
	case "true", "1":
		return true, true
	case "false", "0":
		return false, true
	}
	var b bool
	if err := json.Unmarshal([]byte(val), &b); err == nil {
		return b, true
	}
	return false, false
}

func chargeRedisKey(tenantId string, authType int) string {
	return fmt.Sprintf("charge_tid_type:%s:%d", tenantId, authType)
}

func authResultRedisKey(idCardNum, name string) string {
	return fmt.Sprintf("au:%s:%s", idCardNum, name)
}

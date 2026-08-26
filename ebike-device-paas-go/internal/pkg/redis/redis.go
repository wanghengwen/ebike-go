// Package redis wraps the go-redis client used to read device caches, mirroring
// the Java xyy-redis RedisMapper access patterns (GET / MGET on fixed-length
// device-info strings). Keys are used verbatim (no global prefix), matching the
// Java DevicePassRedisKey.format output.
package redis

import (
	"context"
	"log"
	"strconv"
	"time"

	"ebike-device-paas-go/internal/pkg/config"

	goredis "github.com/redis/go-redis/v9"
)

var (
	// Client is the shared Redis client. Nil until InitRedis succeeds.
	Client *goredis.Client
	bgCtx  = context.Background()
)

// InitRedis connects to Redis using the merged config. Connection failure is
// non-fatal so the service can still start (health probes, shadow setup); query
// handlers degrade gracefully when Client is nil.
func InitRedis() {
	addr := config.RedisAddr()
	if config.GlobalConfig.Redis.Host == "" {
		log.Println("[redis] host empty, skipping Redis initialization")
		return
	}
	c := goredis.NewClient(&goredis.Options{
		Addr:         addr,
		Password:     config.GlobalConfig.Redis.Password,
		DB:           config.GlobalConfig.Redis.Database,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
		MinIdleConns: 5,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Ping(ctx).Err(); err != nil {
		log.Printf("[redis] connect failed (addr=%s db=%d): %v — query endpoints will be unavailable",
			addr, config.GlobalConfig.Redis.Database, err)
		return
	}
	Client = c
	log.Printf("[redis] connected (addr=%s db=%d)", addr, config.GlobalConfig.Redis.Database)
}

// Ctx returns the background context for Redis calls.
func Ctx() context.Context { return bgCtx }

// Get returns the string value at key, or "" if missing / on error.
func Get(key string) string {
	if Client == nil {
		return ""
	}
	v, err := Client.Get(bgCtx, key).Result()
	if err != nil {
		return ""
	}
	return v
}

// MGet returns values for keys in order. Missing keys yield "" entries.
func MGet(keys ...string) []string {
	out := make([]string, len(keys))
	if Client == nil || len(keys) == 0 {
		return out
	}
	vals, err := Client.MGet(bgCtx, keys...).Result()
	if err != nil {
		log.Printf("[redis] MGet failed (%d keys): %v", len(keys), err)
		return out
	}
	for i, v := range vals {
		if s, ok := v.(string); ok {
			out[i] = s
		}
	}
	return out
}

// ZRangeByScoreMin returns members whose score >= min (max = +inf), mirroring
// Java redisMapper.zRangeByScore(key, min, Long.MAX_VALUE).
func ZRangeByScoreMin(key string, min int64) []string {
	if Client == nil {
		return nil
	}
	res, err := Client.ZRangeByScore(bgCtx, key, &goredis.ZRangeBy{
		Min: strconv.FormatInt(min, 10),
		Max: "+inf",
	}).Result()
	if err != nil {
		log.Printf("[redis] ZRangeByScore %s failed: %v", key, err)
		return nil
	}
	return res
}

// ZRangeByScoreRange returns members with score in [min, max], mirroring Java
// redisMapper.zRangeByScore(key, min, max).
func ZRangeByScoreRange(key string, min, max float64) []string {
	if Client == nil {
		return nil
	}
	res, err := Client.ZRangeByScore(bgCtx, key, &goredis.ZRangeBy{
		Min: strconv.FormatFloat(min, 'f', -1, 64),
		Max: strconv.FormatFloat(max, 'f', -1, 64),
	}).Result()
	if err != nil {
		log.Printf("[redis] ZRangeByScore %s failed: %v", key, err)
		return nil
	}
	return res
}

// ScoreMember is a zset member with its score.
type ScoreMember struct {
	Member string
	Score  float64
}

// ZRangeByScoreWithScores returns members + scores with score in [min, max],
// mirroring Java redisMapper.zRangeByScoreWithScores(key, min, max).
func ZRangeByScoreWithScores(key string, min, max float64) []ScoreMember {
	if Client == nil {
		return nil
	}
	res, err := Client.ZRangeByScoreWithScores(bgCtx, key, &goredis.ZRangeBy{
		Min: strconv.FormatFloat(min, 'f', -1, 64),
		Max: strconv.FormatFloat(max, 'f', -1, 64),
	}).Result()
	if err != nil {
		log.Printf("[redis] ZRangeByScoreWithScores %s failed: %v", key, err)
		return nil
	}
	out := make([]ScoreMember, 0, len(res))
	for _, z := range res {
		m, _ := z.Member.(string)
		out = append(out, ScoreMember{Member: m, Score: z.Score})
	}
	return out
}

// Set writes a string value with no expiry, mirroring Java redisMapper.set.
func Set(key, value string) {
	if Client == nil {
		return
	}
	if err := Client.Set(bgCtx, key, value, 0).Err(); err != nil {
		log.Printf("[redis] SET %s failed: %v", key, err)
	}
}

// SetEx writes a string value with a TTL, mirroring Java redisMapper.set(k,v,ttl,unit).
func SetEx(key, value string, ttl time.Duration) {
	if Client == nil {
		return
	}
	if err := Client.Set(bgCtx, key, value, ttl).Err(); err != nil {
		log.Printf("[redis] SETEX %s failed: %v", key, err)
	}
}

// Del deletes a key, mirroring Java redisMapper.delete.
func Del(key string) {
	if Client == nil {
		return
	}
	if err := Client.Del(bgCtx, key).Err(); err != nil {
		log.Printf("[redis] DEL %s failed: %v", key, err)
	}
}

// SetRange overwrites part of the string at key starting at offset, mirroring
// Java redisMapper.setRange(key, value, offset) (Redis SETRANGE). Used for the
// fixed-length device-info partial updates.
func SetRange(key, value string, offset int) {
	if Client == nil {
		return
	}
	if err := Client.SetRange(bgCtx, key, int64(offset), value).Err(); err != nil {
		log.Printf("[redis] SETRANGE %s@%d failed: %v", key, offset, err)
	}
}

// ZAdd adds/updates a single member's score in a zset, mirroring Java
// redisMapper.zAdd(key, member, score).
func ZAdd(key, member string, score float64) {
	if Client == nil {
		return
	}
	if err := Client.ZAdd(bgCtx, key, goredis.Z{Score: score, Member: member}).Err(); err != nil {
		log.Printf("[redis] ZADD %s failed: %v", key, err)
	}
}

// SetNX sets key=value with a TTL only if absent, mirroring Java redisMapper.setNX.
// Returns true when the key was set (lock acquired).
func SetNX(key, value string, ttl time.Duration) bool {
	if Client == nil {
		return false
	}
	ok, err := Client.SetNX(bgCtx, key, value, ttl).Result()
	if err != nil {
		log.Printf("[redis] SETNX %s failed: %v", key, err)
		return false
	}
	return ok
}

// ZRem removes members from a zset, mirroring Java redisMapper.zRem.
func ZRem(key string, members ...string) {
	if Client == nil || len(members) == 0 {
		return
	}
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}
	if err := Client.ZRem(bgCtx, key, args...).Err(); err != nil {
		log.Printf("[redis] ZRem %s failed: %v", key, err)
	}
}

// ZRange returns members in [start, stop], mirroring Java redisMapper.zRange.
func ZRange(key string, start, stop int64) []string {
	if Client == nil {
		return nil
	}
	res, err := Client.ZRange(bgCtx, key, start, stop).Result()
	if err != nil {
		log.Printf("[redis] ZRange %s failed: %v", key, err)
		return nil
	}
	return res
}

// GeoRadius returns member->distance(meters) within radius of (lng,lat),
// mirroring Java redisMapper.georadius(key, lng, lat, radius, METERS, MAX).
func GeoRadius(key string, lng, lat, radius float64) map[string]float64 {
	out := map[string]float64{}
	if Client == nil {
		return out
	}
	res, err := Client.GeoRadius(bgCtx, key, lng, lat, &goredis.GeoRadiusQuery{
		Radius:   radius,
		Unit:     "m",
		WithDist: true,
		Count:    0,
	}).Result()
	if err != nil {
		log.Printf("[redis] GeoRadius %s failed: %v", key, err)
		return out
	}
	for _, loc := range res {
		if _, exists := out[loc.Name]; !exists {
			out[loc.Name] = loc.Dist
		}
	}
	return out
}

// Exists reports whether key exists, mirroring Java redisMapper.exists.
func Exists(key string) bool {
	if Client == nil {
		return false
	}
	n, err := Client.Exists(bgCtx, key).Result()
	if err != nil {
		log.Printf("[redis] EXISTS %s failed: %v", key, err)
		return false
	}
	return n > 0
}

// HGetAll returns all field->value pairs of the hash at key, mirroring Java
// redisMapper.hGetAll. Empty map when missing / on error.
func HGetAll(key string) map[string]string {
	if Client == nil {
		return map[string]string{}
	}
	m, err := Client.HGetAll(bgCtx, key).Result()
	if err != nil {
		log.Printf("[redis] HGETALL %s failed: %v", key, err)
		return map[string]string{}
	}
	return m
}

// HIncrBy increments hash field by delta, mirroring Java redisMapper.hIncrement.
func HIncrBy(key, field string, delta int64) {
	if Client == nil {
		return
	}
	if err := Client.HIncrBy(bgCtx, key, field, delta).Err(); err != nil {
		log.Printf("[redis] HINCRBY %s.%s failed: %v", key, field, err)
	}
}

// Expire sets a TTL on key, mirroring Java redisMapper.expire(key, ttl, unit).
func Expire(key string, ttl time.Duration) {
	if Client == nil {
		return
	}
	if err := Client.Expire(bgCtx, key, ttl).Err(); err != nil {
		log.Printf("[redis] EXPIRE %s failed: %v", key, err)
	}
}

// HDel removes fields from the hash at key, mirroring Java redisMapper.hDel.
func HDel(key string, fields ...string) {
	if Client == nil || len(fields) == 0 {
		return
	}
	if err := Client.HDel(bgCtx, key, fields...).Err(); err != nil {
		log.Printf("[redis] HDEL %s failed: %v", key, err)
	}
}

// Close releases the client.
func Close() error {
	if Client == nil {
		return nil
	}
	return Client.Close()
}

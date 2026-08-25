// Package redis wraps the go-redis client. Two things live here: the per-agent
// rate-limit counters, and the callback subscription set (agentId + event -> URLs)
// which is the only mutable state this service owns.
//
// Keys are used verbatim (no global prefix), matching the Java
// DevicePassRedisKey.format convention of the rest of the platform.
package redis

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"time"

	"ebike-open-paas-go/internal/pkg/config"

	goredis "github.com/redis/go-redis/v9"
)

var (
	// Client is the shared Redis client (device-paas shadow / callbacks). Nil
	// only when no host is configured.
	Client *goredis.Client
	// RegistryClient reads anvelink device_ebike_* hashes. Nil means fall back
	// to Client (same DB as the shadow).
	RegistryClient *goredis.Client
	bgCtx          = context.Background()
	// reachable tracks whether the last command round-tripped, so the health
	// probe and the rate limiter can tell a live client from a stranded one.
	reachable atomic.Bool
)

// InitRedis connects to Redis using the merged config.
//
// A failed initial handshake is not fatal and does not discard the client:
// go-redis dials lazily per command and recovers on its own once Redis is back.
// Throwing the client away here would strand the process — every later command
// would short-circuit on a nil client until someone restarted the pod, even
// though the outage had ended.
func InitRedis() {
	addr := config.RedisAddr()
	rc := config.GlobalConfig().Redis
	if rc.Host == "" {
		log.Println("[redis] host empty, skipping Redis initialization")
		return
	}
	Client = goredis.NewClient(&goredis.Options{
		Addr:         addr,
		Password:     rc.Password,
		DB:           rc.Database,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
		MinIdleConns: 5,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Client.Ping(ctx).Err(); err != nil {
		log.Printf("[redis] initial ping failed (addr=%s db=%d): %v — retrying lazily per command",
			addr, rc.Database, err)
	} else {
		reachable.Store(true)
		log.Printf("[redis] connected (addr=%s db=%d)", addr, rc.Database)
	}

	regDB := rc.RegistryDatabase
	if regDB >= 0 && regDB != rc.Database {
		RegistryClient = goredis.NewClient(&goredis.Options{
			Addr:         addr,
			Password:     rc.Password,
			DB:           regDB,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     10,
			MinIdleConns: 2,
		})
		regCtx, regCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer regCancel()
		if err := RegistryClient.Ping(regCtx).Err(); err != nil {
			log.Printf("[redis] registry ping failed (addr=%s db=%d): %v — retrying lazily per command",
				addr, regDB, err)
		} else {
			log.Printf("[redis] registry connected (addr=%s db=%d) for device_ebike_*", addr, regDB)
		}
	} else {
		RegistryClient = nil
		log.Printf("[redis] registry uses primary db=%d for device_ebike_*", rc.Database)
	}
}

// Available reports whether a client exists at all.
func Available() bool { return Client != nil }

// Healthy reports whether the last command reached Redis. The readiness probe
// uses it so a pod that cannot read the device shadow stops taking traffic.
func Healthy() bool { return Client != nil && reachable.Load() }

// Ping re-checks reachability and refreshes what Healthy reports.
func Ping() error {
	if Client == nil {
		return ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(bgCtx, 2*time.Second)
	defer cancel()
	err := Client.Ping(ctx).Err()
	observe(err)
	return err
}

// observe records whether a command reached Redis. Only transport-level
// failures flip the flag; a miss or a WRONGTYPE means Redis answered.
func observe(err error) {
	if err == nil || errors.Is(err, goredis.Nil) {
		reachable.Store(true)
		return
	}
	reachable.Store(false)
}

// Ctx returns the background context for Redis calls.
func Ctx() context.Context { return bgCtx }

// Get returns the string value at key, or "" if missing / on error.
func Get(key string) string {
	if Client == nil {
		return ""
	}
	v, err := Client.Get(bgCtx, key).Result()
	observe(err)
	if err != nil {
		return ""
	}
	return v
}

// HGet returns one hash field, or "" if missing / on error.
func HGet(key, field string) string {
	if Client == nil {
		return ""
	}
	v, err := Client.HGet(bgCtx, key, field).Result()
	observe(err)
	if err != nil {
		return ""
	}
	return v
}

// RegistryHGet reads a hash field from the IOT registry Redis (device_ebike_*).
// Falls back to the primary Client when no separate registry DB is configured.
func RegistryHGet(key, field string) string {
	c := RegistryClient
	if c == nil {
		c = Client
	}
	if c == nil {
		return ""
	}
	v, err := c.HGet(bgCtx, key, field).Result()
	observe(err)
	if err != nil {
		return ""
	}
	return v
}

// Exists reports whether key exists.
func Exists(key string) bool {
	if Client == nil {
		return false
	}
	n, err := Client.Exists(bgCtx, key).Result()
	observe(err)
	if err != nil {
		log.Printf("[redis] EXISTS %s failed: %v", key, err)
		return false
	}
	return n > 0
}

// SetEx writes a string value with a TTL.
func SetEx(key, value string, ttl time.Duration) {
	if Client == nil {
		return
	}
	err := Client.Set(bgCtx, key, value, ttl).Err()
	observe(err)
	if err != nil {
		log.Printf("[redis] SETEX %s failed: %v", key, err)
	}
}

// HGetAll returns all field->value pairs of the hash at key.
//
// The error is returned rather than folded into an empty map because callers
// that derive state transitions must not read a failed read as "no previous
// state" — that would re-fire an edge on every Redis blip.
func HGetAll(key string) (map[string]string, error) {
	if Client == nil {
		return nil, ErrUnavailable
	}
	m, err := Client.HGetAll(bgCtx, key).Result()
	observe(err)
	if err != nil {
		log.Printf("[redis] HGETALL %s failed: %v", key, err)
		return nil, err
	}
	return m, nil
}

// Del removes keys.
func Del(keys ...string) {
	if Client == nil || len(keys) == 0 {
		return
	}
	err := Client.Del(bgCtx, keys...).Err()
	observe(err)
	if err != nil {
		log.Printf("[redis] DEL %v failed: %v", keys, err)
	}
}

// HSetTTL writes field->value pairs into the hash at key and refreshes its TTL
// in one round trip. The TTL is reapplied on every write so a device that keeps
// reporting keeps its state, while a retired one expires on its own.
func HSetTTL(key string, ttl time.Duration, values ...interface{}) {
	if Client == nil || len(values) == 0 {
		return
	}
	pipe := Client.Pipeline()
	pipe.HSet(bgCtx, key, values...)
	if ttl > 0 {
		pipe.Expire(bgCtx, key, ttl)
	}
	_, err := pipe.Exec(bgCtx)
	observe(err)
	if err != nil {
		log.Printf("[redis] HSET %s failed: %v", key, err)
	}
}

// HDel removes fields from the hash at key.
func HDel(key string, fields ...string) {
	if Client == nil || len(fields) == 0 {
		return
	}
	err := Client.HDel(bgCtx, key, fields...).Err()
	observe(err)
	if err != nil {
		log.Printf("[redis] HDEL %s failed: %v", key, err)
	}
}

// Expire sets a TTL on key.
func Expire(key string, ttl time.Duration) {
	if Client == nil {
		return
	}
	err := Client.Expire(bgCtx, key, ttl).Err()
	observe(err)
	if err != nil {
		log.Printf("[redis] EXPIRE %s failed: %v", key, err)
	}
}

// slidingWindowScript counts the requests already recorded inside the window,
// and only then records this one. Doing it in one script is what makes the limit
// real: the read-then-write it replaces let every replica (and every concurrent
// request on one replica) read the same pre-increment count, so an agent could
// burst well past its quota before any counter caught up.
//
// KEYS[1]   hash of minute-bucket -> count
// ARGV[1]   current bucket, as yyyyMMddHHmm
// ARGV[2]   oldest bucket still inside the window
// ARGV[3]   max requests allowed in the window
// ARGV[4]   key TTL in seconds
// returns   {allowed, countInWindow}
var slidingWindowScript = goredis.NewScript(`
local cur = ARGV[1]
local oldest = ARGV[2]
local limit = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])
local total = 0
local stale = {}
local buckets = redis.call('HGETALL', KEYS[1])
for i = 1, #buckets, 2 do
  local field = buckets[i]
  if string.len(field) == string.len(cur) then
    if field < oldest then
      table.insert(stale, field)
    elseif field <= cur then
      total = total + tonumber(buckets[i + 1])
    end
  end
end
if #stale > 0 then
  redis.call('HDEL', KEYS[1], unpack(stale))
end
if total >= limit then
  return {0, total}
end
redis.call('HINCRBY', KEYS[1], cur, 1)
redis.call('EXPIRE', KEYS[1], ttl)
return {1, total + 1}
`)

// SlidingWindowAllow atomically evaluates and records one request against the
// window. The returned count is the number of requests inside the window
// including this one when allowed.
func SlidingWindowAllow(key, currentBucket, oldestBucket string, limit int, ttl time.Duration) (bool, int64, error) {
	if Client == nil {
		return false, 0, ErrUnavailable
	}
	res, err := slidingWindowScript.Run(bgCtx, Client, []string{key},
		currentBucket, oldestBucket, limit, int64(ttl.Seconds())).Slice()
	observe(err)
	if err != nil {
		return false, 0, err
	}
	if len(res) != 2 {
		return false, 0, errors.New("rate limit script returned an unexpected shape")
	}
	allowed, _ := res[0].(int64)
	count, _ := res[1].(int64)
	return allowed == 1, count, nil
}

// SAdd adds members to the set at key and returns how many were newly added.
// Xiaoan's register-callback response is exactly that count.
func SAdd(key string, members ...string) (int64, error) {
	if Client == nil {
		return 0, ErrUnavailable
	}
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}
	n, err := Client.SAdd(bgCtx, key, args...).Result()
	observe(err)
	if err != nil {
		log.Printf("[redis] SADD %s failed: %v", key, err)
		return 0, err
	}
	return n, nil
}

// SRem removes members from the set at key and returns how many were removed.
func SRem(key string, members ...string) (int64, error) {
	if Client == nil {
		return 0, ErrUnavailable
	}
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}
	n, err := Client.SRem(bgCtx, key, args...).Result()
	observe(err)
	if err != nil {
		log.Printf("[redis] SREM %s failed: %v", key, err)
		return 0, err
	}
	return n, nil
}

// SMembers returns the members of the set at key.
func SMembers(key string) ([]string, error) {
	if Client == nil {
		return nil, ErrUnavailable
	}
	v, err := Client.SMembers(bgCtx, key).Result()
	observe(err)
	if err != nil {
		log.Printf("[redis] SMEMBERS %s failed: %v", key, err)
		return nil, err
	}
	return v, nil
}

// Keys returns keys matching pattern. Only used on the small callback-subscription
// keyspace (a handful of agents times five event types), never on device keys.
func Keys(pattern string) ([]string, error) {
	if Client == nil {
		return nil, ErrUnavailable
	}
	var out []string
	iter := Client.Scan(bgCtx, 0, pattern, 200).Iterator()
	for iter.Next(bgCtx) {
		out = append(out, iter.Val())
	}
	err := iter.Err()
	observe(err)
	if err != nil {
		log.Printf("[redis] SCAN %s failed: %v", pattern, err)
		return nil, err
	}
	return out, nil
}

// Close releases the client.
func Close() error {
	var first error
	if RegistryClient != nil && RegistryClient != Client {
		if err := RegistryClient.Close(); err != nil && first == nil {
			first = err
		}
		RegistryClient = nil
	}
	if Client != nil {
		if err := Client.Close(); err != nil && first == nil {
			first = err
		}
		Client = nil
	}
	return first
}

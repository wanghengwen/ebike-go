package redis

import (
	"errors"

	"ebike-open-paas-go/internal/pkg/config"

	goredis "github.com/redis/go-redis/v9"
)

// RegistryLookupResult is the outcome of reading device_ebike_* from the registry DB.
type RegistryLookupResult struct {
	Value        string
	KeyExists    bool
	FieldExists  bool
	Err          error
}

// registryClient returns the client used for device_ebike_* (separate DB or primary).
func registryClient() *goredis.Client {
	if RegistryClient != nil {
		return RegistryClient
	}
	return Client
}

// RegistryDB returns the Redis DB index used for device_ebike_* lookups.
func RegistryDB() int {
	rc := config.GlobalConfig().Redis
	if rc.RegistryDatabase >= 0 {
		return rc.RegistryDatabase
	}
	return rc.Database
}

// ShadowDB returns the Redis DB index used for device shadow / callbacks.
func ShadowDB() int {
	return config.GlobalConfig().Redis.Database
}

// RegistryAddr returns host:port for registry reads (same host as primary).
func RegistryAddr() string {
	return config.RedisAddr()
}

// UsesSeparateRegistryDB reports whether registry uses a dedicated client/DB.
func UsesSeparateRegistryDB() bool {
	return RegistryClient != nil
}

// RegistryLookup reads one hash field from the registry Redis and reports existence.
func RegistryLookup(key, field string) RegistryLookupResult {
	c := registryClient()
	if c == nil {
		return RegistryLookupResult{Err: ErrUnavailable}
	}
	exists, err := c.Exists(bgCtx, key).Result()
	if err != nil {
		observe(err)
		return RegistryLookupResult{Err: err}
	}
	if exists == 0 {
		observe(nil)
		return RegistryLookupResult{KeyExists: false}
	}
	val, err := c.HGet(bgCtx, key, field).Result()
	if errors.Is(err, goredis.Nil) {
		observe(nil)
		return RegistryLookupResult{KeyExists: true, FieldExists: false}
	}
	observe(err)
	if err != nil {
		return RegistryLookupResult{KeyExists: true, Err: err}
	}
	return RegistryLookupResult{
		Value:       val,
		KeyExists:   true,
		FieldExists: true,
	}
}

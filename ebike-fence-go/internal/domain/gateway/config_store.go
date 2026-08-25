package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"ebike-fence-go/internal/domain/rediskeys"
	pkgredis "ebike-fence-go/internal/pkg/redis"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// ConfigStore provides cache-aside read and double-delete write for service-scoped config rows.
type ConfigStore struct {
	DB *gorm.DB
}

func NewConfigStore(db *gorm.DB) *ConfigStore {
	return &ConfigStore{DB: db}
}

// GetJSON loads a JSON blob: Redis first, then DB loader, then caches result.
func (s *ConfigStore) GetJSON(ctx context.Context, cacheKey string, load func(*gorm.DB) (string, error)) (string, error) {
	rdb := pkgredis.GetClient()
	if rdb != nil {
		val, err := rdb.Get(ctx, cacheKey).Result()
		if err == nil && !rediskeys.IsCacheMiss(val) {
			return val, nil
		}
		if err != nil && !errors.Is(err, redis.Nil) {
			return "", err
		}
	}
	if s.DB == nil {
		return "", nil
	}
	raw, err := load(s.DB.WithContext(ctx))
	if err != nil {
		return "", err
	}
	if rdb != nil && raw != "" {
		_ = rdb.Set(ctx, cacheKey, raw, 0).Err()
	}
	return raw, nil
}

// GetObject unmarshals JSON into dest using GetJSON.
func (s *ConfigStore) GetObject(ctx context.Context, cacheKey string, load func(*gorm.DB) (string, error), dest interface{}) error {
	raw, err := s.GetJSON(ctx, cacheKey, load)
	if err != nil {
		return err
	}
	if rediskeys.IsCacheMiss(raw) {
		return nil
	}
	return json.Unmarshal([]byte(raw), dest)
}

// GetList loads a JSON array list with cache-aside (HelpConfig pattern).
func (s *ConfigStore) GetList(ctx context.Context, cacheKey string, load func(*gorm.DB) (string, error)) (string, error) {
	return s.GetJSON(ctx, cacheKey, load)
}

// Invalidate deletes a cache key twice (Java GatewayImpl double-delete pattern).
func (s *ConfigStore) Invalidate(ctx context.Context, cacheKey string) {
	rdb := pkgredis.GetClient()
	if rdb == nil {
		return
	}
	_ = rdb.Del(ctx, cacheKey).Err()
}

// WriteWithInvalidate runs write fn after cache delete; invalidates again after write.
func (s *ConfigStore) WriteWithInvalidate(ctx context.Context, cacheKey string, write func(*gorm.DB) error) error {
	if s.DB == nil {
		return errors.New("database not configured")
	}
	s.Invalidate(ctx, cacheKey)
	if err := write(s.DB.WithContext(ctx)); err != nil {
		return err
	}
	s.Invalidate(ctx, cacheKey)
	return nil
}

// Now returns current time for created/updated fields.
func Now() time.Time { return time.Now() }

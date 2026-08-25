package redis

import (
	"context"
	"fmt"
	"time"

	"ebike-auth-go/internal/pkg/config"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

func InitRedis() error {
	addr := fmt.Sprintf("%s:%d", config.AppConfig.Redis.Host, config.AppConfig.Redis.Port)
	Rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.AppConfig.Redis.Password,
		DB:       config.AppConfig.Redis.Database,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := Rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("connect to redis %s error: %w", addr, err)
	}
	return nil
}

func Close() error {
	if Rdb != nil {
		return Rdb.Close()
	}
	return nil
}

func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return Rdb.Set(ctx, key, value, expiration).Err()
}

func Get(ctx context.Context, key string) (string, error) {
	val, err := Rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func Delete(ctx context.Context, key string) error {
	return Rdb.Del(ctx, key).Err()
}

// Key formats
func GetLoginDeviceKey(prefix, pin string) string {
	// prefix can be platform (for business) or tenantId (for client)
	return fmt.Sprintf("login_device_%s_%s", prefix, pin)
}

func GetLoginDeviceOldKey(pin string) string {
	return fmt.Sprintf("login_device_%s", pin)
}

func GetUserLoginTagKey(tenantId, pin string) string {
	return fmt.Sprintf("user_login_tag_%s_%s", tenantId, pin)
}

func GetComponentVerifyTicketKey() string {
	return "component_verify_ticket"
}

func GetComponentAccessTokenKey() string {
	return "component_access_token"
}

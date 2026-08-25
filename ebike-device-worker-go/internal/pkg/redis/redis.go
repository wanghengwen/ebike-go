package redis

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"ebike-device-worker-go/internal/config"
	"ebike-device-worker-go/internal/domain/constants"

	goredis "github.com/redis/go-redis/v9"
)

var (
	Client *goredis.Client
	bgCtx  = context.Background()
)

func InitRedis() {
	if config.GlobalConfig.Redis.Addr == "" {
		log.Println("Warning: Redis Addr is empty, skipping Redis initialization")
		return
	}
	connectRedis()
}

// Reinit reconnects Redis after Nacos config hot reload.
func Reinit() error {
	if config.GlobalConfig.Redis.Addr == "" {
		return fmt.Errorf("redis addr is empty")
	}
	if Client != nil {
		_ = Client.Close()
		Client = nil
	}
	connectRedis()
	return nil
}

func connectRedis() {
	db := config.GlobalConfig.Redis.DB
	Client = goredis.NewClient(&goredis.Options{
		Addr:     config.GlobalConfig.Redis.Addr,
		Password: config.GlobalConfig.Redis.Password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := Client.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Printf("Successfully connected to Redis (addr=%s db=%d)", config.GlobalConfig.Redis.Addr, db)
}

func Ctx() context.Context { return bgCtx }

func GetKey(key string) string {
	prefix := config.GlobalConfig.Redis.Prefix
	if prefix != "" {
		return prefix + key
	}
	return key
}

// HSetPipelined mirrors Java RedisMapper.hSetPipeLined.
func HSetPipelined(keys []string, values []map[string]interface{}) error {
	if Client == nil || len(keys) == 0 {
		return nil
	}
	pipe := Client.Pipeline()
	for i, key := range keys {
		if i >= len(values) {
			break
		}
		fullKey := GetKey(key)
		for field, val := range values[i] {
			pipe.HSet(bgCtx, fullKey, field, toRedisValue(val))
		}
	}
	_, err := pipe.Exec(bgCtx)
	return err
}

func toRedisValue(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", t)
	}
}

// StartTenantSubscriber listens on deviceTenantIdQueue, mirroring RedisSubscriber + MessagePushServiceImpl.
func StartTenantSubscriber(onMessage func(string)) {
	if Client == nil || onMessage == nil {
		return
	}
	go func() {
		pubsub := Client.Subscribe(bgCtx, constants.RedisQueueChangeTenant)
		ch := pubsub.Channel()
		log.Printf("[redis] subscribed channel=%s", constants.RedisQueueChangeTenant)
		for msg := range ch {
			onMessage(msg.Payload)
		}
	}()
}

package redis

import (
	"context"
	"log"
	"sync"

	"ebike-analyze-go/internal/pkg/config"

	"github.com/go-redis/redis/v8"
)

var (
	mu     sync.RWMutex
	client *redis.Client
)

// GetClient returns the current Redis client (nil when not configured).
func GetClient() *redis.Client {
	mu.RLock()
	defer mu.RUnlock()
	return client
}

func InitRedis() {
	if err := connectFromConfig(); err != nil {
		log.Printf("[WARN] Failed to connect to Redis: %v. Service will start but cache queries will fail.", err)
	}
}

// Reinit closes the current client and reconnects using the latest GlobalConfig.Redis.
func Reinit() error {
	mu.Lock()
	old := client
	client = nil
	mu.Unlock()
	if old != nil {
		_ = old.Close()
	}
	return connectFromConfig()
}

// Close shuts down the Redis client.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if client != nil {
		_ = client.Close()
		client = nil
	}
}

func connectFromConfig() error {
	cfg := config.GlobalConfig.Redis
	if cfg.Addr == "" {
		log.Println("Redis address not configured, skipping redis initialization.")
		return nil
	}

	newClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	_, err := newClient.Ping(context.Background()).Result()
	if err != nil {
		_ = newClient.Close()
		return err
	}

	mu.Lock()
	client = newClient
	mu.Unlock()
	log.Println("Redis connected successfully to", cfg.Addr)
	return nil
}

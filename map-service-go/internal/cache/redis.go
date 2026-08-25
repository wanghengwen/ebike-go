package cache

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"map-service-go/internal/config"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var (
	Rdb  *redis.Client
	once sync.Once
)

func InitRedis() {
	once.Do(func() {
		if config.GlobalConfig == nil {
			logrus.Fatal("Config not initialized before Redis")
		}
		rc := config.GlobalConfig.Redis

		host := os.Getenv("REDIS_HOST")
		if host == "" {
			host = rc.Host
			// Default to localhost if it's an unresolved Nacos placeholder
			if strings.HasPrefix(host, "${") || host == "" {
				host = "127.0.0.1"
			}
		}

		portStr := os.Getenv("REDIS_PORT")
		port := rc.Port
		if portStr != "" {
			if p, err := strconv.Atoi(portStr); err == nil {
				port = p
			}
		} else if port == 0 {
			port = 6379
		}

		password := os.Getenv("REDIS_PASSWORD")
		if password == "" {
			password = rc.Password
			if strings.HasPrefix(password, "${") {
				password = ""
			}
		}

		dbStr := os.Getenv("REDIS_DATABASE")
		db := rc.Database
		if dbStr != "" {
			if d, err := strconv.Atoi(dbStr); err == nil {
				db = d
			}
		}

		addr := fmt.Sprintf("%s:%d", host, port)
		logrus.Infof("Initializing Redis client with address: %s", addr)

		Rdb = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := Rdb.Ping(ctx).Err(); err != nil {
			logrus.Warnf("Failed to ping Redis, caching may not work: %v", err)
		} else {
			logrus.Info("Redis connected successfully")
		}
	})
}

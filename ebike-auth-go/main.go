package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ebike-auth-go/internal/auth"
	"ebike-auth-go/internal/pkg/config"
	"ebike-auth-go/internal/pkg/logger"
	"ebike-auth-go/internal/pkg/redis"
	"ebike-auth-go/internal/pkg/rpc"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const AppVersion = "v1.0.7"

func main() {
	// 1. Initialize Logger
	logger.InitLogger()
	defer logger.Log.Sync()

	logger.Log.Info("Starting ebike-auth-go...", zap.String("version", AppVersion))

	// 2. Load Local Config File
	if err := config.LoadConfig("conf/application.yml"); err != nil {
		logger.Log.Fatal("Failed to load local conf/application.yml config", zap.Error(err))
	}
	logger.Log.Info("Loaded local configuration", zap.String("mode", config.AppConfig.AuthMode), zap.Int("port", config.AppConfig.Port))

	// 3. Load Configurations dynamically from Nacos (redis, jwt secrets)
	if err := config.InitConfigFromNacos(); err != nil {
		logger.Log.Fatal("Failed to fetch configurations from Nacos", zap.Error(err))
	}
	logger.Log.Info("Successfully fetched configurations from Nacos", zap.String("redis_host", config.AppConfig.Redis.Host))

	// 4. Initialize Redis Client
	if err := redis.InitRedis(); err != nil {
		logger.Log.Fatal("Failed to initialize Redis pool", zap.Error(err))
	}
	logger.Log.Info("Successfully connected to Redis database", zap.Int("db", config.AppConfig.Redis.Database))

	// 5. Initialize RPC Http Client
	rpc.InitRPCClient()

	// 6. Initialize Nacos naming client and Register service instance (if enabled)
	if err := rpc.InitNacosNaming(); err != nil {
		logger.Log.Error("Failed to initialize Nacos Naming client", zap.Error(err))
	}
	rpc.LogServiceURLs()

	// 7. Load Tenant credentials and third-party login configs from ebike-management
	auth.InitTenantAuthManager()

	// 8. Load Google and Apple Public keys certificates cache
	auth.InitPublicKeys()

	// 9. Startup Gin HTTP Server
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Logger middleware for HTTP requests
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		if c.Request.URL.Path == "/actuator/health" {
			return
		}
		logger.Log.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
			zap.String("ip", c.ClientIP()),
		)
	})

	auth.RegisterRoutes(r)

	serverAddr := fmt.Sprintf(":%d", config.AppConfig.Port)
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	go func() {
		logger.Log.Info("HTTP Server listening", zap.String("addr", serverAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("HTTP Server failed to listen", zap.Error(err))
		}
	}()

	// 10. Wait for Shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down ebike-auth-go gracefully...")

	// Deregister from Nacos
	rpc.DeregisterInstance()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("HTTP Server forced shutdown", zap.Error(err))
	}

	if err := redis.Close(); err != nil {
		logger.Log.Error("Redis connection pool failed to close", zap.Error(err))
	}

	logger.Log.Info("ebike-auth-go exited clean.")
}

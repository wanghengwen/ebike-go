package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"ebike-open-paas-go/internal/api"
	"ebike-open-paas-go/internal/client"
	"ebike-open-paas-go/internal/dispatcher"
	"ebike-open-paas-go/internal/event"
	"ebike-open-paas-go/internal/pkg/config"
	"ebike-open-paas-go/internal/pkg/kafka"
	"ebike-open-paas-go/internal/pkg/logger"
	"ebike-open-paas-go/internal/pkg/nacos"
	"ebike-open-paas-go/internal/pkg/redis"
	"ebike-open-paas-go/internal/repository"

	"go.uber.org/zap"
)

func main() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logger.Init(logLevel)
	defer logger.Log.Sync()

	confPath := os.Getenv("CONFIG_PATH")
	if confPath == "" {
		confPath = "conf/application.yaml"
	}
	config.LoadLocalConfig(confPath)

	if os.Getenv("NACOS_DISABLED") == "true" {
		logger.Log.Info("NACOS_DISABLED=true, skipping Nacos; using local config + env")
		config.ApplyEnvOverrides()
	} else {
		nacos.InitConfigClient()
		nacos.InitNamingClient()
	}
	config.LogEffectiveUpstream()
	config.LogEffectiveRedis()
	client.PreloadIotPlatforms()
	client.PreloadConsoleAuth()

	redis.InitRedis()
	defer redis.Close()

	dispatcher.Start()
	// The subscription snapshot must exist before the consumer starts, or the
	// first batch is dropped as "no subscribers".
	repository.StartSubscriptionRefresh()

	consumer := kafka.NewManager(event.Handle)
	consumer.Start()
	nacos.SetKafkaChangeHandler(consumer.Restart)

	port := resolvePort(config.GlobalConfig().Server.Port)
	r := api.NewRouter()
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Log.Info("ebike-open-paas-go starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("server listen error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Log.Info("received shutdown signal", zap.String("signal", sig.String()))

	// Order matters. The consumer stops first so nothing new is committed, then
	// the delivery lanes drain what is already committed — reversing the two
	// would silently discard events Kafka considers acknowledged. HTTP is last
	// so the readiness probe can keep failing while the backlog clears.
	consumer.Stop()
	dispatcher.Shutdown(10 * time.Second)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("graceful shutdown error", zap.Error(err))
	} else {
		logger.Log.Info("shut down gracefully")
	}
}

func resolvePort(configured int) int {
	if env := os.Getenv("SERVER_PORT"); env != "" {
		if p, err := strconv.Atoi(env); err == nil && p > 0 {
			return p
		}
	}
	if configured > 0 {
		return configured
	}
	return 8080
}

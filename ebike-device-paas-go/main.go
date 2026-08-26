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

	"ebike-device-paas-go/internal/api"
	"ebike-device-paas-go/internal/client"
	"ebike-device-paas-go/internal/pkg/config"
	"ebike-device-paas-go/internal/pkg/kafka"
	"ebike-device-paas-go/internal/pkg/logger"
	"ebike-device-paas-go/internal/pkg/nacos"
	"ebike-device-paas-go/internal/pkg/redis"

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
	// Layer 1: local bootstrap config.
	config.LoadLocalConfig(confPath)

	// Layer 2 + 3: Nacos remote configs + env overrides (unless explicitly disabled).
	if os.Getenv("NACOS_DISABLED") == "true" {
		logger.Log.Info("NACOS_DISABLED=true, skipping Nacos; using local config + env")
		config.ApplyEnvOverrides()
	} else {
		nacos.InitConfigClient()
		nacos.InitNamingClient()
	}
	config.LogEffectiveUpstream()
	client.PreloadIotPlatforms()
	cfg := config.GlobalConfig

	// Connect Redis (non-fatal): required by read-only query endpoints.
	redis.InitRedis()
	defer redis.Close()

	// C34 producer (non-fatal): used by state-change / BLE report endpoints.
	kafka.Init()
	defer kafka.Close()

	if cfg.DryRun {
		logger.Log.Info("running in DRY_RUN (shadow) mode",
			zap.String("javaServiceUrl", cfg.Xyy.JavaServiceURL),
			zap.Bool("nacosRegisterEnabled", cfg.Nacos.RegisterEnabled))
	}

	port := resolvePort(cfg.Server.Port)
	r := api.NewRouter()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Log.Info("ebike-device-paas-go starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("server listen error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Log.Info("received shutdown signal", zap.String("signal", sig.String()))

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

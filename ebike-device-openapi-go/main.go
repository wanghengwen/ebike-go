package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"ebike-device-openapi-go/internal/api/controller"
	"ebike-device-openapi-go/internal/middleware"
	"ebike-device-openapi-go/internal/pkg/config"
	"ebike-device-openapi-go/internal/pkg/kafka"
	"ebike-device-openapi-go/internal/pkg/logger"
	"ebike-device-openapi-go/internal/pkg/metrics"
	"ebike-device-openapi-go/internal/pkg/mqtt"
	"ebike-device-openapi-go/internal/pkg/workerpool"
	"ebike-device-openapi-go/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func validateRuntimeMode() {
	dryRunEnv := os.Getenv("DRY_RUN")
	shadowEnabled := dryRunEnv == "" || dryRunEnv == "true"
	if shadowEnabled && os.Getenv("GIN_MODE") == "release" {
		logger.Log.Warn("DRY_RUN shadow mode is enabled in release — Redis/Kafka/ECU side effects are suppressed")
	}
}

func main() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logger.Init(logLevel)
	defer logger.Log.Sync()

	// ── 1. Load Configuration from Nacos ────────────────────────────────────
	httpPort := resolveHTTPServerPort()
	config.SetHTTPServerPort(httpPort)

	config.Init()
	cfg := config.GetConfig()

	validateRuntimeMode()

	// ── 2. Initialize Redis ──────────────────────────────────────────────────
	// Redis connection parameters are now read from Nacos/AppConfig (spring.redis.*),
	// not hardcoded. This allows K8s environment to inject the real Redis address via Nacos.
	redisAddr := fmt.Sprintf("%s:%d", cfg.Spring.Redis.Host, cfg.Spring.Redis.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     cfg.Spring.Redis.Password,
		DB:           cfg.Spring.Redis.Database,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
		MinIdleConns: 5,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Log.Warn("Redis connection failed at startup", zap.String("addr", redisAddr), zap.Error(err))
	} else {
		logger.Log.Info("Redis connected",
			zap.String("addr", redisAddr),
			zap.Int("database", cfg.Spring.Redis.Database),
		)
	}

	// Optional second Redis DB for consume/paas device cache (device_info / imei_car).
	deviceRdb := rdb
	deviceDB := cfg.Spring.Redis.DeviceDatabase
	if deviceDB >= 0 && deviceDB != cfg.Spring.Redis.Database {
		deviceRdb = redis.NewClient(&redis.Options{
			Addr:         redisAddr,
			Password:     cfg.Spring.Redis.Password,
			DB:           deviceDB,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     10,
			MinIdleConns: 2,
		})
		if err := deviceRdb.Ping(context.Background()).Err(); err != nil {
			logger.Log.Warn("Redis device-db connection failed",
				zap.String("addr", redisAddr), zap.Int("deviceDatabase", deviceDB), zap.Error(err))
		} else {
			logger.Log.Info("Redis device-db connected",
				zap.String("addr", redisAddr), zap.Int("deviceDatabase", deviceDB))
		}
	}

	// ── 3. Initialize shared HTTP client for ECU gateway calls ───────────────
	service.InitECUHTTPClient()

	// ── 4. Initialize Kafka Writer ──────────────────────────────────────────
	kafka.InitKafkaWriter()
	pusher := &kafka.Pusher{DryRun: false}

	// ── 5. Initialize Services ──────────────────────────────────────────────
	regService := &service.RegisterService{
		Rdb:       rdb,
		DeviceRdb: deviceRdb,
		Pusher:    pusher,
	}
	luopingMapping := &service.LuopingMappingService{Rdb: rdb}
	luopingSweeper := service.NewLuopingPresenceSweeper(regService, luopingMapping, rdb)
	luopingSvc := &service.LuopingService{
		Register: regService,
		Mapping:  luopingMapping,
		Sweeper:  luopingSweeper,
	}
	service.SetActionRedis(rdb)
	luopingSweeper.Start()
	defer luopingSweeper.Stop()

	// Initialize MQTT client (shared subscription for uplink + publish for downlink).
	// No-op when mqtt.enabled=false. Luoping MQTT always applies side effects.
	// Bounded worker pool + per-message timeout provide backpressure: paho hands each
	// inbound message to Submit, which blocks once the queue is full, so a slow
	// Redis/Kafka downstream cannot spawn unbounded work. Each message is processed
	// under a timeout context so a single stuck handler can't occupy a worker forever.
	msgTimeout := time.Duration(cfg.Mqtt.MessageTimeoutMs) * time.Millisecond
	if msgTimeout <= 0 {
		msgTimeout = 3 * time.Second
	}
	mqttPool := workerpool.New(cfg.Mqtt.WorkerPoolSize, cfg.Mqtt.WorkerQueueSize)
	if err := mqtt.InitClient(func(topic string, payload []byte) {
		// Copy the payload: paho may reuse/free the underlying buffer once the
		// dispatch callback returns, but the job runs asynchronously on a worker.
		buf := make([]byte, len(payload))
		copy(buf, payload)
		t := topic
		mqttPool.Submit(func() {
			ctx, cancel := context.WithTimeout(context.Background(), msgTimeout)
			defer cancel()
			if err := luopingSvc.HandleMqtt(ctx, t, buf); err != nil {
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
					logger.Log.Error("mqtt message handling timed out",
						zap.String("topic", t),
						zap.Duration("timeout", msgTimeout),
						zap.Int("payloadBytes", len(buf)),
						zap.Error(err),
					)
					return
				}
				logger.Log.Error("mqtt message handling failed",
					zap.String("topic", t),
					zap.Int("payloadBytes", len(buf)),
					zap.Error(err),
				)
			}
		})
	}); err != nil {
		logger.Log.Error("mqtt client init failed", zap.Error(err))
	}

	// ── 6. Initialize HTTP Router ───────────────────────────────────────────
	// Use gin.New() + logger middleware instead of gin.Default() for better control.
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.TraceContext())
	r.Use(middleware.RequestMetrics())

	// Kubernetes liveness & readiness probes (required by deployment.yaml)
	r.GET("/actuator/health/liveness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})
	r.GET("/actuator/health/readiness", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "DOWN",
				"redis":  err.Error(),
			})
			return
		}
		// MQTT is only required when configured; an unconfigured/disabled broker is
		// treated as healthy so this service can run without MQTT.
		if !mqtt.Healthy() {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "DOWN",
				"mqtt":   "disconnected",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})
	metrics.RegisterRoutes(r)

	// ── 7. Attach Middlewares ───────────────────────────────────────────────
	// RoutFilter reads the proxy target dynamically from Nacos config.
	// We pass a closure so it reads the latest config on each request.
	r.Use(middleware.RoutFilter(rdb, func() *middleware.RouteConfig {
		c := config.GetConfig()
		return &middleware.RouteConfig{
			Cloud:              c.Xyy.Cloud,
			RouterSwitch:       c.Xyy.Router.RouterSwitch,
			AnvelinkOpenapiUrl: c.Xyy.Router.AnvelinkOpenapiUrl,
			ExcludeUrls:        c.Xyy.Router.ExcludeUrls,
		}
	}))

	// ── 8. Register Routes ──────────────────────────────────────────────────
	controller.InitXiaoanApi(regService)
	controller.InitCmdApi(regService)
	controller.RegisterRoutes(r)
	controller.RegisterXiaoanRoutes(r)

	// ── 9. Start HTTP Server with Graceful Shutdown ─────────────────────────
	// K8s sends SIGTERM when a pod is being terminated (rolling update, scale-down).
	// We listen for it and give in-flight requests up to 15 seconds to complete,
	// then flush the Kafka writer buffer before exiting.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", httpPort),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Log.Info("Ebike Device OpenAPI (Go Edition) starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server listen error", zap.Error(err))
		}
	}()

	// Block until OS signal received
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Log.Info("received shutdown signal", zap.String("signal", sig.String()))

	// Deregister from Nacos first so we stop receiving new requests
	config.DeregisterNacos()

	// Give active HTTP requests up to 15s to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("HTTP server graceful shutdown error", zap.Error(err))
	} else {
		logger.Log.Info("HTTP server shut down gracefully")
	}

	// Stop MQTT ingestion BEFORE draining Kafka/Redis, so in-flight uplink handlers
	// finish writing while those clients are still open (avoids write-after-close).
	mqtt.Close()

	// Drain the worker pool so queued/in-flight uplink jobs finish before we close
	// Kafka/Redis below.
	mqttPool.Close()

	// Flush and close Kafka writer — ensures all buffered messages are delivered
	// before the pod exits. Critical to avoid message loss during rolling updates.
	kafka.CloseKafkaWriter()

	// Close Redis connection pool
	if err := rdb.Close(); err != nil {
		logger.Log.Error("Redis close error", zap.Error(err))
	}

	logger.Log.Info("Ebike Device OpenAPI shut down complete")
}

func resolveHTTPServerPort() int {
	if env := os.Getenv("SERVER_PORT"); env != "" {
		if port, err := strconv.Atoi(env); err == nil && port > 0 {
			return port
		}
	}
	return 8080
}

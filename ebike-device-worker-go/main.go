package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ebike-device-worker-go/internal/api/controller"
	"ebike-device-worker-go/internal/app"
	"ebike-device-worker-go/internal/config"
	nacosreg "ebike-device-worker-go/internal/infrastructure/nacos"
	"ebike-device-worker-go/internal/pkg/async"
	"ebike-device-worker-go/internal/pkg/db"
	"ebike-device-worker-go/internal/pkg/env"
	"ebike-device-worker-go/internal/pkg/fastid"
	"ebike-device-worker-go/internal/pkg/redis"

	"github.com/gin-gonic/gin"
)

const shutdownTimeout = 10 * time.Second

func main() {
	configPath := flag.String("config", config.DefaultConfigPath(), "path to application config file")
	skipNacos := flag.Bool("skip-nacos", false, "use local application.yaml only, skip Nacos extension configs")
	flag.Parse()

	config.LoadConfig(*configPath)
	if env.IsDryRun() {
		log.Println("[shadow] DRY_RUN enabled: client receives Java response; Go handler runs in background for [SHADOW MATCH/DIFF] logs")
	}
	if *skipNacos {
		log.Println("[config] skip-nacos enabled, using local application.yaml")
	} else if err := config.LoadExtensionConfigsFromNacos(); err != nil {
		log.Fatalf("load nacos extension configs: %v", err)
	}
	initFastID(config.GlobalConfig)
	db.InitDB()
	redis.InitRedis()

	rt := app.NewRuntime(config.GlobalConfig)
	rt.Start()

	if !*skipNacos {
		if err := nacosreg.InitNamingClient(); err != nil {
			log.Printf("[WARN] nacos naming client: %v", err)
		} else if err := nacosreg.RegisterInstance(); err != nil {
			log.Printf("[WARN] nacos register: %v", err)
		}
		registerNacosListeners(rt)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/actuator/health", "/actuator/memstats", "/actuator/debug/pprof/heap"},
	}))
	controller.RegisterRoutes(r)

	srv := &http.Server{
		Addr:         ":" + config.GlobalConfig.Server.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting ebike-device-worker-go on port %s...", config.GlobalConfig.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	nacosreg.DeregisterInstance()

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[WARN] http shutdown: %v", err)
	}

	if !async.Wait(5 * time.Second) {
		log.Println("[WARN] background tasks still running after timeout")
	}
	rt.Stop()
	log.Println("shutdown complete")
}

func registerNacosListeners(rt *app.Runtime) {
	client := config.NacosConfigClient()
	if client == nil {
		return
	}
	if err := config.ListenRedisFromNacos(client, func() {
		if err := redis.Reinit(); err != nil {
			log.Printf("[ERROR] redis reinit after nacos change: %v", err)
			return
		}
		log.Println("redis reconnected after nacos change")
	}); err != nil {
		log.Printf("[WARN] listen redis.yaml: %v", err)
	}
	if err := config.ListenKafkaFromNacos(client, func() {
		rt.RestartKafka()
	}); err != nil {
		log.Printf("[WARN] listen kafka.yaml: %v", err)
	}
}

func initFastID(cfg *config.Config) {
	if !cfg.FastIDEnabled() {
		log.Println("Warning: fastid disabled; persistence id generation will fail if used")
		return
	}
	fc := cfg.ResolvedFastID()
	if fc.Secret == "" {
		log.Fatalf("fastid secret is required (Nacos fast_id.yaml or SPRING_XYY_FASTID_SECRET)")
	}
	if err := fastid.Init(fc); err != nil {
		log.Fatalf("fastid init failed: %v", err)
	}
	log.Printf("fastid initialized app=%s url=%s namespace=%s group=%s", fc.AppName, fc.URL, fc.Namespace, fc.GroupID)
}

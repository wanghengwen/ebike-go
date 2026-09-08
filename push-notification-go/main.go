package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"push-notification-go/internal/cache"
	"push-notification-go/internal/handler"
	"push-notification-go/internal/pkg/config"
	"push-notification-go/internal/pkg/mysql"
	"push-notification-go/internal/pkg/nacos"
	"push-notification-go/internal/pkg/validator"
	"push-notification-go/internal/pool"
	"push-notification-go/internal/sender"

	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("========================================")
	log.Println("  push-notification-go starting...")
	log.Println("========================================")

	// 1. Load local config
	config.LoadLocalConfig("conf/application.yml")
	validator.Register()

	// 2. Initialize Nacos config client (fetches remote configs, merges, applies env)
	nacos.InitConfigClient()

	// 3. Initialize MySQL
	mysql.Init()

	// 4. Initialize caches
	cache.Init(mysql.DB)

	// 5. Start periodic cache refresh
	cache.StartRefreshLoop(1 * time.Minute)

	// 6. Create worker pools from config
	msgPoolCfg := config.GlobalConfig.Xyy.MessagePool
	if msgPoolCfg.MaxWorkers <= 0 {
		msgPoolCfg.MaxWorkers = 10
	}
	if msgPoolCfg.QueueSize <= 0 {
		msgPoolCfg.QueueSize = 200
	}
	msgPool := pool.NewWorkerPool(msgPoolCfg.MaxWorkers, msgPoolCfg.QueueSize)

	voicePoolCfg := config.GlobalConfig.Xyy.VoicePool
	if voicePoolCfg.MaxWorkers <= 0 {
		voicePoolCfg.MaxWorkers = 10
	}
	if voicePoolCfg.QueueSize <= 0 {
		voicePoolCfg.QueueSize = 200
	}
	voicePool := pool.NewWorkerPool(voicePoolCfg.MaxWorkers, voicePoolCfg.QueueSize)

	// 7. Register sender handlers
	sender.Register()

	// 8. Initialize Nacos naming client (service registration)
	nacos.InitNamingClient()

	// 9. Setup Gin engine
	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/actuator/health"},
	}))
	r.Use(handler.RecoveryMiddleware())
	handler.RegisterRoutes(r, mysql.DB, msgPool, voicePool)

	// 10. Start HTTP server
	port := config.GlobalConfig.Server.Port
	if port == 0 {
		port = 8080
	}
	addr := fmt.Sprintf(":%d", port)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		log.Printf("========================================")
		log.Printf("  push-notification-go started on %s", addr)
		log.Printf("========================================")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[main] server listen error: %v", err)
		}
	}()

	// 11. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("[main] received signal %v, shutting down...", sig)

	// Shutdown worker pools
	msgPool.Shutdown()
	voicePool.Shutdown()

	// Deregister from Nacos
	nacos.Deregister()

	// Shutdown HTTP server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[main] server shutdown error: %v", err)
	}

	log.Println("[main] push-notification-go stopped")
}

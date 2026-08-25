package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"identity-auth-go/internal/handler"
	"identity-auth-go/internal/model"
	"identity-auth-go/internal/pkg/cache"
	"identity-auth-go/internal/pkg/config"
	"identity-auth-go/internal/pkg/mysql"
	"identity-auth-go/internal/pkg/nacos"
	ossutil "identity-auth-go/internal/pkg/oss"
	goredis "identity-auth-go/internal/pkg/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("========================================")
	log.Println("  identity-auth-go starting...")
	log.Println("========================================")

	// 1. Load local config
	config.LoadLocalConfig("conf/application.yml")

	// 2. Initialize Nacos config client (fetches mysql.yaml, redis.yaml, identity-auth.yml)
	nacos.InitConfigClient()

	// 3. Initialize MySQL
	mysql.Init()

	// 3.1 Initialize local cache (60s refresh, matching Java LocalCache)
	cache.Init(mysql.DB)

	// Auto-migrate the shadow record list (only needed for shadow mode)
	if config.GlobalConfig.DryRun {
		err := mysql.DB.AutoMigrate(&model.ShadowRecordList{})
		if err != nil {
			log.Printf("[main] warning: failed to auto-migrate shadow record list: %v", err)
		}
	}

	// 4. Initialize Redis
	goredis.Init()

	// 5. Initialize OSS (for face image storage)
	ossutil.Init()

	// 6. Initialize Nacos naming client (service registration)
	nacos.InitNamingClient()

	// Log run mode
	if config.GlobalConfig.DryRun {
		log.Println("========================================")
		log.Println("  Running in DRY_RUN (shadow) mode")
		log.Println("  /auth and /charge will NOT execute")
		log.Println("========================================")
	} else {
		log.Println("========================================")
		log.Println("  Running in PRODUCTION mode")
		log.Println("========================================")
	}

	// 7. Setup Gin engine
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Skip: func(c *gin.Context) bool {
			p := c.Request.URL.Path
			if p == "/actuator" || strings.HasPrefix(p, "/actuator/") {
				return true
			}
			// Ignore external scanner noise (GET /, favicon, etc.)
			if c.Request.Method == http.MethodGet && (p == "/" || p == "/favicon.ico") {
				return true
			}
			return false
		},
	}))
	r.Use(gin.Recovery())

	handler.RegisterRoutes(r, mysql.DB)

	// 8. Start HTTP server
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
		log.Printf("  identity-auth-go started on %s", addr)
		log.Printf("========================================")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[main] server listen error: %v", err)
		}
	}()

	// 9. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("[main] received signal %v, shutting down...", sig)

	// Deregister from Nacos
	nacos.Deregister()

	// Shutdown HTTP server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[main] server shutdown error: %v", err)
	}

	log.Println("[main] identity-auth-go stopped")
}

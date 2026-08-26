package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"luopingtech-fastid-register/internal/api/controller"
	"luopingtech-fastid-register/internal/model"
	"luopingtech-fastid-register/internal/pkg/config"
	"luopingtech-fastid-register/internal/repository"
	"luopingtech-fastid-register/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	startTime := time.Now()

	// ---- Load Configuration ----
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// ---- Initialize Database ----
	dsn := cfg.Database.DSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(3 * time.Minute)

	// Auto-migrate is opt-in only (DB_AUTO_MIGRATE=true). Production DBs are managed
	// externally like the Java version; enabling migrate on an existing schema can fail.
	if os.Getenv("DB_AUTO_MIGRATE") == "true" {
		if err := db.AutoMigrate(&model.App{}, &model.Machine{}); err != nil {
			log.Fatalf("Failed to auto-migrate: %v", err)
		}
		log.Println("[INFO] DB auto-migrate completed")
	} else {
		log.Println("[INFO] DB auto-migrate is disabled (set DB_AUTO_MIGRATE=true to enable)")
	}

	// ---- Build Service Layer ----
	repo := repository.NewFastIdRepo(db)
	svc := service.NewFastIdService(repo)

	// ---- Setup HTTP Server ----
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/actuator/health"},
	}), gin.Recovery())

	h := controller.NewFastIdHandler(svc, cfg)
	h.RegisterRoutes(r)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// ---- Nacos Service Registration ----
	namingClient, err := config.RegisterService(cfg)
	if err != nil {
		log.Printf("[WARN] Nacos service registration failed: %v", err)
	}

	// ---- Start Server ----
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	elapsed := time.Since(startTime).Seconds()
	cpuNum := runtime.NumCPU()
	log.Printf("====================STARTING SUCCESS appName=%s port=%d cpuNum=%d time=%.2fseconds ====================",
		cfg.Nacos.AppName, cfg.Server.Port, cpuNum, elapsed)

	// ---- Graceful Shutdown ----
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Deregister from Nacos
	config.DeregisterService(namingClient, cfg)

	// Shutdown HTTP server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close database
	if err := sqlDB.Close(); err != nil {
		log.Printf("Database close error: %v", err)
	}

	log.Println("Server exited")
}

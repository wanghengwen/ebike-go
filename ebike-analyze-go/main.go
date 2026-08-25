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

	"ebike-analyze-go/internal/api/controller"
	"ebike-analyze-go/internal/infrastructure/es"
	"ebike-analyze-go/internal/infrastructure/persistence"
	"ebike-analyze-go/internal/middleware"
	"ebike-analyze-go/internal/pkg/config"
	"ebike-analyze-go/internal/pkg/mysql"
	pkgredis "ebike-analyze-go/internal/pkg/redis"
	pkgrpc "ebike-analyze-go/internal/pkg/rpc"

	"github.com/gin-gonic/gin"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func main() {
	// 1. Load local configuration
	config.LoadConfig("conf/application.yml")

	// 2. Initialize Nacos (service registration + remote config)
	pkgrpc.InitNacos()
	pkgrpc.InitRestClient()

	// 3. Initialize Redis
	pkgredis.InitRedis()

	// 4. Initialize MySQL (dual datasource: analyze + visual)
	mysql.Init()

	// 5. Initialize Elasticsearch
	esCfg, err := es.Init(es.Config{
		Hostname:           config.GlobalConfig.ES.Hostname,
		Username:           config.GlobalConfig.ES.Username,
		Password:           config.GlobalConfig.ES.Password,
		ConnectTimeout:     config.GlobalConfig.ES.ConnectTimeout,
		SocketTimeout:      config.GlobalConfig.ES.SocketTimeout,
		TrackTotalHitsUpTo: config.GlobalConfig.ES.TrackTotalHitsUpTo,
	})
	if err != nil {
		log.Printf("[WARN] ES initialization failed: %v. ES queries will not work.", err)
	} else {
		// Write applied defaults back so SearchCount/SearchAll read the same values.
		config.GlobalConfig.ES.ConnectTimeout = esCfg.ConnectTimeout
		config.GlobalConfig.ES.SocketTimeout = esCfg.SocketTimeout
		config.GlobalConfig.ES.TrackTotalHitsUpTo = esCfg.TrackTotalHitsUpTo
	}

	// 6. Initialize Persistence layer (GORM repositories)
	persistence.Init()

	// 7. Create Gin engine + register middleware + register routes
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.AccessLog())
	r.Use(middleware.GlobalErrorMiddleware())
	r.Use(middleware.CommandContextMiddleware())

	root := r.Group("")
	controller.RegisterRoutes(root)

	// 8. HTTP Server with graceful shutdown
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", config.GlobalConfig.Server.Port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("ebike-analyze-go started on :%d", config.GlobalConfig.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// Nacos deregistration on shutdown
	if pkgrpc.RegisteredToNacos && pkgrpc.NamingClient != nil {
		localIP := pkgrpc.GetLocalIP()
		_, _ = pkgrpc.NamingClient.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          localIP,
			Port:        uint64(config.GlobalConfig.Server.Port),
			ServiceName: config.GlobalConfig.Server.Name,
			GroupName:   config.GlobalConfig.Nacos.Group,
			Ephemeral:   true,
		})
	}

	pkgredis.Close()
	mysql.Close()
	log.Println("ebike-analyze-go exited gracefully")
}

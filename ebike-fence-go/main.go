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

	"ebike-fence-go/internal/api/controller"
	"ebike-fence-go/internal/infrastructure/mq"
	"ebike-fence-go/internal/infrastructure/persistence"
	"ebike-fence-go/internal/middleware"
	"ebike-fence-go/internal/pkg/config"
	"ebike-fence-go/internal/pkg/mysql"
	pkgredis "ebike-fence-go/internal/pkg/redis"
	pkgrpc "ebike-fence-go/internal/pkg/rpc"

	"github.com/gin-gonic/gin"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func main() {
	config.LoadConfig("conf/application.yml")

	pkgrpc.InitNacos()
	config.InitFastID()
	pkgredis.InitRedis()
	mysql.Init()
	persistence.Init()
	persistence.InitConfigGatewayHelpers()
	pkgrpc.InitRestClient()
	persistence.RegisterFenceMissFetcher()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.AccessLog())
	r.Use(middleware.GlobalErrorMiddleware())
	r.Use(middleware.CommandContextMiddleware())

	root := r.Group("")
	controller.RegisterRoutes(root)
	r.NoRoute(middleware.ProxyGateway())

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", config.GlobalConfig.Server.Port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("Server is starting on port %d...", config.GlobalConfig.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	if pkgrpc.RegisteredToNacos && pkgrpc.NamingClient != nil {
		pkgrpc.NamingClient.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          pkgrpc.GetLocalIP(),
			Port:        uint64(config.GlobalConfig.Server.Port),
			ServiceName: config.GlobalConfig.Server.Name,
			GroupName:   config.GlobalConfig.Nacos.Group,
			Ephemeral:   true,
		})
	}

	pkgredis.Close()
	_ = mq.Close()
	log.Println("Redis client closed.")

	log.Println("Server exiting")
}

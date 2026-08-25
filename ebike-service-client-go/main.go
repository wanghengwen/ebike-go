package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/modules/clientconfig"
	"ebike-service-client-go/internal/modules/fence"
	"ebike-service-client-go/internal/modules/management"
	"ebike-service-client-go/internal/modules/marketing"
	"ebike-service-client-go/internal/modules/misc"
	"ebike-service-client-go/internal/modules/operation"
	"ebike-service-client-go/internal/modules/order"
	"ebike-service-client-go/internal/modules/pay"
	"ebike-service-client-go/internal/modules/user"
	"ebike-service-client-go/internal/pkg/buildinfo"
	"ebike-service-client-go/internal/pkg/config"
	"ebike-service-client-go/internal/pkg/env"
	"ebike-service-client-go/internal/pkg/rpc"
	"github.com/gin-gonic/gin"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func main() {
	showVersion := flag.Bool("version", false, "print build version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Fprintf(os.Stderr, "ebike-service-client-go %s\n", buildinfo.Version)
		os.Exit(0)
	}

	config.LoadConfig("conf/application.yml")
	if env.IsDryRun() {
		log.Println("[shadow] DRY_RUN enabled: non-live paths return Java response; Go runs SHADOW compare in logs")
	}
	if config.GlobalConfig.Proxy.TargetURL != "" {
		log.Printf("[proxy] Java upstream=%s live=%d record=%d",
			config.GlobalConfig.Proxy.TargetURL,
			len(config.GlobalConfig.Proxy.LiveList),
			len(config.GlobalConfig.Proxy.RecordList),
		)
	}
	rpc.InitNacos()
	middleware.InitAuthExclude()
	rpc.InitRestClient()

	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/actuator/health"},
	}))
	r.Use(gin.Recovery())
	r.Use(middleware.GlobalErrorMiddleware())
	r.Use(middleware.AuthContextMiddleware())
	r.Use(middleware.ProxyGateway())
	r.Use(middleware.ShadowDiffMiddleware())

	root := r.Group("")
	operation.RegisterRoutes(root)
	user.RegisterRoutes(root)
	order.RegisterRoutes(root)
	management.RegisterRoutes(root)
	pay.RegisterRoutes(root)
	marketing.RegisterRoutes(root)
	fence.RegisterRoutes(root)
	clientconfig.RegisterRoutes(root)
	misc.RegisterRoutes(root)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.GlobalConfig.Server.Port),
		Handler: r,
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

	if os.Getenv("DRY_RUN") != "true" && rpc.NamingClient != nil {
		rpc.NamingClient.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          rpc.GetLocalIP(),
			Port:        uint64(config.GlobalConfig.Server.Port),
			ServiceName: config.GlobalConfig.Server.Name,
			GroupName:   config.GlobalConfig.Nacos.Group,
			Ephemeral:   true,
		})
	}

	log.Println("Server exiting")
}

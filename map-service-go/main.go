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

	"map-service-go/internal/api"
	"map-service-go/internal/api/middleware"
	"map-service-go/internal/cache"
	"map-service-go/internal/config"
	"map-service-go/internal/pkg/log"
	"map-service-go/internal/registry"

	"github.com/gin-gonic/gin"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/sirupsen/logrus"
)

func main() {
	// Setup logging
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	// Register context hook to inject tenantId/traceId automatically into all context logs
	logrus.AddHook(log.NewContextHook())

	logrus.Info("Starting map-service-go...")

	// Initialize Config from Nacos
	config.InitConfig()

	// Initialize Redis
	cache.InitRedis()

	// Initialize Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Add global middlewares
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.SecurityMiddleware())
	r.Use(middleware.LoggerAndContextMiddleware())

	// Controller mapping
	ctrl := api.NewMapController()

	mapGroup := r.Group("/map")
	{
		mapGroup.GET("/location", ctrl.Location)
		mapGroup.POST("/regeo", ctrl.Regeo)
		mapGroup.POST("/batchRegeo", ctrl.BatchRegeo)
		mapGroup.POST("/geo", ctrl.Geo)
		mapGroup.POST("/mapNavigate", ctrl.MapNavigate)
	}

	// Actuator endpoints (Spring Boot Actuator compatible)
	actuator := r.Group("/actuator")
	{
		actuator.GET("/health/liveness", func(c *gin.Context) { c.JSON(200, gin.H{"status": "UP"}) })
		actuator.GET("/health/readiness", func(c *gin.Context) { c.JSON(200, gin.H{"status": "UP"}) })
		actuator.POST("/deregisterService", func(c *gin.Context) {
			if err := registry.Deregister(); err != nil {
				logrus.Errorf("Failed to deregister from Nacos via actuator: %v", err)
				c.JSON(200, gin.H{"success": false, "msg": err.Error()})
				return
			}
			c.JSON(200, gin.H{"success": true, "msg": "deregistered"})
		})
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Construct HTTP Server for graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server asynchronously
	go func() {
		logrus.Infof("Server is running on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to run server: %v", err)
		}
	}()

	// Retrieve local IP for Nacos registration
	localIp := os.Getenv("POD_IP")
	if localIp == "" {
		localIp = "127.0.0.1" // Fallback for local development
	}

	// Register instance to Nacos discovery here if NACOS_REGISTER_ENABLED == true
	var namingClient naming_client.INamingClient
	var err error
	if os.Getenv("NACOS_REGISTER_ENABLED") == "true" {
		serverAddr := os.Getenv("NACOS_SERVER_ADDR")
		if serverAddr == "" {
			serverAddr = "192.168.2.20:8848"
		}
		namespace := os.Getenv("NACOS_NAMESPACE")
		if namespace == "" {
			namespace = "prod"
		}
		group := os.Getenv("NACOS_GROUP")
		if group == "" {
			group = "xyy"
		}

		ip := serverAddr
		nport := 8848
		for i := len(serverAddr) - 1; i >= 0; i-- {
			if serverAddr[i] == ':' {
				ip = serverAddr[:i]
				if p, e := strconv.Atoi(serverAddr[i+1:]); e == nil {
					nport = p
				}
				break
			}
		}

		sc := []constant.ServerConfig{
			*constant.NewServerConfig(ip, uint64(nport), constant.WithContextPath("/nacos")),
		}
		cc := *constant.NewClientConfig(
			constant.WithNamespaceId(namespace),
			constant.WithTimeoutMs(5000),
			constant.WithLogLevel("error"),
		)

		namingClient, err = clients.NewNamingClient(vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		})
		if err != nil {
			logrus.Fatalf("Failed to create naming client: %v", err)
		}

		p, _ := strconv.Atoi(port)
		success, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
			Ip:          localIp,
			Port:        uint64(p),
			ServiceName: "map-service",
			Weight:      10,
			Enable:      true,
			Healthy:     true,
			Ephemeral:   true,
			GroupName:   group,
		})
		if !success || err != nil {
			logrus.Errorf("Failed to register Nacos instance: %v", err)
		} else {
			logrus.Infof("Successfully registered to Nacos as %s:%d", localIp, p)
			registry.SetDeregister(func() error {
				ok, derr := namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
					Ip:          localIp,
					Port:        uint64(p),
					ServiceName: "map-service",
					GroupName:   group,
					Ephemeral:   true,
				})
				if !ok || derr != nil {
					return fmt.Errorf("deregister failed: %v", derr)
				}
				logrus.Info("Successfully deregistered from Nacos via actuator")
				return nil
			})
		}
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down server...")

	// Deregister from Nacos
	if err := registry.Deregister(); err != nil {
		logrus.Errorf("Failed to deregister from Nacos: %v", err)
	}

	// Shutdown HTTP server gracefully
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logrus.Errorf("HTTP Server forced to shutdown: %v", err)
	}
	logrus.Info("Server exited gracefully")
}

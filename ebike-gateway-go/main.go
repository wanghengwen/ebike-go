package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ebike-gateway-go/config"
	"ebike-gateway-go/logger"
	"ebike-gateway-go/metrics"
	"ebike-gateway-go/middleware"
	"ebike-gateway-go/proxy"

	"github.com/gin-gonic/gin"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var Version = "1.0.25-debug"

func main() {
	logger.InitLogger()
	defer logger.Log.Sync()

	if err := config.LoadLocalConfig("application.yml"); err != nil {
		logger.Log.Fatal("Failed to load local config", zap.Error(err))
	}

	if err := config.InitConfigFromNacos(); err != nil {
		logger.Log.Fatal("Failed to load nacos config", zap.Error(err))
	}

	if err := middleware.InitRedisWithPolicy(); err != nil {
		logger.Log.Fatal("Failed to init redis", zap.Error(err))
	}

	if err := config.ListenRemoteConfig(func() {
		if err := middleware.ReinitRedis(); err != nil {
			logger.Log.Error("Failed to reinit redis after config hot reload", zap.Error(err))
		}
	}); err != nil {
		logger.Log.Error("Failed to listen for remote config hot reload", zap.Error(err))
	}

	proxyEngine := proxy.NewProxyEngine()

	configClient, err := clients.NewConfigClient(config.NacosConfigClient)
	if err != nil {
		logger.Log.Fatal("Failed to create nacos config client for listening", zap.Error(err))
	}

	routeDataId := fmt.Sprintf("ebike-gateway-%s-route.yml", config.AppConfig.GatewayMode)
	routeGroup := config.AppConfig.Nacos.Group

	logger.Log.Info("Fetching initial routing table from Nacos", zap.String("dataId", routeDataId))
	routeContent, err := configClient.GetConfig(vo.ConfigParam{
		DataId: routeDataId,
		Group:  routeGroup,
	})
	if err == nil && routeContent != "" {
		routes, parseErr := proxy.ParseRoutes(routeContent)
		if parseErr != nil {
			logger.Log.Error("Failed to parse initial routing table", zap.Error(parseErr))
		} else {
			proxyEngine.UpdateRoutes(routes)
			metrics.SetRoutesLoaded(len(routes))
			metrics.SetProxyCacheSize(proxyEngine.CachedProxyCount())
			logger.Log.Info("Initial routing table applied")
		}
	}

	err = configClient.ListenConfig(vo.ConfigParam{
		DataId: routeDataId,
		Group:  routeGroup,
		OnChange: func(namespace, group, dataId, data string) {
			logger.Log.Info("Nacos dynamic route configuration updated by remote", zap.String("dataId", dataId))
			routes, parseErr := proxy.ParseRoutes(data)
			if parseErr != nil {
				logger.Log.Error("Failed to parse updated routing table", zap.Error(parseErr))
				return
			}
			proxyEngine.UpdateRoutes(routes)
			metrics.SetRoutesLoaded(len(routes))
			metrics.SetProxyCacheSize(proxyEngine.CachedProxyCount())
			logger.Log.Info("Applied new routing table to proxy engine successfully")
		},
	})
	if err != nil {
		logger.Log.Error("Failed to listen for route changes", zap.Error(err))
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	_ = r.SetTrustedProxies([]string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"100.64.0.0/10",
	})

	r.GET("/actuator/prometheus", gin.WrapH(promhttp.Handler()))

	r.Use(metrics.Middleware())
	r.Use(middleware.CorsMiddleware())
	r.Use(proxyEngine.MatchRouteMiddleware())
	r.Use(middleware.ReadBodyMiddleware())
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.DebugHeadersMiddleware())
	r.Use(middleware.LogMiddleware())
	r.Use(middleware.SignMiddleware())
	r.Use(middleware.SecurityMiddleware())
	r.Use(middleware.MirrorMiddleware())

	if config.AppConfig.GatewayMode == "business" {
		middleware.InitPermissionManager()
	}

	r.NoRoute(func(c *gin.Context) {
		proxyEngine.HandleRequest(c)
		metrics.SetProxyCacheSize(proxyEngine.CachedProxyCount())
	})

	port := config.AppConfig.Port
	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		logger.Log.Info("Starting ebike-gateway-go",
			zap.String("version", Version),
			zap.Int("port", port),
			zap.String("mode", config.AppConfig.GatewayMode),
			zap.Bool("debugHeaders", config.AppConfig.DebugHeaders),
			zap.Bool("mirrorEnabled", config.AppConfig.Mirror.Enabled),
			zap.String("mirrorTarget", config.AppConfig.Mirror.Target),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server forced to shutdown", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("Shutdown signal received, draining connections")

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server shutdown failed", zap.Error(err))
	}
	logger.Log.Info("Server exited gracefully")
}

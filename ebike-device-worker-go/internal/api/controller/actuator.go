package controller

import (
	"context"
	"net"
	"net/http"
	"time"

	"ebike-device-worker-go/internal/api/dto"
	nacosreg "ebike-device-worker-go/internal/infrastructure/nacos"
	"ebike-device-worker-go/internal/config"
	"ebike-device-worker-go/internal/pkg/db"
	"ebike-device-worker-go/internal/pkg/fastid"
	redispkg "ebike-device-worker-go/internal/pkg/redis"

	"github.com/gin-gonic/gin"
)

type componentHealth struct {
	Status string `json:"status"`
}

type healthResponse struct {
	Status     string                     `json:"status"`
	Components map[string]componentHealth `json:"components"`
}

// Health mirrors Spring Boot /actuator/health for probes.
func Health(c *gin.Context) {
	components := map[string]componentHealth{
		"fastid": fastidComponent(),
		"redis":  redisComponent(),
		"db":     dbComponent(),
		"kafka":  kafkaComponent(),
	}

	overall := "UP"
	for _, comp := range components {
		if comp.Status == "DOWN" {
			overall = "DOWN"
			break
		}
	}

	code := http.StatusOK
	if overall == "DOWN" {
		code = http.StatusServiceUnavailable
	}
	c.JSON(code, healthResponse{Status: overall, Components: components})
}

// DeregisterService mirrors Java /actuator/deregisterService (localhost only, for K8s preStop).
func DeregisterService(c *gin.Context) {
	if !isLocalhostRequest(c) {
		c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeException, "只允许localhost请求"))
		return
	}
	nacosreg.DeregisterInstance()
	c.JSON(http.StatusOK, dto.NewSuccessResult("ok"))
}

func isLocalhostRequest(c *gin.Context) bool {
	host := c.Request.URL.Hostname()
	if host == "" {
		var err error
		host, _, err = net.SplitHostPort(c.Request.Host)
		if err != nil {
			host = c.Request.Host
		}
	}
	return host == "localhost" || host == "127.0.0.1"
}

func fastidComponent() componentHealth {
	if !config.GlobalConfig.FastIDEnabled() {
		return componentHealth{Status: "UP"}
	}
	if fastid.Ready() {
		return componentHealth{Status: "UP"}
	}
	return componentHealth{Status: "DOWN"}
}

func redisComponent() componentHealth {
	if config.GlobalConfig.Redis.Addr == "" {
		return componentHealth{Status: "UP"}
	}
	if redispkg.Client == nil {
		return componentHealth{Status: "DOWN"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := redispkg.Client.Ping(ctx).Err(); err != nil {
		return componentHealth{Status: "DOWN"}
	}
	return componentHealth{Status: "UP"}
}

func dbComponent() componentHealth {
	if config.GlobalConfig.Database.Dsn == "" {
		return componentHealth{Status: "UP"}
	}
	if db.DB == nil {
		return componentHealth{Status: "DOWN"}
	}
	sqlDB, err := db.DB.DB()
	if err != nil {
		return componentHealth{Status: "DOWN"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return componentHealth{Status: "DOWN"}
	}
	return componentHealth{Status: "UP"}
}

func kafkaComponent() componentHealth {
	if !config.GlobalConfig.Kafka.Enabled {
		return componentHealth{Status: "UP"}
	}
	if len(config.GlobalConfig.Kafka.Brokers) == 0 {
		return componentHealth{Status: "DOWN"}
	}
	return componentHealth{Status: "UP"}
}

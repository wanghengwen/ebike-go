package config

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"ebike-gateway-go/logger"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	NacosConfigClient vo.NacosClientParam
	configMu          sync.RWMutex
)

func InitConfigFromNacos() error {
	if AppConfig.Nacos.ServerAddr == "" {
		return fmt.Errorf("nacos server address is empty")
	}

	parts := strings.Split(AppConfig.Nacos.ServerAddr, ":")
	ip := parts[0]
	portVal := uint64(8848)
	if len(parts) > 1 {
		if p, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
			portVal = p
		}
	}

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(ip, portVal, constant.WithContextPath("/nacos")),
	}

	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(AppConfig.Nacos.Namespace),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("./logs/nacos"),
		constant.WithCacheDir("./logs/nacos/cache"),
		constant.WithLogLevel("error"),
	)

	NacosConfigClient = vo.NacosClientParam{
		ClientConfig:  &cc,
		ServerConfigs: sc,
	}

	configClient, err := clients.NewConfigClient(NacosConfigClient)
	if err != nil {
		return fmt.Errorf("create nacos client error: %w", err)
	}

	redisContent, err := configClient.GetConfig(vo.ConfigParam{
		DataId: "redis.yaml",
		Group:  AppConfig.Nacos.Group + "_ops",
	})
	if err != nil {
		return fmt.Errorf("fetch redis.yaml from nacos error: %w", err)
	}

	secretContent, err := configClient.GetConfig(vo.ConfigParam{
		DataId: "gateway-secret.yaml",
		Group:  AppConfig.Nacos.Group + "_ops",
	})
	if err != nil {
		return fmt.Errorf("fetch gateway-secret.yaml from nacos error: %w", err)
	}

	if err := ApplyRemoteConfig(redisContent, secretContent); err != nil {
		return err
	}

	logger.Log.Info("Loaded configuration from Nacos",
		zap.String("mode", AppConfig.GatewayMode),
		zap.String("redis", GetRedisConfig().Host),
	)
	return nil
}

// ApplyRemoteConfig parses redis.yaml and gateway-secret.yaml and updates AppConfig.
func ApplyRemoteConfig(redisContent, secretContent string) error {
	var nacosRedis NacosRedisConfig
	if err := yaml.Unmarshal([]byte(redisContent), &nacosRedis); err != nil {
		return fmt.Errorf("unmarshal redis.yaml error: %w", err)
	}

	var nacosSecret NacosSecretConfig
	if err := yaml.Unmarshal([]byte(secretContent), &nacosSecret); err != nil {
		return fmt.Errorf("unmarshal gateway-secret.yaml error: %w", err)
	}

	configMu.Lock()
	defer configMu.Unlock()

	if AppConfig.GatewayMode == "business" {
		AppConfig.Redis = nacosRedis.Redis.EbikeManagement
		AppConfig.JwtSecret = nacosSecret.Xyy.Secret.Jwt.Business
	} else {
		AppConfig.Redis = nacosRedis.Redis.EbikeUser
		AppConfig.JwtSecret = nacosSecret.Xyy.Secret.Jwt.Client
	}
	return nil
}

// GetRedisConfig returns a snapshot of the current Redis configuration.
func GetRedisConfig() RedisInstance {
	configMu.RLock()
	defer configMu.RUnlock()
	return AppConfig.Redis
}

// GetJwtSecret returns the current JWT secret for the active gateway mode.
func GetJwtSecret() string {
	configMu.RLock()
	defer configMu.RUnlock()
	return AppConfig.JwtSecret
}

// ListenRemoteConfig watches redis.yaml and gateway-secret.yaml for hot reload.
func ListenRemoteConfig(onChange func()) error {
	configClient, err := clients.NewConfigClient(NacosConfigClient)
	if err != nil {
		return fmt.Errorf("create nacos config client for remote listen error: %w", err)
	}

	opsGroup := AppConfig.Nacos.Group + "_ops"
	latest := struct {
		redis  string
		secret string
		mu     sync.Mutex
	}{}

	applyIfReady := func() {
		latest.mu.Lock()
		redisContent := latest.redis
		secretContent := latest.secret
		latest.mu.Unlock()

		if redisContent == "" || secretContent == "" {
			return
		}

		if err := ApplyRemoteConfig(redisContent, secretContent); err != nil {
			logger.Log.Error("Failed to apply hot-reloaded nacos config", zap.Error(err))
			return
		}

		logger.Log.Info("Hot-reloaded redis/jwt config from Nacos",
			zap.String("mode", AppConfig.GatewayMode),
			zap.String("redis", GetRedisConfig().Host),
		)
		if onChange != nil {
			onChange()
		}
	}

	listen := func(dataId string, kind string) error {
		content, err := configClient.GetConfig(vo.ConfigParam{DataId: dataId, Group: opsGroup})
		if err != nil {
			return fmt.Errorf("fetch initial %s error: %w", dataId, err)
		}
		latest.mu.Lock()
		if kind == "redis" {
			latest.redis = content
		} else {
			latest.secret = content
		}
		latest.mu.Unlock()

		return configClient.ListenConfig(vo.ConfigParam{
			DataId: dataId,
			Group:  opsGroup,
			OnChange: func(_, _, _, data string) {
				logger.Log.Info("Nacos remote config updated", zap.String("dataId", dataId))
				latest.mu.Lock()
				if kind == "redis" {
					latest.redis = data
				} else {
					latest.secret = data
				}
				latest.mu.Unlock()
				applyIfReady()
			},
		})
	}

	if err := listen("redis.yaml", "redis"); err != nil {
		return err
	}
	if err := listen("gateway-secret.yaml", "secret"); err != nil {
		return err
	}

	return nil
}

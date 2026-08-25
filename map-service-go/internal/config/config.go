package config

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/goccy/go-yaml"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/sirupsen/logrus"
)

var (
	GlobalConfig *Config
	once         sync.Once
	configMu     sync.RWMutex
	lastProfileYAML string
)

func GetConfig() *Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return GlobalConfig
}


type Config struct {
	Xyy   XyyConfig   `yaml:"xyy"`
	Redis RedisConfig `yaml:"redis"` // Assuming standard Redis config structure
}

type XyyConfig struct {
	AMap      MapUrl       `yaml:"amap"`
	TMap      MapUrl       `yaml:"tmap"`
	GeoLength int          `yaml:"geoLength"`
	System    SystemConfig `yaml:"system"`
}

type MapUrl struct {
	Keys         map[string]interface{} `yaml:"keys"`
	Key          map[string]interface{} `yaml:"key"` // Fallback for singular 'key' mapping
	GeoUrl       string                 `yaml:"geoUrl"`
	RegeoUrl     string                 `yaml:"regeoUrl"`
	WalkingUrl   string                 `yaml:"walkingUrl"`
	BicyclingUrl string                 `yaml:"bicyclingUrl"`
}

type SystemConfig struct {
	Secrets            []string `yaml:"secrets"`
	SecurityIgnoreUrls []string `yaml:"securityIgnoreUrls"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
}

func getActiveProfile() string {
	profile := os.Getenv("SPRING_PROFILES_ACTIVE")
	if profile == "" {
		profile = os.Getenv("APP_PROFILE")
	}
	if profile == "" {
		profile = "prod"
	}
	return profile
}

func InitConfig() {
	once.Do(func() {
		serverAddr := os.Getenv("NACOS_SERVER_ADDR")
		if serverAddr == "" {
			serverAddr = "192.168.2.20:8848" // Default fallback
		}
		namespace := os.Getenv("NACOS_NAMESPACE")
		if namespace == "" {
			namespace = "prod"
		}
		group := os.Getenv("NACOS_GROUP")
		if group == "" {
			group = "xyy"
		}
		dataId := "map-service.yml"
		profile := getActiveProfile()
		profileDataId := fmt.Sprintf("map-service-%s.yml", profile)

		// Parse Server Addr
		ip := serverAddr
		port := 8848
		// Basic split for ip:port
		for i := len(serverAddr) - 1; i >= 0; i-- {
			if serverAddr[i] == ':' {
				ip = serverAddr[:i]
				if p, err := strconv.Atoi(serverAddr[i+1:]); err == nil {
					port = p
				}
				break
			}
		}

		sc := []constant.ServerConfig{
			*constant.NewServerConfig(ip, uint64(port), constant.WithContextPath("/nacos")),
		}

		logDir := os.TempDir() + string(os.PathSeparator) + "nacos" + string(os.PathSeparator) + "log"
		cacheDir := os.TempDir() + string(os.PathSeparator) + "nacos" + string(os.PathSeparator) + "cache"

		cc := *constant.NewClientConfig(
			constant.WithNamespaceId(namespace),
			constant.WithTimeoutMs(5000),
			constant.WithNotLoadCacheAtStart(true),
			constant.WithLogDir(logDir),
			constant.WithCacheDir(cacheDir),
			constant.WithLogLevel("error"),
		)

		client, err := clients.NewConfigClient(
			vo.NacosClientParam{
				ClientConfig:  &cc,
				ServerConfigs: sc,
			},
		)
		if err != nil {
			logrus.Fatalf("Failed to create Nacos config client: %v", err)
		}

		content, err := client.GetConfig(vo.ConfigParam{
			DataId: dataId,
			Group:  group,
		})
		if err != nil {
			logrus.Fatalf("Failed to get config from Nacos: %v", err)
		}

		err = parseConfig(content)
		if err != nil {
			logrus.Fatalf("Failed to parse config: %v", err)
		}

		profileContent, err := client.GetConfig(vo.ConfigParam{
			DataId: profileDataId,
			Group:  group,
		})
		if err != nil {
			logrus.Warnf("Failed to get profile config %s from Nacos: %v", profileDataId, err)
		} else if profileContent != "" {
			if err = mergeProfileConfig(profileContent); err != nil {
				logrus.Fatalf("Failed to merge profile config: %v", err)
			}
			logrus.Infof("Merged Nacos profile config: %s", profileDataId)
		}

		// Fetch redis.yaml from xyy_ops
		redisContent, err := client.GetConfig(vo.ConfigParam{
			DataId: "redis.yaml",
			Group:  "xyy_ops",
		})
		if err != nil {
			logrus.Warnf("Failed to get redis.yaml from Nacos: %v", err)
		} else {
			_ = parseRedisConfig(redisContent)
		}

		// Listen for config changes
		err = client.ListenConfig(vo.ConfigParam{
			DataId: dataId,
			Group:  group,
			OnChange: func(namespace, group, dataId, data string) {
				logrus.Infof("Config changed in Nacos: %s/%s", group, dataId)
				if err := parseConfig(data); err != nil {
					logrus.Errorf("Failed to parse base config: %v", err)
				}
			},
		})
		if err != nil {
			logrus.Errorf("Failed to listen config from Nacos: %v", err)
		}

		err = client.ListenConfig(vo.ConfigParam{
			DataId: profileDataId,
			Group:  group,
			OnChange: func(namespace, group, dataId, data string) {
				logrus.Infof("Profile config changed in Nacos: %s/%s", group, dataId)
				_ = mergeProfileConfig(data)
			},
		})
		if err != nil {
			logrus.Errorf("Failed to listen profile config from Nacos: %v", err)
		}

		// Listen for redis.yaml changes
		err = client.ListenConfig(vo.ConfigParam{
			DataId: "redis.yaml",
			Group:  "xyy_ops",
			OnChange: func(namespace, group, dataId, data string) {
				logrus.Infof("Config changed in Nacos: %s/%s", group, dataId)
				_ = parseRedisConfig(data)
			},
		})
		if err != nil {
			logrus.Errorf("Failed to listen redis config from Nacos: %v", err)
		}
	})
}

func parseConfig(content string) error {
	var c Config
	decoder := yaml.NewDecoder(bytes.NewBufferString(content))
	if err := decoder.Decode(&c); err != nil {
		return fmt.Errorf("yaml decode error: %w", err)
	}
	configMu.Lock()
	if GlobalConfig != nil {
		c.Redis = GlobalConfig.Redis // Preserve existing Redis config
	}
	if lastProfileYAML != "" {
		var overlay Config
		overlayDecoder := yaml.NewDecoder(bytes.NewBufferString(lastProfileYAML))
		if err := overlayDecoder.Decode(&overlay); err == nil {
			c = mergeConfig(c, overlay)
		}
	}
	GlobalConfig = &c
	configMu.Unlock()
	return nil
}

func mergeProfileConfig(content string) error {
	lastProfileYAML = content
	var overlay Config
	decoder := yaml.NewDecoder(bytes.NewBufferString(content))
	if err := decoder.Decode(&overlay); err != nil {
		return fmt.Errorf("profile yaml decode error: %w", err)
	}
	configMu.Lock()
	defer configMu.Unlock()
	if GlobalConfig == nil {
		GlobalConfig = &overlay
		return nil
	}
	merged := mergeConfig(*GlobalConfig, overlay)
	GlobalConfig = &merged
	return nil
}

func mergeConfig(base, overlay Config) Config {
	result := base
	if overlay.Xyy.GeoLength != 0 {
		result.Xyy.GeoLength = overlay.Xyy.GeoLength
	}
	result.Xyy.AMap = mergeMapUrl(base.Xyy.AMap, overlay.Xyy.AMap)
	result.Xyy.TMap = mergeMapUrl(base.Xyy.TMap, overlay.Xyy.TMap)
	if len(overlay.Xyy.System.Secrets) > 0 {
		result.Xyy.System.Secrets = overlay.Xyy.System.Secrets
	}
	if len(overlay.Xyy.System.SecurityIgnoreUrls) > 0 {
		result.Xyy.System.SecurityIgnoreUrls = overlay.Xyy.System.SecurityIgnoreUrls
	}
	return result
}

func mergeMapUrl(base, overlay MapUrl) MapUrl {
	result := base
	if overlay.GeoUrl != "" {
		result.GeoUrl = overlay.GeoUrl
	}
	if overlay.RegeoUrl != "" {
		result.RegeoUrl = overlay.RegeoUrl
	}
	if overlay.WalkingUrl != "" {
		result.WalkingUrl = overlay.WalkingUrl
	}
	if overlay.BicyclingUrl != "" {
		result.BicyclingUrl = overlay.BicyclingUrl
	}
	if len(overlay.Keys) > 0 {
		result.Keys = overlay.Keys
	}
	if len(overlay.Key) > 0 {
		result.Key = overlay.Key
	}
	return result
}

type RedisOpsYaml struct {
	Redis map[string]RedisConfig `yaml:"redis"`
}

func parseRedisConfig(content string) error {
	if content == "" {
		return nil
	}
	var r RedisOpsYaml
	decoder := yaml.NewDecoder(bytes.NewBufferString(content))
	if err := decoder.Decode(&r); err != nil {
		return fmt.Errorf("redis yaml decode error: %w", err)
	}

	if rc, ok := r.Redis["map_service"]; ok {
		configMu.Lock()
		if GlobalConfig != nil {
			GlobalConfig.Redis = rc
		}
		configMu.Unlock()
	}
	return nil
}

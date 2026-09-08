package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"ebike-auth-go/internal/pkg/logger"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type RedisInstance struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
}

type NacosRedisConfig struct {
	Redis struct {
		EbikeManagement RedisInstance `yaml:"ebike_management"`
		EbikeUser       RedisInstance `yaml:"ebike_user"`
	} `yaml:"redis"`
}

type NacosSecretConfig struct {
	Xyy struct {
		Secret struct {
			Jwt struct {
				Client   string `yaml:"client"`
				Business string `yaml:"business"`
			} `yaml:"jwt"`
		} `yaml:"secret"`
	} `yaml:"xyy"`
}

type NacosAppProperties struct {
	Xyy struct {
		EnablePhoneVerify        bool              `yaml:"enablePhoneVerify"`
		EnableEmailVerify        bool              `yaml:"enableEmailVerify"`
		EnableTenantSecretVerify bool              `yaml:"enableTenantSecretVerify"`
		Aks                      map[string]string `yaml:"aks"`
		GrayTenantIds            []string          `yaml:"grayTenantIds"`
		Wechat                   struct {
			ComponentAppId     string `yaml:"componentAppId"`
			ComponentAppSecret string `yaml:"componentAppSecret"`
		} `yaml:"wechat"`
		// 玉环公共电单车客户定制：玉岛行登录实名拦截与手机号格式化（对应 ebike-auth-client ydx 分支）
		YuhuanYudaoxingCustom *bool `yaml:"yuhuanYudaoxingCustom"`
	} `yaml:"xyy"`
	Yudaoxing nacosYudaoxingProperties `yaml:"yudaoxing"`
	Weixin    nacosWeixinProperties    `yaml:"weixin"`
	Feign     struct {
		Weixin struct {
			URL string `yaml:"url"`
		} `yaml:"weixin"`
	} `yaml:"feign"`
}

type NacosConfig struct {
	ServerAddr      string `yaml:"serverAddr"`
	Namespace       string `yaml:"namespace"`
	Group           string `yaml:"group"`
	RegisterEnabled bool   `yaml:"registerEnabled"`
}

type AppConfigStruct struct {
	Port                     int
	AuthMode                 string // "business" or "client"
	Nacos                    NacosConfig
	Redis                    RedisInstance
	JwtSecret                string
	EnablePhoneVerify        bool
	EnableEmailVerify        bool
	EnableTenantSecretVerify bool
	Aks                      map[string]string
	GrayTenantIds            []string
	WechatComponentAppId     string
	WechatComponentAppSecret string
	// YuhuanYudaoxingCustom 玉环公共电单车客户定制开关（玉岛行登录扩展，见 auth/yuhuan_yudaoxing.go）
	YuhuanYudaoxingCustom bool
	Yudaoxing             YudaoxingConfig
	WeixinPublicPlatform  WeixinPublicPlatformConfig
}

type YudaoxingConfig struct {
	BaseUrl            string
	Account            string
	MerchantPrivateKey string
	ServerPublicKey    string
}

var AppConfig = &AppConfigStruct{
	Port:                     8080,
	AuthMode:                 "client",
	EnablePhoneVerify:        true,
	EnableEmailVerify:        true,
	EnableTenantSecretVerify: false,
	YuhuanYudaoxingCustom:    YuhuanYudaoxingCompileDefault(),
	WeixinPublicPlatform: WeixinPublicPlatformConfig{
		AppId:     defaultWeixinPublicAppID,
		AppSecret: defaultWeixinPublicAppSecret,
	},
}

type rootFileConfig struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Xyy struct {
		Auth struct {
			Mode string `yaml:"mode"`
		} `yaml:"auth"`
		Nacos NacosConfig `yaml:"nacos"`
	} `yaml:"xyy"`
}

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read application.yml error: %w", err)
	}

	var root rootFileConfig
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("parse application.yml error: %w", err)
	}

	if root.Server.Port > 0 {
		AppConfig.Port = root.Server.Port
	}
	if root.Xyy.Auth.Mode != "" {
		AppConfig.AuthMode = root.Xyy.Auth.Mode
	}
	AppConfig.Nacos = root.Xyy.Nacos

	overrideWithEnv()
	return nil
}

func overrideWithEnv() {
	if val := os.Getenv("AUTH_SERVICE_MODE"); val != "" {
		AppConfig.AuthMode = strings.ToLower(val)
	} else if val := os.Getenv("XYY_AUTH_MODE"); val != "" {
		AppConfig.AuthMode = strings.ToLower(val)
	}

	if val := os.Getenv("PORT"); val != "" {
		if p, err := strconv.Atoi(val); err == nil {
			AppConfig.Port = p
		}
	} else if val := os.Getenv("XYY_HTTP_PORT"); val != "" {
		if p, err := strconv.Atoi(val); err == nil {
			AppConfig.Port = p
		}
	}

	if val := os.Getenv("SPRING_CLOUD_NACOS_DISCOVERY_SERVER_ADDR"); val != "" {
		AppConfig.Nacos.ServerAddr = val
	} else if val := os.Getenv("NACOS_SERVER_ADDR"); val != "" {
		AppConfig.Nacos.ServerAddr = val
	}

	if val := os.Getenv("SPRING_CLOUD_NACOS_DISCOVERY_NAMESPACE"); val != "" {
		AppConfig.Nacos.Namespace = val
	} else if val := os.Getenv("NACOS_NAMESPACE"); val != "" {
		AppConfig.Nacos.Namespace = val
	}

	if val := os.Getenv("SPRING_CLOUD_NACOS_DISCOVERY_GROUP"); val != "" {
		AppConfig.Nacos.Group = val
	} else if val := os.Getenv("NACOS_GROUP"); val != "" {
		AppConfig.Nacos.Group = val
	}

	applyYuhuanYudaoxingCustomFromEnv()

	if val := os.Getenv("SPRING_CLOUD_NACOS_DISCOVERY_ENABLED"); val != "" {
		AppConfig.Nacos.RegisterEnabled = strings.ToLower(val) == "true"
	} else if val := os.Getenv("NACOS_REGISTER_ENABLED"); val != "" {
		AppConfig.Nacos.RegisterEnabled = strings.ToLower(val) == "true"
	}
}

func InitConfigFromNacos() error {
	if AppConfig.Nacos.ServerAddr == "" {
		return fmt.Errorf("nacos server address is empty, cannot load configurations")
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

	logger.Log.Info("Initializing Nacos config client",
		zap.String("server_addr", AppConfig.Nacos.ServerAddr),
		zap.String("namespace", AppConfig.Nacos.Namespace),
		zap.String("group", AppConfig.Nacos.Group),
	)

	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return fmt.Errorf("create nacos client error: %w", err)
	}

	// 1. Fetch redis.yaml config
	redisDataId := "redis.yaml"
	redisGroup := AppConfig.Nacos.Group + "_ops"
	redisContent, err := configClient.GetConfig(vo.ConfigParam{
		DataId: redisDataId,
		Group:  redisGroup,
	})
	if err != nil {
		return fmt.Errorf("fetch redis.yaml from nacos error: %w", err)
	}
	logger.Log.Info("Successfully pulled nacos config file",
		zap.String("data_id", redisDataId),
		zap.String("group", redisGroup),
	)
	var nacosRedis NacosRedisConfig
	if err := yaml.Unmarshal([]byte(redisContent), &nacosRedis); err != nil {
		return fmt.Errorf("unmarshal redis.yaml error: %w", err)
	}

	// 2. Fetch gateway-secret.yaml config
	secretDataId := "gateway-secret.yaml"
	secretGroup := AppConfig.Nacos.Group + "_ops"
	secretContent, err := configClient.GetConfig(vo.ConfigParam{
		DataId: secretDataId,
		Group:  secretGroup,
	})
	if err != nil {
		return fmt.Errorf("fetch gateway-secret.yaml from nacos error: %w", err)
	}
	logger.Log.Info("Successfully pulled nacos config file",
		zap.String("data_id", secretDataId),
		zap.String("group", secretGroup),
	)
	var nacosSecret NacosSecretConfig
	if err := yaml.Unmarshal([]byte(secretContent), &nacosSecret); err != nil {
		return fmt.Errorf("unmarshal gateway-secret.yaml error: %w", err)
	}

	// 3. Fetch ebike-auth-{mode}.yml config
	propertiesDataId := "ebike-auth-" + AppConfig.AuthMode + ".yml"
	propertiesGroup := AppConfig.Nacos.Group
	appPropertiesContent, err := configClient.GetConfig(vo.ConfigParam{
		DataId: propertiesDataId,
		Group:  propertiesGroup,
	})
	if err == nil && appPropertiesContent != "" {
		logger.Log.Info("Successfully pulled nacos config file",
			zap.String("data_id", propertiesDataId),
			zap.String("group", propertiesGroup),
		)
		var appProps NacosAppProperties
		if err := yaml.Unmarshal([]byte(appPropertiesContent), &appProps); err == nil {
			AppConfig.EnablePhoneVerify = appProps.Xyy.EnablePhoneVerify
			AppConfig.EnableEmailVerify = appProps.Xyy.EnableEmailVerify
			AppConfig.EnableTenantSecretVerify = appProps.Xyy.EnableTenantSecretVerify
			AppConfig.Aks = appProps.Xyy.Aks
			AppConfig.GrayTenantIds = appProps.Xyy.GrayTenantIds
			AppConfig.WechatComponentAppId = appProps.Xyy.Wechat.ComponentAppId
			AppConfig.WechatComponentAppSecret = appProps.Xyy.Wechat.ComponentAppSecret
			if appProps.Xyy.YuhuanYudaoxingCustom != nil {
				AppConfig.YuhuanYudaoxingCustom = *appProps.Xyy.YuhuanYudaoxingCustom
			}
			AppConfig.Yudaoxing = appProps.Yudaoxing.toYudaoxingConfig()
			// 与 Java @Value 默认值语义一致：仅当 Nacos 显式配置时才覆盖默认凭据，
			// 未配置则保留 defaultWeixinPublicAppID/Secret。
			weixinCfg := appProps.Weixin.toWeixinPublicPlatformConfig()
			if weixinCfg.AppId != "" {
				AppConfig.WeixinPublicPlatform.AppId = weixinCfg.AppId
			}
			if weixinCfg.AppSecret != "" {
				AppConfig.WeixinPublicPlatform.AppSecret = weixinCfg.AppSecret
			}
			if appProps.Feign.Weixin.URL != "" {
				AppConfig.WeixinPublicPlatform.APIBaseURL = appProps.Feign.Weixin.URL
			}
		}
	} else {
		logger.Log.Info("Optional configuration file not found on Nacos, using default values",
			zap.String("data_id", propertiesDataId),
			zap.String("group", propertiesGroup),
			zap.Error(err),
		)
	}

	// Apply mode specific configurations
	if AppConfig.AuthMode == "business" {
		AppConfig.Redis = nacosRedis.Redis.EbikeManagement
		AppConfig.JwtSecret = nacosSecret.Xyy.Secret.Jwt.Business
	} else {
		AppConfig.Redis = nacosRedis.Redis.EbikeUser
		AppConfig.JwtSecret = nacosSecret.Xyy.Secret.Jwt.Client
	}

	if val := os.Getenv("REDIS_HOST"); val != "" {
		AppConfig.Redis.Host = val
	}
	if val := os.Getenv("REDIS_PORT"); val != "" {
		if p, err := strconv.Atoi(val); err == nil {
			AppConfig.Redis.Port = p
		}
	}

	applyOptionalEnvOverrides()
	return nil
}

func applyOptionalEnvOverrides() {
	if val := os.Getenv("GRAY_TENANT_IDS"); val != "" {
		parts := strings.Split(val, ",")
		ids := make([]string, 0, len(parts))
		for _, part := range parts {
			if id := strings.TrimSpace(part); id != "" {
				ids = append(ids, id)
			}
		}
		if len(ids) > 0 {
			AppConfig.GrayTenantIds = ids
		}
	}
	if val := os.Getenv("WECHAT_COMPONENT_APP_ID"); val != "" {
		AppConfig.WechatComponentAppId = val
	}
	if val := os.Getenv("WECHAT_COMPONENT_APP_SECRET"); val != "" {
		AppConfig.WechatComponentAppSecret = val
	}
	if val := os.Getenv("YUDAOXING_BASE_URL"); val != "" {
		AppConfig.Yudaoxing.BaseUrl = val
	}
	if val := os.Getenv("YUDAOXING_ACCOUNT"); val != "" {
		AppConfig.Yudaoxing.Account = val
	}
	if val := os.Getenv("YUDAOXING_PRIVATE_KEY"); val != "" {
		AppConfig.Yudaoxing.MerchantPrivateKey = val
	}
	if val := os.Getenv("YUDAOXING_SERVER_PUBLIC_KEY"); val != "" {
		AppConfig.Yudaoxing.ServerPublicKey = val
	}
	applyWeixinPublicPlatformFromEnv()
	applyYuhuanYudaoxingCustomFromEnv()
}

func applyWeixinPublicPlatformFromEnv() {
	if val := os.Getenv("WEIXIN_PUBLIC_APP_ID"); val != "" {
		AppConfig.WeixinPublicPlatform.AppId = val
	}
	if val := os.Getenv("WEIXIN_PUBLIC_APP_SECRET"); val != "" {
		AppConfig.WeixinPublicPlatform.AppSecret = val
	}
	if val := os.Getenv("FEIGN_WEIXIN_URL"); val != "" {
		AppConfig.WeixinPublicPlatform.APIBaseURL = val
	}
}

func applyYuhuanYudaoxingCustomFromEnv() {
	if val := os.Getenv("YUHUAN_YUDAOXING_CUSTOM"); val != "" {
		if enabled, ok := parseBoolEnv(val); ok {
			AppConfig.YuhuanYudaoxingCustom = enabled
		}
		return
	}
	if val := os.Getenv("AUTH_YUHUAN_YUDAOXING_CUSTOM"); val != "" {
		if enabled, ok := parseBoolEnv(val); ok {
			AppConfig.YuhuanYudaoxingCustom = enabled
		}
	}
}

func parseBoolEnv(val string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	default:
		return false, false
	}
}

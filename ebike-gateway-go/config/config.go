package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type NacosConfig struct {
	ServerAddr      string `yaml:"serverAddr"`
	Namespace       string `yaml:"namespace"`
	Group           string `yaml:"group"`
	RegisterEnabled bool   `yaml:"registerEnabled"`
}

// NacosRedisConfig maps to redis.yaml
type NacosRedisConfig struct {
	Redis struct {
		EbikeManagement RedisInstance `yaml:"ebike_management"`
		EbikeUser       RedisInstance `yaml:"ebike_user"`
	} `yaml:"redis"`
}

type RedisInstance struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
}

// NacosSecretConfig maps to gateway-secret.yaml
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

type MirrorConfig struct {
	Enabled          bool   // MIRROR_ENABLED=true
	Target           string // MIRROR_TARGET=http://ebike-service-client-go:8081
	RoutePattern     string // MIRROR_ROUTE_PATTERN=/client/** (Ant path, empty = all routes)
	ConnectTimeoutMs int    // MIRROR_CONNECT_TIMEOUT_MS=500
	ReadTimeoutMs    int    // MIRROR_READ_TIMEOUT_MS=500
}

type AppConfigStruct struct {
	Port            int
	GatewayMode     string // "business" or "client"
	Nacos           NacosConfig
	Redis           RedisInstance
	RedisRequired   bool
	JwtSecret       string
	ManagementUrl   string
	EnableApiVerify bool
	ExcludeApis     map[string]struct{}
	LogExcludePaths []string
	DebugHeaders    bool
	Mirror          MirrorConfig
}

var AppConfig = &AppConfigStruct{
	Port:          8080,
	GatewayMode:   "client",
	ManagementUrl: "http://ebike-management:8080",
	RedisRequired: true,
}

type rootFileConfig struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Xyy struct {
		Gateway struct {
			Mode string `yaml:"mode"`
		} `yaml:"gateway"`
		Nacos NacosConfig `yaml:"nacos"`
	} `yaml:"xyy"`
}

func LoadLocalConfig(path string) error {
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
	if root.Xyy.Gateway.Mode != "" {
		AppConfig.GatewayMode = root.Xyy.Gateway.Mode
	}
	AppConfig.Nacos = root.Xyy.Nacos

	overrideWithEnv()
	return validateConfig()
}

func validateConfig() error {
	switch AppConfig.GatewayMode {
	case "client", "business":
	default:
		return fmt.Errorf("invalid GATEWAY_MODE %q, must be client or business", AppConfig.GatewayMode)
	}
	return nil
}

func overrideWithEnv() {
	if val := os.Getenv("GATEWAY_MODE"); val != "" {
		AppConfig.GatewayMode = strings.ToLower(val)
	} else if val := os.Getenv("XYY_GATEWAY_MODE"); val != "" {
		AppConfig.GatewayMode = strings.ToLower(val)
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

	if val := os.Getenv("SPRING_CLOUD_NACOS_DISCOVERY_ENABLED"); val != "" {
		AppConfig.Nacos.RegisterEnabled = strings.ToLower(val) == "true"
	} else if val := os.Getenv("NACOS_REGISTER_ENABLED"); val != "" {
		AppConfig.Nacos.RegisterEnabled = strings.ToLower(val) == "true"
	}

	if val := os.Getenv("MANAGEMENT_URL"); val != "" {
		AppConfig.ManagementUrl = val
	}

	// Mirror (shadow traffic) configuration
	if val := os.Getenv("MIRROR_ENABLED"); val != "" {
		AppConfig.Mirror.Enabled = strings.ToLower(val) == "true"
	}
	if val := os.Getenv("MIRROR_TARGET"); val != "" {
		AppConfig.Mirror.Target = val
	}
	if val := os.Getenv("MIRROR_ROUTE_PATTERN"); val != "" {
		AppConfig.Mirror.RoutePattern = val
	}
	if val := os.Getenv("MIRROR_CONNECT_TIMEOUT_MS"); val != "" {
		if ms, err := strconv.Atoi(val); err == nil {
			AppConfig.Mirror.ConnectTimeoutMs = ms
		}
	}
	if val := os.Getenv("MIRROR_READ_TIMEOUT_MS"); val != "" {
		if ms, err := strconv.Atoi(val); err == nil {
			AppConfig.Mirror.ReadTimeoutMs = ms
		}
	}
	if AppConfig.Mirror.ConnectTimeoutMs <= 0 {
		AppConfig.Mirror.ConnectTimeoutMs = 500 // default 500ms
	}
	if AppConfig.Mirror.ReadTimeoutMs <= 0 {
		AppConfig.Mirror.ReadTimeoutMs = 500 // default 500ms
	}

	if val := os.Getenv("ENABLE_API_VERIFY"); val != "" {
		AppConfig.EnableApiVerify = strings.ToLower(val) == "true"
	}

	if val := os.Getenv("EXCLUDE_APIS"); val != "" {
		AppConfig.ExcludeApis = make(map[string]struct{})
		for _, api := range strings.Split(val, ",") {
			api = strings.TrimSpace(api)
			if api != "" {
				AppConfig.ExcludeApis[api] = struct{}{}
			}
		}
	}
	applyDefaultBusinessExcludeApis()

	if val := os.Getenv("REDIS_REQUIRED"); val != "" {
		AppConfig.RedisRequired = strings.ToLower(val) == "true"
	} else if AppConfig.GatewayMode == "client" {
		AppConfig.RedisRequired = true
	}

	if val := os.Getenv("LOG_EXCLUDE_PATHS"); val != "" {
		AppConfig.LogExcludePaths = splitAndTrim(val)
	} else if AppConfig.GatewayMode == "client" {
		// Default from Nacos template ebike-gateway-client.yml
		AppConfig.LogExcludePaths = []string{"/callback/pay/score/"}
	}

	if val := os.Getenv("DEBUG_HEADERS"); val != "" {
		AppConfig.DebugHeaders = strings.ToLower(val) == "true"
	}
}

func splitAndTrim(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func IsExcludeApi(api string) bool {
	if api == "" || len(AppConfig.ExcludeApis) == 0 {
		return false
	}
	_, ok := AppConfig.ExcludeApis[api]
	return ok
}

// applyDefaultBusinessExcludeApis adds login/session APIs that must bypass RBAC when API verify is on.
func applyDefaultBusinessExcludeApis() {
	if AppConfig.GatewayMode != "business" || !AppConfig.EnableApiVerify {
		return
	}
	if AppConfig.ExcludeApis == nil {
		AppConfig.ExcludeApis = make(map[string]struct{})
	}
	for _, api := range []string{
		"/business/ebike-management/user/getUserByToken",
		"/oauth/logout",
	} {
		AppConfig.ExcludeApis[api] = struct{}{}
	}
}

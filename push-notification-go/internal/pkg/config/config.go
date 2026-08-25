package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// GlobalConfig holds the merged application configuration.
var GlobalConfig AppConfig

// AppConfig is the top-level configuration structure.
type AppConfig struct {
	Server ServerConfig `yaml:"server"`
	Nacos  NacosConfig  `yaml:"nacos"`
	MySQL  MySQLConfig  `yaml:"mysql"`
	Xyy    XyyConfig    `yaml:"xyy"`
	DryRun bool         `yaml:"dryRun"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port int    `yaml:"port"`
	Name string `yaml:"name"`
}

// NacosConfig holds Nacos connection settings.
type NacosConfig struct {
	ServerAddr string `yaml:"serverAddr"`
	Namespace  string `yaml:"namespace"`
	Group      string `yaml:"group"`
	Port       uint64 `yaml:"port"`
}

// MySQLConfig holds database connection settings.
type MySQLConfig struct {
	DSN          string `yaml:"dsn"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
}

// XyyConfig holds business-specific settings.
type XyyConfig struct {
	Secrets     []string        `yaml:"secrets"`
	MessagePool PoolConfig      `yaml:"messagePool"`
	VoicePool   PoolConfig      `yaml:"voicePool"`
	Chuanglan   ChuanglanConfig `yaml:"chuanglan"`
}

// ChuanglanConfig holds Chuanglan SMS gateway settings.
type ChuanglanConfig struct {
	ApiUrl string `yaml:"apiUrl"`
}

// PoolConfig holds worker pool settings.
type PoolConfig struct {
	MaxWorkers int `yaml:"maxWorkers"`
	QueueSize  int `yaml:"queueSize"`
}

// ---------------------------------------------------------------------------
// Nacos YAML structures (Spring-style format)
// ---------------------------------------------------------------------------

// nacosAppYAML maps the push-notification.yml format from Nacos.
type nacosAppYAML struct {
	Spring struct {
		Xyy struct {
			Secrets           []string `yaml:"secrets"` // YAML list: - secret1\n- secret2
			MessageThreadPool struct {
				MaxPoolSize   int `yaml:"maxPoolSize"`
				QueueCapacity int `yaml:"queueCapacity"`
			} `yaml:"messageThreadPool"`
			VoiceThreadPool struct {
				MaxPoolSize   int `yaml:"maxPoolSize"`
				QueueCapacity int `yaml:"queueCapacity"`
			} `yaml:"voiceThreadPool"`
		} `yaml:"xyy"`
	} `yaml:"spring"`
}

// nacosMySQLYAML maps the mysql.yaml format from Nacos.
type nacosMySQLYAML struct {
	MySQL struct {
		PushNotification struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Database string `yaml:"database"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"push_notification"`
	} `yaml:"mysql"`
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// LoadLocalConfig reads the local YAML file and populates GlobalConfig.
func LoadLocalConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("[config] failed to read local config %s: %v", path, err)
	}
	if err := yaml.Unmarshal(data, &GlobalConfig); err != nil {
		log.Fatalf("[config] failed to parse local config %s: %v", path, err)
	}
	log.Printf("[config] loaded local config from %s", path)
}

// MergeNacosAppConfig parses the push-notification.yml content from Nacos and
// merges relevant fields into GlobalConfig.
func MergeNacosAppConfig(yamlContent string) {
	var nc nacosAppYAML
	if err := yaml.Unmarshal([]byte(yamlContent), &nc); err != nil {
		log.Printf("[config] failed to parse nacos app config: %v", err)
		return
	}

	// Secrets – YAML list in the Nacos config.
	if len(nc.Spring.Xyy.Secrets) > 0 {
		GlobalConfig.Xyy.Secrets = nc.Spring.Xyy.Secrets
	}

	// Message pool
	if nc.Spring.Xyy.MessageThreadPool.MaxPoolSize > 0 {
		GlobalConfig.Xyy.MessagePool.MaxWorkers = nc.Spring.Xyy.MessageThreadPool.MaxPoolSize
	}
	if nc.Spring.Xyy.MessageThreadPool.QueueCapacity > 0 {
		GlobalConfig.Xyy.MessagePool.QueueSize = nc.Spring.Xyy.MessageThreadPool.QueueCapacity
	}

	// Voice pool
	if nc.Spring.Xyy.VoiceThreadPool.MaxPoolSize > 0 {
		GlobalConfig.Xyy.VoicePool.MaxWorkers = nc.Spring.Xyy.VoiceThreadPool.MaxPoolSize
	}
	if nc.Spring.Xyy.VoiceThreadPool.QueueCapacity > 0 {
		GlobalConfig.Xyy.VoicePool.QueueSize = nc.Spring.Xyy.VoiceThreadPool.QueueCapacity
	}

	log.Printf("[config] merged nacos app config: secrets=%d, msgPool=%d/%d, voicePool=%d/%d",
		len(GlobalConfig.Xyy.Secrets),
		GlobalConfig.Xyy.MessagePool.MaxWorkers, GlobalConfig.Xyy.MessagePool.QueueSize,
		GlobalConfig.Xyy.VoicePool.MaxWorkers, GlobalConfig.Xyy.VoicePool.QueueSize,
	)
}

// MergeNacosMySQLConfig parses the mysql.yaml content from Nacos and assembles
// a Go-style DSN into GlobalConfig.
func MergeNacosMySQLConfig(yamlContent string) {
	var nc nacosMySQLYAML
	if err := yaml.Unmarshal([]byte(yamlContent), &nc); err != nil {
		log.Printf("[config] failed to parse nacos mysql config: %v", err)
		return
	}

	m := nc.MySQL.PushNotification
	if m.Host == "" || m.Database == "" {
		log.Printf("[config] nacos mysql config incomplete, skipping DSN assembly")
		return
	}

	port := m.Port
	if port == 0 {
		port = 3306
	}

	GlobalConfig.MySQL.DSN = fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FShanghai",
		m.Username, m.Password, m.Host, port, m.Database,
	)
	log.Printf("[config] merged nacos mysql config: host=%s, db=%s", m.Host, m.Database)
}

// ApplyEnvOverrides reads environment variables and overrides the corresponding
// fields in GlobalConfig. This is the final layer of configuration.
func ApplyEnvOverrides() {
	if v := os.Getenv("MYSQL_DSN"); v != "" {
		GlobalConfig.MySQL.DSN = v
		log.Printf("[config] env override: MYSQL_DSN")
	}
	if v := os.Getenv("NACOS_SERVER_ADDR"); v != "" {
		if strings.Contains(v, ":") {
			parts := strings.Split(v, ":")
			GlobalConfig.Nacos.ServerAddr = parts[0]
			if port, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
				GlobalConfig.Nacos.Port = port
			}
		} else {
			GlobalConfig.Nacos.ServerAddr = v
		}
		log.Printf("[config] env override: NACOS_SERVER_ADDR=%s", v)
	}
	if v := os.Getenv("NACOS_NAMESPACE"); v != "" {
		GlobalConfig.Nacos.Namespace = v
		log.Printf("[config] env override: NACOS_NAMESPACE=%s", v)
	}
	if v := os.Getenv("NACOS_GROUP"); v != "" {
		GlobalConfig.Nacos.Group = v
		log.Printf("[config] env override: NACOS_GROUP=%s", v)
	}
	if v := os.Getenv("XYY_SECRETS"); v != "" {
		GlobalConfig.Xyy.Secrets = splitAndTrim(v)
		log.Printf("[config] env override: XYY_SECRETS (count=%d)", len(GlobalConfig.Xyy.Secrets))
	}
	if v := os.Getenv("DRY_RUN"); v != "" {
		GlobalConfig.DryRun = strings.EqualFold(v, "true") || v == "1"
		log.Printf("[config] env override: DRY_RUN=%v", GlobalConfig.DryRun)
	}
	if v := os.Getenv("CHUANGLAN_API_URL"); v != "" {
		GlobalConfig.Xyy.Chuanglan.ApiUrl = v
		log.Printf("[config] env override: CHUANGLAN_API_URL=%s", v)
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			GlobalConfig.Server.Port = port
			log.Printf("[config] env override: SERVER_PORT=%d", port)
		}
	}
}

// VerifySecret checks whether the given secret is present in the configured
// list of valid secrets.
func VerifySecret(secret string) bool {
	for _, s := range GlobalConfig.Xyy.Secrets {
		if s == secret {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

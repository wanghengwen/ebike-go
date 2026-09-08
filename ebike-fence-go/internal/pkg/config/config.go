package config

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// XyyConfig mirrors Java ApplicationProperties (prefix "xyy").
type XyyConfig struct {
	RemoteLockDistance       *int     `yaml:"remoteLockDistance"`
	HelmetAuditWhite         []string `yaml:"helmetAuditWhite"`
	DirectionIgnoreTenantIds []string `yaml:"directionIgnoreTenantIds"`
	CancelAuthTenantIds      []string `yaml:"cancelAuthTenantIds"`
	Proxy                    struct {
		TargetUrl  string   `yaml:"targetUrl"`
		LiveList   []string `yaml:"liveList"`
		RecordList []string `yaml:"recordList"`
	} `yaml:"proxy"`
}

type AppConfig struct {
	Server struct {
		Port int    `yaml:"port"`
		Name string `yaml:"name"`
	} `yaml:"server"`

	Nacos struct {
		ServerAddr string `yaml:"serverAddr"`
		Namespace  string `yaml:"namespace"`
		Group      string `yaml:"group"`
		DataId     string `yaml:"dataId"`
		Port       uint64 `yaml:"port"`
	} `yaml:"nacos"`

	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`

	MySQL struct {
		DSN          string `yaml:"dsn"`
		MaxOpenConns int    `yaml:"maxOpenConns"`
		MaxIdleConns int    `yaml:"maxIdleConns"`
	} `yaml:"mysql"`

	Xyy XyyConfig `yaml:"xyy"`

	Spring struct {
		Application struct {
			Name string `yaml:"name"`
		} `yaml:"application"`
		Xyy struct {
			Fastid FastIDSettings `yaml:"fastid"`
		} `yaml:"xyy"`
	} `yaml:"spring"`

	// FastID is deprecated; use spring.xyy.fastid.
	FastID FastIDSettings `yaml:"fastid"`

	// Kafka mirrors spring.kafka (topic-prefix + brokers).
	Kafka struct {
		Brokers      []string `yaml:"brokers"`
		TopicPrefix  string   `yaml:"topicPrefix"`
	} `yaml:"kafka"`
}

var GlobalConfig AppConfig

func LoadConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read config file %s: %v", path, err)
	}

	err = yaml.Unmarshal(data, &GlobalConfig)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	log.Printf("Loaded local config successfully. Service: %s, Port: %d", GlobalConfig.Server.Name, GlobalConfig.Server.Port)
	ApplyEnvOverrides()
}

// ApplyEnvOverrides lets NACOS_SERVER_ADDR / NACOS_NAMESPACE / NACOS_GROUP
// override values loaded from the local YAML (highest precedence).
// Must run before creating any Nacos client.
func ApplyEnvOverrides() {
	if v := strings.TrimSpace(os.Getenv("NACOS_SERVER_ADDR")); v != "" {
		host, port := splitHostPort(v, 8848)
		GlobalConfig.Nacos.ServerAddr = host
		GlobalConfig.Nacos.Port = port
	} else if v := strings.TrimSpace(os.Getenv("NACOS_SERVERADDR")); v != "" {
		GlobalConfig.Nacos.ServerAddr = v
	}
	if v := strings.TrimSpace(os.Getenv("NACOS_PORT")); v != "" {
		if port, err := strconv.ParseUint(v, 10, 64); err == nil {
			GlobalConfig.Nacos.Port = port
		}
	}
	if v := strings.TrimSpace(os.Getenv("NACOS_NAMESPACE")); v != "" {
		GlobalConfig.Nacos.Namespace = v
	}
	if v := strings.TrimSpace(os.Getenv("NACOS_GROUP")); v != "" {
		GlobalConfig.Nacos.Group = v
	}
}

func splitHostPort(raw string, defaultPort uint64) (host string, port uint64) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", defaultPort
	}
	if idx := strings.Index(raw, "://"); idx >= 0 {
		raw = raw[idx+3:]
	}
	if h, p, err := net.SplitHostPort(raw); err == nil {
		portNum, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			return h, defaultPort
		}
		return h, portNum
	}
	return raw, defaultPort
}

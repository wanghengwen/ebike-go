package config

import (
	"log"
	"os"

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
}

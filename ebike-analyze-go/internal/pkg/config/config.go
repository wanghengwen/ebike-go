package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// ESConfig holds Elasticsearch connection settings.
type ESConfig struct {
	Hostname           string `yaml:"hostname"`
	Username           string `yaml:"username"`
	Password           string `yaml:"password"`
	ConnectTimeout     int    `yaml:"connectTimeout"`
	SocketTimeout      int    `yaml:"socketTimeout"`
	TrackTotalHitsUpTo int    `yaml:"trackTotalHitsUpTo"`
}

// XyyConfig mirrors Java ApplicationProperties (prefix "xyy").
type XyyConfig struct {
	TableSplit struct {
		Enable     bool     `yaml:"enable"`
		TableNames []string `yaml:"tableNames"`
	} `yaml:"tableSplit"`
	Proxy struct {
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
		AnalyzeDSN   string `yaml:"analyzeDSN"`
		VisualDSN    string `yaml:"visualDSN"`
		MaxOpenConns int    `yaml:"maxOpenConns"`
		MaxIdleConns int    `yaml:"maxIdleConns"`
	} `yaml:"mysql"`

	ES ESConfig `yaml:"es"`

	Xyy XyyConfig `yaml:"xyy"`

	Spring struct {
		Application struct {
			Name string `yaml:"name"`
		} `yaml:"application"`
	} `yaml:"spring"`
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

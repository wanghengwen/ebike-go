package config

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Server struct {
		Port int    `yaml:"port"`
		Name string `yaml:"name"`
	} `yaml:"server"`

	Nacos struct {
		ServerAddr       string   `yaml:"serverAddr"`
		Namespace        string   `yaml:"namespace"`
		Group            string   `yaml:"group"`
		DataId           string   `yaml:"dataId"`
		Port             uint64   `yaml:"port"`
		ContextPath      string   `yaml:"contextPath"` // MSE Nacos: /nacos
		ExtensionDataIds []string `yaml:"extensionDataIds"`
		ExtensionGroups  []string `yaml:"extensionGroups"`
	} `yaml:"nacos"`

	// Xyy mirrors the Java ApplicationProperties (prefix "xyy").
	Xyy struct {
		MapServiceConfig struct {
			Url    string `yaml:"url"`
			Secret string `yaml:"secret"`
		} `yaml:"mapServiceConfig"`
		GatewayFilter struct {
			Exclude string `yaml:"exclude"`
		} `yaml:"gatewayfilter"`
	} `yaml:"xyy"`

	// Gaode mirrors gaode.navigate.url from Nacos ebike-service-client.yml.
	Gaode struct {
		Navigate struct {
			Url string `yaml:"url"`
		} `yaml:"navigate"`
	} `yaml:"gaode"`

	// Proxy configures Java BFF reverse proxy and live/record/shadow routing.
	Proxy ProxyConfig `yaml:"proxy"`
}

// ProxyConfig mirrors ebike-device-worker-go proxy settings.
type ProxyConfig struct {
	TargetURL  string   `yaml:"target_url"`
	LiveList   []string `yaml:"live_list"`
	RecordList []string `yaml:"record_list"`
}

var GlobalConfig AppConfig

// LoadConfig reads the local configuration file (e.g. conf/application.yml)
func LoadConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read config file %s: %v", path, err)
	}

	err = yaml.Unmarshal(data, &GlobalConfig)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	ApplyNacosEnvOverrides()
	ApplyProxyEnvOverrides()
	loadProxyLists(filepath.Join(filepath.Dir(path), "proxy_lists.yml"))
	log.Printf("Config loaded: service=%s port=%d proxy=%q live=%d record=%d",
		GlobalConfig.Server.Name,
		GlobalConfig.Server.Port,
		GlobalConfig.Proxy.TargetURL,
		len(GlobalConfig.Proxy.LiveList),
		len(GlobalConfig.Proxy.RecordList),
	)
}

// ApplyProxyEnvOverrides applies PROXY_TARGET_URL over yaml proxy.target_url.
func ApplyProxyEnvOverrides() {
	if v := os.Getenv("PROXY_TARGET_URL"); v != "" {
		GlobalConfig.Proxy.TargetURL = v
	}
}

func loadProxyLists(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var lists struct {
		LiveList   []string `yaml:"live_list"`
		RecordList []string `yaml:"record_list"`
	}
	if err := yaml.Unmarshal(data, &lists); err != nil {
		log.Printf("Warning: failed to parse %s: %v", path, err)
		return
	}
	if len(lists.LiveList) > 0 {
		GlobalConfig.Proxy.LiveList = lists.LiveList
	}
	if len(lists.RecordList) > 0 {
		GlobalConfig.Proxy.RecordList = lists.RecordList
	}
}

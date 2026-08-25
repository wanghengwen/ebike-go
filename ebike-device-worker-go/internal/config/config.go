package config

import (
	"log"
	"strconv"
	"strings"

	"ebike-device-worker-go/internal/pkg/fastid"
	"ebike-device-worker-go/internal/pkg/utils"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"server"`
	Nacos struct {
		ServerAddr      string `mapstructure:"serverAddr"`
		Port            uint64 `mapstructure:"port"`
		Namespace       string `mapstructure:"namespace"`
		Group           string `mapstructure:"group"`
		RegisterEnabled bool   `mapstructure:"registerEnabled"`
	} `mapstructure:"nacos"`
	Proxy struct {
		TargetUrl  string   `mapstructure:"target_url"`
		LiveList   []string `mapstructure:"live_list"`
		RecordList []string `mapstructure:"record_list"`
	} `mapstructure:"proxy"`
	Database struct {
		Dsn string `mapstructure:"dsn"`
	} `mapstructure:"database"`
	Redis struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
		Prefix   string `mapstructure:"prefix"`
		DB       int    `mapstructure:"db"`
	} `mapstructure:"redis"`
	Kafka struct {
		Brokers     []string `mapstructure:"brokers"`
		TopicPrefix string   `mapstructure:"topic_prefix"`
		Topics      struct {
			DataTopic    string `mapstructure:"data_topic"`
			EventTopic   string `mapstructure:"event_topic"`
			AlarmTopic   string `mapstructure:"alarm_topic"`
			ToSaasTopic  string `mapstructure:"to_saas_topic"`
			GrayToSaas   string `mapstructure:"gray_to_saas_topic"`
		} `mapstructure:"topics"`
		Consumers struct {
			GroupData         string `mapstructure:"group_data"`
			GroupEvent        string `mapstructure:"group_event"`
			GroupAlarm        string `mapstructure:"group_alarm"`
			DataConcurrency   int    `mapstructure:"data_concurrency"`
			EventConcurrency  int    `mapstructure:"event_concurrency"`
			AlarmConcurrency  int    `mapstructure:"alarm_concurrency"`
			BatchSize         int    `mapstructure:"batch_size"`
			BatchWaitMs       int    `mapstructure:"batch_wait_ms"`
			StartOffset       string `mapstructure:"start_offset"` // earliest|latest, only when group has no committed offset
		} `mapstructure:"consumers"`
		Enabled     bool `mapstructure:"enabled"`
		PushEnabled bool `mapstructure:"push_enabled"` // forward to to-saas / gray-to-saas; disable on verify consumer
	} `mapstructure:"kafka"`
	DbConfig struct {
		// Reserved for compatibility with Java Nacos; Go always writes PostgreSQL.
		WriteType int `mapstructure:"write_type"`
	} `mapstructure:"db_config"`
	Spring struct {
		Application struct {
			Name string `mapstructure:"name"`
		} `mapstructure:"application"`
		Xyy struct {
			Fastid FastIDSettings `mapstructure:"fastid"`
			Gray   struct {
				TenantIds []string `mapstructure:"tenant-ids"`
			} `mapstructure:"gray"`
		} `mapstructure:"xyy"`
	} `mapstructure:"spring"`
	// FastID is deprecated; use spring.xyy.fastid. Kept for backward-compatible flat config.
	FastID FastIDSettings `mapstructure:"fastid"`
	PersistConfig struct {
		PersistBatchSize  int    `mapstructure:"persist_batch_size"`
		PersistLimitSize  int    `mapstructure:"persist_limit_size"`
		PersistIntervalMs int    `mapstructure:"persist_interval_ms"`
		TableSuffix       string `mapstructure:"table_suffix"` // e.g. "_new"; empty = legacy ebike_gps
	} `mapstructure:"persist_config"`
	Caffeine struct {
		DeviceTenantMappingCacheSize int `mapstructure:"device_tenant_mapping_cache_size"`
		TenantCacheSize              int `mapstructure:"tenant_cache_size"`
		TimeoutSec                   int `mapstructure:"timeout_sec"`
	} `mapstructure:"caffeine"`
	Xyy struct {
		GrayTenantIds []string `mapstructure:"gray_tenant_ids"`
	} `mapstructure:"xyy"`
	Paas struct {
		BaseURL string `mapstructure:"base_url"`
		Enabled bool   `mapstructure:"enabled"`
	} `mapstructure:"paas"`
}

// FastIDSettings mirrors spring.xyy.fastid in Java Nacos fast_id.yaml.
type FastIDSettings struct {
	Enabled                  bool   `mapstructure:"enabled" yaml:"enabled"`
	ServerAddr               string `mapstructure:"server-addr" yaml:"server-addr"`
	URL                      string `mapstructure:"url" yaml:"url"`
	Namespace                string `mapstructure:"namespace" yaml:"namespace"`
	GroupID                  string `mapstructure:"groupId" yaml:"groupId"`
	Secret                   string `mapstructure:"secret" yaml:"secret"`
	AppName                  string `mapstructure:"app-name" yaml:"app-name"`
	Port                     int    `mapstructure:"port" yaml:"port"`
	InstanceNoLocalDirectory string `mapstructure:"instance-no-local-directory" yaml:"instance-no-local-directory"`
	MachineUUID              string `mapstructure:"machine-uuid" yaml:"machine-uuid"`
	DriftTime                int    `mapstructure:"drift-time" yaml:"drift-time"`
	UseHTTPS                 bool   `mapstructure:"use-https" yaml:"use-https"`
	UseHTTPSExplicit         bool   `mapstructure:"-" yaml:"-"`
}

var GlobalConfig *Config

const defaultConfigPath = "conf/application.yaml"

// DefaultConfigPath returns the default bootstrap config file path.
func DefaultConfigPath() string {
	return defaultConfigPath
}

// LoadConfig reads the local bootstrap file (default conf/application.yaml).
func LoadConfig(configPath string) {
	if configPath == "" {
		configPath = defaultConfigPath
	}

	viper.Reset()
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	viper.SetDefault("server.port", "8080")
	viper.SetDefault("proxy.target_url", "http://127.0.0.1:10080")
	viper.SetDefault("kafka.enabled", false)
	viper.SetDefault("kafka.push_enabled", true)
	viper.SetDefault("kafka.consumers.batch_size", 500)
	viper.SetDefault("kafka.consumers.batch_wait_ms", 100)
	viper.SetDefault("kafka.consumers.start_offset", "latest")
	viper.SetDefault("kafka.consumers.data_concurrency", 1)
	viper.SetDefault("spring.application.name", "ebike-device-worker")
	viper.SetDefault("spring.xyy.fastid.enabled", true)
	viper.SetDefault("spring.xyy.fastid.server-addr", "fastid.luopingtech.com")
	viper.SetDefault("spring.xyy.fastid.namespace", "prod")
	viper.SetDefault("spring.xyy.fastid.groupId", "xyy")
	viper.SetDefault("spring.xyy.fastid.drift-time", 10)
	viper.SetDefault("persist_config.persist_batch_size", 20)
	viper.SetDefault("persist_config.persist_limit_size", 50)
	viper.SetDefault("caffeine.device_tenant_mapping_cache_size", 100000)
	viper.SetDefault("caffeine.tenant_cache_size", 2000)
	viper.SetDefault("caffeine.timeout_sec", 120)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config file %s: %v", configPath, err)
	}
	log.Printf("loaded config from %s", configPath)

	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		log.Fatalf("Error unmarshaling config: %v", err)
	}
	normalizeConfig(GlobalConfig)
	// Env overrides run in FinalizeConfig after Nacos (see nacos_bootstrap.go).
}

func normalizeConfig(c *Config) {
	if len(c.Xyy.GrayTenantIds) == 0 && len(c.Spring.Xyy.Gray.TenantIds) > 0 {
		c.Xyy.GrayTenantIds = c.Spring.Xyy.Gray.TenantIds
	}
}

// FastIDEnabled reports whether FastId client should start.
func (c *Config) FastIDEnabled() bool {
	_, enabled := c.resolvedFastIDSettings()
	return enabled
}

// ResolvedFastID builds fastid.Config from spring.xyy.fastid (Nacos-aligned).
func (c *Config) ResolvedFastID() fastid.Config {
	cfg, _ := c.resolvedFastIDSettings()
	return cfg
}

func (c *Config) resolvedFastIDSettings() (fastid.Config, bool) {
	src := c.Spring.Xyy.Fastid
	if !src.Enabled && c.FastID.Enabled {
		src = c.FastID
	}
	if src.ServerAddr == "" && src.URL == "" && c.FastID.URL != "" {
		src = c.FastID
	}
	if !src.Enabled && !c.FastID.Enabled && (src.ServerAddr != "" || c.FastID.URL != "") {
		src.Enabled = true
	}

	url := firstNonEmpty(src.ServerAddr, src.URL, c.FastID.ServerAddr, c.FastID.URL)
	appName := resolveAppName(src.AppName, c.Spring.Application.Name)
	namespace := firstNonEmpty(src.Namespace, c.FastID.Namespace)
	groupID := firstNonEmpty(src.GroupID, c.FastID.GroupID)
	secret := firstNonEmpty(src.Secret, c.FastID.Secret)

	port := src.Port
	if port == 0 {
		port = c.FastID.Port
	}
	if port == 0 {
		if p, err := strconv.Atoi(c.Server.Port); err == nil {
			port = p
		}
	}
	if port == 0 {
		port = 8080
	}

	localDir := firstNonEmpty(src.InstanceNoLocalDirectory, c.FastID.InstanceNoLocalDirectory)
	machineUUID := firstNonEmpty(src.MachineUUID, c.FastID.MachineUUID)
	drift := src.DriftTime
	if drift == 0 {
		drift = c.FastID.DriftTime
	}
	if drift == 0 {
		drift = 10
	}

	enabled := src.Enabled || c.FastID.Enabled

	return fastid.Config{
		URL:                      url,
		Namespace:                namespace,
		GroupID:                  groupID,
		AppName:                  appName,
		Port:                     port,
		Secret:                   secret,
		InstanceNoLocalDirectory: localDir,
		MachineUUID:              machineUUID,
		DriftTime:                drift,
		UseHTTPS:                 src.UseHTTPS,
		UseHTTPSExplicit:         src.UseHTTPSExplicit,
	}, enabled
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func (c *Config) IsGrayTenant(tenantID int) bool {
	for _, id := range c.Xyy.GrayTenantIds {
		if id == utils.Itoa(tenantID) {
			return true
		}
	}
	return false
}

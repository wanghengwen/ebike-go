package config

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"ebike-device-openapi-go/internal/pkg/logger"
)

// AppConfig is the full application configuration structure.
// Fields map to Nacos YAML keys. Any field that can change at runtime is read
// through the atomic pointer below to prevent data races.
type AppConfig struct {
	Xyy struct {
		Bin2UpdateHost bool   `yaml:"bin2UpdateHost"`
		Cloud          string `yaml:"cloud"`
		Router         struct {
			RouterSwitch       bool     `yaml:"routerSwitch"`
			AnvelinkOpenapiUrl string   `yaml:"anvelinkOpenapiUrl"`
			ExcludeUrls        []string `yaml:"excludeUrls"`
		} `yaml:"router"`
	} `yaml:"xyy"`
	TimeoutConfig struct {
		ConnectTimeoutMs int `yaml:"connectTimeoutMs"`
		ReadTimeoutMs    int `yaml:"readTimeoutMs"`
	} `yaml:"timeout-config"`
	Spring struct {
		Kafka struct {
			BootstrapServers string `yaml:"bootstrap-servers"`
			Switch           bool   `yaml:"switch"`
			Producer         struct {
				DataTopic  string `yaml:"data-topic"`
				AlarmTopic string `yaml:"alarm-topic"`
				EventTopic string `yaml:"event-topic"`
			} `yaml:"producer"`
		} `yaml:"kafka"`
		Redis struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Password string `yaml:"password"`
			Database int    `yaml:"database"`
			// DeviceDatabase is the Redis DB for consume/paas device cache
			// (device_info_*, imei_car_*, offline_time_ex_*). -1 means same as Database.
			// Prod often keeps ecu_login on DB0 and device_info on DB4.
			DeviceDatabase int `yaml:"device-database"`
		} `yaml:"redis"`
	} `yaml:"spring"`
	// Luoping MQTT (EMQX) access — type stored in Redis is "luoping".
	// 收发通过内置 MQTT 客户端 + 共享订阅完成（不再走 EMQX HTTP API/Webhook）。
	Mqtt struct {
		Enabled          bool `yaml:"enabled"`
		PublishQos       int  `yaml:"publishQos"`
		SubscribeQos     int  `yaml:"subscribeQos"`
		OnlineTTLSeconds int  `yaml:"onlineTtlSeconds"` // lastSeen stale threshold; default 1800s (~3x 10m idle heartbeat)
		SyncWaitMs       int  `yaml:"syncWaitMs"`       // sync cmd wait for rsp; 0 → use readTimeout

		// 上行消息处理背压：有界 worker 池 + 单条处理超时。
		WorkerPoolSize   int `yaml:"workerPoolSize"`   // 并发处理上行消息的 worker 数；默认 32
		WorkerQueueSize  int `yaml:"workerQueueSize"`  // 有界队列长度，满则接收侧阻塞形成背压；默认 256
		MessageTimeoutMs int `yaml:"messageTimeoutMs"` // 单条上行消息处理超时；默认 3000ms

		// MQTT broker 连接（支持逗号分隔多地址），如 tcp://emqx:1883 或 ssl://emqx:8883。
		Broker   string `yaml:"broker"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`

		// clientId = ClientIdPrefix + 机器名（K8s Pod 名）。
		ClientIdPrefix   string `yaml:"clientIdPrefix"`
		SharedGroup      string `yaml:"sharedGroup"` // 共享订阅组名，$share/{group}/...
		KeepAliveSeconds int    `yaml:"keepAliveSeconds"`
		CleanSession     bool   `yaml:"cleanSession"`

		// 上行订阅主题（会自动加 $share/{group}/ 前缀）。
		RptTopic string `yaml:"rptTopic"` // 默认 ecu/brpt/ebike/#（二进制上报+应答）
		RspTopic string `yaml:"rspTopic"` // deprecated: 应答已并入 brpt，可留空

		// 可选：EMQX 下线事件主题（自动加 $share/{group}/ 前缀）。
		// ConnectedTopic 已废弃：不再订阅/处理 connected；上线仅由透传上行驱动。
		// DisconnectedTopic 形如 $SYS/brokers/+/clients/+/disconnected；留空则仅靠
		// lastSeen ZSET 扫描离线。注意：$SYS 订阅需在 EMQX ACL 放开本客户端。
		ConnectedTopic    string `yaml:"connectedTopic"` // deprecated: ignored (not subscribed)
		DisconnectedTopic string `yaml:"disconnectedTopic"`

		// Deprecated: 旧的 EMQX HTTP API 配置，保留以兼容历史 yaml，当前已不使用。
		Emqx struct {
			APIBase   string `yaml:"apiBase"`
			APIKey    string `yaml:"apiKey"`
			APISecret string `yaml:"apiSecret"`
		} `yaml:"emqx"`
	} `yaml:"mqtt"`
}

// atomicConfig holds a pointer to AppConfig. Using atomic.Value for concurrent-safe
// reads during dynamic Nacos config refresh without holding a mutex in hot paths.
//
// NOTE: We use unsafe.Pointer + atomic operations because atomic.Value requires the
// stored type to be exactly the same every Store() call — *AppConfig satisfies this.
var (
	atomicConfig   unsafe.Pointer
	namingClient   naming_client.INamingClient
	httpServerPort int32 = 8080
)

// SetHTTPServerPort records the bound HTTP listen port (Java SpringUtil.getHttpServerPort()).
func SetHTTPServerPort(port int) {
	if port > 0 {
		atomic.StoreInt32(&httpServerPort, int32(port))
	}
}

// GetHTTPServerPort returns the HTTP port used for WildCmd callback addressing.
// Falls back to SERVER_PORT env, then 8080.
func GetHTTPServerPort() int {
	if p := atomic.LoadInt32(&httpServerPort); p > 0 {
		return int(p)
	}
	if env := os.Getenv("SERVER_PORT"); env != "" {
		if v, err := strconv.Atoi(env); err == nil && v > 0 {
			return v
		}
	}
	return 8080
}

// GetConfig returns the latest loaded application configuration.
// Callers should not cache the returned pointer for long periods,
// as Nacos reloads will generate a new AppConfig object.
// This is safe to call concurrently from any goroutine.
func GetConfig() *AppConfig {
	p := atomic.LoadPointer(&atomicConfig)
	if p == nil {
		// Return a default config rather than panicking if not initialized.
		cfg := defaultConfig()
		return cfg
	}
	return (*AppConfig)(p)
}

// storeConfig atomically replaces the current configuration.
func storeConfig(cfg *AppConfig) {
	atomic.StorePointer(&atomicConfig, unsafe.Pointer(cfg))
}

// Init loads application configuration from Nacos and environment variables.
// The function is designed to be called once at startup.
func Init() {
	// Layer 1: built-in defaults + optional local bootstrap file (conf/application.yaml).
	cfg := defaultConfig()
	loadLocalConfig(cfg)

	// NACOS_DISABLED: run with local config + env overrides only (local dev / no Nacos).
	if os.Getenv("NACOS_DISABLED") == "true" {
		applyEnvOverrides(cfg)
		storeConfig(cfg)
		logger.Log.Info("NACOS_DISABLED=true; using local config + env overrides only",
			zap.Bool("mqttEnabled", cfg.Mqtt.Enabled),
			zap.String("redisHost", cfg.Spring.Redis.Host),
		)
		return
	}

	namespaceId := os.Getenv("NACOS_NAMESPACE")
	if namespaceId == "" {
		namespaceId = "prod"
	}

	nacosAddr := os.Getenv("NACOS_SERVER_ADDR")
	if nacosAddr == "" {
		nacosAddr = "127.0.0.1:8848"
	}

	parts := strings.Split(nacosAddr, ":")
	host := parts[0]
	port := uint64(8848)
	if len(parts) > 1 {
		p, _ := strconv.ParseUint(parts[1], 10, 64)
		if p > 0 {
			port = p
		}
	}

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(host, port, constant.WithContextPath("/nacos")),
	}
	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(namespaceId),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir(os.TempDir()+"/nacos/log"),
		constant.WithCacheDir(os.TempDir()+"/nacos/cache"),
		constant.WithLogLevel("error"),
	)

	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		logger.Log.Warn("failed to connect to Nacos config client, using local config + env", zap.Error(err))
		applyEnvOverrides(cfg)
		storeConfig(cfg)
		return
	}
	logger.Log.Info("Nacos config client initialized")

	// Layers: defaults+local (cfg) → Nacos overrides → env overrides (highest).
	cfg = loadFromNacos(configClient, cfg)
	applyEnvOverrides(cfg)
	storeConfig(cfg)

	// Register dynamic listeners so that configuration changes in Nacos are applied
	// at runtime without restarting the pod.
	registerNacosListeners(configClient)

	logger.Log.Info("loaded config",
		zap.String("cloud", cfg.Xyy.Cloud),
		zap.Bool("bin2UpdateHost", cfg.Xyy.Bin2UpdateHost),
		zap.Bool("kafkaSwitch", cfg.Spring.Kafka.Switch),
		zap.String("redisHost", cfg.Spring.Redis.Host),
		zap.Int("redisPort", cfg.Spring.Redis.Port),
		zap.Bool("mqttEnabled", cfg.Mqtt.Enabled),
		zap.String("mqttBroker", GetMqttBroker()),
	)

	// --- ADDED: Register Nacos Naming Client (Service Discovery) ---
	var errNaming error
	namingClient, errNaming = clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if errNaming == nil {
		groupName := os.Getenv("NACOS_GROUP")
		if groupName == "" {
			groupName = "xyy"
		}
		success, _ := namingClient.RegisterInstance(vo.RegisterInstanceParam{
			Ip:          GetLocalIP(),
			Port:        uint64(GetHTTPServerPort()),
			ServiceName: "ebike-device-openapi",
			GroupName:   groupName,
			Enable:      true,
			Healthy:     true,
			Weight:      1.0,
			Ephemeral:   true,
		})
		if success {
			logger.Log.Info("Nacos naming client registered", zap.String("ip", GetLocalIP()), zap.Int("port", GetHTTPServerPort()))
		} else {
			logger.Log.Warn("Nacos naming client registration failed")
		}
	} else {
		logger.Log.Warn("failed to connect to Nacos naming client", zap.Error(errNaming))
	}
}

// DeregisterNacos deregisters the current instance from Nacos
func DeregisterNacos() {
	if namingClient != nil {
		groupName := os.Getenv("NACOS_GROUP")
		if groupName == "" {
			groupName = "xyy"
		}
		success, err := namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          GetLocalIP(),
			Port:        uint64(GetHTTPServerPort()),
			ServiceName: "ebike-device-openapi",
			GroupName:   groupName,
			Ephemeral:   true,
		})
		if success {
			logger.Log.Info("successfully deregistered from Nacos")
		} else {
			logger.Log.Warn("failed to deregister from Nacos", zap.Error(err))
		}
	}
}

func GetLocalIP() string {
	// 优先使用 K8s Downward API 注入的 POD_IP 环境变量
	// 这是最可靠的方式，IP 由 K8s 直接提供，不受 CNI 虚拟网卡干扰
	if podIP := os.Getenv("POD_IP"); podIP != "" {
		return podIP
	}

	// Fallback: 扫描网卡，只取私有地址（与 Java InetAddress.isSiteLocalAddress 一致）
	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipnet.IP.To4()
			if ip == nil {
				continue
			}
			if isSiteLocal(ip) {
				return ip.String()
			}
		}
	}
	return "127.0.0.1"
}

// isSiteLocal checks if an IPv4 address is a private/site-local address,
// equivalent to Java's InetAddress.isSiteLocalAddress().
// Matches: 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16
func isSiteLocal(ip net.IP) bool {
	ip = ip.To4()
	if ip == nil {
		return false
	}
	return ip[0] == 10 ||
		(ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31) ||
		(ip[0] == 192 && ip[1] == 168)
}

// defaultConfig returns a config struct populated with safe production defaults.
func defaultConfig() *AppConfig {
	cfg := &AppConfig{}
	cfg.Xyy.Bin2UpdateHost = true
	cfg.Xyy.Cloud = "aliyun"
	cfg.Xyy.Router.RouterSwitch = true
	cfg.Xyy.Router.AnvelinkOpenapiUrl = "http://127.0.0.1:8081"
	cfg.Xyy.Router.ExcludeUrls = []string{}
	cfg.Spring.Kafka.Switch = true
	cfg.Spring.Kafka.Producer.DataTopic = "dataTopic"
	cfg.Spring.Kafka.Producer.AlarmTopic = "alarmTopic"
	cfg.Spring.Kafka.Producer.EventTopic = "eventTopic"
	// Redis defaults (will be overridden by Nacos/env)
	cfg.Spring.Redis.Host = "127.0.0.1"
	cfg.Spring.Redis.Port = 6379
	cfg.Spring.Redis.Password = ""
	cfg.Spring.Redis.Database = 0
	cfg.Spring.Redis.DeviceDatabase = -1 // same as Database unless overridden
	cfg.TimeoutConfig.ConnectTimeoutMs = 2000
	cfg.TimeoutConfig.ReadTimeoutMs = 15000
	cfg.Mqtt.Enabled = false
	cfg.Mqtt.PublishQos = 0
	cfg.Mqtt.SubscribeQos = 0
	cfg.Mqtt.OnlineTTLSeconds = 1800
	cfg.Mqtt.SyncWaitMs = 0
	cfg.Mqtt.ClientIdPrefix = "ebike-device-openapi-"
	cfg.Mqtt.SharedGroup = "openapi"
	cfg.Mqtt.KeepAliveSeconds = 30
	cfg.Mqtt.CleanSession = true
	cfg.Mqtt.RptTopic = "ecu/brpt/ebike/#" // V8 二进制上报+应答
	cfg.Mqtt.RspTopic = ""                 // V8 应答走 brpt，无需单独 rsp
	cfg.Mqtt.WorkerPoolSize = 32
	cfg.Mqtt.WorkerQueueSize = 256
	cfg.Mqtt.MessageTimeoutMs = 3000
	return cfg
}

// GetMqttOnlineTTL returns how long a luoping device may stay "online" without uplink.
// Default 1800s ≈ 3× idle heartbeat (10m).
func GetMqttOnlineTTL() time.Duration {
	sec := GetConfig().Mqtt.OnlineTTLSeconds
	if sec <= 0 {
		sec = 1800
	}
	return time.Duration(sec) * time.Second
}

// GetMqttBroker returns the MQTT broker address(es).
// Priority: env MQTT_BROKER > Nacos/config file.
// Unresolved ${ENV:default} leftovers (e.g. from a local YAML that skipped
// placeholder expansion) are resolved here so paho never sees a literal "${...}".
func GetMqttBroker() string {
	if v := os.Getenv("MQTT_BROKER"); v != "" {
		return v
	}
	return resolveConfigString(GetConfig().Mqtt.Broker)
}

// GetMqttUsername returns the MQTT username.
// Priority: env MQTT_USERNAME (highest) > Nacos > config file.
func GetMqttUsername() string {
	if v := os.Getenv("MQTT_USERNAME"); v != "" {
		return v
	}
	return resolveConfigString(GetConfig().Mqtt.Username)
}

// GetMqttPassword returns the MQTT password.
// Priority: env MQTT_PASSWORD (highest) > Nacos > config file.
func GetMqttPassword() string {
	if v, ok := os.LookupEnv("MQTT_PASSWORD"); ok {
		return v
	}
	return resolveConfigString(GetConfig().Mqtt.Password)
}

// resolveConfigString expands leftover ${key:default} / ${key} placeholders in a
// single config value. Safe for plain values (returned unchanged).
func resolveConfigString(v string) string {
	if v == "" || !strings.Contains(v, "${") {
		return v
	}
	return resolvePlaceholders(v, nil)
}

// GetMqttClientID builds the MQTT clientId as {prefix}{hostname}. In K8s the
// hostname is the Pod name, guaranteeing per-Pod uniqueness. Falls back to
// HOSTNAME env then the local IP.
func GetMqttClientID() string {
	prefix := GetConfig().Mqtt.ClientIdPrefix
	if prefix == "" {
		prefix = "ebike-device-openapi-"
	}
	host, _ := os.Hostname()
	if host == "" {
		host = os.Getenv("HOSTNAME")
	}
	if host == "" {
		host = GetLocalIP()
	}
	return prefix + host
}

// GetMqttSyncWait returns how long sync luoping commands wait for rsp.
func GetMqttSyncWait() time.Duration {
	ms := GetConfig().Mqtt.SyncWaitMs
	if ms > 0 {
		return time.Duration(ms) * time.Millisecond
	}
	return GetReadTimeout()
}

// GetConnectTimeout returns HTTP connect timeout (default 2s, matches Java RestTemplateConfig).
func GetConnectTimeout() time.Duration {
	ms := GetConfig().TimeoutConfig.ConnectTimeoutMs
	if ms <= 0 {
		ms = 2000
	}
	return time.Duration(ms) * time.Millisecond
}

// GetReadTimeout returns HTTP read timeout (default 15s, matches Java RestTemplateConfig).
func GetReadTimeout() time.Duration {
	ms := GetConfig().TimeoutConfig.ReadTimeoutMs
	if ms <= 0 {
		ms = 15000
	}
	return time.Duration(ms) * time.Millisecond
}

// resolvePlaceholders replaces ${key:default} or ${key} with values from props map or environment variables.
func resolvePlaceholders(content string, props map[string]string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		inner := match[2 : len(match)-1]
		parts := strings.SplitN(inner, ":", 2)
		key := strings.TrimSpace(parts[0])

		if val, ok := props[key]; ok && val != "" {
			return val
		}

		// Fallback to environment variables
		envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		envKey = strings.ReplaceAll(envKey, "-", "_")
		if val := os.Getenv(envKey); val != "" {
			return val
		}

		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
		return match
	})
}

// flattenMap recursively flattens a map into dot-separated keys.
func flattenMap(prefix string, v interface{}, dest map[string]string) {
	switch m := v.(type) {
	case map[string]interface{}:
		for k, val := range m {
			newKey := k
			if prefix != "" {
				newKey = prefix + "." + k
			}
			flattenMap(newKey, val, dest)
		}
	case map[interface{}]interface{}:
		for k, val := range m {
			newKey := fmt.Sprintf("%v", k)
			if prefix != "" {
				newKey = prefix + "." + newKey
			}
			flattenMap(newKey, val, dest)
		}
	default:
		dest[prefix] = fmt.Sprintf("%v", v)
	}
}

// loadLocalConfig merges an optional local bootstrap YAML onto cfg. The path is
// CONFIG_PATH, defaulting to conf/application.yaml. A missing file is tolerated
// (defaults still apply). This is the base layer below Nacos and env overrides.
// ${ENV:default} placeholders are resolved here so baked-in k8s profiles work
// even when Nacos has no mqtt/cloud overrides.
func loadLocalConfig(cfg *AppConfig) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "conf/application.yaml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Log.Info("local config not loaded; using built-in defaults",
			zap.String("path", path), zap.Error(err))
		return
	}
	resolved := resolvePlaceholders(string(data), nil)
	if err := yaml.Unmarshal([]byte(resolved), cfg); err != nil {
		logger.Log.Warn("failed to parse local config", zap.String("path", path), zap.Error(err))
		return
	}
	logger.Log.Info("loaded local config", zap.String("path", path),
		zap.Bool("mqttEnabled", cfg.Mqtt.Enabled),
		zap.String("mqttBroker", cfg.Mqtt.Broker),
	)
}

// applyEnvOverrides applies environment variables as the highest-priority layer,
// overriding both local file and Nacos values. Only ops-relevant keys are mapped.
// MQTT credentials are additionally honored at read time via GetMqtt* getters.
func applyEnvOverrides(cfg *AppConfig) {
	if v := os.Getenv("CLOUD"); v != "" {
		cfg.Xyy.Cloud = v
	}
	// Redis
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.Spring.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			cfg.Spring.Redis.Port = p
		}
	}
	if v, ok := os.LookupEnv("REDIS_PASSWORD"); ok {
		cfg.Spring.Redis.Password = v
	}
	if v := os.Getenv("REDIS_DATABASE"); v != "" {
		if d, err := strconv.Atoi(v); err == nil {
			cfg.Spring.Redis.Database = d
		}
	}
	if v := os.Getenv("REDIS_DEVICE_DATABASE"); v != "" {
		if d, err := strconv.Atoi(v); err == nil {
			cfg.Spring.Redis.DeviceDatabase = d
		}
	}
	// Kafka
	if v := os.Getenv("KAFKA_BOOTSTRAP_SERVERS"); v != "" {
		cfg.Spring.Kafka.BootstrapServers = v
	}
	// MQTT
	if v := os.Getenv("MQTT_BROKER"); v != "" {
		cfg.Mqtt.Broker = v
	}
	if v := os.Getenv("MQTT_USERNAME"); v != "" {
		cfg.Mqtt.Username = v
	}
	if v, ok := os.LookupEnv("MQTT_PASSWORD"); ok {
		cfg.Mqtt.Password = v
	}
	if v := os.Getenv("MQTT_ENABLED"); v != "" {
		cfg.Mqtt.Enabled = strings.EqualFold(v, "true")
	}
}

// loadFromNacos fetches configuration from Nacos data IDs and merges them
// into the provided base config. Returns the merged config.
func loadFromNacos(client config_client.IConfigClient, base *AppConfig) *AppConfig {
	props := make(map[string]string)

	// Fetch redis.yaml first to resolve placeholders
	redisContent, err := client.GetConfig(vo.ConfigParam{
		DataId: "redis.yaml",
		Group:  "xyy_ops",
	})
	if err == nil && redisContent != "" {
		var redisMap map[string]interface{}
		if parseErr := yaml.Unmarshal([]byte(redisContent), &redisMap); parseErr == nil {
			flattenMap("", redisMap, props)
		}
	}

	// DataID: ebike-device-openapi.yml — application-specific config
	appGroup := os.Getenv("NACOS_GROUP")
	if appGroup == "" {
		appGroup = "xyy"
	}
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: "ebike-device-openapi.yml",
		Group:  appGroup,
	})
	if err == nil && content != "" {
		resolvedContent := resolvePlaceholders(content, props)
		if parseErr := yaml.Unmarshal([]byte(resolvedContent), base); parseErr != nil {
			logger.Log.Warn("failed to parse ebike-device-openapi.yml", zap.Error(parseErr))
		}
	} else if err != nil {
		logger.Log.Warn("cannot fetch ebike-device-openapi.yml from Nacos", zap.Error(err))
	}

	// DataID: kafka.yaml — Kafka broker address and topic config
	kafkaContent, err := client.GetConfig(vo.ConfigParam{
		DataId: "kafka.yaml",
		Group:  "xyy_ops",
	})
	if err == nil && kafkaContent != "" {
		resolvedKafka := resolvePlaceholders(kafkaContent, props)
		if parseErr := yaml.Unmarshal([]byte(resolvedKafka), base); parseErr != nil {
			logger.Log.Warn("failed to parse kafka.yaml", zap.Error(parseErr))
		}
	} else if err != nil {
		logger.Log.Warn("cannot fetch kafka.yaml from Nacos", zap.Error(err))
	}

	return base
}

// registerNacosListeners registers config change listeners for dynamic reload.
func registerNacosListeners(client config_client.IConfigClient) {
	listener := func(namespace, group, dataId, data string) {
		logger.Log.Info("Nacos config changed", zap.String("namespace", namespace), zap.String("group", group), zap.String("dataId", dataId))
		// On any change, reload all configurations from scratch to ensure placeholder consistency,
		// preserving the same layering as Init: defaults+local → Nacos → env.
		cfg := defaultConfig()
		loadLocalConfig(cfg)
		cfg = loadFromNacos(client, cfg)
		applyEnvOverrides(cfg)
		storeConfig(cfg)
		logger.Log.Info("config hot-reloaded",
			zap.String("cloud", cfg.Xyy.Cloud),
			zap.Bool("kafkaSwitch", cfg.Spring.Kafka.Switch),
		)
	}

	listenParam := func(dataId, group string) {
		err := client.ListenConfig(vo.ConfigParam{
			DataId:   dataId,
			Group:    group,
			OnChange: listener,
		})
		if err != nil {
			logger.Log.Warn("failed to register Nacos listener", zap.String("group", group), zap.String("dataId", dataId), zap.Error(err))
		}
	}

	appGroup := os.Getenv("NACOS_GROUP")
	if appGroup == "" {
		appGroup = "xyy"
	}

	listenParam("ebike-device-openapi.yml", appGroup)
	listenParam("kafka.yaml", "xyy_ops")
	listenParam("redis.yaml", "xyy_ops")
}

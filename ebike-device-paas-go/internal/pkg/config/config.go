// Package config loads runtime configuration in three layers, mirroring the
// approach used by identity-auth-go:
//
//  1. LoadLocalConfig  — local bootstrap YAML (conf/application.yaml).
//  2. Nacos merge      — remote ops configs (redis.yaml @ {group}_ops) and the
//     app config (ebike-device-paas.yml @ {group}); see internal/pkg/nacos.
//  3. ApplyEnvOverrides — environment variables (highest precedence).
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
var GlobalConfig = defaultConfig()

// Config is the top-level configuration structure.
//
// Spring/Java-only sections of the Nacos app config (server.tomcat, spring.*,
// feign.*, logging.*) are intentionally not mapped here; yaml.Unmarshal ignores
// unknown keys, so e.g. spring.redis.port's "${...}" placeholder never reaches
// an int field.
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Nacos      NacosConfig      `yaml:"nacos"`
	Redis      RedisConfig      `yaml:"redis"`
	Xyy        XyyConfig        `yaml:"xyy"`
	DryRun     bool             `yaml:"dryRun"`
	Shadow     ShadowConfig     `yaml:"shadow"`
	Coordinate CoordinateConfig `yaml:"coordinate"`
	Ecu        EcuConfig        `yaml:"ecu"`
	Anvelink   AnvelinkConfig   `yaml:"anvelink"`
	Feign      FeignConfig      `yaml:"feign"`
	// Kafka holds the C34 producer settings. Brokers come from kafka.yaml in the
	// {group}_ops Nacos group (spring.kafka.bootstrap-servers); empty means the
	// producer degrades to a no-op (write/report endpoints still respond).
	Kafka KafkaConfig `yaml:"-"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port int    `yaml:"port"`
	Name string `yaml:"name"`
}

// NacosConfig holds Nacos connection settings. Defaults target a local Nacos.
type NacosConfig struct {
	ServerAddr      string `yaml:"serverAddr"`
	Port            uint64 `yaml:"port"`
	Namespace       string `yaml:"namespace"`
	Group           string `yaml:"group"`
	RegisterEnabled bool   `yaml:"registerEnabled"`
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
}

// XyyConfig holds business-specific settings.
type XyyConfig struct {
	// JavaServiceURL is the legacy Java ebike-device-paas base URL for shadow diff.
	JavaServiceURL string `yaml:"javaServiceUrl"`
	// OpenapiURL is the ebike-device-openapi gateway base URL for forwarded
	// device-command queries (innerParam / deviceInfo / blueTBeacon / bleHelmet).
	OpenapiURL string `yaml:"openapiUrl"`
	// ManagementURL is the ebike-management base URL for computeRestBattery /
	// user carInfos (device/list role filter).
	ManagementURL string `yaml:"managementUrl"`
	// FenceURL is the ebike-fence base URL for service-area / back-car / use-car
	// config (device/eBikeLocation).
	FenceURL string `yaml:"fenceUrl"`
	// WorkerURL is the ebike-device-worker base URL for trajectory / metric
	// queries (device/trajectory/*).
	WorkerURL string `yaml:"workerUrl"`
	// MapServiceConfig is the map-service (amap reverse-geocode) endpoint for
	// device/page address/scanAddress (mirrors Java xyy.mapServiceConfig.*).
	MapServiceConfig MapServiceConfig `yaml:"mapServiceConfig"`
	// Proxy holds fence-style gray routing (see middleware.ProxyGateway):
	// liveList → Go only; recordList → Java+[RECORD]; else DryRun shadow dual-run.
	Proxy ProxyConfig `yaml:"proxy"`
}

// ProxyConfig mirrors ebike-fence-go xyy.proxy for gray rollout.
type ProxyConfig struct {
	// TargetURL is the Java ebike-device-paas base URL for record/shadow proxy.
	// Falls back to JavaServiceURL when empty.
	TargetURL string `yaml:"targetUrl"`
	// LiveList paths are served natively by Go with no shadow comparison
	// (graduated after shadow verification).
	LiveList []string `yaml:"liveList"`
	// RecordList paths are reverse-proxied to Java only ([RECORD] log).
	RecordList []string `yaml:"recordList"`
}

// MapServiceConfig mirrors Java's xyy.mapServiceConfig.{url,api,secret}. API
// selects the upstream map provider and defaults to "aMap"
// (Java ${xyy.mapServiceConfig.api:aMap}). Secret is sent as the "secret" header
// on every map-service call (mirrors FeignClientConfig's map-service branch).
type MapServiceConfig struct {
	URL    string `yaml:"url"`
	API    string `yaml:"api"`
	Secret string `yaml:"secret"`
}

// ShadowConfig tunes shadow-mode response comparison.
type ShadowConfig struct {
	IgnoreFields   []string `yaml:"ignoreFields"`
	FloatTolerance float64  `yaml:"floatTolerance"`
	// UnorderedFields are object keys whose array values are compared
	// order-insensitively (Java HashMap.values() lists, top-level "data").
	UnorderedFields []string `yaml:"unorderedFields"`
}

// CoordinateConfig selects the coordinate system used for GPS output (P2).
// Type: 0 = WGS-84, 1 = GCJ-02.
type CoordinateConfig struct {
	Type int `yaml:"type"`
}

// EcuConfig holds ECU-related settings (Kafka C34 upload in P5).
type EcuConfig struct {
	Kafka EcuKafkaConfig `yaml:"kafka"`
	Debug EcuDebugConfig `yaml:"debug"`
}

// EcuDebugConfig maps ecu.debug.isEnable; when true, getDeviceInfo mirrors
// helmet6Lock into helmet6React (matches Java RpcResultToCo.toDeviceInfoCo).
type EcuDebugConfig struct {
	IsEnable bool `yaml:"isEnable"`
}

// EcuKafkaConfig holds the Kafka parent topic (Spring key: ecu.kafka.parent-topic).
type EcuKafkaConfig struct {
	ParentTopic string `yaml:"parent-topic"`
}

// AnvelinkConfig holds downstream Anvelink OpenAPI settings.
type AnvelinkConfig struct {
	Openapi AnvelinkOpenapiConfig `yaml:"openapi"`
}

// FeignConfig mirrors the Spring feign.* tree; only feign.client.auth-url-regex
// is consumed (it selects which forwarded URLs receive the AES auth headers).
type FeignConfig struct {
	Client FeignClientConfig `yaml:"client"`
}

// FeignClientConfig maps feign.client.auth-url-regex: the list of path regexes
// (joined like Java FeignClientFilterUrlConfig) whose requests get authKey /
// authSecret / appId headers. Prod sets ["/ebike/.*"].
type FeignClientConfig struct {
	AuthURLRegex []string `yaml:"auth-url-regex"`
}

// AnvelinkOpenapiConfig holds the OpenAPI URL and AES key (Spring key: anvelink.openapi.aes-key).
type AnvelinkOpenapiConfig struct {
	URL    string `yaml:"url"`
	AesKey string `yaml:"aes-key"`
}

// KafkaConfig holds the C34 producer bootstrap servers (Spring key:
// spring.kafka.bootstrap-servers, from kafka.yaml @ {group}_ops).
type KafkaConfig struct {
	Brokers []string
}

func defaultConfig() *Config {
	c := &Config{}
	c.Server.Port = 8080
	c.Server.Name = "ebike-device-paas"
	c.Nacos.ServerAddr = "127.0.0.1"
	c.Nacos.Port = 8848
	c.Nacos.Namespace = "prod"
	c.Nacos.Group = "xyy"
	c.Nacos.RegisterEnabled = true
	c.Redis.Host = "127.0.0.1"
	c.Redis.Port = 6379
	c.Shadow.FloatTolerance = 1e-6
	c.Shadow.IgnoreFields = defaultIgnoreFields()
	c.Shadow.UnorderedFields = defaultUnorderedFields()
	// Default to WGS-84 (0) to match Java's ${coordinate.type:0} fallback. Prod
	// Nacos yml sets coordinate.type=1 explicitly, which overrides this on merge;
	// the default only governs the missing-key case, which must equal Java's.
	c.Coordinate.Type = 0
	return c
}

func defaultIgnoreFields() []string {
	return []string{
		"reportTime", "timestamp", "bmsTimeStamp", "rfidTimestamp",
		"lockTime", "unlockTime", "lat", "lng", "wgs84Lat", "wgs84Lng",
		"scanLat", "scanLng", "speed", "course", "soc", "restBattery",
		"restMileage", "voltage", "cusTime",
		"noOrderTime", "staticTime", // now-derived in queryDeviceBy{NoOrderTime,StaticTime}
		"address", "scanAddress", // amap reverse-geocode (not ported in getDevicePage)
	}
}

// defaultUnorderedFields lists array keys Java emits in HashMap (non-deterministic)
// order, so shadow compares them as multisets. Only uniquely-named keys belong
// here; endpoints whose top-level "data" list is unordered pass it per-call so
// genuinely-ordered list endpoints keep strict order comparison.
func defaultUnorderedFields() []string {
	return []string{
		"serviceStatistics", // carStatisticsByService
		"parkingStatistics", // carStatisticsByService
	}
}

// OpsGroup returns the Nacos group for shared ops configs (redis.yaml, kafka.yaml, …).
func OpsGroup() string {
	if GlobalConfig.Nacos.Group == "" {
		return "xyy_ops"
	}
	return GlobalConfig.Nacos.Group + "_ops"
}

// AppDataID returns the Nacos dataId for this app's config (prefix + .yml).
func AppDataID() string {
	name := GlobalConfig.Server.Name
	if name == "" {
		name = "ebike-device-paas"
	}
	return name + ".yml"
}

// LoadLocalConfig reads the local bootstrap YAML into GlobalConfig. Missing file
// is tolerated (defaults + env still apply).
func LoadLocalConfig(path string) {
	if path == "" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[config] local config %s not read (%v); using defaults", path, err)
		return
	}
	if err := yaml.Unmarshal(data, GlobalConfig); err != nil {
		log.Fatalf("[config] failed to parse local config %s: %v", path, err)
	}
	log.Printf("[config] loaded local config from %s", path)
}

// ---------------------------------------------------------------------------
// Nacos YAML structures + merge functions
// ---------------------------------------------------------------------------

// redisNacosFile maps redis.yaml (dataId=redis.yaml, group={group}_ops):
//
//	redis:
//	  ebike_device_paas:
//	    host: ...
//	    port: 6379
//	    password: ...
//	    database: 0
type redisNacosFile struct {
	Redis map[string]redisDataSource `yaml:"redis"`
}

type redisDataSource struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	Database int    `yaml:"database"`
}

// redisPaasDataSourceKeys are the relaxed-binding keys we accept for this app.
var redisPaasDataSourceKeys = []string{"ebike_device_paas", "ebike-device-paas"}

// MergeNacosRedisConfig parses redis.yaml content from Nacos and merges the
// ebike_device_paas datasource into GlobalConfig.Redis.
func MergeNacosRedisConfig(yamlContent string) {
	var raw redisNacosFile
	if err := yaml.Unmarshal([]byte(yamlContent), &raw); err != nil {
		log.Printf("[config] failed to parse nacos redis.yaml: %v", err)
		return
	}
	var ds redisDataSource
	var key string
	for _, k := range redisPaasDataSourceKeys {
		if v, ok := raw.Redis[k]; ok && v.Host != "" {
			ds, key = v, k
			break
		}
	}
	if key == "" {
		log.Printf("[config] redis.yaml missing datasource (tried %v)", redisPaasDataSourceKeys)
		return
	}
	GlobalConfig.Redis.Host = ds.Host
	if ds.Port != 0 {
		GlobalConfig.Redis.Port = ds.Port
	}
	if ds.Password != "" {
		GlobalConfig.Redis.Password = ds.Password
	}
	db := ds.DB
	if db == 0 {
		db = ds.Database
	}
	GlobalConfig.Redis.Database = db
	log.Printf("[config] merged nacos redis (key=%s, host=%s, port=%d, db=%d)",
		key, GlobalConfig.Redis.Host, GlobalConfig.Redis.Port, GlobalConfig.Redis.Database)
}

// MergeNacosAppConfig merges the app's own Nacos config (ebike-device-paas.yml).
// Only keys present in our Config schema are picked up (relaxed via yaml tags).
func MergeNacosAppConfig(yamlContent string) {
	if strings.TrimSpace(yamlContent) == "" {
		return
	}
	if err := yaml.Unmarshal([]byte(yamlContent), GlobalConfig); err != nil {
		log.Printf("[config] failed to parse nacos app config: %v", err)
		return
	}
	log.Printf("[config] merged nacos app config (%s)", AppDataID())
}

// kafkaNacosFile maps kafka.yaml (dataId=kafka.yaml, group={group}_ops):
//
//	spring:
//	  kafka:
//	    bootstrap-servers: host1:9092,host2:9092
//	    topic-prefix: ebike
type kafkaNacosFile struct {
	Spring struct {
		Kafka struct {
			BootstrapServers string `yaml:"bootstrap-servers"`
		} `yaml:"kafka"`
	} `yaml:"spring"`
}

// MergeNacosKafkaConfig parses kafka.yaml content from Nacos and merges the
// bootstrap servers into GlobalConfig.Kafka.Brokers. A placeholder value (e.g.
// "${...}") or empty content leaves the producer disabled.
func MergeNacosKafkaConfig(yamlContent string) {
	if strings.TrimSpace(yamlContent) == "" {
		return
	}
	var raw kafkaNacosFile
	if err := yaml.Unmarshal([]byte(yamlContent), &raw); err != nil {
		log.Printf("[config] failed to parse nacos kafka.yaml: %v", err)
		return
	}
	servers := strings.TrimSpace(raw.Spring.Kafka.BootstrapServers)
	if servers == "" || strings.Contains(servers, "${") {
		log.Printf("[config] kafka.yaml has no usable bootstrap-servers; C34 producer disabled")
		return
	}
	GlobalConfig.Kafka.Brokers = splitCSV(servers)
	log.Printf("[config] merged nacos kafka (brokers=%v)", GlobalConfig.Kafka.Brokers)
}

// splitCSV splits a comma-separated list, trimming spaces and dropping empties.
func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Environment overrides (highest precedence)
// ---------------------------------------------------------------------------

// ApplyEnvOverrides applies environment variables as the final config layer.
func ApplyEnvOverrides() {
	if v := os.Getenv("NACOS_SERVER_ADDR"); v != "" {
		if strings.Contains(v, ":") {
			parts := strings.SplitN(v, ":", 2)
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
	}
	if v := os.Getenv("NACOS_GROUP"); v != "" {
		GlobalConfig.Nacos.Group = v
	}
	if v := os.Getenv("NACOS_REGISTER_ENABLED"); v != "" {
		GlobalConfig.Nacos.RegisterEnabled = strings.EqualFold(v, "true") || v == "1"
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			GlobalConfig.Server.Port = port
		}
	}
	if v := os.Getenv("DRY_RUN"); v != "" {
		GlobalConfig.DryRun = strings.EqualFold(v, "true") || v == "1"
		log.Printf("[config] env override: DRY_RUN=%v", GlobalConfig.DryRun)
	}
	if v := os.Getenv("JAVA_SERVICE_URL"); v != "" {
		GlobalConfig.Xyy.JavaServiceURL = v
	}
	if v := os.Getenv("OPENAPI_URL"); v != "" {
		GlobalConfig.Xyy.OpenapiURL = v
		log.Printf("[config] env override: OPENAPI_URL=%s", v)
	}
	if v := os.Getenv("MANAGEMENT_URL"); v != "" {
		GlobalConfig.Xyy.ManagementURL = v
		log.Printf("[config] env override: MANAGEMENT_URL=%s", v)
	}
	if v := os.Getenv("FENCE_URL"); v != "" {
		GlobalConfig.Xyy.FenceURL = v
		log.Printf("[config] env override: FENCE_URL=%s", v)
	}
	if v := os.Getenv("WORKER_URL"); v != "" {
		GlobalConfig.Xyy.WorkerURL = v
		log.Printf("[config] env override: WORKER_URL=%s", v)
	}
	if v := os.Getenv("REDIS_HOST"); v != "" {
		GlobalConfig.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			GlobalConfig.Redis.Port = port
		}
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		GlobalConfig.Redis.Password = v
	}
	if v := os.Getenv("REDIS_DATABASE"); v != "" {
		if db, err := strconv.Atoi(v); err == nil {
			GlobalConfig.Redis.Database = db
		}
	}
	if v := os.Getenv("KAFKA_BROKERS"); v != "" {
		GlobalConfig.Kafka.Brokers = splitCSV(v)
		log.Printf("[config] env override: KAFKA_BROKERS=%v", GlobalConfig.Kafka.Brokers)
	}
}

// LogEffectiveUpstream prints the resolved upstream URLs and auth prerequisites
// after all config layers merge (local + Nacos + env).
func LogEffectiveUpstream() {
	cfg := GlobalConfig.Xyy
	aesKey := GlobalConfig.Anvelink.Openapi.AesKey
	aesOK := len(aesKey) > 16
	log.Printf("[config] upstream urls: managementUrl=%q openapiUrl=%q workerUrl=%q fenceUrl=%q",
		cfg.ManagementURL, cfg.OpenapiURL, cfg.WorkerURL, cfg.FenceURL)
	log.Printf("[config] anvelink openapi aes-key configured=%v (len=%d)", aesOK, len(aesKey))
}

// IgnoreFieldSet returns the configured volatile-field ignore list as a set.
func IgnoreFieldSet() map[string]bool {
	out := make(map[string]bool, len(GlobalConfig.Shadow.IgnoreFields))
	for _, f := range GlobalConfig.Shadow.IgnoreFields {
		out[f] = true
	}
	return out
}

// UnorderedFieldSet returns the configured order-insensitive array keys as a set.
func UnorderedFieldSet() map[string]bool {
	out := make(map[string]bool, len(GlobalConfig.Shadow.UnorderedFields))
	for _, f := range GlobalConfig.Shadow.UnorderedFields {
		out[f] = true
	}
	return out
}

// RedisAddr returns the host:port address for the Redis client.
func RedisAddr() string {
	return fmt.Sprintf("%s:%d", GlobalConfig.Redis.Host, GlobalConfig.Redis.Port)
}

// ProxyTargetURL returns the Java base URL for recordList reverse proxy.
func ProxyTargetURL() string {
	if u := strings.TrimSpace(GlobalConfig.Xyy.Proxy.TargetURL); u != "" {
		return u
	}
	return strings.TrimSpace(GlobalConfig.Xyy.JavaServiceURL)
}

// PathInProxyList reports whether path matches any entry in list (exact or prefix).
func PathInProxyList(list []string, path string) bool {
	for _, v := range list {
		if path == v || strings.HasPrefix(path, v) || strings.HasPrefix(v, path) {
			return true
		}
	}
	return false
}

// IsLivePath reports whether path has graduated from shadow comparison.
func IsLivePath(path string) bool {
	return PathInProxyList(GlobalConfig.Xyy.Proxy.LiveList, path)
}

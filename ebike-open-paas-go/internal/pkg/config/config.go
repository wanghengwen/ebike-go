// Package config loads runtime configuration in three layers, mirroring
// ebike-device-paas-go:
//
//  1. LoadLocalConfig  — local bootstrap YAML (conf/application.yaml).
//  2. Nacos merge      — remote ops configs (redis.yaml @ {group}_ops) and the
//     app config (ebike-open-paas.yml @ {group}); see internal/pkg/nacos.
//  3. ApplyEnvOverrides — environment variables (highest precedence).
//
// The open-platform agent registry lives in the Nacos app config (open.agents)
// rather than MySQL, so credentials and quotas can be changed without a release.
// Callback subscriptions are the one piece of mutable state and live in Redis
// (see internal/callback), because third parties register them over the API.
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"gopkg.in/yaml.v3"
)

// live is the published configuration snapshot. Nacos re-merges arrive on the
// listener goroutine while requests are reading, so a snapshot is never edited
// in place: every merge clones, edits the clone, and swaps it in. Editing in
// place would race on the maps and slices inside (a concurrent read/write of
// Open.VoiceIndexMap is a fatal runtime error, not a torn read).
var live atomic.Pointer[Config]

// baseline is the snapshot as it stood before any Nacos merge: code defaults
// plus the local bootstrap YAML. A Nacos merge has to reset the collections it
// owns before unmarshalling, because yaml.Unmarshal appends to slices instead of
// replacing them; baseline is what those collections fall back to when the
// remote config does not define them.
var baseline atomic.Pointer[Config]

// mergeMu serialises merges so two listeners cannot both clone the same
// snapshot and lose one another's edits.
var mergeMu sync.Mutex

// agentIndex caches the agentId -> AgentConfig lookup built from
// Open.Agents. It is rebuilt on every config merge so Nacos hot-reload takes
// effect without a restart, and read atomically because requests resolve agents
// concurrently with the Nacos listener goroutine.
var agentIndex atomic.Pointer[map[string]AgentConfig]

func init() {
	c := defaultConfig()
	live.Store(c)
	baseline.Store(c)
}

// GlobalConfig returns the current configuration snapshot. The returned pointer
// is immutable: callers must not write through it. Hold the result in a local
// when reading several fields that should agree with each other, since a Nacos
// reload can publish a new snapshot between two calls.
func GlobalConfig() *Config { return live.Load() }

// mutate publishes a modified copy of the live snapshot. fn must replace maps
// and slices wholesale rather than writing into the ones it receives, which are
// still shared with the snapshot readers already hold.
func mutate(fn func(*Config)) {
	_ = mutateE(func(c *Config) error {
		fn(c)
		return nil
	})
}

// mutateE is mutate for merges that can fail. The clone is discarded when fn
// returns an error, so a malformed remote config leaves the running snapshot
// intact instead of publishing a half-applied one.
func mutateE(fn func(*Config) error) error {
	mergeMu.Lock()
	defer mergeMu.Unlock()
	next := *live.Load()
	if err := fn(&next); err != nil {
		return err
	}
	live.Store(&next)
	return nil
}

// Config is the top-level configuration structure.
//
// Spring/Java-only sections of shared Nacos configs are intentionally not mapped
// here; yaml.Unmarshal ignores unknown keys.
type Config struct {
	Server ServerConfig `yaml:"server"`
	Nacos  NacosConfig  `yaml:"nacos"`
	Redis  RedisConfig  `yaml:"redis"`
	Xyy    XyyConfig    `yaml:"xyy"`
	// Anvelink carries the AES key used to sign the tenant auth headers on
	// upstream /ebike/.* calls (Spring key: anvelink.openapi.aes-key).
	Anvelink AnvelinkConfig `yaml:"anvelink"`
	Feign    FeignConfig    `yaml:"feign"`
	Open     OpenConfig     `yaml:"open"`
	// Kafka configures the saas_0 consumer that sources device events; see kafka.go.
	Kafka KafkaConfig `yaml:"kafka"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port int    `yaml:"port"`
	Name string `yaml:"name"`
}

// NacosConfig holds Nacos connection settings.
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
	// RegistryDatabase is where anvelink's device_ebike_{imei} hashes live
	// (openapi/IOT Redis; production is db 12). The main Database is usually the
	// device-paas shadow DB (often 4). -1 means use Database for both.
	RegistryDatabase int `yaml:"registry-database"`
}

// XyyConfig holds the upstream service base URLs. Like ebike-device-paas-go this
// service does not use Nacos service discovery for outbound calls; each upstream
// is a K8s ClusterIP configured here or via env.
type XyyConfig struct {
	// PaasURL is ebike-device-paas (Go or Java): device shadow queries and the
	// device-command orchestration endpoints under /device/paas/*.
	PaasURL string `yaml:"paasUrl"`
	// OpenapiURL is ebike-device-openapi: /set_limit_speed and raw /ebike/cmd/*.
	OpenapiURL string `yaml:"openapiUrl"`
	// WorkerURL is ebike-device-worker: GPS trajectory queries.
	WorkerURL string `yaml:"workerUrl"`
	// ConsoleURL is anvelink-console: tenant device registry (allDevices).
	// In-cluster: http://anvelink-console.prod.svc.cluster.local:8080
	ConsoleURL string `yaml:"consoleUrl"`
	// ConsoleAuth holds /auth/login credentials when console JWT is required.
	// Phone is the login account; username is informational only.
	ConsoleAuth ConsoleAuthConfig `yaml:"consoleAuth"`
	// ConsoleAuthToken is an optional static Bearer JWT. When empty, consoleAuth
	// phone/password are used and refreshed on expiry (5003) or auth failure.
	ConsoleAuthToken string `yaml:"consoleAuthToken"`
	// ManagementURL is ebike-management: tenant IoT credentials for AES headers
	// on /ebike/* upstream calls (e.g. worker trajectory). Not used by allDevices.
	ManagementURL string `yaml:"managementUrl"`
	// MapServiceConfig is the map-service reverse-geocode endpoint.
	MapServiceConfig MapServiceConfig `yaml:"mapServiceConfig"`
}

// ConsoleAuthConfig is anvelink-console /auth/login credentials.
type ConsoleAuthConfig struct {
	Phone    string `yaml:"phone"`
	Password string `yaml:"password"`
	Username string `yaml:"username"`
}

// MapServiceConfig mirrors Java's xyy.mapServiceConfig.{url,api,secret}. Secret
// is sent as the "secret" header on every map-service call.
type MapServiceConfig struct {
	URL    string `yaml:"url"`
	API    string `yaml:"api"`
	Secret string `yaml:"secret"`
}

// AnvelinkConfig holds downstream Anvelink OpenAPI settings.
type AnvelinkConfig struct {
	Openapi AnvelinkOpenapiConfig `yaml:"openapi"`
}

// AnvelinkOpenapiConfig holds the AES key (Spring key: anvelink.openapi.aes-key).
// Only the substring from index 16 is used as key+iv, matching Java
// AesEncryptUtil.
type AnvelinkOpenapiConfig struct {
	AesKey string `yaml:"aes-key"`
}

// FeignConfig mirrors the Spring feign.* tree; only feign.client.auth-url-regex
// is consumed (it selects which upstream URLs receive the AES auth headers).
type FeignConfig struct {
	Client FeignClientConfig `yaml:"client"`
}

// FeignClientConfig maps feign.client.auth-url-regex. Prod sets ["/ebike/.*"].
type FeignClientConfig struct {
	AuthURLRegex []string `yaml:"auth-url-regex"`
}

// OpenConfig holds everything specific to the third-party open platform.
type OpenConfig struct {
	// Agents is the third-party credential registry (Nacos-managed, hot-reloaded).
	Agents []AgentConfig `yaml:"agents"`
	// RateLimit is the default per-agent quota applied when an agent does not
	// override it.
	RateLimit RateLimitConfig `yaml:"rateLimit"`
	// Callback tunes the outbound event-callback dispatcher.
	Callback CallbackConfig `yaml:"callback"`
	// CommandPin is the audit `pin` put on commandContext for upstream calls.
	CommandPin string `yaml:"commandPin"`
	// VoiceIndexMap translates a Xiaoan ringtone slot to our internal slot.
	// Keys and values are both slot numbers; slots absent from the map pass
	// through unchanged. See docs/CONTRACT.md for the two tables.
	VoiceIndexMap map[int]int `yaml:"voiceIndexMap"`
	// Notify tunes the derived-notify state machine (fence crossings, SOC steps).
	Notify NotifyConfig `yaml:"notify"`
	// InternalToken guards /internal/*. Leaving it empty denies those routes
	// outright rather than opening them, so a forgotten secret fails closed.
	InternalToken string `yaml:"internalToken"`
}

// AgentConfig is one third-party credential entry.
type AgentConfig struct {
	// AgentID is the Xiaoan `agentId` third parties send in query/body.
	AgentID string `yaml:"agentId"`
	// Token is the `xc-access-token` header value.
	Token string `yaml:"token"`
	// TenantID is the internal tenant this agent maps to. Every upstream call
	// and Redis shadow key is scoped by it.
	TenantID string `yaml:"tenantId"`
	// Name is a human label for logs/audit.
	Name string `yaml:"name"`
	// Enabled allows revoking an agent without deleting the entry.
	Enabled *bool `yaml:"enabled"`
	// RateLimit overrides Open.RateLimit for this agent.
	RateLimit *RateLimitConfig `yaml:"rateLimit"`
}

// IsEnabled reports whether the agent may call the API. Absent means enabled, so
// a minimal config entry works.
func (a AgentConfig) IsEnabled() bool { return a.Enabled == nil || *a.Enabled }

// RateLimitConfig is a sliding-window quota. The Xiaoan doc states the realtime
// budget as "30k / 5 minutes / 1000 online vehicles"; expressed per agent that
// is MaxRequests over WindowSeconds.
type RateLimitConfig struct {
	WindowSeconds int `yaml:"windowSeconds"`
	MaxRequests   int `yaml:"maxRequests"`
}

// CallbackConfig tunes the outbound dispatcher. The queue is deliberately
// bounded: a slow third-party URL must never build unbounded backlog, and it
// must never push backpressure into our saas_0 consumer.
type CallbackConfig struct {
	// QueueSize is the total bounded delivery depth, split evenly across the
	// per-URL lanes. Overflow drops the event and increments a counter.
	QueueSize int `yaml:"queueSize"`
	// Workers is the number of delivery lanes; a URL is pinned to one of them,
	// so a stalled customer cannot occupy the capacity of the others.
	Workers int `yaml:"workers"`
	// EnqueueWaitMs is how long a full lane is given to drain before the event
	// is dropped. Short by design: it absorbs a burst without letting the
	// slowest customer set the pace of the Kafka consumer.
	EnqueueWaitMs int `yaml:"enqueueWaitMs"`
	// TimeoutSeconds is the per-attempt HTTP timeout for a customer URL.
	TimeoutSeconds int `yaml:"timeoutSeconds"`
	// MaxAttempts includes the first try (1 = no retry).
	MaxAttempts int `yaml:"maxAttempts"`
	// RetryBackoffMs is the base backoff, multiplied by the attempt number.
	RetryBackoffMs int `yaml:"retryBackoffMs"`
	// SubscriptionTTLSeconds bounds the in-process subscription cache. Redis is
	// the source of truth; this only avoids a Redis round-trip per event.
	SubscriptionTTLSeconds int `yaml:"subscriptionTtlSeconds"`
	// CircuitEnabled turns on the per-URL breaker. A repeatedly timing-out
	// callback is muted in-process for CircuitCooldownSeconds; the Redis
	// subscription is left intact so list/unregister still work.
	CircuitEnabled bool `yaml:"circuitEnabled"`
	// CircuitFailThreshold is consecutive exhausted deliveries that open the
	// circuit (default 20).
	CircuitFailThreshold int `yaml:"circuitFailThreshold"`
	// CircuitCooldownSeconds is how long an open circuit stays muted before a
	// single probe is allowed (default 600).
	CircuitCooldownSeconds int `yaml:"circuitCooldownSeconds"`
}

// NotifyConfig tunes the notify codes that saas_0 does not report directly and
// that we therefore derive from state transitions.
//
// These are opt-in because a derived notify is weaker than a device-reported
// one: we infer it from consecutive packets, so a missed packet means a missed
// edge. See internal/event/state.go.
type NotifyConfig struct {
	// FenceEnabled derives notify 17/18 from the GPS out-of-service-area bit.
	FenceEnabled bool `yaml:"fenceEnabled"`
	// SocStepsEnabled derives notify 21/22 from BMS state-of-charge crossings.
	SocStepsEnabled bool `yaml:"socStepsEnabled"`
	// SocSteps are the descending SOC percentages that fire notify 21 and 22.
	// Xiaoan defines them as "50% remaining" and "30% remaining".
	SocSteps []SocStep `yaml:"socSteps"`
	// StateTTLSeconds is how long a device's last-known state is kept for edge
	// detection. It must exceed the device's reporting interval; too short
	// re-fires an edge, too long wastes Redis memory on retired devices.
	StateTTLSeconds int `yaml:"stateTtlSeconds"`
}

// SocStep maps a state-of-charge threshold to the notify code fired when a
// device's SOC crosses it downward.
type SocStep struct {
	Percent int `yaml:"percent"`
	Notify  int `yaml:"notify"`
}

func defaultConfig() *Config {
	c := &Config{}
	c.Server.Port = 8080
	c.Server.Name = "ebike-open-paas"
	c.Nacos.ServerAddr = "127.0.0.1"
	c.Nacos.Port = 8848
	c.Nacos.Namespace = "prod"
	c.Nacos.Group = "xyy"
	c.Nacos.RegisterEnabled = true
	c.Redis.Host = "127.0.0.1"
	c.Redis.Port = 6379
	c.Redis.RegistryDatabase = -1
	c.Open.RateLimit = RateLimitConfig{WindowSeconds: 300, MaxRequests: 30000}
	c.Open.Callback = CallbackConfig{
		QueueSize:              20000,
		Workers:                16,
		EnqueueWaitMs:          200,
		TimeoutSeconds:         5,
		MaxAttempts:            3,
		RetryBackoffMs:         500,
		SubscriptionTTLSeconds: 10,
		CircuitEnabled:         true,
		CircuitFailThreshold:   20,
		CircuitCooldownSeconds: 600,
	}
	c.Open.Notify = NotifyConfig{
		FenceEnabled:    true,
		SocStepsEnabled: true,
		SocSteps:        []SocStep{{Percent: 50, Notify: 21}, {Percent: 30, Notify: 22}},
		StateTTLSeconds: 86400,
	}
	c.Open.CommandPin = "open_paas"
	c.Kafka = defaultKafkaConfig()
	return c
}

// OpsGroup returns the Nacos group for shared ops configs (redis.yaml, …).
func OpsGroup() string {
	if GlobalConfig().Nacos.Group == "" {
		return "xyy_ops"
	}
	return GlobalConfig().Nacos.Group + "_ops"
}

// AppDataID returns the Nacos dataId for this app's config (prefix + .yml).
func AppDataID() string {
	name := GlobalConfig().Server.Name
	if name == "" {
		name = "ebike-open-paas"
	}
	return name + ".yml"
}

// LoadLocalConfig reads the local bootstrap YAML. A missing file is tolerated
// (defaults + env still apply); a malformed one is fatal, because starting with
// half-applied credentials is worse than not starting.
func LoadLocalConfig(path string) {
	if path == "" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[config] local config %s not read (%v); using defaults", path, err)
		snapshotBaseline()
		rebuildAgentIndex()
		return
	}
	if err := mutateE(func(c *Config) error { return yaml.Unmarshal(data, c) }); err != nil {
		log.Fatalf("[config] failed to parse local config %s: %v", path, err)
	}
	log.Printf("[config] loaded local config from %s", path)
	snapshotBaseline()
	rebuildAgentIndex()
}

// snapshotBaseline freezes the pre-Nacos snapshot for MergeNacosAppConfig to
// fall back to.
func snapshotBaseline() { baseline.Store(live.Load()) }

// ---------------------------------------------------------------------------
// Nacos YAML structures + merge functions
// ---------------------------------------------------------------------------

// redisNacosFile maps redis.yaml (dataId=redis.yaml, group={group}_ops).
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

// redisDataSourceKeys are the relaxed-binding keys we accept for this app. The
// open platform shares the device-paas Redis instance because it reads the same
// device shadow keys; a dedicated `ebike_open_paas` datasource takes precedence
// when ops define one.
var redisDataSourceKeys = []string{
	"ebike_open_paas", "ebike-open-paas",
	"ebike_device_paas", "ebike-device-paas",
}

// MergeNacosRedisConfig parses redis.yaml content from Nacos and merges the
// first matching datasource into GlobalConfig.Redis.
func MergeNacosRedisConfig(yamlContent string) {
	var raw redisNacosFile
	if err := yaml.Unmarshal([]byte(yamlContent), &raw); err != nil {
		log.Printf("[config] failed to parse nacos redis.yaml: %v", err)
		return
	}
	var ds redisDataSource
	var key string
	for _, k := range redisDataSourceKeys {
		if v, ok := raw.Redis[k]; ok && v.Host != "" {
			ds, key = v, k
			break
		}
	}
	if key == "" {
		log.Printf("[config] redis.yaml missing datasource (tried %v)", redisDataSourceKeys)
		return
	}
	mutate(func(c *Config) {
		c.Redis.Host = ds.Host
		if ds.Port != 0 {
			c.Redis.Port = ds.Port
		}
		if ds.Password != "" {
			c.Redis.Password = ds.Password
		}
		db := ds.DB
		if db == 0 {
			db = ds.Database
		}
		c.Redis.Database = db
	})
	r := GlobalConfig().Redis
	log.Printf("[config] merged nacos redis (key=%s, host=%s, port=%d, db=%d)",
		key, r.Host, r.Port, r.Database)
}

// MergeNacosAppConfig merges the app's own Nacos config (ebike-open-paas.yml).
//
// Every collection the remote file owns is cleared first, because yaml.Unmarshal
// appends to slices and merges into maps instead of replacing them. Without the
// reset, open.agents would resurrect revoked credentials and
// feign.client.auth-url-regex would grow a duplicate entry on every hot-reload.
// Collections the remote file does not define fall back to the pre-Nacos
// baseline so a partial app config cannot blank them — except open.agents, where
// an empty remote list must win so revoking every credential is possible.
func MergeNacosAppConfig(yamlContent string) {
	if strings.TrimSpace(yamlContent) == "" {
		return
	}
	err := mutateE(func(c *Config) error {
		base := baseline.Load()
		c.Open.Agents = nil
		c.Open.VoiceIndexMap = nil
		c.Open.Notify.SocSteps = nil
		c.Feign.Client.AuthURLRegex = nil
		if err := yaml.Unmarshal([]byte(yamlContent), c); err != nil {
			return err
		}
		if len(c.Open.VoiceIndexMap) == 0 {
			c.Open.VoiceIndexMap = base.Open.VoiceIndexMap
		}
		if len(c.Open.Notify.SocSteps) == 0 {
			c.Open.Notify.SocSteps = base.Open.Notify.SocSteps
		}
		if len(c.Feign.Client.AuthURLRegex) == 0 {
			c.Feign.Client.AuthURLRegex = base.Feign.Client.AuthURLRegex
		}
		return nil
	})
	if err != nil {
		log.Printf("[config] failed to parse nacos app config, keeping previous: %v", err)
		return
	}
	log.Printf("[config] merged nacos app config (%s)", AppDataID())
	rebuildAgentIndex()
}

// ---------------------------------------------------------------------------
// Agent registry
// ---------------------------------------------------------------------------

// rebuildAgentIndex refreshes the agentId lookup from the live Open.Agents.
func rebuildAgentIndex() {
	agents := GlobalConfig().Open.Agents
	m := make(map[string]AgentConfig, len(agents))
	for _, a := range agents {
		id := strings.TrimSpace(a.AgentID)
		if id == "" {
			log.Printf("[config] skipping open.agents entry with empty agentId")
			continue
		}
		if a.TenantID == "" {
			log.Printf("[config] skipping open.agents entry agentId=%s with empty tenantId", id)
			continue
		}
		m[id] = a
	}
	agentIndex.Store(&m)
	log.Printf("[config] open platform agent registry: %d entries", len(m))
	for id, a := range m {
		if a.IsEnabled() {
			log.Printf("[config] open agent: agentId=%s tenantId=%s name=%q enabled=true",
				id, a.TenantID, a.Name)
		}
	}
}

// Agent returns the credential entry for agentID.
func Agent(agentID string) (AgentConfig, bool) {
	m := agentIndex.Load()
	if m == nil {
		return AgentConfig{}, false
	}
	a, ok := (*m)[strings.TrimSpace(agentID)]
	return a, ok
}

// AgentsByTenant returns every configured agentId for tenantID. The callback
// dispatcher uses it to resolve a device event's tenant back to its subscribers.
func AgentsByTenant(tenantID string) []string {
	m := agentIndex.Load()
	if m == nil {
		return nil
	}
	var out []string
	for id, a := range *m {
		if a.TenantID == tenantID && a.IsEnabled() {
			out = append(out, id)
		}
	}
	return out
}

// EnabledAgents returns every enabled agent entry. The subscription cache walks
// it to enumerate callback keys, which is both cheaper and less ambiguous than
// SCANning Redis and parsing agentIds back out of the key names.
func EnabledAgents() []AgentConfig {
	m := agentIndex.Load()
	if m == nil {
		return nil
	}
	out := make([]AgentConfig, 0, len(*m))
	for _, a := range *m {
		if a.IsEnabled() {
			out = append(out, a)
		}
	}
	return out
}

// AgentRateLimit returns the effective quota for an agent.
func AgentRateLimit(a AgentConfig) RateLimitConfig {
	rl := GlobalConfig().Open.RateLimit
	if a.RateLimit != nil {
		if a.RateLimit.WindowSeconds > 0 {
			rl.WindowSeconds = a.RateLimit.WindowSeconds
		}
		if a.RateLimit.MaxRequests > 0 {
			rl.MaxRequests = a.RateLimit.MaxRequests
		}
	}
	if rl.WindowSeconds <= 0 {
		rl.WindowSeconds = 300
	}
	if rl.MaxRequests <= 0 {
		rl.MaxRequests = 30000
	}
	return rl
}

// ---------------------------------------------------------------------------
// Environment overrides (highest precedence)
// ---------------------------------------------------------------------------

// ApplyEnvOverrides applies environment variables as the final config layer.
func ApplyEnvOverrides() {
	mutate(applyEnvTo)
}

func applyEnvTo(c *Config) {
	if v := os.Getenv("NACOS_SERVER_ADDR"); v != "" {
		if strings.Contains(v, ":") {
			parts := strings.SplitN(v, ":", 2)
			c.Nacos.ServerAddr = parts[0]
			if port, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
				c.Nacos.Port = port
			}
		} else {
			c.Nacos.ServerAddr = v
		}
		log.Printf("[config] env override: NACOS_SERVER_ADDR=%s", v)
	}
	if v := os.Getenv("NACOS_NAMESPACE"); v != "" {
		c.Nacos.Namespace = v
	}
	if v := os.Getenv("NACOS_GROUP"); v != "" {
		c.Nacos.Group = v
	}
	if v := os.Getenv("NACOS_REGISTER_ENABLED"); v != "" {
		c.Nacos.RegisterEnabled = strings.EqualFold(v, "true") || v == "1"
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Server.Port = port
		}
	}
	if v := os.Getenv("PAAS_URL"); v != "" {
		c.Xyy.PaasURL = v
		log.Printf("[config] env override: PAAS_URL=%s", v)
	}
	if v := os.Getenv("OPENAPI_URL"); v != "" {
		c.Xyy.OpenapiURL = v
		log.Printf("[config] env override: OPENAPI_URL=%s", v)
	}
	if v := os.Getenv("WORKER_URL"); v != "" {
		c.Xyy.WorkerURL = v
		log.Printf("[config] env override: WORKER_URL=%s", v)
	}
	if v := os.Getenv("CONSOLE_URL"); v != "" {
		c.Xyy.ConsoleURL = v
		log.Printf("[config] env override: CONSOLE_URL=%s", v)
	}
	if v := os.Getenv("CONSOLE_AUTH_TOKEN"); v != "" {
		c.Xyy.ConsoleAuthToken = v
		log.Printf("[config] env override: CONSOLE_AUTH_TOKEN set (len=%d)", len(v))
	}
	if v := os.Getenv("CONSOLE_AUTH_PHONE"); v != "" {
		c.Xyy.ConsoleAuth.Phone = v
		log.Printf("[config] env override: CONSOLE_AUTH_PHONE=%s", v)
	}
	if v := os.Getenv("CONSOLE_AUTH_PASSWORD"); v != "" {
		c.Xyy.ConsoleAuth.Password = v
		log.Printf("[config] env override: CONSOLE_AUTH_PASSWORD set (len=%d)", len(v))
	}
	if v := os.Getenv("CONSOLE_AUTH_USERNAME"); v != "" {
		c.Xyy.ConsoleAuth.Username = v
	}
	if v := os.Getenv("MANAGEMENT_URL"); v != "" {
		c.Xyy.ManagementURL = v
		log.Printf("[config] env override: MANAGEMENT_URL=%s", v)
	}
	if v := os.Getenv("MAP_SERVICE_URL"); v != "" {
		c.Xyy.MapServiceConfig.URL = v
	}
	if v := os.Getenv("MAP_SERVICE_SECRET"); v != "" {
		c.Xyy.MapServiceConfig.Secret = v
	}
	if v := os.Getenv("REDIS_HOST"); v != "" {
		c.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Redis.Port = port
		}
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		c.Redis.Password = v
	}
	if v := os.Getenv("REDIS_DATABASE"); v != "" {
		if db, err := strconv.Atoi(v); err == nil {
			c.Redis.Database = db
		}
	}
	if v := os.Getenv("REDIS_REGISTRY_DATABASE"); v != "" {
		if db, err := strconv.Atoi(v); err == nil {
			c.Redis.RegistryDatabase = db
			log.Printf("[config] env override: REDIS_REGISTRY_DATABASE=%d", db)
		}
	}
	if v := os.Getenv("INTERNAL_TOKEN"); v != "" {
		c.Open.InternalToken = v
	}
	applyKafkaEnvOverrides(c)
}

// LogEffectiveRedis prints resolved Redis settings after all config layers merge.
func LogEffectiveRedis() {
	r := GlobalConfig().Redis
	regDB := r.RegistryDatabase
	if regDB < 0 {
		regDB = r.Database
	}
	separate := regDB != r.Database
	log.Printf("[config] redis: addr=%s shadowDb=%d registryDb=%d separateRegistryClient=%v",
		RedisAddr(), r.Database, regDB, separate)
}

// LogEffectiveUpstream prints the resolved upstream URLs and auth prerequisites
// after all config layers merge.
func LogEffectiveUpstream() {
	g := GlobalConfig()
	cfg := g.Xyy
	aesKey := g.Anvelink.Openapi.AesKey
	log.Printf("[config] upstream urls: paasUrl=%q openapiUrl=%q workerUrl=%q consoleUrl=%q managementUrl=%q mapServiceUrl=%q",
		cfg.PaasURL, cfg.OpenapiURL, cfg.WorkerURL, cfg.ConsoleURL, cfg.ManagementURL, cfg.MapServiceConfig.URL)
	log.Printf("[config] console auth: tokenConfigured=%v phoneConfigured=%v username=%q",
		cfg.ConsoleAuthToken != "", cfg.ConsoleAuth.Phone != "", cfg.ConsoleAuth.Username)
	log.Printf("[config] anvelink openapi aes-key configured=%v (len=%d)", len(aesKey) > 16, len(aesKey))
	log.Printf("[config] internal token configured=%v", g.Open.InternalToken != "")
}

// RedisAddr returns the host:port address for the Redis client.
func RedisAddr() string {
	r := GlobalConfig().Redis
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

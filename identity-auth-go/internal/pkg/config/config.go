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
	Redis  RedisConfig  `yaml:"redis"`
	OSS    OSSConfig    `yaml:"oss"`
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
	Secrets        []string        `yaml:"secrets"`
	MessagePool    PoolConfig      `yaml:"messagePool"`
	VoicePool      PoolConfig      `yaml:"voicePool"`
	Chuanglan      ChuanglanConfig `yaml:"chuanglan"`
	JavaServiceUrl string          `yaml:"javaServiceUrl"` // Added for shadow testing
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

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
}

// OSSConfig holds Alibaba Cloud OSS settings.
type OSSConfig struct {
	AccessKeyId     string `yaml:"accessKeyId"`
	AccessKeySecret string `yaml:"accessKeySecret"`
	Endpoint        string `yaml:"endpoint"`
	BucketName      string `yaml:"bucketName"`
	UrlPrefix       string `yaml:"urlPrefix"`
}

// ---------------------------------------------------------------------------
// Nacos YAML structures (Spring-style format)
// ---------------------------------------------------------------------------

// nacosIdentityAuthYAML maps the identity-auth.yml / aliyun_oss.yaml format from
// Nacos. The aliyun.oss section is parsed into a generic map so we can apply
// Spring-style "relaxed binding": the ops-maintained YAML may use camelCase
// (accessKeyId), kebab-case (access-key-id) or snake_case (access_key_id), all
// of which Spring's @ConfigurationProperties(prefix="aliyun.oss") accepts.
type nacosIdentityAuthYAML struct {
	Aliyun struct {
		OSS map[string]interface{} `yaml:"oss"`
	} `yaml:"aliyun"`
}

// normalizeKey lowercases a key and strips '-'/'_' so that camelCase,
// kebab-case and snake_case variants collapse to the same canonical form,
// mirroring Spring Boot relaxed property binding.
func normalizeKey(k string) string {
	k = strings.ToLower(k)
	k = strings.ReplaceAll(k, "-", "")
	k = strings.ReplaceAll(k, "_", "")
	return k
}

// nacosRedisYAML maps the redis.yaml format from Nacos (dataId=redis.yaml, group=xyy_ops).
type nacosRedisYAML struct {
	Redis struct {
		IdentityAuth struct {
			Host     string `yaml:"host"`
			Password string `yaml:"password"`
			Port     int    `yaml:"port"`
			Database int    `yaml:"database"`
		} `yaml:"identity_auth"`
	} `yaml:"redis"`
}

// nacosMySQLYAML maps the mysql.yaml format from Nacos (dataId=mysql.yaml, group=xyy_ops).
type nacosMySQLYAML struct {
	MySQL struct {
		IdentityAuth struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Database string `yaml:"database"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"identity_auth"`
		ConnectionParam string `yaml:"connection-param"`
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

// MergeNacosIdentityAuthConfig parses the identity-auth.yml content from Nacos
// and merges OSS settings into GlobalConfig. Redis config comes from redis.yaml
// via MergeNacosRedisConfig, not from this file.
func MergeNacosIdentityAuthConfig(yamlContent string) {
	var nc nacosIdentityAuthYAML
	if err := yaml.Unmarshal([]byte(yamlContent), &nc); err != nil {
		log.Printf("[config] failed to parse nacos identity-auth config: %v", err)
		return
	}

	// OSS (relaxed-binding key matching). Collect normalized keys first so the
	// assignment order is deterministic regardless of YAML/map iteration order.
	norm := make(map[string]string, len(nc.Aliyun.OSS))
	for k, raw := range nc.Aliyun.OSS {
		if raw == nil {
			continue
		}
		v := strings.TrimSpace(fmt.Sprintf("%v", raw))
		if v == "" {
			continue
		}
		norm[normalizeKey(k)] = v
	}

	if v := norm["accesskeyid"]; v != "" {
		GlobalConfig.OSS.AccessKeyId = v
	}
	if v := norm["accesskeysecret"]; v != "" {
		GlobalConfig.OSS.AccessKeySecret = v
	}
	if v := norm["endpoint"]; v != "" {
		GlobalConfig.OSS.Endpoint = v
	}
	if v := norm["bucketname"]; v != "" {
		GlobalConfig.OSS.BucketName = v
	}
	// Java's AliyunOssProperties only binds `aliyun.oss.urlPrefix`. The
	// authoritative value lives in identity-auth.yml (the main app config, which
	// outranks the aliyun_oss.yaml extension config in Spring). The `url` and
	// `key` entries in aliyun_oss.yaml are not consumed by identity-auth, so we
	// deliberately ignore them to stay byte-for-byte consistent with Java's
	// getUrl() output (note: the configured prefix has no scheme).
	if v := norm["urlprefix"]; v != "" {
		GlobalConfig.OSS.UrlPrefix = v
	}

	log.Printf("[config] merged nacos identity-auth config: oss.endpoint=%s, oss.bucketName=%s, oss.urlPrefix=%s",
		GlobalConfig.OSS.Endpoint, GlobalConfig.OSS.BucketName, GlobalConfig.OSS.UrlPrefix)
}

// MergeNacosRedisConfig parses the redis.yaml content from Nacos
// (dataId=redis.yaml, group=xyy_ops) and merges into GlobalConfig.
func MergeNacosRedisConfig(yamlContent string) {
	var nc nacosRedisYAML
	if err := yaml.Unmarshal([]byte(yamlContent), &nc); err != nil {
		log.Printf("[config] failed to parse nacos redis config: %v", err)
		return
	}

	r := nc.Redis.IdentityAuth
	if r.Host != "" {
		GlobalConfig.Redis.Host = r.Host
	}
	if r.Port != 0 {
		GlobalConfig.Redis.Port = r.Port
	}
	if r.Password != "" {
		GlobalConfig.Redis.Password = r.Password
	}
	// Database can be 0 legitimately, so always set if host is present.
	if r.Host != "" {
		GlobalConfig.Redis.Database = r.Database
	}

	log.Printf("[config] merged nacos redis config: host=%s, port=%d, db=%d",
		GlobalConfig.Redis.Host, GlobalConfig.Redis.Port, GlobalConfig.Redis.Database)
}

// MergeNacosMySQLConfig parses the mysql.yaml content from Nacos and assembles
// a Go-style DSN into GlobalConfig.
func MergeNacosMySQLConfig(yamlContent string) {
	var nc nacosMySQLYAML
	if err := yaml.Unmarshal([]byte(yamlContent), &nc); err != nil {
		log.Printf("[config] failed to parse nacos mysql config: %v", err)
		return
	}

	m := nc.MySQL.IdentityAuth
	if m.Host == "" || m.Database == "" {
		log.Printf("[config] nacos mysql config incomplete, skipping DSN assembly")
		return
	}

	port := m.Port
	if port == 0 {
		port = 3306
	}

	// Build DSN with Go-compatible query params. Nacos mysql.connection-param is
	// JDBC-style (useSSL, serverTimezone, …) and must not be passed through to
	// go-sql-driver/mysql verbatim.
	connParam := toGoMySQLParams(nc.MySQL.ConnectionParam)

	GlobalConfig.MySQL.DSN = fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?%s",
		m.Username, m.Password, m.Host, port, m.Database, connParam,
	)
	log.Printf("[config] merged nacos mysql config: host=%s, db=%s", m.Host, m.Database)
}

// jdbcMySQLParamsBlacklist lists JDBC-only query keys from mysql.connection-param
// that go-sql-driver/mysql does not understand.
var jdbcMySQLParamsBlacklist = map[string]bool{
	"usessl":               true,
	"useunicode":           true,
	"characterencoding":    true,
	"servertimezone":       true,
	"zerodatetimebehavior": true,
}

// toGoMySQLParams converts Nacos mysql.connection-param (JDBC style) into
// go-sql-driver/mysql DSN query parameters.
func toGoMySQLParams(jdbcParams string) string {
	if strings.TrimSpace(jdbcParams) == "" {
		return "charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai&tls=false"
	}

	parts := strings.Split(jdbcParams, "&")
	var valid []string
	hasLoc, hasParseTime, hasCharset, hasTLS := false, false, false, false

	for _, part := range parts {
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		key := strings.ToLower(strings.TrimSpace(kv[0]))
		if key == "" {
			continue
		}

		if jdbcMySQLParamsBlacklist[key] {
			if key == "usessl" && len(kv) > 1 {
				if strings.EqualFold(strings.TrimSpace(kv[1]), "true") {
					valid = append(valid, "tls=true")
					hasTLS = true
				} else if strings.EqualFold(strings.TrimSpace(kv[1]), "false") {
					valid = append(valid, "tls=false")
					hasTLS = true
				}
			}
			continue
		}

		switch key {
		case "loc":
			hasLoc = true
		case "parsetime":
			hasParseTime = true
		case "charset":
			hasCharset = true
		case "tls":
			hasTLS = true
		}
		valid = append(valid, part)
	}

	if !hasParseTime {
		valid = append(valid, "parseTime=True")
	}
	if !hasLoc {
		valid = append(valid, "loc=Asia%2FShanghai")
	}
	if !hasCharset {
		valid = append(valid, "charset=utf8mb4")
	}
	if !hasTLS {
		valid = append(valid, "tls=false")
	}

	return strings.Join(valid, "&")
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
	if v := os.Getenv("DRY_RUN"); v != "" {
		GlobalConfig.DryRun = strings.EqualFold(v, "true") || v == "1"
		log.Printf("[config] env override: DRY_RUN=%v", GlobalConfig.DryRun)
	}
	if v := os.Getenv("JAVA_SERVICE_URL"); v != "" {
		GlobalConfig.Xyy.JavaServiceUrl = v
		log.Printf("[config] env override: JAVA_SERVICE_URL=%s", v)
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

	// Redis env overrides
	if v := os.Getenv("REDIS_HOST"); v != "" {
		GlobalConfig.Redis.Host = v
		log.Printf("[config] env override: REDIS_HOST=%s", v)
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			GlobalConfig.Redis.Port = port
			log.Printf("[config] env override: REDIS_PORT=%d", port)
		}
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		GlobalConfig.Redis.Password = v
		log.Printf("[config] env override: REDIS_PASSWORD")
	}
	if v := os.Getenv("REDIS_DATABASE"); v != "" {
		if db, err := strconv.Atoi(v); err == nil {
			GlobalConfig.Redis.Database = db
			log.Printf("[config] env override: REDIS_DATABASE=%d", db)
		}
	}

	// OSS env overrides
	if v := os.Getenv("OSS_ACCESS_KEY_ID"); v != "" {
		GlobalConfig.OSS.AccessKeyId = v
		log.Printf("[config] env override: OSS_ACCESS_KEY_ID")
	}
	if v := os.Getenv("OSS_ACCESS_KEY_SECRET"); v != "" {
		GlobalConfig.OSS.AccessKeySecret = v
		log.Printf("[config] env override: OSS_ACCESS_KEY_SECRET")
	}
	if v := os.Getenv("OSS_ENDPOINT"); v != "" {
		GlobalConfig.OSS.Endpoint = v
		log.Printf("[config] env override: OSS_ENDPOINT=%s", v)
	}
	if v := os.Getenv("OSS_BUCKET_NAME"); v != "" {
		GlobalConfig.OSS.BucketName = v
		log.Printf("[config] env override: OSS_BUCKET_NAME=%s", v)
	}
	if v := os.Getenv("OSS_URL_PREFIX"); v != "" {
		GlobalConfig.OSS.UrlPrefix = v
		log.Printf("[config] env override: OSS_URL_PREFIX=%s", v)
	}
}

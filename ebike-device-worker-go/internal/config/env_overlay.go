package config

import (
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

func envSet(keys ...string) bool {
	for _, k := range keys {
		if strings.TrimSpace(os.Getenv(k)) != "" {
			return true
		}
	}
	return false
}

func envFirst(keys ...string) (string, bool) {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v, true
		}
	}
	return "", false
}

func applyNacosEnv(c *Config) {
	if v, ok := envFirst("NACOS_SERVER_ADDR"); ok {
		host, port := splitHostPort(v, 8848)
		c.Nacos.ServerAddr = host
		c.Nacos.Port = port
	} else if v, ok := envFirst("NACOS_SERVERADDR"); ok {
		c.Nacos.ServerAddr = v
	}
	if v, ok := envFirst("NACOS_PORT"); ok {
		if port, err := strconv.ParseUint(v, 10, 64); err == nil {
			c.Nacos.Port = port
		}
	}
	if v, ok := envFirst("NACOS_NAMESPACE"); ok {
		c.Nacos.Namespace = v
	}
	if v, ok := envFirst("NACOS_GROUP"); ok {
		c.Nacos.Group = v
	}
	if v, ok := envFirst("NACOS_REGISTER_ENABLED", "SPRING_CLOUD_NACOS_DISCOVERY_REGISTER_ENABLED", "NACOS_REGISTERENABLED"); ok {
		c.Nacos.RegisterEnabled = envBool(v)
	}
}

func applyFastIDEnv(c *Config) {
	if v, ok := envFirst("SPRING_XYY_FASTID_SECRET"); ok {
		c.Spring.Xyy.Fastid.Secret = v
	}
	serverURL, hasServerURL := envFirst("SPRING_XYY_FASTID_SERVER_URL", "SPRING_XYY_FASTID_URL")
	if v, ok := envFirst("SPRING_XYY_FASTID_SERVER_ADDR"); ok {
		c.Spring.Xyy.Fastid.ServerAddr = v
		if !hasServerURL {
			c.Spring.Xyy.Fastid.ServerURL = ""
		}
		applyFastIDSchemeFromAddr(&c.Spring.Xyy.Fastid)
	}
	if hasServerURL {
		c.Spring.Xyy.Fastid.ServerURL = serverURL
		applyFastIDSchemeFromAddr(&c.Spring.Xyy.Fastid)
	}
	if v, ok := envFirst("SPRING_XYY_FASTID_NAMESPACE"); ok {
		c.Spring.Xyy.Fastid.Namespace = v
	}
	if v, ok := envFirst("SPRING_XYY_FASTID_GROUPID"); ok {
		c.Spring.Xyy.Fastid.GroupID = v
	}
	if v, ok := envFirst("SPRING_XYY_FASTID_APP_NAME"); ok {
		c.Spring.Xyy.Fastid.AppName = v
	}
	if v, ok := envFirst("SPRING_XYY_FASTID_PORT"); ok {
		if port, err := strconv.Atoi(v); err == nil {
			c.Spring.Xyy.Fastid.Port = port
		}
	}
	if v, ok := envFirst("SPRING_XYY_FASTID_ENABLED"); ok {
		c.Spring.Xyy.Fastid.Enabled = envBool(v)
	}
	if v, ok := envFirst("SPRING_XYY_FASTID_USE_HTTPS"); ok {
		c.Spring.Xyy.Fastid.UseHTTPS = envBool(v)
		c.Spring.Xyy.Fastid.UseHTTPSExplicit = true
	}
}

func applyRedisEnv(c *Config) {
	if v, ok := envFirst("REDIS_ADDR"); ok {
		c.Redis.Addr = v
	}
	if v, ok := envFirst("REDIS_PASSWORD"); ok {
		c.Redis.Password = v
	}
	if v, ok := envFirst("REDIS_PREFIX"); ok {
		c.Redis.Prefix = v
	}
	if envSet("REDIS_DB") {
		c.Redis.DB = viper.GetInt("redis.db")
	}
}

func applyServerEnv(c *Config) {
	if v, ok := envFirst("SERVER_PORT"); ok {
		c.Server.Port = v
	}
}

func applyProxyEnv(c *Config) {
	if v, ok := envFirst("PROXY_TARGET_URL"); ok {
		c.Proxy.TargetUrl = v
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

// applyEnvOverrides merges environment variables (highest precedence).
func applyPersistEnv(c *Config) {
	if v, ok := envFirst("PERSIST_CONFIG_TABLE_SUFFIX", "TABLE_SUFFIX"); ok {
		c.PersistConfig.TableSuffix = v
	}
}

func applyEnvOverrides(c *Config) {
	if c == nil {
		return
	}
	applyNacosEnv(c)
	applyFastIDEnv(c)
	applyRedisEnv(c)
	applyServerEnv(c)
	applyProxyEnv(c)
	applyKafkaEnvOverrides(c)
	applyDatabaseEnvOverrides(c)
	applyPersistEnv(c)
}

func envBool(v string) bool {
	switch strings.ToLower(strings.TrimSpace(strings.Trim(v, `"'`))) {
	case "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}

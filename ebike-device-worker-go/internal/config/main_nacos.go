package config

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"
)

var mainNacosDataIDs = []string{"ebike-device-worker.yml", "ebike-device-worker.yaml"}

type mainNacosFile struct {
	Server struct {
		Port interface{} `yaml:"port"`
	} `yaml:"server"`
	Spring struct {
		Application struct {
			Name string `yaml:"name"`
		} `yaml:"application"`
		Datasource struct {
			URL      string `yaml:"url"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"datasource"`
		Kafka struct {
			Topic struct {
				DataTopic   string `yaml:"data-topic"`
				EventTopic  string `yaml:"event-topic"`
				AlarmTopic  string `yaml:"alarm-topic"`
				ToSaasTopic string `yaml:"to-saas-topic"`
				GrayToSaas  string `yaml:"gray-to-saas-topic"`
			} `yaml:"topic"`
			Consumer struct {
				GroupIDData           string `yaml:"group-id-data"`
				GroupIDEvent          string `yaml:"group-id-event"`
				GroupIDAlarm          string `yaml:"group-id-alarm"`
				DataTopicConcurrency  int    `yaml:"data-topic-concurrency"`
				EventTopicConcurrency int    `yaml:"event-topic-concurrency"`
				AlarmTopicConcurrency int    `yaml:"alarm-topic-concurrency"`
			} `yaml:"consumer"`
		} `yaml:"kafka"`
	} `yaml:"spring"`
	Database struct {
		Dsn string `yaml:"dsn"`
	} `yaml:"database"`
	PersistConfig struct {
		PersistLimitSize    int     `yaml:"persist-limit-size"`
		PersistBatchSize    int     `yaml:"persist-batch-size"`
		PersistIntervalMill int     `yaml:"persist-interval-mill"`
		TableSuffix         *string `yaml:"table-suffix"`
	} `yaml:"persist-config"`
	Caffeine struct {
		DeviceTenantMappingCacheSize int `yaml:"device-tenant-mapping-cache-size"`
		TenantCacheSize              int `yaml:"tenant-cache-size"`
		Timeout                      int `yaml:"timeout"`
	} `yaml:"caffeine"`
}

func applyMainFromNacos(client config_client.IConfigClient) error {
	if client == nil {
		return fmt.Errorf("nacos config client is nil")
	}
	group := MainGroup()
	content, dataID, err := fetchMainNacosContent(client, group)
	if err != nil {
		return err
	}
	if content == "" {
		return fmt.Errorf("main config not found in nacos group %s (tried %v)", group, mainNacosDataIDs)
	}
	if err := applyMainYAML(content); err != nil {
		return err
	}
	log.Printf("[nacos] loaded main config from %s (group=%s, app=%s, database=%t)",
		dataID, group, applicationName(GlobalConfig), GlobalConfig.Database.Dsn != "")
	return nil
}

func fetchMainNacosContent(client config_client.IConfigClient, group string) (content, dataID string, err error) {
	for _, id := range mainNacosDataIDs {
		content, err = client.GetConfig(vo.ConfigParam{
			DataId: id,
			Group:  group,
		})
		if err != nil {
			log.Printf("[WARN] get %s from nacos group %s: %v", id, group, err)
			continue
		}
		if content != "" {
			return content, id, nil
		}
	}
	return "", "", nil
}

func applyMainYAML(content string) error {
	var raw mainNacosFile
	if err := yaml.Unmarshal([]byte(content), &raw); err != nil {
		return fmt.Errorf("parse main config: %w", err)
	}
	mergeMainSettings(raw)
	return nil
}

func mergeMainSettings(raw mainNacosFile) {
	if name := strings.TrimSpace(raw.Spring.Application.Name); name != "" && !isUnresolvedPlaceholder(name) {
		GlobalConfig.Spring.Application.Name = name
	}
	if port := stringifyPort(raw.Server.Port); port != "" {
		GlobalConfig.Server.Port = port
	}
	if raw.Database.Dsn != "" {
		GlobalConfig.Database.Dsn = raw.Database.Dsn
	}
	ds := raw.Spring.Datasource
	if ds.URL != "" {
		dsn, err := buildPostgresDSN(ds.URL, ds.Username, ds.Password)
		if err != nil {
			log.Printf("[WARN] build postgres dsn from spring.datasource: %v", err)
		} else if dsn != "" {
			GlobalConfig.Database.Dsn = dsn
		}
	}
	mergeKafkaTopicsFromMain(raw)
	mergeKafkaConsumersFromMain(raw)
	mergePersistAndCaffeineFromMain(raw)
}

func mergePersistAndCaffeineFromMain(raw mainNacosFile) {
	pc := raw.PersistConfig
	if pc.PersistBatchSize > 0 {
		GlobalConfig.PersistConfig.PersistBatchSize = pc.PersistBatchSize
	}
	if pc.PersistLimitSize > 0 {
		GlobalConfig.PersistConfig.PersistLimitSize = pc.PersistLimitSize
	}
	if pc.PersistIntervalMill > 0 {
		GlobalConfig.PersistConfig.PersistIntervalMs = pc.PersistIntervalMill
	}
	if pc.TableSuffix != nil {
		GlobalConfig.PersistConfig.TableSuffix = *pc.TableSuffix
	}
	cf := raw.Caffeine
	if cf.DeviceTenantMappingCacheSize > 0 {
		GlobalConfig.Caffeine.DeviceTenantMappingCacheSize = cf.DeviceTenantMappingCacheSize
	}
	if cf.TenantCacheSize > 0 {
		GlobalConfig.Caffeine.TenantCacheSize = cf.TenantCacheSize
	}
	if cf.Timeout > 0 {
		GlobalConfig.Caffeine.TimeoutSec = cf.Timeout
	}
}

func stringifyPort(v interface{}) string {
	switch p := v.(type) {
	case nil:
		return ""
	case int:
		if p > 0 {
			return fmt.Sprintf("%d", p)
		}
	case int64:
		if p > 0 {
			return fmt.Sprintf("%d", p)
		}
	case float64:
		if p > 0 {
			return fmt.Sprintf("%d", int(p))
		}
	case string:
		return strings.TrimSpace(p)
	}
	return ""
}

func buildPostgresDSN(jdbcURL, username, password string) (string, error) {
	jdbcURL = strings.TrimSpace(jdbcURL)
	if jdbcURL == "" {
		return "", nil
	}
	if strings.HasPrefix(jdbcURL, "postgres://") || strings.HasPrefix(jdbcURL, "postgresql://") {
		return normalizePostgresDSN(jdbcURL, username, password)
	}

	raw := jdbcURL
	if strings.HasPrefix(raw, "jdbc:") {
		raw = raw[len("jdbc:"):]
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse jdbc url: %w", err)
	}
	if username != "" {
		parsed.User = url.UserPassword(username, password)
	}
	parsed.Scheme = "postgres"
	parsed.RawQuery = sanitizePostgresQuery(parsed.Query()).Encode()
	return parsed.String(), nil
}

func normalizePostgresDSN(rawURL, username, password string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if parsed.User == nil && username != "" {
		parsed.User = url.UserPassword(username, password)
	}
	parsed.Scheme = "postgres"
	parsed.RawQuery = sanitizePostgresQuery(parsed.Query()).Encode()
	return parsed.String(), nil
}

// postgresAllowedQueryParams are libpq/pgx-safe URI query keys (lowercase).
var postgresAllowedQueryParams = map[string]struct{}{
	"sslmode":                   {},
	"sslcert":                   {},
	"sslkey":                    {},
	"sslrootcert":               {},
	"sslcrl":                    {},
	"sslcompression":            {},
	"sslpassword":               {},
	"sslsni":                    {},
	"search_path":               {},
	"connect_timeout":           {},
	"application_name":          {},
	"fallback_application_name": {},
	"client_encoding":           {},
	"timezone":                  {},
	"options":                   {},
	"keepalives":                {},
	"keepalives_idle":           {},
	"keepalives_interval":       {},
	"keepalives_count":          {},
	"tcp_user_timeout":          {},
	"replication":               {},
	"gssencmode":                {},
	"channel_binding":           {},
	"target_session_attrs":      {},
}

func sanitizePostgresQuery(q url.Values) url.Values {
	out := make(url.Values)
	for key, vals := range q {
		lower := strings.ToLower(key)
		if lower == "currentschema" {
			if len(vals) > 0 && vals[0] != "" {
				out.Set("search_path", vals[0])
			}
			continue
		}
		if _, ok := postgresAllowedQueryParams[lower]; !ok {
			continue
		}
		out[key] = vals
	}
	if out.Get("sslmode") == "" {
		out.Set("sslmode", "disable")
	}
	return out
}

// MainGroup returns the Nacos group for the primary application config (same as Java bootstrap group).
func MainGroup() string {
	if GlobalConfig == nil || GlobalConfig.Nacos.Group == "" {
		return "xyy"
	}
	return GlobalConfig.Nacos.Group
}

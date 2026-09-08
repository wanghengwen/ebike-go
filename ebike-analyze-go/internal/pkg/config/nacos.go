package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"
)

// ──────────────────────────────────────────────
// Constants
// ──────────────────────────────────────────────

const (
	redisNacosDataID = "redis.yaml"
	mysqlNacosDataID = "mysql.yaml"

	mysqlAnalyzeDataSource = "ebike_analyze"
	mysqlVisualDataSource  = "ebike_visual"

	defaultRedisPort      = 6379
	defaultRedisDB        = 0
	defaultAnalyzeRedisDB = 13 // mirrors Java ebike-analyze.yml spring.redis.database
)

var redisAnalyzeDataSourceKeys = []string{"ebike_order", "ebike-analyze", "ebike_analyze"}

// ──────────────────────────────────────────────
// OpsGroup
// ──────────────────────────────────────────────

// OpsGroup returns the Nacos group for shared ops configs (redis.yaml, mysql.yaml).
func OpsGroup() string {
	return GlobalConfig.Nacos.Group + "_ops"
}

// ──────────────────────────────────────────────
// ApplyMainNacosYAML
// ──────────────────────────────────────────────

// ApplyMainNacosYAML merges ebike-analyze-go.yaml from Nacos into GlobalConfig
// without clearing MySQL/Redis values that were already loaded from ops extension configs.
func ApplyMainNacosYAML(content string) error {
	var incoming AppConfig
	if err := yaml.Unmarshal([]byte(content), &incoming); err != nil {
		return err
	}
	if incoming.Server.Port > 0 || incoming.Server.Name != "" {
		GlobalConfig.Server = incoming.Server
	}
	// Merge xyy config
	mergeXyyIncoming(&GlobalConfig.Xyy, incoming.Xyy)
	if incoming.Xyy.Proxy.TargetUrl != "" {
		GlobalConfig.Xyy.Proxy.TargetUrl = incoming.Xyy.Proxy.TargetUrl
	}
	if len(incoming.Xyy.Proxy.LiveList) > 0 {
		GlobalConfig.Xyy.Proxy.LiveList = incoming.Xyy.Proxy.LiveList
	}
	if len(incoming.Xyy.Proxy.RecordList) > 0 {
		GlobalConfig.Xyy.Proxy.RecordList = incoming.Xyy.Proxy.RecordList
	}
	// Merge Redis
	if incoming.Redis.Addr != "" {
		GlobalConfig.Redis = incoming.Redis
	}
	// Merge MySQL
	if incoming.MySQL.AnalyzeDSN != "" {
		GlobalConfig.MySQL.AnalyzeDSN = incoming.MySQL.AnalyzeDSN
	}
	if incoming.MySQL.VisualDSN != "" {
		GlobalConfig.MySQL.VisualDSN = incoming.MySQL.VisualDSN
	}
	if incoming.MySQL.MaxOpenConns > 0 {
		GlobalConfig.MySQL.MaxOpenConns = incoming.MySQL.MaxOpenConns
	}
	if incoming.MySQL.MaxIdleConns > 0 {
		GlobalConfig.MySQL.MaxIdleConns = incoming.MySQL.MaxIdleConns
	}
	// Merge ES
	mergeESIncoming(&GlobalConfig.ES, incoming.ES)
	return nil
}

// mergeXyyIncoming overlays non-empty xyy fields from incoming onto dst.
func mergeXyyIncoming(dst *XyyConfig, src XyyConfig) {
	if len(src.TableSplit.TableNames) > 0 {
		dst.TableSplit.Enable = src.TableSplit.Enable
		dst.TableSplit.TableNames = src.TableSplit.TableNames
	}
}

// mergeESIncoming overlays non-empty ES fields from src onto dst.
func mergeESIncoming(dst *ESConfig, src ESConfig) {
	if src.Hostname != "" {
		dst.Hostname = src.Hostname
	}
	if src.Username != "" {
		dst.Username = src.Username
	}
	if src.Password != "" {
		dst.Password = src.Password
	}
	if src.ConnectTimeout > 0 {
		dst.ConnectTimeout = src.ConnectTimeout
	}
	if src.SocketTimeout > 0 {
		dst.SocketTimeout = src.SocketTimeout
	}
	if src.TrackTotalHitsUpTo > 0 {
		dst.TrackTotalHitsUpTo = src.TrackTotalHitsUpTo
	}
}

// ──────────────────────────────────────────────
// ApplyRedisFromNacos
// ──────────────────────────────────────────────

type redisNacosFile struct {
	Redis redisNacosRoot `yaml:"redis"`
}

type redisNacosRoot struct {
	EbikeOrder        redisDataSource `yaml:"ebike_order"`
	EbikeAnalyze      redisDataSource `yaml:"ebike-analyze"`
	EbikeAnalyzeSnake redisDataSource `yaml:"ebike_analyze"`
}

type redisDataSource struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	Database int    `yaml:"database"`
}

func (d redisDataSource) toAddr() string {
	port := d.Port
	if port == 0 {
		port = defaultRedisPort
	}
	addr := d.Host
	if addr != "" && !strings.Contains(addr, ":") {
		addr = fmt.Sprintf("%s:%d", d.Host, port)
	}
	return addr
}

func (d redisDataSource) resolveDB(fallback int) int {
	if d.DB != 0 {
		return d.DB
	}
	if d.Database != 0 {
		return d.Database
	}
	return fallback
}

// ApplyRedisFromNacos loads redis config from Nacos redis.yaml (group={group}_ops).
func ApplyRedisFromNacos(client config_client.IConfigClient) error {
	if client == nil {
		return fmt.Errorf("nacos config client is nil")
	}
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: redisNacosDataID,
		Group:  OpsGroup(),
	})
	if err != nil {
		return fmt.Errorf("get %s from nacos group %s: %w", redisNacosDataID, OpsGroup(), err)
	}
	if content == "" {
		return fmt.Errorf("%s is empty in nacos group %s", redisNacosDataID, OpsGroup())
	}
	return applyRedisYAML(content)
}

func applyRedisYAML(content string) error {
	var raw redisNacosFile
	if err := yaml.Unmarshal([]byte(content), &raw); err != nil {
		return fmt.Errorf("parse redis.yaml: %w", err)
	}
	ds, key, dbFallback := pickRedisDataSource(raw.Redis)
	if key == "" {
		return fmt.Errorf("redis.yaml has no analyze data source (tried %v)", redisAnalyzeDataSourceKeys)
	}
	if ds.Host == "" {
		return fmt.Errorf("redis.yaml %q has empty host", key)
	}
	GlobalConfig.Redis.Addr = ds.toAddr()
	GlobalConfig.Redis.Password = ds.Password
	GlobalConfig.Redis.DB = ds.resolveDB(dbFallback)
	log.Printf("Redis config loaded from Nacos %s (group=%s, key=%s, addr=%s, db=%d)",
		redisNacosDataID, OpsGroup(), key, GlobalConfig.Redis.Addr, GlobalConfig.Redis.DB)
	return nil
}

func pickRedisDataSource(root redisNacosRoot) (redisDataSource, string, int) {
	candidates := []struct {
		key string
		ds  redisDataSource
		db  int
	}{
		{"ebike_order", root.EbikeOrder, defaultAnalyzeRedisDB},
		{"ebike-analyze", root.EbikeAnalyze, defaultAnalyzeRedisDB},
		{"ebike_analyze", root.EbikeAnalyzeSnake, defaultAnalyzeRedisDB},
	}
	for _, c := range candidates {
		if c.ds.Host != "" {
			return c.ds, c.key, c.db
		}
	}
	return redisDataSource{}, "", defaultRedisDB
}

// ListenRedisFromNacos watches redis.yaml and updates GlobalConfig.Redis on change.
func ListenRedisFromNacos(client config_client.IConfigClient, onChange func()) error {
	if client == nil {
		return fmt.Errorf("nacos config client is nil")
	}
	return client.ListenConfig(vo.ConfigParam{
		DataId: redisNacosDataID,
		Group:  OpsGroup(),
		OnChange: func(namespace, group, dataId, data string) {
			log.Printf("Nacos redis config changed: dataId=%s group=%s", dataId, group)
			if err := applyRedisYAML(data); err != nil {
				log.Printf("[ERROR] Failed to apply redis.yaml change: %v", err)
				return
			}
			if onChange != nil {
				onChange()
			}
		},
	})
}

// ──────────────────────────────────────────────
// ApplyMySQLFromNacos
// ──────────────────────────────────────────────

type mysqlNacosFile struct {
	MySQL mysqlNacosRoot `yaml:"mysql"`
}

type mysqlNacosRoot struct {
	ConnectionParam string          `yaml:"connection-param"`
	DriverClassName string          `yaml:"driver-class-name"`
	Type            string          `yaml:"type"`
	EbikeAnalyze    mysqlDataSource `yaml:"ebike_analyze"`
	EbikeVisual     mysqlDataSource `yaml:"ebike_visual"`
}

type mysqlDataSource struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (d mysqlDataSource) buildDSN() string {
	goParam := "parseTime=true&loc=Asia%2FShanghai&charset=utf8mb4&timeout=5s&readTimeout=20s&writeTimeout=10s"
	port := d.Port
	if port == 0 {
		port = 3306
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		d.Username, d.Password, d.Host, port, d.Database, goParam)
}

// ApplyMySQLFromNacos loads dual data-source DSNs (ebike_analyze + ebike_visual) from Nacos mysql.yaml.
// Matches Java bootstrap extension-config: data-id=mysql.yaml, group={group}_ops.
func ApplyMySQLFromNacos(client config_client.IConfigClient) error {
	if client == nil {
		return fmt.Errorf("nacos config client is nil")
	}
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: mysqlNacosDataID,
		Group:  OpsGroup(),
	})
	if err != nil {
		return fmt.Errorf("get %s from nacos group %s: %w", mysqlNacosDataID, OpsGroup(), err)
	}
	if content == "" {
		return fmt.Errorf("%s is empty in nacos group %s", mysqlNacosDataID, OpsGroup())
	}
	return applyMySQLYAML(content)
}

func applyMySQLYAML(content string) error {
	var raw mysqlNacosFile
	if err := yaml.Unmarshal([]byte(content), &raw); err != nil {
		return fmt.Errorf("parse mysql.yaml: %w", err)
	}

	analyzeDS := raw.MySQL.EbikeAnalyze
	if analyzeDS.Host == "" || analyzeDS.Database == "" || analyzeDS.Username == "" {
		return fmt.Errorf("mysql.yaml %q has incomplete connection fields", mysqlAnalyzeDataSource)
	}
	GlobalConfig.MySQL.AnalyzeDSN = analyzeDS.buildDSN()
	log.Printf("MySQL analyzeDSN loaded from Nacos %s (group=%s, database=%s)",
		mysqlNacosDataID, OpsGroup(), analyzeDS.Database)

	visualDS := raw.MySQL.EbikeVisual
	if visualDS.Host == "" || visualDS.Database == "" || visualDS.Username == "" {
		log.Printf("[WARN] mysql.yaml %q has incomplete connection fields, skipping visualDSN", mysqlVisualDataSource)
	} else {
		GlobalConfig.MySQL.VisualDSN = visualDS.buildDSN()
		log.Printf("MySQL visualDSN loaded from Nacos %s (group=%s, database=%s)",
			mysqlNacosDataID, OpsGroup(), visualDS.Database)
	}

	return nil
}

// ListenMySQLFromNacos watches mysql.yaml and updates GlobalConfig MySQL DSNs on change.
func ListenMySQLFromNacos(client config_client.IConfigClient, onChange func()) error {
	if client == nil {
		return fmt.Errorf("nacos config client is nil")
	}
	return client.ListenConfig(vo.ConfigParam{
		DataId: mysqlNacosDataID,
		Group:  OpsGroup(),
		OnChange: func(namespace, group, dataId, data string) {
			log.Printf("Nacos mysql config changed: dataId=%s group=%s", dataId, group)
			if err := applyMySQLYAML(data); err != nil {
				log.Printf("[ERROR] Failed to apply mysql.yaml change: %v", err)
				return
			}
			if onChange != nil {
				onChange()
			}
		},
	})
}

// ──────────────────────────────────────────────
// ApplyESFromNacos
// ──────────────────────────────────────────────

// esNacosYAML represents the ES config in the main Nacos config.
// ES config may appear under top-level "es" or "xyy.es".
type esNacosYAML struct {
	Xyy struct {
		ES ESConfig `yaml:"es"`
	} `yaml:"xyy"`
	ES ESConfig `yaml:"es"`
}

// ApplyESFromNacos extracts ES configuration from the main Nacos YAML content.
// ES config may appear under xyy.es or top-level es in the Nacos config.
func ApplyESFromNacos(content string) error {
	var raw esNacosYAML
	if err := yaml.Unmarshal([]byte(content), &raw); err != nil {
		return fmt.Errorf("parse ES config from nacos: %w", err)
	}

	// Prefer top-level es, fall back to xyy.es
	es := raw.ES
	if es.Hostname == "" {
		es = raw.Xyy.ES
	}

	if es.Hostname == "" {
		log.Printf("[WARN] No ES config found in Nacos content")
		return nil
	}

	mergeESIncoming(&GlobalConfig.ES, es)
	log.Printf("ES config loaded from Nacos (hostname=%s, username=%s)",
		GlobalConfig.ES.Hostname, GlobalConfig.ES.Username)
	return nil
}

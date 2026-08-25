package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"
)

const (
	redisNacosDataID     = "redis.yaml"
	redisFenceDataSource = "ebike_fence"
	defaultRedisPort     = 6379
	defaultRedisDB       = 6 // production ebike_fence uses database 6 (see Nacos redis.yaml)
)

type redisNacosFile struct {
	Redis redisNacosRoot `yaml:"redis"`
}

type redisNacosRoot struct {
	EbikeFence redisDataSource `yaml:"ebike_fence"`
}

type redisDataSource struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	Database int    `yaml:"database"` // Nacos redis.yaml uses database (same as Java spring.redis.database)
}

func (d redisDataSource) toRedisConfig() RedisConfig {
	port := d.Port
	if port == 0 {
		port = defaultRedisPort
	}
	addr := d.Host
	if addr != "" && !strings.Contains(addr, ":") {
		addr = fmt.Sprintf("%s:%d", d.Host, port)
	}
	db := d.DB
	if db == 0 {
		db = d.Database
	}
	if db == 0 {
		db = defaultRedisDB
	}
	return RedisConfig{
		Addr:     addr,
		Password: d.Password,
		DB:       db,
	}
}

// RedisConfig is the runtime Redis connection settings.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// ApplyRedisFromNacos loads redis.ebike_fence from Nacos redis.yaml.
// Matches Java bootstrap extension-config: data-id=redis.yaml, group={group}_ops.
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
	ds := raw.Redis.EbikeFence
	if ds.Host == "" {
		return fmt.Errorf("redis.yaml %q has empty host", redisFenceDataSource)
	}
	cfg := ds.toRedisConfig()
	GlobalConfig.Redis.Addr = cfg.Addr
	GlobalConfig.Redis.Password = cfg.Password
	GlobalConfig.Redis.DB = cfg.DB
	log.Printf("Redis config loaded from Nacos %s (group=%s, addr=%s, db=%d)",
		redisNacosDataID, OpsGroup(), cfg.Addr, cfg.DB)
	return nil
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

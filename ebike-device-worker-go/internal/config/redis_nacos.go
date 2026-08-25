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
	redisNacosDataID = "redis.yaml"
	defaultRedisPort = 6379
)

var redisWorkerDataSourceKeys = []string{"ebike_device_worker", "ebike-device-worker"}

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

func (d redisDataSource) toRuntime() (addr, password string, db int) {
	port := d.Port
	if port == 0 {
		port = defaultRedisPort
	}
	addr = d.Host
	if addr != "" && !strings.Contains(addr, ":") {
		addr = fmt.Sprintf("%s:%d", d.Host, port)
	}
	db = d.DB
	if db == 0 {
		db = d.Database
	}
	return addr, d.Password, db
}

func applyRedisFromNacos(client config_client.IConfigClient) error {
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
	ds, key, ok := pickRedisWorkerDataSource(raw.Redis)
	if !ok {
		return fmt.Errorf("redis.yaml missing datasource (tried %v)", redisWorkerDataSourceKeys)
	}
	addr, password, db := ds.toRuntime()
	if addr == "" {
		return fmt.Errorf("redis.yaml %q has empty host", key)
	}
	GlobalConfig.Redis.Addr = addr
	GlobalConfig.Redis.Password = password
	GlobalConfig.Redis.DB = db
	log.Printf("[nacos] loaded redis from %s (group=%s, key=%s, addr=%s, db=%d)",
		redisNacosDataID, OpsGroup(), key, addr, db)
	return nil
}

func pickRedisWorkerDataSource(redisRoot map[string]redisDataSource) (redisDataSource, string, bool) {
	if redisRoot == nil {
		return redisDataSource{}, "", false
	}
	for _, key := range redisWorkerDataSourceKeys {
		if ds, ok := redisRoot[key]; ok && ds.Host != "" {
			return ds, key, true
		}
	}
	return redisDataSource{}, "", false
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
			log.Printf("[nacos] redis config changed: dataId=%s group=%s", dataId, group)
			if err := applyRedisYAML(data); err != nil {
				log.Printf("[ERROR] apply redis.yaml change: %v", err)
				return
			}
			if onChange != nil {
				onChange()
			}
		},
	})
}

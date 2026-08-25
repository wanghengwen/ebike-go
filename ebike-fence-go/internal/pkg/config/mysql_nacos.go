package config

import (
	"fmt"
	"log"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"
)

const (
	mysqlNacosDataID     = "mysql.yaml"
	mysqlFenceDataSource = "ebike_fence"
	defaultMySQLParams   = "useSSL=false&useUnicode=true&characterEncoding=UTF-8&serverTimezone=GMT%2B8&zeroDateTimeBehavior=convertToNull"
)

type mysqlNacosFile struct {
	MySQL mysqlNacosRoot `yaml:"mysql"`
}

type mysqlNacosRoot struct {
	ConnectionParam string           `yaml:"connection-param"`
	EbikeFence      mysqlDataSource  `yaml:"ebike_fence"`
}

type mysqlDataSource struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (d mysqlDataSource) buildDSN(connectionParam string) string {
	// Java JDBC parameters are incompatible with go-sql-driver/mysql.
	// We replace them with standard Go parameters: parseTime=true, loc, charset.
	goParam := "parseTime=true&loc=Asia%2FShanghai&charset=utf8mb4"

	port := d.Port
	if port == 0 {
		port = 3306
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		d.Username, d.Password, d.Host, port, d.Database, goParam)
}

// OpsGroup returns the Nacos group for shared ops configs (redis.yaml, mysql.yaml).
func OpsGroup() string {
	return GlobalConfig.Nacos.Group + "_ops"
}

// ApplyMySQLFromNacos loads ebike_fence datasource from Nacos mysql.yaml.
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
	ds := raw.MySQL.EbikeFence
	if ds.Host == "" || ds.Database == "" || ds.Username == "" {
		return fmt.Errorf("mysql.yaml %q has incomplete connection fields", mysqlFenceDataSource)
	}
	GlobalConfig.MySQL.DSN = ds.buildDSN(raw.MySQL.ConnectionParam)
	log.Printf("MySQL DSN loaded from Nacos %s (group=%s, database=%s)",
		mysqlNacosDataID, OpsGroup(), ds.Database)
	return nil
}

// ListenMySQLFromNacos watches mysql.yaml and updates GlobalConfig.MySQL.DSN on change.
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

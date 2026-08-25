package config

import "gopkg.in/yaml.v3"

// ApplyMainNacosYAML merges ebike-fence-go.yaml without clearing MySQL/Redis loaded from ops extension configs.
func ApplyMainNacosYAML(content string) error {
	var incoming AppConfig
	if err := yaml.Unmarshal([]byte(content), &incoming); err != nil {
		return err
	}
	if incoming.Server.Port > 0 || incoming.Server.Name != "" {
		GlobalConfig.Server = incoming.Server
	}
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
	if len(incoming.Kafka.Brokers) > 0 || incoming.Kafka.TopicPrefix != "" {
		GlobalConfig.Kafka = incoming.Kafka
	}
	if incoming.Redis.Addr != "" {
		GlobalConfig.Redis = incoming.Redis
	}
	if incoming.MySQL.DSN != "" {
		GlobalConfig.MySQL.DSN = incoming.MySQL.DSN
	}
	if incoming.MySQL.MaxOpenConns > 0 {
		GlobalConfig.MySQL.MaxOpenConns = incoming.MySQL.MaxOpenConns
	}
	if incoming.MySQL.MaxIdleConns > 0 {
		GlobalConfig.MySQL.MaxIdleConns = incoming.MySQL.MaxIdleConns
	}
	return nil
}

package config

import (
	"log"
	"os"
	"strings"
)

// applyKafkaEnvOverrides applies Kafka-related env vars after file/Nacos merge.
// Use KAFKA_CONSUMERS_GROUP_* to pin consumer groups for verify deployments so
// Nacos kafka.yaml (Java groups) does not overwrite them on reload.
func applyKafkaEnvOverrides(c *Config) {
	if c == nil {
		return
	}
	if v := os.Getenv("KAFKA_ENABLED"); v != "" {
		c.Kafka.Enabled = envBool(v)
	}
	if v := strings.TrimSpace(os.Getenv("KAFKA_CONSUMERS_GROUP_DATA")); v != "" {
		c.Kafka.Consumers.GroupData = v
	}
	if v := strings.TrimSpace(os.Getenv("KAFKA_CONSUMERS_GROUP_EVENT")); v != "" {
		c.Kafka.Consumers.GroupEvent = v
	}
	if v := strings.TrimSpace(os.Getenv("KAFKA_CONSUMERS_GROUP_ALARM")); v != "" {
		c.Kafka.Consumers.GroupAlarm = v
	}
	if v := os.Getenv("KAFKA_PUSH_ENABLED"); v != "" {
		c.Kafka.PushEnabled = envBool(v)
	}
	if v := strings.TrimSpace(os.Getenv("KAFKA_CONSUMERS_START_OFFSET")); v != "" {
		c.Kafka.Consumers.StartOffset = v
	}
	applyKafkaTopicEnvOverrides(c)
}

// LogKafkaConsumerGroups prints effective consumer groups at startup/restart.
func LogKafkaConsumerGroups() {
	if GlobalConfig == nil {
		return
	}
	k := GlobalConfig.Kafka
	log.Printf("[kafka] consumers data=%s event=%s alarm=%s enabled=%v push_enabled=%v start_offset=%s",
		k.Consumers.GroupData, k.Consumers.GroupEvent, k.Consumers.GroupAlarm, k.Enabled, k.PushEnabled, k.Consumers.StartOffset)
}

package config

import (
	"log"
	"os"
	"strings"
)

func mergeKafkaTopicsFromMain(raw mainNacosFile) {
	k := &GlobalConfig.Kafka
	t := raw.Spring.Kafka.Topic
	if v := strings.TrimSpace(t.DataTopic); v != "" {
		k.Topics.DataTopic = v
	}
	if v := strings.TrimSpace(t.EventTopic); v != "" {
		k.Topics.EventTopic = v
	}
	if v := strings.TrimSpace(t.AlarmTopic); v != "" {
		k.Topics.AlarmTopic = v
	}
	if v := strings.TrimSpace(t.ToSaasTopic); v != "" {
		k.Topics.ToSaasTopic = v
	}
	if v := strings.TrimSpace(t.GrayToSaas); v != "" {
		k.Topics.GrayToSaas = v
	}
}

func mergeKafkaConsumersFromMain(raw mainNacosFile) {
	k := &GlobalConfig.Kafka
	c := raw.Spring.Kafka.Consumer
	if c.GroupIDData != "" {
		k.Consumers.GroupData = c.GroupIDData
	}
	if c.GroupIDEvent != "" {
		k.Consumers.GroupEvent = c.GroupIDEvent
	}
	if c.GroupIDAlarm != "" {
		k.Consumers.GroupAlarm = c.GroupIDAlarm
	}
	if c.DataTopicConcurrency > 0 {
		k.Consumers.DataConcurrency = c.DataTopicConcurrency
	}
	if c.EventTopicConcurrency > 0 {
		k.Consumers.EventConcurrency = c.EventTopicConcurrency
	}
	if c.AlarmTopicConcurrency > 0 {
		k.Consumers.AlarmConcurrency = c.AlarmTopicConcurrency
	}
}

func applyKafkaTopicEnvOverrides(c *Config) {
	if c == nil {
		return
	}
	if v := strings.TrimSpace(os.Getenv("KAFKA_TOPICS_DATA_TOPIC")); v != "" {
		c.Kafka.Topics.DataTopic = v
	}
	if v := strings.TrimSpace(os.Getenv("KAFKA_TOPICS_EVENT_TOPIC")); v != "" {
		c.Kafka.Topics.EventTopic = v
	}
	if v := strings.TrimSpace(os.Getenv("KAFKA_TOPICS_ALARM_TOPIC")); v != "" {
		c.Kafka.Topics.AlarmTopic = v
	}
	if v := strings.TrimSpace(os.Getenv("KAFKA_TOPICS_TO_SAAS_TOPIC")); v != "" {
		c.Kafka.Topics.ToSaasTopic = v
	}
}

// resolveKafkaTopics expands ${spring.kafka.topic-prefix} placeholders after kafka.yaml is loaded.
func resolveKafkaTopics(c *Config) {
	if c == nil {
		return
	}
	prefix := strings.TrimSpace(c.Kafka.TopicPrefix)
	props := map[string]string{"spring.kafka.topic-prefix": prefix}
	t := &c.Kafka.Topics
	t.DataTopic = resolveKafkaPlaceholder(t.DataTopic, props)
	t.EventTopic = resolveKafkaPlaceholder(t.EventTopic, props)
	t.AlarmTopic = resolveKafkaPlaceholder(t.AlarmTopic, props)
	t.ToSaasTopic = resolveKafkaPlaceholder(t.ToSaasTopic, props)
	t.GrayToSaas = resolveKafkaPlaceholder(t.GrayToSaas, props)
}

func resolveKafkaPlaceholder(value string, props map[string]string) string {
	value = strings.TrimSpace(value)
	if value == "" || !strings.Contains(value, "${") {
		return value
	}
	out := value
	for key, val := range props {
		out = strings.ReplaceAll(out, "${"+key+"}", val)
	}
	if strings.Contains(out, "${") {
		return value
	}
	return out
}

// LogKafkaTopics prints effective topic names (call after resolveKafkaTopics).
func LogKafkaTopics() {
	if GlobalConfig == nil {
		return
	}
	t := GlobalConfig.Kafka.Topics
	log.Printf("[kafka] topics data=%s event=%s alarm=%s to_saas=%s gray_to_saas=%s prefix=%q",
		t.DataTopic, t.EventTopic, t.AlarmTopic, t.ToSaasTopic, t.GrayToSaas, GlobalConfig.Kafka.TopicPrefix)
}

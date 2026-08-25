package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"
)

const kafkaNacosDataID = "kafka.yaml"

type kafkaNacosFile struct {
	Spring struct {
		Kafka kafkaNacosSettings `yaml:"kafka"`
	} `yaml:"spring"`
}

type kafkaNacosSettings struct {
	BootstrapServers string `yaml:"bootstrap-servers"`
	TopicPrefix      string `yaml:"topic-prefix"`
	Topic            struct {
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
}

func applyKafkaFromNacos(client config_client.IConfigClient) error {
	if client == nil {
		return fmt.Errorf("nacos config client is nil")
	}
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: kafkaNacosDataID,
		Group:  OpsGroup(),
	})
	if err != nil {
		return fmt.Errorf("get %s from nacos group %s: %w", kafkaNacosDataID, OpsGroup(), err)
	}
	if content == "" {
		return fmt.Errorf("%s is empty in nacos group %s", kafkaNacosDataID, OpsGroup())
	}
	return applyKafkaYAML(content)
}

func applyKafkaYAML(content string) error {
	var raw kafkaNacosFile
	if err := yaml.Unmarshal([]byte(content), &raw); err != nil {
		return fmt.Errorf("parse kafka.yaml: %w", err)
	}
	mergeKafkaSettings(raw.Spring.Kafka)
	log.Printf("[nacos] loaded kafka from %s (group=%s, brokers=%v)",
		kafkaNacosDataID, OpsGroup(), GlobalConfig.Kafka.Brokers)
	return nil
}

func mergeKafkaSettings(from kafkaNacosSettings) {
	k := &GlobalConfig.Kafka
	if brokers := parseBootstrapServers(from.BootstrapServers); len(brokers) > 0 {
		k.Brokers = brokers
	}
	if prefix := strings.TrimSpace(from.TopicPrefix); prefix != "" {
		k.TopicPrefix = prefix
	}
	if from.Topic.DataTopic != "" {
		k.Topics.DataTopic = from.Topic.DataTopic
	}
	if from.Topic.EventTopic != "" {
		k.Topics.EventTopic = from.Topic.EventTopic
	}
	if from.Topic.AlarmTopic != "" {
		k.Topics.AlarmTopic = from.Topic.AlarmTopic
	}
	if from.Topic.ToSaasTopic != "" {
		k.Topics.ToSaasTopic = from.Topic.ToSaasTopic
	}
	if from.Topic.GrayToSaas != "" {
		k.Topics.GrayToSaas = from.Topic.GrayToSaas
	}
	if from.Consumer.GroupIDData != "" {
		k.Consumers.GroupData = from.Consumer.GroupIDData
	}
	if from.Consumer.GroupIDEvent != "" {
		k.Consumers.GroupEvent = from.Consumer.GroupIDEvent
	}
	if from.Consumer.GroupIDAlarm != "" {
		k.Consumers.GroupAlarm = from.Consumer.GroupIDAlarm
	}
	if from.Consumer.DataTopicConcurrency > 0 {
		k.Consumers.DataConcurrency = from.Consumer.DataTopicConcurrency
	}
	if from.Consumer.EventTopicConcurrency > 0 {
		k.Consumers.EventConcurrency = from.Consumer.EventTopicConcurrency
	}
	if from.Consumer.AlarmTopicConcurrency > 0 {
		k.Consumers.AlarmConcurrency = from.Consumer.AlarmTopicConcurrency
	}
}

func parseBootstrapServers(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ListenKafkaFromNacos watches kafka.yaml and invokes onChange after config is applied.
func ListenKafkaFromNacos(client config_client.IConfigClient, onChange func()) error {
	if client == nil {
		return fmt.Errorf("nacos config client is nil")
	}
	return client.ListenConfig(vo.ConfigParam{
		DataId: kafkaNacosDataID,
		Group:  OpsGroup(),
		OnChange: func(namespace, group, dataId, data string) {
			log.Printf("[nacos] kafka config changed: dataId=%s group=%s", dataId, group)
			if err := applyKafkaYAML(data); err != nil {
				log.Printf("[ERROR] apply kafka.yaml change: %v", err)
				return
			}
			FinalizeConfig(GlobalConfig)
			if onChange != nil {
				onChange()
			}
		},
	})
}

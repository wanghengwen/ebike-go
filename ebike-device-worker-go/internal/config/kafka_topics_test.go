package config_test

import (
	"testing"

	"ebike-device-worker-go/internal/config"
)

func TestKafkaTopicsFromMainAndPrefix(t *testing.T) {
	config.GlobalConfig = &config.Config{}
	main := `
spring:
  kafka:
    topic:
      data-topic: dataTopic
      event-topic: eventTopic
      alarm-topic: alarmTopic
`
	if err := config.ApplyMainYAMLForTest(main); err != nil {
		t.Fatal(err)
	}
	kafka := `
spring:
  kafka:
    bootstrap-servers: 192.168.2.20:9092
    topic-prefix: ebike
`
	if err := config.ApplyKafkaYAMLForTest(kafka); err != nil {
		t.Fatal(err)
	}
	config.FinalizeConfigForTest(config.GlobalConfig)
	k := config.GlobalConfig.Kafka
	if k.Topics.DataTopic != "dataTopic" || k.Topics.EventTopic != "eventTopic" || k.Topics.AlarmTopic != "alarmTopic" {
		t.Fatalf("topics=%+v", k.Topics)
	}
}

func TestKafkaTopicPrefixPlaceholder(t *testing.T) {
	config.GlobalConfig = &config.Config{}
	main := `
spring:
  kafka:
    topic:
      data-topic: ${spring.kafka.topic-prefix}_device_data
`
	if err := config.ApplyMainYAMLForTest(main); err != nil {
		t.Fatal(err)
	}
	kafka := `
spring:
  kafka:
    topic-prefix: ebike
`
	if err := config.ApplyKafkaYAMLForTest(kafka); err != nil {
		t.Fatal(err)
	}
	config.FinalizeConfigForTest(config.GlobalConfig)
	if config.GlobalConfig.Kafka.Topics.DataTopic != "ebike_device_data" {
		t.Fatalf("data topic=%s", config.GlobalConfig.Kafka.Topics.DataTopic)
	}
}

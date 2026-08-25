package config_test

import (
	"os"
	"testing"

	"ebike-device-worker-go/internal/config"
)

func TestKafkaConsumerGroupEnvOverridesNacos(t *testing.T) {
	t.Setenv("KAFKA_CONSUMERS_GROUP_DATA", "ebike-device-worker-go-verify-data")
	t.Setenv("KAFKA_CONSUMERS_GROUP_EVENT", "ebike-device-worker-go-verify-event")
	t.Setenv("KAFKA_CONSUMERS_GROUP_ALARM", "ebike-device-worker-go-verify-alarm")
	t.Setenv("KAFKA_PUSH_ENABLED", "false")

	config.GlobalConfig = &config.Config{}
	content := `
spring:
  kafka:
    bootstrap-servers: broker:9092
    consumer:
      group-id-data: ebike-device-worker-data
      group-id-event: ebike-device-worker-event
      group-id-alarm: ebike-device-worker-alarm
`
	if err := config.ApplyKafkaYAMLForTest(content); err != nil {
		t.Fatal(err)
	}
	config.FinalizeConfigForTest(config.GlobalConfig)
	k := config.GlobalConfig.Kafka
	if k.Consumers.GroupData != "ebike-device-worker-go-verify-data" {
		t.Fatalf("data group=%s", k.Consumers.GroupData)
	}
	if k.Consumers.GroupEvent != "ebike-device-worker-go-verify-event" {
		t.Fatalf("event group=%s", k.Consumers.GroupEvent)
	}
	if k.Consumers.GroupAlarm != "ebike-device-worker-go-verify-alarm" {
		t.Fatalf("alarm group=%s", k.Consumers.GroupAlarm)
	}
	if k.PushEnabled {
		t.Fatal("push should be disabled")
	}
}

func TestKafkaConsumerGroupFromNacosWithoutEnv(t *testing.T) {
	os.Unsetenv("KAFKA_CONSUMERS_GROUP_DATA")
	config.GlobalConfig = &config.Config{}
	content := `
spring:
  kafka:
    consumer:
      group-id-data: ebike-device-worker-data
`
	if err := config.ApplyKafkaYAMLForTest(content); err != nil {
		t.Fatal(err)
	}
	if config.GlobalConfig.Kafka.Consumers.GroupData != "ebike-device-worker-data" {
		t.Fatalf("group=%s", config.GlobalConfig.Kafka.Consumers.GroupData)
	}
}

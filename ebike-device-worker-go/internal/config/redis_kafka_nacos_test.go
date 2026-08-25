package config_test

import (
	"testing"

	"ebike-device-worker-go/internal/config"
)

func TestApplyRedisYAMLDeviceWorker(t *testing.T) {
	config.GlobalConfig = &config.Config{}
	content := `
redis:
  ebike_device_worker:
    host: 127.0.0.1
    port: 6379
    password: secret
    database: 4
`
	if err := config.ApplyRedisYAMLForTest(content); err != nil {
		t.Fatal(err)
	}
	if config.GlobalConfig.Redis.Addr != "127.0.0.1:6379" {
		t.Fatalf("addr=%s", config.GlobalConfig.Redis.Addr)
	}
	if config.GlobalConfig.Redis.DB != 4 || config.GlobalConfig.Redis.Password != "secret" {
		t.Fatalf("db/password mismatch: db=%d", config.GlobalConfig.Redis.DB)
	}
}

func TestApplyKafkaYAML(t *testing.T) {
	config.GlobalConfig = &config.Config{}
	content := `
spring:
  kafka:
    bootstrap-servers: broker1:9092,broker2:9092
    topic:
      data-topic: prod-data
      event-topic: prod-event
      alarm-topic: prod-alarm
      to-saas-topic: prod-saas
      gray-to-saas-topic: prod-gray-saas
    consumer:
      group-id-data: ebike-device-worker-data
      group-id-event: ebike-device-worker-event
      group-id-alarm: ebike-device-worker-alarm
      data-topic-concurrency: 2
`
	if err := config.ApplyKafkaYAMLForTest(content); err != nil {
		t.Fatal(err)
	}
	k := config.GlobalConfig.Kafka
	if len(k.Brokers) != 2 || k.Brokers[0] != "broker1:9092" {
		t.Fatalf("brokers=%v", k.Brokers)
	}
	if k.Topics.DataTopic != "prod-data" || k.Topics.GrayToSaas != "prod-gray-saas" {
		t.Fatalf("topics mismatch")
	}
	if k.Consumers.GroupData != "ebike-device-worker-data" || k.Consumers.DataConcurrency != 2 {
		t.Fatalf("consumer mismatch")
	}
}

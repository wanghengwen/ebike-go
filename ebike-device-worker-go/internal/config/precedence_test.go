package config_test

import (
	"os"
	"testing"

	"ebike-device-worker-go/internal/config"
)

// Precedence: local < Nacos (main + kafka.yaml) < env.
func TestConfigPrecedenceLocalNacosEnv(t *testing.T) {
	t.Setenv("KAFKA_TOPICS_TO_SAAS_TOPIC", "env_saas_topic")
	t.Setenv("KAFKA_PUSH_ENABLED", "false")

	config.GlobalConfig = &config.Config{}
	config.GlobalConfig.Kafka.Topics.ToSaasTopic = "local_to_saas"
	config.GlobalConfig.Kafka.PushEnabled = true

	main := `
spring:
  kafka:
    topic:
      to-saas-topic: saas_0
`
	if err := config.ApplyMainYAMLForTest(main); err != nil {
		t.Fatal(err)
	}
	if config.GlobalConfig.Kafka.Topics.ToSaasTopic != "saas_0" {
		t.Fatalf("after main nacos to_saas=%s want saas_0", config.GlobalConfig.Kafka.Topics.ToSaasTopic)
	}

	kafka := `
spring:
  kafka:
    topic:
      to-saas-topic: kafka_yaml_saas
`
	if err := config.ApplyKafkaYAMLForTest(kafka); err != nil {
		t.Fatal(err)
	}
	if config.GlobalConfig.Kafka.Topics.ToSaasTopic != "kafka_yaml_saas" {
		t.Fatalf("after kafka.yaml to_saas=%s want kafka_yaml_saas", config.GlobalConfig.Kafka.Topics.ToSaasTopic)
	}

	config.FinalizeConfigForTest(config.GlobalConfig)
	if config.GlobalConfig.Kafka.Topics.ToSaasTopic != "env_saas_topic" {
		t.Fatalf("after env to_saas=%s want env_saas_topic", config.GlobalConfig.Kafka.Topics.ToSaasTopic)
	}
	if config.GlobalConfig.Kafka.PushEnabled {
		t.Fatal("push should be disabled by env")
	}
}

func TestConfigPrecedenceNacosOverridesLocalWithoutEnv(t *testing.T) {
	os.Unsetenv("KAFKA_TOPICS_TO_SAAS_TOPIC")
	os.Unsetenv("KAFKA_CONSUMERS_GROUP_DATA")

	config.GlobalConfig = &config.Config{}
	config.GlobalConfig.Kafka.Topics.ToSaasTopic = "local_to_saas"
	config.GlobalConfig.Kafka.Consumers.GroupData = "local_group"

	main := `
spring:
  kafka:
    topic:
      to-saas-topic: saas_0
    consumer:
      group-id-data: anvelink_worker_data_group
`
	if err := config.ApplyMainYAMLForTest(main); err != nil {
		t.Fatal(err)
	}
	config.FinalizeConfigForTest(config.GlobalConfig)

	if config.GlobalConfig.Kafka.Topics.ToSaasTopic != "saas_0" {
		t.Fatalf("to_saas=%s want saas_0", config.GlobalConfig.Kafka.Topics.ToSaasTopic)
	}
	if config.GlobalConfig.Kafka.Consumers.GroupData != "anvelink_worker_data_group" {
		t.Fatalf("group=%s", config.GlobalConfig.Kafka.Consumers.GroupData)
	}
}

func TestMainNacosPersistAndCaffeineMerge(t *testing.T) {
	config.GlobalConfig = &config.Config{}
	main := `
persist-config:
  persist-limit-size: 100
  persist-batch-size: 50
  persist-interval-mill: 1000
caffeine:
  device-tenant-mapping-cache-size: 100000
  tenant-cache-size: 1000
  timeout: 100
`
	if err := config.ApplyMainYAMLForTest(main); err != nil {
		t.Fatal(err)
	}
	pc := config.GlobalConfig.PersistConfig
	if pc.PersistLimitSize != 100 || pc.PersistBatchSize != 50 || pc.PersistIntervalMs != 1000 {
		t.Fatalf("persist=%+v", pc)
	}
	cf := config.GlobalConfig.Caffeine
	if cf.DeviceTenantMappingCacheSize != 100000 || cf.TenantCacheSize != 1000 || cf.TimeoutSec != 100 {
		t.Fatalf("caffeine=%+v", cf)
	}
}

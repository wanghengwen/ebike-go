package config

import "testing"

func withKafka(t *testing.T, k KafkaConfig) {
	t.Helper()
	prev := GlobalConfig().Kafka
	t.Cleanup(func() { mutate(func(c *Config) { c.Kafka = prev }) })
	mutate(func(c *Config) { c.Kafka = k })
}

func validKafka() KafkaConfig {
	k := defaultKafkaConfig()
	k.Brokers = []string{"kafka-0:9092"}
	return k
}

// TestValidateKafkaRejectsForeignGroupID is the important one: sharing a consumer
// group with ebike-device-worker or ebike-device-consume makes Kafka split the
// topic's partitions between us and them, silently starving the other service.
func TestValidateKafkaRejectsForeignGroupID(t *testing.T) {
	for _, group := range []string{
		"ebike-device-worker-saas",
		"ebike-device-consume",
		"saas_0_group",
		"",
	} {
		k := validKafka()
		k.GroupID = group
		withKafka(t, k)
		if err := ValidateKafka(); err == nil {
			t.Errorf("groupId %q was accepted, want rejection", group)
		}
	}
}

func TestValidateKafkaAcceptsOwnGroupID(t *testing.T) {
	withKafka(t, validKafka())
	if err := ValidateKafka(); err != nil {
		t.Errorf("default config rejected: %v", err)
	}
}

func TestValidateKafkaRejectsUnresolvedTopicPlaceholder(t *testing.T) {
	k := validKafka()
	k.Topics = []string{"${spring.kafka.topic-prefix}_saas_0"}
	withKafka(t, k)
	// Subscribing to a literal "${...}" name would look healthy while receiving
	// nothing, so this must fail loudly instead.
	if err := ValidateKafka(); err == nil {
		t.Error("unresolved placeholder was accepted, want rejection")
	}
}

func TestValidateKafkaRequiresBrokersAndTopics(t *testing.T) {
	k := validKafka()
	k.Brokers = nil
	withKafka(t, k)
	if err := ValidateKafka(); err == nil {
		t.Error("empty brokers accepted, want rejection")
	}

	k = validKafka()
	k.Topics = nil
	withKafka(t, k)
	if err := ValidateKafka(); err == nil {
		t.Error("empty topics accepted, want rejection")
	}
}

// TestMergeNacosKafkaConfigSubscribesGrayTopic covers the gray-tenant split:
// ebike-device-worker routes a gray tenant's events to gray-to-saas-topic, so
// missing it loses every event for those tenants.
func TestMergeNacosKafkaConfigSubscribesGrayTopic(t *testing.T) {
	withKafka(t, defaultKafkaConfig())
	MergeNacosKafkaConfig(`
spring:
  kafka:
    bootstrap-servers: kafka-0:9092,kafka-1:9092
    topic:
      to-saas-topic: saas_0
      gray-to-saas-topic: gray_saas_0
`)
	k := GlobalConfig().Kafka
	if len(k.Brokers) != 2 {
		t.Errorf("Brokers = %v, want two entries", k.Brokers)
	}
	if len(k.Topics) != 2 || k.Topics[0] != "saas_0" || k.Topics[1] != "gray_saas_0" {
		t.Errorf("Topics = %v, want [saas_0 gray_saas_0]", k.Topics)
	}
}

// TestMergeNacosKafkaConfigIgnoresForeignGroupIDs proves kafka.yaml cannot inject
// a group id: the file's group-id-* keys belong to worker and consume.
func TestMergeNacosKafkaConfigIgnoresForeignGroupIDs(t *testing.T) {
	withKafka(t, defaultKafkaConfig())
	want := GlobalConfig().Kafka.GroupID
	MergeNacosKafkaConfig(`
spring:
  kafka:
    bootstrap-servers: kafka-0:9092
    topic:
      to-saas-topic: saas_0
    consumer:
      group-id-data: ebike-device-worker-data
`)
	if got := GlobalConfig().Kafka.GroupID; got != want {
		t.Errorf("GroupID = %q, want it left at %q", got, want)
	}
}

func TestMergeNacosKafkaConfigExpandsTopicPrefix(t *testing.T) {
	withKafka(t, defaultKafkaConfig())
	MergeNacosKafkaConfig(`
spring:
  kafka:
    bootstrap-servers: kafka-0:9092
    topic-prefix: prod
    topic:
      to-saas-topic: ${spring.kafka.topic-prefix}_saas_0
`)
	if got := GlobalConfig().Kafka.Topics; len(got) != 1 || got[0] != "prod_saas_0" {
		t.Errorf("Topics = %v, want [prod_saas_0]", got)
	}
}

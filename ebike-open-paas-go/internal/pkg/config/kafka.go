package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// KafkaNacosDataID is the shared ops config that carries brokers and topic names.
const KafkaNacosDataID = "kafka.yaml"

// groupIDGuard is the substring every consumer group must contain. Sharing a
// group with ebike-device-worker or ebike-device-consume would make Kafka split
// saas_0's partitions between us and them, silently starving the other service
// of the events it needs to maintain the device shadow. The check is cheap
// insurance against a copy-pasted group id in Nacos.
const groupIDGuard = "open-paas"

// KafkaConfig configures the saas_0 consumer that feeds the callback dispatcher.
//
// Topic and brokers are read from the shared kafka.yaml so we automatically
// follow ebike-device-worker's producer target. The consumer group is
// deliberately NOT read from kafka.yaml: that file's group-id-* keys belong to
// worker and consume, and inheriting one would steal their partitions.
type KafkaConfig struct {
	// Enabled gates the consumer entirely; callbacks are simply not delivered
	// when false, which is the safe default for a fresh deployment.
	Enabled bool `yaml:"enabled"`
	// Brokers is the bootstrap server list (kafka.yaml spring.kafka.bootstrap-servers).
	Brokers []string `yaml:"brokers"`
	// Topics are the topics ebike-device-worker pushes flattened device events
	// to. Both to-saas-topic and gray-to-saas-topic must be subscribed: worker
	// routes a gray tenant's events to the gray topic instead of the main one
	// (push.Service.push), so subscribing to only one silently loses every
	// event for whichever set of tenants is on the other.
	Topics []string `yaml:"topics"`
	// TopicPrefix expands a "${spring.kafka.topic-prefix}" placeholder in Topics.
	TopicPrefix string `yaml:"topicPrefix"`
	// GroupID must be unique to this service; see groupIDGuard.
	GroupID string `yaml:"groupId"`
	// Concurrency is the number of consumer-group members this process joins
	// with. Keep it at or below the topic's partition count.
	Concurrency int `yaml:"concurrency"`
	// BatchSize bounds how many records one claim drains before committing,
	// mirroring the Java max-poll-records semantics.
	BatchSize int `yaml:"batchSize"`
	// BatchWaitMs is how long a partial batch waits before being processed.
	BatchWaitMs int `yaml:"batchWaitMs"`
	// StartOffset applies only when the group has no committed offset:
	// "earliest" replays the retention window, anything else starts at the tail.
	// Default "latest" — a fresh subscriber must not flood third parties with
	// historical events.
	StartOffset string `yaml:"startOffset"`
}

func defaultKafkaConfig() KafkaConfig {
	return KafkaConfig{
		Enabled:     false,
		Topics:      []string{"saas_0"},
		GroupID:     "ebike-open-paas-go-saas",
		Concurrency: 1,
		BatchSize:   500,
		BatchWaitMs: 100,
		StartOffset: "latest",
	}
}

// kafkaNacosFile maps the subset of kafka.yaml this service consumes. The
// consumer.group-id-* keys are intentionally absent, see KafkaConfig.
type kafkaNacosFile struct {
	Spring struct {
		Kafka struct {
			BootstrapServers string `yaml:"bootstrap-servers"`
			TopicPrefix      string `yaml:"topic-prefix"`
			Topic            struct {
				ToSaasTopic string `yaml:"to-saas-topic"`
				GrayToSaas  string `yaml:"gray-to-saas-topic"`
			} `yaml:"topic"`
		} `yaml:"kafka"`
	} `yaml:"spring"`
}

// MergeNacosKafkaConfig merges brokers and the to-saas topic from kafka.yaml.
func MergeNacosKafkaConfig(yamlContent string) {
	if strings.TrimSpace(yamlContent) == "" {
		return
	}
	var raw kafkaNacosFile
	if err := yaml.Unmarshal([]byte(yamlContent), &raw); err != nil {
		log.Printf("[config] failed to parse nacos %s: %v", KafkaNacosDataID, err)
		return
	}
	mutate(func(c *Config) {
		k := &c.Kafka
		if brokers := splitCSV(raw.Spring.Kafka.BootstrapServers); len(brokers) > 0 {
			k.Brokers = brokers
		}
		if prefix := strings.TrimSpace(raw.Spring.Kafka.TopicPrefix); prefix != "" {
			k.TopicPrefix = prefix
		}
		if topics := dedupeNonEmpty(raw.Spring.Kafka.Topic.ToSaasTopic, raw.Spring.Kafka.Topic.GrayToSaas); len(topics) > 0 {
			k.Topics = topics
		}
		resolveKafkaTopics(k)
	})
	k := GlobalConfig().Kafka
	log.Printf("[config] merged nacos %s: brokers=%v topics=%v", KafkaNacosDataID, k.Brokers, k.Topics)
}

// resolveKafkaTopics expands the "${spring.kafka.topic-prefix}" placeholder that
// kafka.yaml uses, leaving the raw value untouched when the prefix is unset so
// ValidateKafka can report it instead of silently subscribing to a bogus name.
//
// The expansion writes a fresh slice: the one it receives is still shared with
// the config snapshot that readers are holding.
func resolveKafkaTopics(k *KafkaConfig) {
	if k.TopicPrefix == "" || len(k.Topics) == 0 {
		return
	}
	out := make([]string, len(k.Topics))
	for i, t := range k.Topics {
		out[i] = strings.ReplaceAll(t, "${spring.kafka.topic-prefix}", k.TopicPrefix)
	}
	k.Topics = out
}

func dedupeNonEmpty(values ...string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, dup := seen[v]; dup {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// ValidateKafka reports why the consumer cannot start, naming the config key to
// fix. Callers should treat a non-nil error as "leave the consumer off" rather
// than aborting the process: the HTTP API stays useful without callbacks.
func ValidateKafka() error {
	k := GlobalConfig().Kafka
	if len(k.Brokers) == 0 {
		return fmt.Errorf("no brokers (kafka.yaml spring.kafka.bootstrap-servers or KAFKA_BROKERS)")
	}
	if len(k.Topics) == 0 {
		return fmt.Errorf("no topics (kafka.yaml spring.kafka.topic.to-saas-topic or KAFKA_TOPICS)")
	}
	for _, t := range k.Topics {
		if strings.TrimSpace(t) == "" {
			return fmt.Errorf("empty topic in %v", k.Topics)
		}
		if strings.Contains(t, "${") {
			return fmt.Errorf("unresolved placeholder in topic %q (spring.kafka.topic-prefix is unset)", t)
		}
	}
	if strings.TrimSpace(k.GroupID) == "" {
		return fmt.Errorf("empty groupId (kafka.groupId or KAFKA_GROUP_ID)")
	}
	if !strings.Contains(k.GroupID, groupIDGuard) {
		return fmt.Errorf("groupId %q must contain %q so it cannot collide with the ebike-device-worker/consume groups on %v",
			k.GroupID, groupIDGuard, k.Topics)
	}
	return nil
}

// LogEffectiveKafka prints the resolved consumer settings after all layers merge.
func LogEffectiveKafka() {
	k := GlobalConfig().Kafka
	log.Printf("[config] effective kafka enabled=%v brokers=%v topics=%v group=%q concurrency=%d batch=%d/%dms startOffset=%s",
		k.Enabled, k.Brokers, k.Topics, k.GroupID, k.Concurrency, k.BatchSize, k.BatchWaitMs, k.StartOffset)
}

// applyKafkaEnvOverrides is called from ApplyEnvOverrides, already inside a
// config mutation, so it edits the clone it is handed.
func applyKafkaEnvOverrides(c *Config) {
	k := &c.Kafka
	if v := os.Getenv("KAFKA_ENABLED"); v != "" {
		k.Enabled = strings.EqualFold(v, "true") || v == "1"
		log.Printf("[config] env override: KAFKA_ENABLED=%v", k.Enabled)
	}
	if v := splitCSV(os.Getenv("KAFKA_BROKERS")); len(v) > 0 {
		k.Brokers = v
		log.Printf("[config] env override: KAFKA_BROKERS=%v", v)
	}
	if v := splitCSV(os.Getenv("KAFKA_TOPICS")); len(v) > 0 {
		k.Topics = v
		log.Printf("[config] env override: KAFKA_TOPICS=%v", v)
	}
	if v := strings.TrimSpace(os.Getenv("KAFKA_GROUP_ID")); v != "" {
		k.GroupID = v
		log.Printf("[config] env override: KAFKA_GROUP_ID=%s", v)
	}
	if v := os.Getenv("KAFKA_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			k.Concurrency = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("KAFKA_START_OFFSET")); v != "" {
		k.StartOffset = v
	}
	resolveKafkaTopics(k)
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

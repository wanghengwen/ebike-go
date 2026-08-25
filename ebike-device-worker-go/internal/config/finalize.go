package config

import "log"

// FinalizeConfig resolves placeholders and applies environment overrides after all
// config layers are merged. Precedence (low → high): local file < Nacos < env.
func FinalizeConfig(c *Config) {
	if c == nil {
		return
	}
	resolvePlaceholders(c)
	resolveKafkaTopics(c)
	normalizeConfig(c)
	applyEnvOverrides(c)
	LogKafkaTopics()
	LogKafkaConsumerGroups()
}

// LogEffectiveConfigSource logs a one-line summary of kafka/redis push settings after finalize.
func LogEffectiveConfigSource() {
	if GlobalConfig == nil {
		return
	}
	log.Printf("[config] effective kafka to_saas=%s push_enabled=%v redis_db=%d table_suffix=%q (precedence: local < nacos < env)",
		GlobalConfig.Kafka.Topics.ToSaasTopic,
		GlobalConfig.Kafka.PushEnabled,
		GlobalConfig.Redis.DB,
		GlobalConfig.PersistConfig.TableSuffix,
	)
}

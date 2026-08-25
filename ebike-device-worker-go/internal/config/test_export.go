package config

// ParseFastIDNacosYAMLForTest exposes fast_id.yaml parsing for unit tests.
func ParseFastIDNacosYAMLForTest(content string) (FastIDSettings, error) {
	return parseFastIDNacosYAML(content)
}

// MergeFastIDSettingsForTest exposes fastid merge for unit tests.
func MergeFastIDSettingsForTest(dst *FastIDSettings, from FastIDSettings) {
	mergeFastIDSettings(dst, from)
}

// ApplyRedisYAMLForTest exposes redis.yaml parsing for unit tests.
func ApplyRedisYAMLForTest(content string) error {
	return applyRedisYAML(content)
}

// ApplyKafkaYAMLForTest exposes kafka.yaml parsing for unit tests.
func ApplyKafkaYAMLForTest(content string) error {
	return applyKafkaYAML(content)
}

// ApplyMainYAMLForTest exposes main application config parsing for unit tests.
func ApplyMainYAMLForTest(content string) error {
	return applyMainYAML(content)
}

// BuildPostgresDSNForTest exposes jdbc->postgres dsn conversion for unit tests.
func BuildPostgresDSNForTest(jdbcURL, username, password string) (string, error) {
	return buildPostgresDSN(jdbcURL, username, password)
}

// ResolveAppNameForTest exposes app-name placeholder resolution for unit tests.
func ResolveAppNameForTest(fastidAppName, springAppName string) string {
	return resolveAppName(fastidAppName, springAppName)
}

// ApplyEnvOverridesForTest exposes env merge for unit tests.
func ApplyEnvOverridesForTest(c *Config) {
	applyEnvOverrides(c)
}

// FinalizeConfigForTest runs placeholder resolution + env overrides (same as production).
func FinalizeConfigForTest(c *Config) {
	FinalizeConfig(c)
}

// ApplyDatabaseEnvOverridesForTest exposes database env merge for unit tests.
func ApplyDatabaseEnvOverridesForTest(c *Config) {
	applyDatabaseEnvOverrides(c)
}

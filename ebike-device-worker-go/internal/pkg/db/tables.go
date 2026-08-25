package db

import "ebike-device-worker-go/internal/config"

// PgTable returns the PostgreSQL table name for an ebike persist target,
// appending persist_config.table_suffix when set (e.g. "_new" for TimescaleDB hypertables).
func PgTable(base string) string {
	if base == "" {
		return base
	}
	return base + TableSuffix()
}

// TableSuffix returns the configured PG table suffix (empty for legacy table names).
func TableSuffix() string {
	if config.GlobalConfig == nil {
		return ""
	}
	return config.GlobalConfig.PersistConfig.TableSuffix
}

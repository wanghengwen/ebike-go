package config_test

import (
	"os"
	"strings"
	"testing"

	"ebike-device-worker-go/internal/config"
)

func TestReplacePostgresDatabaseName(t *testing.T) {
	dsn := "postgres://ebike_user:secret@192.168.2.24:5432/anvelink?sslmode=disable&search_path=public"
	got, err := config.ReplacePostgresDatabaseNameForTest(dsn, "anvelink_shadow")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "/anvelink_shadow") {
		t.Fatalf("db name not replaced: %s", got)
	}
	if !strings.Contains(got, "192.168.2.24") || !strings.Contains(got, "ebike_user") {
		t.Fatalf("host/user changed: %s", got)
	}
}

func TestDatabaseNameEnvOverridesAfterNacos(t *testing.T) {
	t.Setenv("DATABASE_NAME", "anvelink_shadow")
	config.GlobalConfig = &config.Config{}
	content := `
spring:
  datasource:
    url: jdbc:postgresql://192.168.2.24:5432/anvelink
    username: ebike_user
    password: secret
`
	if err := config.ApplyMainYAMLForTest(content); err != nil {
		t.Fatal(err)
	}
	config.ApplyDatabaseEnvOverridesForTest(config.GlobalConfig)
	if config.GlobalConfig.Database.Dsn == "" {
		t.Fatal("dsn empty")
	}
	if !strings.Contains(config.GlobalConfig.Database.Dsn, "/anvelink_shadow") {
		t.Fatalf("dsn=%s", config.GlobalConfig.Database.Dsn)
	}
}

func TestDatabaseDSNEnvFullOverride(t *testing.T) {
	t.Setenv("DATABASE_DSN", "postgres://u:p@host:5432/custom")
	t.Setenv("DATABASE_NAME", "ignored")
	config.GlobalConfig = &config.Config{}
	config.GlobalConfig.Database.Dsn = "postgres://old/old"
	config.ApplyDatabaseEnvOverridesForTest(config.GlobalConfig)
	if config.GlobalConfig.Database.Dsn != "postgres://u:p@host:5432/custom" {
		t.Fatalf("dsn=%s", config.GlobalConfig.Database.Dsn)
	}
	os.Unsetenv("DATABASE_DSN")
}

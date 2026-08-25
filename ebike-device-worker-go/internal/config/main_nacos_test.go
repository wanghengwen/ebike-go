package config_test

import (
	"strings"
	"testing"

	"ebike-device-worker-go/internal/config"
)

func TestBuildPostgresDSNFromJDBC(t *testing.T) {
	dsn, err := config.BuildPostgresDSNForTest(
		"jdbc:postgresql://192.168.2.24:5432/ebike_device?currentSchema=public",
		"ebike_user",
		"s3cret",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(dsn, "192.168.2.24:5432") {
		t.Fatalf("host missing: %s", dsn)
	}
	if !strings.Contains(dsn, "/ebike_device") {
		t.Fatalf("db missing: %s", dsn)
	}
	if !strings.Contains(dsn, "ebike_user") || !strings.Contains(dsn, "sslmode=disable") {
		t.Fatalf("unexpected dsn: %s", dsn)
	}
	if !strings.Contains(dsn, "search_path=public") {
		t.Fatalf("expected search_path from currentSchema: %s", dsn)
	}
}

func TestBuildPostgresDSNStripsJavaJDBCParams(t *testing.T) {
	dsn, err := config.BuildPostgresDSNForTest(
		"jdbc:postgresql://192.168.2.25:5432/anvelink?useUnicode=true&characterEncoding=UTF-8&serverTimezone=GMT%2B8&zeroDateTimeBehavior=convertToNull",
		"luoping",
		"secret",
	)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(dsn, "useUnicode") || strings.Contains(dsn, "characterEncoding") || strings.Contains(dsn, "serverTimezone") {
		t.Fatalf("jdbc-only params should be stripped: %s", dsn)
	}
	if !strings.Contains(dsn, "sslmode=disable") || !strings.Contains(dsn, "/anvelink") {
		t.Fatalf("unexpected dsn: %s", dsn)
	}
}

func TestBuildPostgresDSNStripsStringtype(t *testing.T) {
	dsn, err := config.BuildPostgresDSNForTest(
		"jdbc:postgresql://192.168.2.25:5432/anvelink?stringtype=unspecified&prepareThreshold=0",
		"luoping",
		"secret",
	)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(dsn, "stringtype") || strings.Contains(dsn, "prepareThreshold") {
		t.Fatalf("jdbc driver params should be stripped: %s", dsn)
	}
}

func TestApplyMainYAMLDatabaseAndAppName(t *testing.T) {
	config.GlobalConfig = &config.Config{}
	content := `
spring:
  application:
    name: ebike-device-worker
  datasource:
    url: jdbc:postgresql://192.168.2.24:5432/ebike_device
    username: ebike_user
    password: s3cret
`
	if err := config.ApplyMainYAMLForTest(content); err != nil {
		t.Fatal(err)
	}
	if config.GlobalConfig.Spring.Application.Name != "ebike-device-worker" {
		t.Fatalf("app=%s", config.GlobalConfig.Spring.Application.Name)
	}
	if config.GlobalConfig.Database.Dsn == "" {
		t.Fatal("database dsn should be populated")
	}
}

func TestApplyMainYAMLTableSuffix(t *testing.T) {
	config.GlobalConfig = &config.Config{}
	content := `
persist-config:
  table-suffix: "_new"
`
	if err := config.ApplyMainYAMLForTest(content); err != nil {
		t.Fatal(err)
	}
	if config.GlobalConfig.PersistConfig.TableSuffix != "_new" {
		t.Fatalf("suffix=%q", config.GlobalConfig.PersistConfig.TableSuffix)
	}
}

func TestTableSuffixEnvOverride(t *testing.T) {
	t.Setenv("TABLE_SUFFIX", "_new")
	cfg := &config.Config{}
	config.ApplyEnvOverridesForTest(cfg)
	if cfg.PersistConfig.TableSuffix != "_new" {
		t.Fatalf("suffix=%q", cfg.PersistConfig.TableSuffix)
	}
}

func TestResolveAppNamePlaceholder(t *testing.T) {
	got := config.ResolveAppNameForTest("${spring.application.name}", "ebike-device-worker")
	if got != "ebike-device-worker" {
		t.Fatalf("app=%s", got)
	}
}

func TestResolvedFastIDIgnoresPlaceholderAppName(t *testing.T) {
	cfg := &config.Config{}
	cfg.Spring.Application.Name = "ebike-device-worker"
	cfg.Spring.Xyy.Fastid = config.FastIDSettings{
		Enabled:    true,
		ServerAddr: "fastid.luopingtech.com",
		Namespace:  "prod",
		GroupID:    "xyy",
		Secret:     "test-secret",
		AppName:    "${spring.application.name}",
	}
	cfg.Server.Port = "8080"

	fc := cfg.ResolvedFastID()
	if fc.AppName != "ebike-device-worker" {
		t.Fatalf("app=%s", fc.AppName)
	}
}

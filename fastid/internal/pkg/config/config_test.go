package config

import (
	"os"
	"testing"
)

func TestResolvePlaceholders(t *testing.T) {
	cfg := &AppConfig{}
	cfg.Database.Host = "10.0.0.1"
	cfg.Database.Port = 3306
	cfg.Database.Username = "myuser"
	cfg.Database.Password = "mypass"

	// Test case 1: Known properties
	input := "jdbc:mysql://${mysql.fastid.host}:${mysql.fastid.port}/db?user=${mysql.fastid.username}&pass=${mysql.fastid.password}"
	expected := "jdbc:mysql://10.0.0.1:3306/db?user=myuser&pass=mypass"
	got := resolvePlaceholders(input, cfg, "")
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}

	// Test case 2: Environment variables
	os.Setenv("MY_ENV_VAR", "env_val")
	defer os.Unsetenv("MY_ENV_VAR")
	input = "val-${MY_ENV_VAR}"
	expected = "val-env_val"
	got = resolvePlaceholders(input, cfg, "")
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}

	// Test case 3: Default values
	input = "${NON_EXISTENT_VAR:default_val}"
	expected = "default_val"
	got = resolvePlaceholders(input, cfg, "")
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}

	// Test case 4: Combined and unresolved
	input = "${mysql.fastid.host}-${UNRESOLVED_VAR}"
	expected = "10.0.0.1-${UNRESOLVED_VAR}"
	got = resolvePlaceholders(input, cfg, "")
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestParseJDBCURL(t *testing.T) {
	cfg := &AppConfig{}

	// Test standard JDBC URL with params
	url := "jdbc:mysql://localhost:3306/fastid?useSSL=false&characterEncoding=UTF-8"
	parseJDBCURL(url, cfg)

	if cfg.Database.Host != "localhost" {
		t.Errorf("expected host localhost, got %q", cfg.Database.Host)
	}
	if cfg.Database.Port != 3306 {
		t.Errorf("expected port 3306, got %d", cfg.Database.Port)
	}
	if cfg.Database.Name != "fastid" {
		t.Errorf("expected database name fastid, got %q", cfg.Database.Name)
	}
	if !testing.Short() {
		// Params should have charset=utf8mb4 prepended
		expectedParams := "charset=utf8mb4&useSSL=false&characterEncoding=UTF-8&parseTime=True&loc=Local"
		if cfg.Database.Params != expectedParams {
			t.Errorf("expected params %q, got %q", expectedParams, cfg.Database.Params)
		}
	}
}

func TestOverrideWithEnv(t *testing.T) {
	cfg := &AppConfig{}

	os.Setenv("SPRING_NACOS_SERVER_ADDR", "nacos.local:8848")
	os.Setenv("SPRING_NACOS_NAMESPACE", "ns-prod")
	os.Setenv("SPRING_NACOS_GROUP", "prod-group")
	os.Setenv("SPRING_APPLICATION_NAME", "my-app")
	os.Setenv("SERVER_PORT", "9090")
	defer func() {
		os.Unsetenv("SPRING_NACOS_SERVER_ADDR")
		os.Unsetenv("SPRING_NACOS_NAMESPACE")
		os.Unsetenv("SPRING_NACOS_GROUP")
		os.Unsetenv("SPRING_APPLICATION_NAME")
		os.Unsetenv("SERVER_PORT")
	}()

	overrideWithEnv(cfg)

	if cfg.Nacos.ServerAddr != "nacos.local:8848" {
		t.Errorf("expected server-addr nacos.local:8848, got %q", cfg.Nacos.ServerAddr)
	}
	if cfg.Nacos.Namespace != "ns-prod" {
		t.Errorf("expected namespace ns-prod, got %q", cfg.Nacos.Namespace)
	}
	if cfg.Nacos.Group != "prod-group" {
		t.Errorf("expected group prod-group, got %q", cfg.Nacos.Group)
	}
	if cfg.Nacos.AppName != "my-app" {
		t.Errorf("expected app-name my-app, got %q", cfg.Nacos.AppName)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if !cfg.Nacos.Config.Enabled || !cfg.Nacos.Discovery.Enabled {
		t.Errorf("expected Nacos config and discovery to be auto-enabled")
	}
}

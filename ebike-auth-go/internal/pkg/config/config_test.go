package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	yamlContent := `
server:
  port: 8081
xyy:
  auth:
    mode: business
  nacos:
    serverAddr: "192.168.1.100:8848"
    namespace: "test-ns"
    group: "test-group"
    registerEnabled: true
`
	tmpFile, err := os.CreateTemp("", "application_*.yml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Clear env vars to make sure they don't interfere
	os.Unsetenv("AUTH_SERVICE_MODE")
	os.Unsetenv("PORT")
	os.Unsetenv("NACOS_SERVER_ADDR")

	err = LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if AppConfig.Port != 8081 {
		t.Errorf("expected port 8081, got %d", AppConfig.Port)
	}
	if AppConfig.AuthMode != "business" {
		t.Errorf("expected AuthMode 'business', got %q", AppConfig.AuthMode)
	}
	if AppConfig.Nacos.ServerAddr != "192.168.1.100:8848" {
		t.Errorf("expected Nacos ServerAddr '192.168.1.100:8848', got %q", AppConfig.Nacos.ServerAddr)
	}
	if AppConfig.Nacos.Namespace != "test-ns" {
		t.Errorf("expected Nacos Namespace 'test-ns', got %q", AppConfig.Nacos.Namespace)
	}
	if AppConfig.Nacos.Group != "test-group" {
		t.Errorf("expected Nacos Group 'test-group', got %q", AppConfig.Nacos.Group)
	}
	if !AppConfig.Nacos.RegisterEnabled {
		t.Errorf("expected Nacos RegisterEnabled true, got false")
	}
}

func TestOverrideWithEnv(t *testing.T) {
	os.Setenv("AUTH_SERVICE_MODE", "client")
	os.Setenv("PORT", "9090")
	os.Setenv("NACOS_SERVER_ADDR", "localhost:8848")
	os.Setenv("NACOS_NAMESPACE", "env-ns")
	os.Setenv("NACOS_GROUP", "env-group")
	os.Setenv("NACOS_REGISTER_ENABLED", "false")

	defer func() {
		os.Unsetenv("AUTH_SERVICE_MODE")
		os.Unsetenv("PORT")
		os.Unsetenv("NACOS_SERVER_ADDR")
		os.Unsetenv("NACOS_NAMESPACE")
		os.Unsetenv("NACOS_GROUP")
		os.Unsetenv("NACOS_REGISTER_ENABLED")
	}()

	overrideWithEnv()

	if AppConfig.AuthMode != "client" {
		t.Errorf("expected AuthMode 'client', got %q", AppConfig.AuthMode)
	}
	if AppConfig.Port != 9090 {
		t.Errorf("expected Port 9090, got %d", AppConfig.Port)
	}
	if AppConfig.Nacos.ServerAddr != "localhost:8848" {
		t.Errorf("expected Nacos ServerAddr 'localhost:8848', got %q", AppConfig.Nacos.ServerAddr)
	}
	if AppConfig.Nacos.Namespace != "env-ns" {
		t.Errorf("expected Nacos Namespace 'env-ns', got %q", AppConfig.Nacos.Namespace)
	}
	if AppConfig.Nacos.Group != "env-group" {
		t.Errorf("expected Nacos Group 'env-group', got %q", AppConfig.Nacos.Group)
	}
	if AppConfig.Nacos.RegisterEnabled {
		t.Errorf("expected Nacos RegisterEnabled false, got true")
	}
}

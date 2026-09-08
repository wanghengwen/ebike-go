package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"ebike-device-worker-go/internal/config"
)

func TestParseFastIDNacosYAML(t *testing.T) {
	t.Parallel()
	content := `
spring:
  xyy:
    fastid:
      enabled: true
      server-addr: fastid.luopingtech.com
      server-url: http://xyy-fastid:8080
      namespace: prod
      groupId: xyy
      secret: nacos-secret
      app-name: ebike-device-worker
`
	fastid, err := config.ParseFastIDNacosYAMLForTest(content)
	if err != nil {
		t.Fatal(err)
	}
	if fastid.Secret != "nacos-secret" || fastid.ServerAddr != "fastid.luopingtech.com" {
		t.Fatalf("unexpected fastid: %+v", fastid)
	}
	if fastid.ServerURL != "http://xyy-fastid:8080" {
		t.Fatalf("server-url=%s", fastid.ServerURL)
	}
}

func TestMergeFastIDSettings(t *testing.T) {
	t.Parallel()
	dst := config.FastIDSettings{
		ServerAddr: "local-host",
		Namespace:  "local-ns",
	}
	from := config.FastIDSettings{
		Secret:     "from-nacos",
		GroupID:    "xyy",
		ServerAddr: "fastid.luopingtech.com",
		ServerURL:  "http://xyy-fastid:8080",
	}
	config.MergeFastIDSettingsForTest(&dst, from)
	if dst.Secret != "from-nacos" || dst.GroupID != "xyy" {
		t.Fatalf("merge failed: %+v", dst)
	}
	if dst.ServerAddr != "fastid.luopingtech.com" {
		t.Fatalf("addr=%s", dst.ServerAddr)
	}
	if dst.ServerURL != "xyy-fastid:8080" {
		t.Fatalf("server-url=%s", dst.ServerURL)
	}
	if dst.UseHTTPS {
		t.Fatal("expected UseHTTPS=false from http://server-url")
	}
}

func TestLoadConfigFromCustomPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "application.yaml")
	if err := os.WriteFile(path, []byte(`
server:
  port: "9099"
spring:
  application:
    name: test-worker
`), 0o644); err != nil {
		t.Fatal(err)
	}
	config.LoadConfig(path)
	if config.GlobalConfig.Server.Port != "9099" {
		t.Fatalf("port=%s", config.GlobalConfig.Server.Port)
	}
	if config.GlobalConfig.Spring.Application.Name != "test-worker" {
		t.Fatalf("app=%s", config.GlobalConfig.Spring.Application.Name)
	}
}

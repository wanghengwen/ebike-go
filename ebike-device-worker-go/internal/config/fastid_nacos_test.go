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
}

func TestMergeFastIDSettings(t *testing.T) {
	t.Parallel()
	dst := config.FastIDSettings{
		ServerAddr: "local-host",
		Namespace:  "local-ns",
	}
	from := config.FastIDSettings{Secret: "from-nacos", GroupID: "xyy"}
	config.MergeFastIDSettingsForTest(&dst, from)
	if dst.Secret != "from-nacos" || dst.ServerAddr != "local-host" || dst.GroupID != "xyy" {
		t.Fatalf("merge failed: %+v", dst)
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

package config

import "testing"

func TestResolveFastIDAppName(t *testing.T) {
	got := resolveFastIDAppName("${spring.application.name}", "ebike-fence", "ebike-fence")
	if got != "ebike-fence" {
		t.Fatalf("got %q", got)
	}
}

func TestResolvedFastIDDefaults(t *testing.T) {
	GlobalConfig = AppConfig{}
	GlobalConfig.Server.Name = "ebike-fence"
	GlobalConfig.Server.Port = 8080
	GlobalConfig.Nacos.Namespace = "prod"
	GlobalConfig.Spring.Xyy.Fastid = FastIDSettings{
		Enabled:    true,
		ServerAddr: "fastid.luopingtech.com",
		GroupID:    "xyy",
		Secret:     "test-secret",
	}
	cfg, enabled := resolvedFastIDSettings()
	if !enabled {
		t.Fatal("expected enabled")
	}
	if cfg.AppName != "ebike-fence" || cfg.Secret != "test-secret" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}

func TestParseFastIDNacosYAML(t *testing.T) {
	content := `
spring:
  xyy:
    fastid:
      enabled: true
      server-addr: fastid.luopingtech.com
      secret: nacos-secret
`
	fastid, err := parseFastIDNacosYAML(content)
	if err != nil {
		t.Fatal(err)
	}
	if !fastid.Enabled || fastid.Secret != "nacos-secret" {
		t.Fatalf("unexpected: %+v", fastid)
	}
}

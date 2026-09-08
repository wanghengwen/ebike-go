package config

import "testing"

func TestResolveFastIDAppName(t *testing.T) {
	got := resolveFastIDAppName("${spring.application.name}", "ebike-fence", "ebike-fence")
	if got != "ebike-fence" {
		t.Fatalf("got %q", got)
	}
}

func TestResolvedFastIDDefaults(t *testing.T) {
	t.Setenv("SPRING_XYY_FASTID_SERVER_ADDR", "")
	t.Setenv("SPRING_XYY_FASTID_SERVER_URL", "")
	t.Setenv("SPRING_XYY_FASTID_URL", "")
	t.Setenv("SPRING_XYY_FASTID_SECRET", "")
	t.Setenv("SPRING_XYY_FASTID_USE_HTTPS", "")

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
	if cfg.URL != "fastid.luopingtech.com" {
		t.Fatalf("url=%s", cfg.URL)
	}
}

func TestServerURLOverridesServerAddr(t *testing.T) {
	t.Setenv("SPRING_XYY_FASTID_SERVER_ADDR", "")
	t.Setenv("SPRING_XYY_FASTID_SERVER_URL", "")
	t.Setenv("SPRING_XYY_FASTID_URL", "")
	t.Setenv("SPRING_XYY_FASTID_SECRET", "")
	t.Setenv("SPRING_XYY_FASTID_USE_HTTPS", "")

	GlobalConfig = AppConfig{}
	GlobalConfig.Server.Name = "ebike-fence"
	GlobalConfig.Spring.Xyy.Fastid = FastIDSettings{
		Enabled:    true,
		ServerAddr: "fastid.luopingtech.com",
		ServerURL:  "http://xyy-fastid:8080",
		Secret:     "from-nacos",
	}
	cfg, enabled := resolvedFastIDSettings()
	if !enabled {
		t.Fatal("expected enabled")
	}
	if cfg.URL != "xyy-fastid:8080" {
		t.Fatalf("url=%s want xyy-fastid:8080 (server-url must override server-addr)", cfg.URL)
	}
	if cfg.UseHTTPS {
		t.Fatal("expected UseHTTPS=false for http://server-url")
	}
}

func TestFastIDEnvServerURLOverridesNacos(t *testing.T) {
	t.Setenv("SPRING_XYY_FASTID_SERVER_URL", "http://env-fastid:8080")
	t.Setenv("SPRING_XYY_FASTID_SERVER_ADDR", "")
	t.Setenv("SPRING_XYY_FASTID_URL", "")
	t.Setenv("SPRING_XYY_FASTID_SECRET", "")
	t.Setenv("SPRING_XYY_FASTID_USE_HTTPS", "")

	GlobalConfig = AppConfig{}
	GlobalConfig.Server.Name = "ebike-fence"
	GlobalConfig.Spring.Xyy.Fastid = FastIDSettings{
		Enabled:    true,
		ServerAddr: "fastid.luopingtech.com",
		ServerURL:  "http://xyy-fastid:8080",
		Secret:     "from-nacos",
	}
	cfg, enabled := resolvedFastIDSettings()
	if !enabled {
		t.Fatal("expected enabled")
	}
	if cfg.URL != "env-fastid:8080" {
		t.Fatalf("url=%s", cfg.URL)
	}
	if cfg.UseHTTPS {
		t.Fatal("expected UseHTTPS=false")
	}
}

func TestParseFastIDNacosYAMLServerURL(t *testing.T) {
	content := `
spring:
  xyy:
    fastid:
      enabled: true
      server-addr: fastid.luopingtech.com
      server-url: http://xyy-fastid:8080
      secret: nacos-secret
`
	fastid, err := parseFastIDNacosYAML(content)
	if err != nil {
		t.Fatal(err)
	}
	if !fastid.Enabled || fastid.Secret != "nacos-secret" {
		t.Fatalf("unexpected: %+v", fastid)
	}
	if fastid.ServerAddr != "fastid.luopingtech.com" || fastid.ServerURL != "http://xyy-fastid:8080" {
		t.Fatalf("addr=%q url=%q", fastid.ServerAddr, fastid.ServerURL)
	}
}

func TestMergeFastIDSettingsServerURL(t *testing.T) {
	dst := FastIDSettings{ServerAddr: "local-host"}
	from := FastIDSettings{
		ServerAddr: "fastid.luopingtech.com",
		ServerURL:  "http://xyy-fastid:8080",
		Secret:     "from-nacos",
	}
	mergeFastIDSettings(&dst, from)
	if dst.ServerAddr != "fastid.luopingtech.com" {
		t.Fatalf("addr=%s", dst.ServerAddr)
	}
	if dst.ServerURL != "xyy-fastid:8080" {
		t.Fatalf("server-url=%s (scheme should be stripped)", dst.ServerURL)
	}
	if dst.UseHTTPS || !dst.UseHTTPSExplicit {
		t.Fatalf("https=%v explicit=%v", dst.UseHTTPS, dst.UseHTTPSExplicit)
	}
}

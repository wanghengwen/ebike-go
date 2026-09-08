package config_test

import (
	"testing"

	"ebike-device-worker-go/internal/config"
)

func TestNacosEnvOverridesNacosConfig(t *testing.T) {
	t.Setenv("NACOS_SERVERADDR", "10.9.1.152")
	t.Setenv("NACOS_PORT", "8848")

	cfg := &config.Config{}
	cfg.Nacos.ServerAddr = "192.168.2.20"
	cfg.Nacos.Port = 9848
	config.FinalizeConfigForTest(cfg)

	if cfg.Nacos.ServerAddr != "10.9.1.152" {
		t.Fatalf("serverAddr=%s", cfg.Nacos.ServerAddr)
	}
	if cfg.Nacos.Port != 8848 {
		t.Fatalf("port=%d", cfg.Nacos.Port)
	}
}

func TestNacosServerAddrEnvOverridesCombined(t *testing.T) {
	t.Setenv("NACOS_SERVER_ADDR", "10.9.1.152:8848")

	cfg := &config.Config{}
	cfg.Nacos.ServerAddr = "192.168.2.20"
	cfg.Nacos.Port = 9848
	config.FinalizeConfigForTest(cfg)

	if cfg.Nacos.ServerAddr != "10.9.1.152" || cfg.Nacos.Port != 8848 {
		t.Fatalf("nacos=%s:%d", cfg.Nacos.ServerAddr, cfg.Nacos.Port)
	}
}

// Bootstrap Nacos client must see NACOS_SERVER_ADDR before FinalizeConfig
// (deployment sets 192.168.1.108 while local yaml may still list an old addr).
func TestNacosServerAddrEnvAppliesBeforeClientCreate(t *testing.T) {
	t.Setenv("NACOS_SERVER_ADDR", "192.168.1.108:8848")

	cfg := &config.Config{}
	cfg.Nacos.ServerAddr = "192.168.2.20"
	cfg.Nacos.Port = 8848
	cfg.Nacos.Namespace = "prod"
	cfg.Nacos.Group = "xyy"
	config.ApplyNacosEnvForTest(cfg)

	if cfg.Nacos.ServerAddr != "192.168.1.108" || cfg.Nacos.Port != 8848 {
		t.Fatalf("nacos=%s:%d (env must override yaml before InitNacosConfigClient)", cfg.Nacos.ServerAddr, cfg.Nacos.Port)
	}
}

func TestFastIDEnvOverridesNacosAfterFinalize(t *testing.T) {
	t.Setenv("SPRING_XYY_FASTID_SERVER_ADDR", "http://xyy-fastid:8080")
	t.Setenv("SPRING_XYY_FASTID_USE_HTTPS", "false")

	cfg := &config.Config{}
	cfg.Spring.Xyy.Fastid.ServerAddr = "xyy-fastid.prod.svc.cluster.local:8080"
	config.FinalizeConfigForTest(cfg)

	fc := cfg.ResolvedFastID()
	if fc.URL != "xyy-fastid:8080" {
		t.Fatalf("url=%s", fc.URL)
	}
	if fc.UseHTTPS {
		t.Fatal("expected UseHTTPS=false")
	}
}

func TestFastIDSchemeFromNacosHTTPAddr(t *testing.T) {
	cfg := &config.Config{}
	config.MergeFastIDSettingsForTest(&cfg.Spring.Xyy.Fastid, config.FastIDSettings{
		ServerAddr: "http://xyy-fastid:8080",
	})
	config.FinalizeConfigForTest(cfg)

	fc := cfg.ResolvedFastID()
	if fc.URL != "xyy-fastid:8080" {
		t.Fatalf("url=%s", fc.URL)
	}
	if fc.UseHTTPS {
		t.Fatal("expected UseHTTPS=false for http:// nacos addr")
	}
}

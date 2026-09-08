package config_test

import (
	"testing"

	"ebike-device-worker-go/internal/config"
)

func TestFastIDSecretFromEnv(t *testing.T) {
	t.Setenv("SPRING_XYY_FASTID_SECRET", "env-secret")
	cfg := &config.Config{}
	config.ApplyEnvOverridesForTest(cfg)
	fc := cfg.ResolvedFastID()
	if fc.Secret != "env-secret" {
		t.Fatalf("secret=%s", fc.Secret)
	}
}

func TestFastIDServerAddrEnvOverridesNacos(t *testing.T) {
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
		t.Fatal("expected UseHTTPS=false from env")
	}
	if !fc.UseHTTPSExplicit {
		t.Fatal("expected UseHTTPSExplicit=true from env")
	}
}

func TestResolvedFastIDFromSpringXyy(t *testing.T) {
	cfg := &config.Config{}
	cfg.Spring.Application.Name = "ebike-device-worker"
	cfg.Spring.Xyy.Fastid = config.FastIDSettings{
		Enabled:    true,
		ServerAddr: "fastid.luopingtech.com",
		Namespace:  "prod",
		GroupID:    "xyy",
		Secret:     "test-secret",
		AppName:    "ebike-device-worker",
	}
	cfg.Server.Port = "8080"

	fc := cfg.ResolvedFastID()
	if fc.URL != "fastid.luopingtech.com" {
		t.Fatalf("url=%s", fc.URL)
	}
	if fc.Namespace != "prod" || fc.GroupID != "xyy" {
		t.Fatalf("namespace/group mismatch: %s %s", fc.Namespace, fc.GroupID)
	}
	if fc.AppName != "ebike-device-worker" || fc.Secret != "test-secret" {
		t.Fatalf("app/secret mismatch")
	}
	if fc.Port != 8080 {
		t.Fatalf("port=%d", fc.Port)
	}
}

func TestServerURLOverridesServerAddr(t *testing.T) {
	t.Setenv("SPRING_XYY_FASTID_SERVER_ADDR", "")
	t.Setenv("SPRING_XYY_FASTID_SERVER_URL", "")
	t.Setenv("SPRING_XYY_FASTID_URL", "")

	cfg := &config.Config{}
	cfg.Spring.Application.Name = "ebike-device-worker"
	cfg.Spring.Xyy.Fastid = config.FastIDSettings{
		Enabled:    true,
		ServerAddr: "fastid.luopingtech.com",
		ServerURL:  "http://xyy-fastid:8080",
		Secret:     "from-nacos",
	}
	config.FinalizeConfigForTest(cfg)
	fc := cfg.ResolvedFastID()
	if fc.URL != "xyy-fastid:8080" {
		t.Fatalf("url=%s", fc.URL)
	}
	if fc.UseHTTPS {
		t.Fatal("expected UseHTTPS=false")
	}
}

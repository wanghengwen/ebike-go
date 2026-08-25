package config

import "testing"

func TestParseHostPort(t *testing.T) {
	host, port := parseHostPort("mse.example.com:8848", 8848)
	if host != "mse.example.com" || port != 8848 {
		t.Fatalf("host=%q port=%d", host, port)
	}
	host, port = parseHostPort("127.0.0.1", 8848)
	if host != "127.0.0.1" || port != 8848 {
		t.Fatalf("host=%q port=%d", host, port)
	}
}

func TestApplyNacosEnvOverridesServerAddr(t *testing.T) {
	GlobalConfig = AppConfig{}
	GlobalConfig.Nacos.Port = 8848
	t.Setenv("NACOS_SERVER_ADDR", "nacos.prod.internal:9848")
	ApplyNacosEnvOverrides()
	if GlobalConfig.Nacos.ServerAddr != "nacos.prod.internal" {
		t.Fatalf("serverAddr = %q", GlobalConfig.Nacos.ServerAddr)
	}
	if GlobalConfig.Nacos.Port != 9848 {
		t.Fatalf("port = %d", GlobalConfig.Nacos.Port)
	}
}

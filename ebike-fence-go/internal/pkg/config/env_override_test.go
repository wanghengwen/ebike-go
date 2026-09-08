package config

import "testing"

func TestApplyEnvOverridesNacosServerAddr(t *testing.T) {
	t.Setenv("NACOS_SERVER_ADDR", "192.168.1.108:8848")
	t.Setenv("NACOS_NAMESPACE", "prod")
	t.Setenv("NACOS_GROUP", "xyy")

	GlobalConfig = AppConfig{}
	GlobalConfig.Nacos.ServerAddr = "192.168.2.20"
	GlobalConfig.Nacos.Port = 9848
	ApplyEnvOverrides()

	if GlobalConfig.Nacos.ServerAddr != "192.168.1.108" || GlobalConfig.Nacos.Port != 8848 {
		t.Fatalf("nacos=%s:%d", GlobalConfig.Nacos.ServerAddr, GlobalConfig.Nacos.Port)
	}
	if GlobalConfig.Nacos.Namespace != "prod" || GlobalConfig.Nacos.Group != "xyy" {
		t.Fatalf("ns=%s group=%s", GlobalConfig.Nacos.Namespace, GlobalConfig.Nacos.Group)
	}
}

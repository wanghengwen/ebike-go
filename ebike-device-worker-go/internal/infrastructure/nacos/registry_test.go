package nacos_test

import (
	"testing"

	"ebike-device-worker-go/internal/config"
	nacosreg "ebike-device-worker-go/internal/infrastructure/nacos"
)

func TestServiceNameAndPort(t *testing.T) {
	config.GlobalConfig = &config.Config{}
	config.GlobalConfig.Spring.Application.Name = "ebike-device-worker"
	config.GlobalConfig.Server.Port = "8080"

	if nacosreg.ServiceName() != "ebike-device-worker" {
		t.Fatalf("name=%s", nacosreg.ServiceName())
	}
	port, err := nacosreg.ServerPort()
	if err != nil || port != 8080 {
		t.Fatalf("port=%d err=%v", port, err)
	}
}

func TestRegisterSkippedWhenDisabled(t *testing.T) {
	config.GlobalConfig = &config.Config{}
	config.GlobalConfig.Nacos.RegisterEnabled = false
	if err := nacosreg.RegisterInstance(); err != nil {
		t.Fatal(err)
	}
}

func TestLocalIPPrefersPodIP(t *testing.T) {
	t.Setenv("POD_IP", "10.244.1.88")
	if ip := nacosreg.LocalIP(); ip != "10.244.1.88" {
		t.Fatalf("LocalIP()=%q want POD_IP", ip)
	}
}

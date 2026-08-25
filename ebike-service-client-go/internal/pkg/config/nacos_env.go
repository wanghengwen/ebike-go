package config

import (
	"os"
	"strconv"
	"strings"
)

// ApplyNacosEnvOverrides applies Nacos connection settings from environment variables,
// mirroring Spring Cloud Nacos and other Go services in this repo.
func ApplyNacosEnvOverrides() {
	if v := firstNonEmpty(
		os.Getenv("SPRING_CLOUD_NACOS_CONFIG_SERVER_ADDR"),
		os.Getenv("SPRING_CLOUD_NACOS_DISCOVERY_SERVER_ADDR"),
		os.Getenv("NACOS_SERVER_ADDR"),
	); v != "" {
		host, port := parseHostPort(v, GlobalConfig.Nacos.Port)
		GlobalConfig.Nacos.ServerAddr = host
		if port > 0 {
			GlobalConfig.Nacos.Port = port
		}
	}

	if v := firstNonEmpty(
		os.Getenv("SPRING_CLOUD_NACOS_CONFIG_NAMESPACE"),
		os.Getenv("SPRING_CLOUD_NACOS_DISCOVERY_NAMESPACE"),
		os.Getenv("NACOS_NAMESPACE"),
	); v != "" {
		GlobalConfig.Nacos.Namespace = v
	}

	if v := firstNonEmpty(
		os.Getenv("SPRING_CLOUD_NACOS_CONFIG_GROUP"),
		os.Getenv("SPRING_CLOUD_NACOS_DISCOVERY_GROUP"),
		os.Getenv("NACOS_GROUP"),
	); v != "" {
		GlobalConfig.Nacos.Group = v
	}

	if v := os.Getenv("NACOS_CONTEXT_PATH"); v != "" {
		GlobalConfig.Nacos.ContextPath = v
	}

	normalizeNacosServerAddr()
}

func normalizeNacosServerAddr() {
	addr := strings.TrimSpace(GlobalConfig.Nacos.ServerAddr)
	if addr == "" || !strings.Contains(addr, ":") {
		return
	}
	host, port := parseHostPort(addr, GlobalConfig.Nacos.Port)
	GlobalConfig.Nacos.ServerAddr = host
	if port > 0 {
		GlobalConfig.Nacos.Port = port
	}
}

func parseHostPort(addr string, defaultPort uint64) (string, uint64) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "", defaultPort
	}
	if !strings.Contains(addr, ":") {
		return addr, defaultPort
	}
	host, portStr, ok := strings.Cut(addr, ":")
	if !ok || host == "" {
		return addr, defaultPort
	}
	port, err := strconv.ParseUint(strings.TrimSpace(portStr), 10, 64)
	if err != nil || port == 0 {
		return host, defaultPort
	}
	return host, port
}

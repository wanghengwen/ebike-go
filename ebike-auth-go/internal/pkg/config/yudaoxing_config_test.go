package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNacosYudaoxingProperties_KebabCase(t *testing.T) {
	// 与 Nacos ebike-auth-client.yml 中 Spring Boot 风格字段一致。
	yamlContent := `
yudaoxing:
  account: "331083002"
  base-url: "https://www.yhznbc.com/oauth2"
  base-url1: "http://183.134.75.180:80/oauth2"
  server-public-key: "server-pub-key"
  merchant-private-key: "merchant-priv-key"
`
	var props NacosAppProperties
	if err := yaml.Unmarshal([]byte(yamlContent), &props); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	cfg := props.Yudaoxing.toYudaoxingConfig()
	if cfg.Account != "331083002" {
		t.Fatalf("account = %q", cfg.Account)
	}
	if cfg.BaseUrl != "https://www.yhznbc.com/oauth2" {
		t.Fatalf("baseUrl = %q", cfg.BaseUrl)
	}
	if cfg.ServerPublicKey != "server-pub-key" {
		t.Fatalf("serverPublicKey = %q", cfg.ServerPublicKey)
	}
	if cfg.MerchantPrivateKey != "merchant-priv-key" {
		t.Fatalf("merchantPrivateKey = %q", cfg.MerchantPrivateKey)
	}
}

func TestNacosYudaoxingProperties_CamelCaseFallback(t *testing.T) {
	yamlContent := `
yudaoxing:
  account: "acc"
  baseUrl: "https://example.com/oauth2"
  serverPublicKey: "pub"
  merchantPrivateKey: "priv"
`
	var props NacosAppProperties
	if err := yaml.Unmarshal([]byte(yamlContent), &props); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	cfg := props.Yudaoxing.toYudaoxingConfig()
	if cfg.BaseUrl != "https://example.com/oauth2" {
		t.Fatalf("baseUrl = %q", cfg.BaseUrl)
	}
	if cfg.MerchantPrivateKey != "priv" {
		t.Fatalf("merchantPrivateKey = %q", cfg.MerchantPrivateKey)
	}
}

func TestNacosYudaoxingProperties_KebabPreferredOverCamel(t *testing.T) {
	yamlContent := `
yudaoxing:
  base-url: "https://kebab.example.com"
  baseUrl: "https://camel.example.com"
`
	var props NacosAppProperties
	if err := yaml.Unmarshal([]byte(yamlContent), &props); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	cfg := props.Yudaoxing.toYudaoxingConfig()
	if cfg.BaseUrl != "https://kebab.example.com" {
		t.Fatalf("expected kebab-case preferred, got %q", cfg.BaseUrl)
	}
}

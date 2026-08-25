package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNacosWeixinProperties_KebabCase(t *testing.T) {
	yamlContent := `
weixin:
  publicplatform:
    app-id: wx-test-app
    app-secret: test-secret
feign:
  weixin:
    url: https://api.weixin.qq.com
`
	var props NacosAppProperties
	if err := yaml.Unmarshal([]byte(yamlContent), &props); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	cfg := props.Weixin.toWeixinPublicPlatformConfig()
	if cfg.AppId != "wx-test-app" {
		t.Fatalf("appId = %q", cfg.AppId)
	}
	if cfg.AppSecret != "test-secret" {
		t.Fatalf("appSecret = %q", cfg.AppSecret)
	}
}

func TestNacosWeixinProperties_CamelCaseFallback(t *testing.T) {
	yamlContent := `
weixin:
  publicplatform:
    appid: wx-camel
    appSecret: secret-camel
`
	var props NacosAppProperties
	if err := yaml.Unmarshal([]byte(yamlContent), &props); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	cfg := props.Weixin.toWeixinPublicPlatformConfig()
	if cfg.AppId != "wx-camel" {
		t.Fatalf("appId = %q", cfg.AppId)
	}
	if cfg.AppSecret != "secret-camel" {
		t.Fatalf("appSecret = %q", cfg.AppSecret)
	}
}

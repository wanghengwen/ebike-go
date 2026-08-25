package config

import "testing"

func TestMergeRemoteYAMLMapServiceFromXyy(t *testing.T) {
	GlobalConfig = AppConfig{}
	content := `
xyy:
  mapServiceConfig:
    url: http://map-service.test:8080
    secret: test-secret
  gatewayfilter:
    exclude: /foo
`
	MergeRemoteYAML(content)
	if GlobalConfig.Xyy.MapServiceConfig.Url != "http://map-service.test:8080" {
		t.Fatalf("url = %q", GlobalConfig.Xyy.MapServiceConfig.Url)
	}
	if GlobalConfig.Xyy.MapServiceConfig.Secret != "test-secret" {
		t.Fatalf("secret = %q", GlobalConfig.Xyy.MapServiceConfig.Secret)
	}
	if GlobalConfig.Xyy.GatewayFilter.Exclude != "/foo" {
		t.Fatalf("exclude = %q", GlobalConfig.Xyy.GatewayFilter.Exclude)
	}
}

func TestMergeRemoteYAMLMapServiceFromXiaoantechFallback(t *testing.T) {
	GlobalConfig = AppConfig{}
	content := `
xiaoantech:
  mapServiceConfig:
    url: http://map-service.legacy:8080
    secret: legacy-secret
gaode:
  navigate:
    url: https://gaode-navigate.example
`
	MergeRemoteYAML(content)
	if GlobalConfig.Xyy.MapServiceConfig.Url != "http://map-service.legacy:8080" {
		t.Fatalf("url = %q", GlobalConfig.Xyy.MapServiceConfig.Url)
	}
	if GlobalConfig.Gaode.Navigate.Url != "https://gaode-navigate.example" {
		t.Fatalf("gaode url = %q", GlobalConfig.Gaode.Navigate.Url)
	}
}

func TestMergeRemoteYAMLMapServiceFromRootLevel(t *testing.T) {
	GlobalConfig = AppConfig{}
	content := `
mapServiceConfig:
  url: http://map-service.root:8080
  secret: root-secret
`
	MergeRemoteYAML(content)
	if GlobalConfig.Xyy.MapServiceConfig.Url != "http://map-service.root:8080" {
		t.Fatalf("url = %q", GlobalConfig.Xyy.MapServiceConfig.Url)
	}
	if GlobalConfig.Xyy.MapServiceConfig.Secret != "root-secret" {
		t.Fatalf("secret = %q", GlobalConfig.Xyy.MapServiceConfig.Secret)
	}
}

func TestMergeRemoteYAMLMapServiceFromMapServiceYaml(t *testing.T) {
	GlobalConfig = AppConfig{}
	// prod: group=xyy_ops, dataId=map-service.yaml (Java bootstrap extension-config)
	content := `
xyy:
  mapServiceConfig:
    url: http://map.luopingtech.com
    secret: test-secret-from-map-service-yaml
    api: aMap
`
	MergeRemoteYAML(content)
	if GlobalConfig.Xyy.MapServiceConfig.Url != "http://map.luopingtech.com" {
		t.Fatalf("url = %q", GlobalConfig.Xyy.MapServiceConfig.Url)
	}
	if GlobalConfig.Xyy.MapServiceConfig.Secret != "test-secret-from-map-service-yaml" {
		t.Fatalf("secret = %q", GlobalConfig.Xyy.MapServiceConfig.Secret)
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	GlobalConfig = AppConfig{}
	t.Setenv("XYY_MAP_SERVICE_URL", "http://env-map:8080")
	t.Setenv("XYY_MAP_SERVICE_SECRET", "env-secret")
	ApplyEnvOverrides()
	if GlobalConfig.Xyy.MapServiceConfig.Url != "http://env-map:8080" {
		t.Fatalf("url = %q", GlobalConfig.Xyy.MapServiceConfig.Url)
	}
	if GlobalConfig.Xyy.MapServiceConfig.Secret != "env-secret" {
		t.Fatalf("secret = %q", GlobalConfig.Xyy.MapServiceConfig.Secret)
	}
}

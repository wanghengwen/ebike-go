package rpc

import (
	"testing"

	"ebike-service-client-go/internal/pkg/config"
)

func TestNacosExtensionDataIdsIncludesJavaParity(t *testing.T) {
	cfg := config.AppConfig{}
	cfg.Nacos.DataId = "ebike-service-client.yaml"
	ids := nacosExtensionDataIds(cfg)
	want := map[string]bool{
		"ebike-service-client.yaml":    false,
		"ebike-service-client.yml":     false,
		"map-service.yaml":             false,
		"ebike-service-client-go.yaml": false,
	}
	for _, id := range ids {
		if _, ok := want[id]; ok {
			want[id] = true
		}
	}
	for id, found := range want {
		if !found {
			t.Fatalf("missing dataId %q in %v", id, ids)
		}
	}
}

func TestNacosExtensionGroupsIncludesOps(t *testing.T) {
	cfg := config.AppConfig{}
	cfg.Nacos.Group = "xyy"
	groups := nacosExtensionGroups(cfg)
	hasXYy, hasOps := false, false
	for _, g := range groups {
		if g == "xyy" {
			hasXYy = true
		}
		if g == "xyy_ops" {
			hasOps = true
		}
	}
	if !hasXYy || !hasOps {
		t.Fatalf("groups = %v", groups)
	}
}

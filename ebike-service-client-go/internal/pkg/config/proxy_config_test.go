package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProxyLists(t *testing.T) {
	dir := t.TempDir()
	appPath := filepath.Join(dir, "application.yml")
	listsPath := filepath.Join(dir, "proxy_lists.yml")
	if err := os.WriteFile(appPath, []byte("server:\n  port: 1\n  name: test\nproxy:\n  target_url: http://java\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(listsPath, []byte("live_list:\n  - /a\nrecord_list:\n  - /b\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	LoadConfig(appPath)
	if len(GlobalConfig.Proxy.LiveList) != 1 || GlobalConfig.Proxy.LiveList[0] != "/a" {
		t.Fatalf("live_list=%v", GlobalConfig.Proxy.LiveList)
	}
	if len(GlobalConfig.Proxy.RecordList) != 1 || GlobalConfig.Proxy.RecordList[0] != "/b" {
		t.Fatalf("record_list=%v", GlobalConfig.Proxy.RecordList)
	}
}

func TestApplyProxyEnvOverrides(t *testing.T) {
	GlobalConfig.Proxy.TargetURL = "http://yaml"
	t.Setenv("PROXY_TARGET_URL", "http://env")
	ApplyProxyEnvOverrides()
	if GlobalConfig.Proxy.TargetURL != "http://env" {
		t.Fatalf("target=%q", GlobalConfig.Proxy.TargetURL)
	}
}

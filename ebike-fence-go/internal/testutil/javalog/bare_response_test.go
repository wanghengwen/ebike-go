package javalog

import "testing"

func TestIsBareResponseURL(t *testing.T) {
	if !IsBareResponseURL("/fence/tags/getAll") {
		t.Fatal("expected bare response for fence tags")
	}
	if IsBareResponseURL("/config/usecar/getConfigByServiceId") {
		t.Fatal("usecar config should use Result envelope")
	}
}

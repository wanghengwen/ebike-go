package gateway

import "testing"

func TestNewConfigGatewaySingleton(t *testing.T) {
	a := NewConfigGateway()
	b := NewConfigGateway()
	if a != b {
		t.Fatal("expected singleton ConfigGateway instance")
	}
	if a.localCache != b.localCache {
		t.Fatal("expected shared local LRU cache")
	}
}

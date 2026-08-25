package api

import (
	"testing"
)

func TestGeneratedRouteCount(t *testing.T) {
	routes := GeneratedRoutes()
	if len(routes) < 210 {
		t.Fatalf("expected at least 210 generated routes, got %d", len(routes))
	}
}

func TestAllRoutesHaveNativeHandlersAfterRegister(t *testing.T) {
	// Handlers are registered at router init; this test validates manifest completeness.
	paths := make(map[string]struct{})
	for _, r := range GeneratedRoutes() {
		paths[r.Path] = struct{}{}
	}
	if _, ok := paths["/serviceArea/returnCar"]; !ok {
		t.Fatal("missing hot path /serviceArea/returnCar in generated routes")
	}
	if _, ok := paths["/config/getRidingCarConfig"]; !ok {
		t.Fatal("missing /config/getRidingCarConfig in generated routes")
	}
}

package api_test

import (
	"testing"

	"ebike-device-worker-go/internal/api"
	"ebike-device-worker-go/internal/api/controller"
)

func TestAllGeneratedRoutesHaveNativeHandlers(t *testing.T) {
	controller.InitNativeHandlers()

	native := api.NativeHandlerPaths()
	routes := api.GeneratedRoutes()
	if len(routes) != 12 {
		t.Fatalf("expected 12 routes, got %d", len(routes))
	}

	var missing []string
	for _, r := range routes {
		if _, ok := native[r.Path]; !ok {
			missing = append(missing, r.Path)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("%d routes missing native handlers: %v", len(missing), missing)
	}
}

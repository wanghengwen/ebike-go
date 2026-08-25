package api_test

import (
	"testing"

	"ebike-fence-go/internal/api"
	"ebike-fence-go/internal/api/controller"
)

func TestAllGeneratedRoutesHaveNativeHandlers(t *testing.T) {
	controller.InitNativeHandlers()

	native := api.NativeHandlerPaths()
	routes := api.GeneratedRoutes()
	if len(routes) < 210 {
		t.Fatalf("expected >=210 routes, got %d", len(routes))
	}

	var missing []string
	for _, r := range routes {
		if _, ok := native[r.Path]; !ok {
			missing = append(missing, r.Path)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("%d routes missing native handlers, first 10: %v", len(missing), missing[:min(10, len(missing))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

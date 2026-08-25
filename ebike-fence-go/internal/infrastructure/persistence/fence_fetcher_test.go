package persistence

import (
	"testing"

	"ebike-fence-go/internal/domain/fence"
)

func TestFenceTypeFromPrefixIncludesCustom(t *testing.T) {
	typ, ok := fenceTypeFromPrefix("fence_custom")
	if !ok || typ != fence.TypeCustom {
		t.Fatalf("fence_custom = %d %v want %d true", typ, ok, fence.TypeCustom)
	}
}

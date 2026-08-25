package geohash_test

import (
	"testing"

	"ebike-analyze-go/internal/common/geohash"
)

func TestGetSpaceCoordinate_JavaSample(t *testing.T) {
	coord := geohash.GetSpaceCoordinate("wt3ms03nf")
	if coord[0] == 0 && coord[1] == 0 {
		t.Fatalf("expected non-zero coordinate for wt3ms03nf, got %v", coord)
	}
}

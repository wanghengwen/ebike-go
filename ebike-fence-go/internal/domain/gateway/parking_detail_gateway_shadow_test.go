package gateway

import (
	"context"
	"testing"

	"ebike-fence-go/internal/pkg/shadow"
)

func TestSetDetailCacheSkipsInShadow(t *testing.T) {
	ctx := shadow.WithShadowTest(context.Background())
	if err := SetDetailCache(ctx, "1000", &ParkingDetailE{
		CarId:     "car-1",
		ParkingId: 123,
	}); err != nil {
		t.Fatalf("shadow mode should skip Redis write: %v", err)
	}
}

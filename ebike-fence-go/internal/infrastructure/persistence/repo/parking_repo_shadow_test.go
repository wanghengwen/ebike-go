package repo

import (
	"context"
	"testing"

	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/pkg/shadow"
)

func TestSaveOrUpdateV1SkipsInShadow(t *testing.T) {
	ctx := shadow.WithShadowTest(context.Background())
	r := &ParkingRepository{}
	if err := r.SaveOrUpdateV1(ctx, "1000", "", &gateway.ParkingDetailE{
		CarId:     "car-1",
		ParkingId: 123,
	}); err != nil {
		t.Fatalf("shadow mode should skip DB write: %v", err)
	}
}

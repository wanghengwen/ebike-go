package service

import (
	"context"

	"ebike-fence-go/internal/infrastructure/persistence"
	"ebike-fence-go/internal/infrastructure/rpc"
)

// ParkingCarCount returns vehicles bound to a parking site (DB first, RPC fallback).
func ParkingCarCount(ctx context.Context, tenantId string, parkingId int64) (int64, error) {
	if persistence.Parking != nil {
		return persistence.Parking.CountByParkingID(ctx, tenantId, parkingId)
	}
	return rpc.ParkingCarCount(ctx, tenantId, parkingId)
}

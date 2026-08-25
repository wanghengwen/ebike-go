package service

import (
	"context"

	"ebike-fence-go/internal/domain/config"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/geo"
)

// GetBanRidingByLocation finds which ban-riding area the location belongs to
func GetBanRidingByLocation(ctx context.Context, tenantId string, loc geo.Location) (*gateway.FenceE, error) {
	return FindZoneByLocation(ctx, tenantId, loc, ZoneCheckOption{
		FencePrefix:     BanRidingPrefix,
		GeoPrefix:       BanRidingGeo,
		Radius:          1000.0,
		Count:           20,
		UseBuffer:       false,
		FilterByService: false, // Java doesn't filter BanRiding by ServiceAreaId
	})
}

// izInBanRiding mirrors Java BanRidingServiceImpl.izInBanRiding.
func izInBanRiding(ctx context.Context, tenantId string, loc geo.Location) (*gateway.FenceE, error) {
	if loc.Lng == 0 && loc.Lat == 0 {
		return nil, nil
	}
	return GetBanRidingByLocation(ctx, tenantId, loc)
}

// BanRidingReturnCar mirrors Java BanRidingServiceImpl.returnCar.
// Java currently returns null (switch logic commented out); align with production behavior.
func BanRidingReturnCar(ctx context.Context, tenantId string, carLoc, userLoc geo.Location, backcarConfig *config.ConfigBackcarCO) (*gateway.FenceE, error) {
	_ = ctx
	_ = tenantId
	_ = carLoc
	_ = userLoc
	_ = backcarConfig
	return nil, nil
}

package service

import (
	"context"

	"ebike-fence-go/internal/domain/config"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/geo"
)

// NoParkingReturnCar mirrors Java NoParkingServiceImpl.returnCar.
func NoParkingReturnCar(ctx context.Context, tenantId string, carLoc, userLoc geo.Location, serviceAreaId int64, backcarConfig *config.ConfigBackcarCO) (*gateway.FenceE, error) {
	mode := 0
	if backcarConfig != nil && backcarConfig.NoParkingBackcar != nil {
		mode = *backcarConfig.NoParkingBackcar
	}
	switch mode {
	case 1:
		return izInNoParking(ctx, tenantId, carLoc, serviceAreaId, backcarConfig)
	case 2:
		return izInNoParking(ctx, tenantId, userLoc, serviceAreaId, backcarConfig)
	default:
		np, err := izInNoParking(ctx, tenantId, carLoc, serviceAreaId, backcarConfig)
		if err != nil {
			return nil, err
		}
		if np != nil {
			return np, nil
		}
		return izInNoParking(ctx, tenantId, userLoc, serviceAreaId, backcarConfig)
	}
}

func izInNoParking(ctx context.Context, tenantId string, loc geo.Location, serviceAreaId int64, backcarConfig *config.ConfigBackcarCO) (*gateway.FenceE, error) {
	if loc.Lng == 0 && loc.Lat == 0 {
		return nil, nil
	}
	fences, err := gateway.GetFencesByLocation(ctx, NoParkingPrefix, NoParkingGeo+"_"+tenantId, tenantId, loc.Lng, loc.Lat, 5000, 20, serviceAreaId)
	if err != nil {
		return nil, err
	}
	useOther := backcarConfig != nil && backcarConfig.GetIzUseOtherParking()
	for _, fence := range fences {
		if !useOther && fence.ServiceId != serviceAreaId {
			continue
		}
		if geo.IsPointInParsedPolygon(loc, fence.ParsedPolygon) {
			f := fence
			return &f, nil
		}
	}
	return nil, nil
}

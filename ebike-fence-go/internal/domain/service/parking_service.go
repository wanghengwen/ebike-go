package service

import (
	"context"

	"ebike-fence-go/internal/domain/config"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/geo"
)

// ParkingReturnCar mirrors Java ParkingServiceImpl.returnCar.
func ParkingReturnCar(ctx context.Context, tenantId string, carLoc, userLoc geo.Location, serviceAreaId int64, backcarConfig *config.ConfigBackcarCO) (*gateway.FenceE, error) {
	mode := 0
	if backcarConfig != nil && backcarConfig.ParkingBackcar != nil {
		mode = *backcarConfig.ParkingBackcar
	}
	switch mode {
	case 1:
		return izInParking(ctx, tenantId, carLoc, serviceAreaId, backcarConfig)
	case 2:
		return izInParking(ctx, tenantId, userLoc, serviceAreaId, backcarConfig)
	default:
		p, err := izInParking(ctx, tenantId, carLoc, serviceAreaId, backcarConfig)
		if err != nil {
			return nil, err
		}
		if p != nil {
			return p, nil
		}
		return izInParking(ctx, tenantId, userLoc, serviceAreaId, backcarConfig)
	}
}

func izInParking(ctx context.Context, tenantId string, loc geo.Location, serviceAreaId int64, backcarConfig *config.ConfigBackcarCO) (*gateway.FenceE, error) {
	if loc.Lng == 0 && loc.Lat == 0 {
		return nil, nil
	}
	fences, err := gateway.GetFencesByLocation(ctx, ParkingPrefix, ParkingGeo+"_"+tenantId, tenantId, loc.Lng, loc.Lat, 1000, 20, serviceAreaId)
	if err != nil {
		return nil, err
	}
	useOther := backcarConfig != nil && backcarConfig.GetIzUseOtherParking()
	for _, fence := range fences {
		if !useOther && fence.ServiceId != serviceAreaId {
			continue
		}
		if !fence.IzEnable {
			continue
		}
		if !isInOpeningHours(&fence) {
			continue
		}
		bufferDistance := fence.BufferDistance
		if bufferDistance <= 0 && backcarConfig != nil && backcarConfig.BufferDistance != nil {
			bufferDistance = *backcarConfig.BufferDistance
		}
		if bufferDistance <= 0 {
			bufferDistance = 5
		}
		if geo.Intersect(loc, fence.ParsedPolygon, bufferDistance) {
			f := applyParkingBackcarDefaults(&fence, backcarConfig)
			return f, nil
		}
	}
	return nil, nil
}

func applyParkingBackcarDefaults(f *gateway.FenceE, backcarConfig *config.ConfigBackcarCO) *gateway.FenceE {
	if f == nil {
		return nil
	}
	out := *f
	if out.IzFullPileNoStop == nil && backcarConfig != nil && backcarConfig.IzFullPileNoStop != nil {
		v := *backcarConfig.IzFullPileNoStop
		out.IzFullPileNoStop = &v
	}
	return &out
}

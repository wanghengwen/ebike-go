package service

import (
	"context"

	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/geo"
)

// FenceRedisKeys
const (
	ServiceAreaPrefix = "fence_serviceArea"
	ServiceAreaGeo    = "fence_serviceArea_geo"
	ParkingPrefix     = "fence_parking"
	ParkingGeo        = "fence_parking_geo"
	NoParkingPrefix   = "fence_noParking"
	NoParkingGeo      = "fence_noParking_geo"
	BanRidingPrefix   = "fence_banRiding"
	BanRidingGeo      = "fence_banRiding_geo"
	MaintainAreaPrefix = "fence_maintainArea"
	MaintainAreaGeo    = "fence_maintainArea_geo"
	FenceCustomPrefix  = "fence_custom"
)

// FindMaintainAreaNear returns the nearest maintain area within radius meters.
func FindMaintainAreaNear(ctx context.Context, tenantId string, loc geo.Location, radius float64, count int) (*gateway.FenceE, error) {
	fences, err := gateway.GetFencesByLocation(ctx, MaintainAreaPrefix, MaintainAreaGeo+"_"+tenantId, tenantId, loc.Lng, loc.Lat, radius, count, 0)
	if err != nil || len(fences) == 0 {
		return nil, err
	}
	f := fences[0]
	return &f, nil
}

// FindNearestServiceArea returns the nearest service area (Java getNearServiceByLocation).
func FindNearestServiceArea(ctx context.Context, tenantId string, loc geo.Location) (*gateway.FenceE, error) {
	fences, err := gateway.GetFencesByLocation(ctx, ServiceAreaPrefix, ServiceAreaGeo+"_"+tenantId, tenantId, loc.Lng, loc.Lat, 20000000, 20, 0)
	if err != nil || len(fences) == 0 {
		return nil, err
	}
	var inService []gateway.FenceE
	for _, f := range fences {
		if ok, _ := geo.IsPointInPolygon(loc, f.PointList); ok {
			inService = append(inService, f)
		}
	}
	if len(inService) > 0 {
		f := inService[0]
		return &f, nil
	}
	f := fences[0]
	return &f, nil
}

// GetServiceAreaByLocation finds which service area the location belongs to
func GetServiceAreaByLocation(ctx context.Context, tenantId string, loc geo.Location) (*gateway.FenceE, error) {
	return FindZoneByLocation(ctx, tenantId, loc, ZoneCheckOption{
		FencePrefix:     ServiceAreaPrefix,
		GeoPrefix:       ServiceAreaGeo,
		Radius:          200000.0, // 200 km
		Count:           20,
		UseBuffer:       false,
		FilterByService: false,
	})
}

// GetServiceAreaById finds a specific service area by ID
func GetServiceAreaById(ctx context.Context, tenantId string, serviceAreaId int64) (*gateway.FenceE, error) {
	return gateway.GetFenceById(ctx, ServiceAreaPrefix, tenantId, serviceAreaId)
}

package fenceadmin

import (
	"context"
	"fmt"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/fence"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/geo"
	domainsvc "ebike-fence-go/internal/domain/service"
)

type BanRidingAdmin struct{}

func NewBanRidingAdmin() *BanRidingAdmin { return &BanRidingAdmin{} }

func (b *BanRidingAdmin) GetByID(ctx context.Context, tenantID string, id int64) (dto.BanRidingCO, error) {
	fe, err := gateway.GetFenceById(ctx, domainsvc.BanRidingPrefix, tenantID, id)
	if err != nil {
		return dto.BanRidingCO{}, err
	}
	if fe == nil {
		fe, err = getFenceOrErr(ctx, id)
		if err != nil {
			return dto.BanRidingCO{}, err
		}
	}
	return toBanRidingCO(*fe), nil
}

func (b *BanRidingAdmin) GetListByServiceID(ctx context.Context, tenantID string, serviceID int64) ([]dto.BanRidingCO, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return nil, err
	}
	list, err := r.ListByTypeAndServiceID(ctx, tenantID, fence.TypeBanRiding, serviceID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.BanRidingCO, 0, len(list))
	for _, fe := range list {
		out = append(out, toBanRidingCO(fe))
	}
	return out, nil
}

func (b *BanRidingAdmin) Create(ctx context.Context, tenantID, pin string, cmd dto.BanRidingCmd) (int64, error) {
	fe := cmdToFenceE(fence.TypeBanRiding, cmd.Name, cmd.ShapeType, cmd.CenterLat, cmd.CenterLng, cmd.PointList, cmd.ServiceId)
	sid := fe.ServiceId
	return saveFenceNoDedupe(ctx, tenantID, pin, &fe, writeOpts{kind: gateway.CacheBanRiding, serviceID: &sid})
}

func (b *BanRidingAdmin) Update(ctx context.Context, tenantID, pin string, cmd dto.BanRidingCmd) error {
	if cmd.Id == 0 {
		return newBizError("00004", "id must not be null")
	}
	fe := cmdToFenceE(fence.TypeBanRiding, cmd.Name, cmd.ShapeType, cmd.CenterLat, cmd.CenterLng, cmd.PointList, cmd.ServiceId)
	fe.Id = cmd.Id
	sid := fe.ServiceId
	return updateFenceNoDedupe(ctx, tenantID, pin, &fe, writeOpts{kind: gateway.CacheBanRiding, serviceID: &sid})
}

func (b *BanRidingAdmin) Delete(ctx context.Context, tenantID string, id int64) error {
	if err := deleteFence(ctx, tenantID, id, writeOpts{kind: gateway.CacheBanRiding}); err != nil {
		return err
	}
	// Java BanRidingServiceImpl.deleteBanRiding: clear t_parking.ban_riding_id.
	return removeParkingBinding(ctx, tenantID, id, fence.TypeBanRiding)
}

func (b *BanRidingAdmin) DeleteBatch(ctx context.Context, tenantID string, ids []int64) error {
	return deleteFences(ctx, tenantID, ids, writeOpts{kind: gateway.CacheBanRiding})
}

func (b *BanRidingAdmin) GetNearAreas(ctx context.Context, tenantID string, cmd dto.NearLocationCmd) ([]dto.BanRidingCO, error) {
	radius := 5000.0
	if cmd.Radius != nil {
		radius = *cmd.Radius
	}
	fences, err := gateway.GetFencesByLocation(ctx, domainsvc.BanRidingPrefix, domainsvc.BanRidingGeo+"_"+tenantID, tenantID, cmd.Location.Lng, cmd.Location.Lat, radius, 20, cmd.ServiceId)
	if err != nil {
		return nil, err
	}
	out := make([]dto.BanRidingCO, 0)
	for _, fe := range fences {
		if cmd.ServiceId > 0 && fe.ServiceId != cmd.ServiceId {
			continue
		}
		out = append(out, toNearBanRidingCO(fe))
	}
	return out, nil
}

// GetNearest returns the nearest ban-riding area, or a Java-shaped empty CO when none
// exist nearby. The empty shape MUST emit distance:null (not 0): ebike-device-consume
// maps null→100m and distance<=0→accOff; a zero distance falsely powers the bike off.
func (b *BanRidingAdmin) GetNearest(ctx context.Context, tenantID string, cmd dto.NearLocationCmd) (any, error) {
	radius := 1000.0
	fences, err := gateway.GetFencesByLocation(ctx, domainsvc.BanRidingPrefix, domainsvc.BanRidingGeo+"_"+tenantID, tenantID, cmd.Location.Lng, cmd.Location.Lat, radius, 20, cmd.ServiceId)
	if err != nil {
		return nil, err
	}
	filtered := make([]gateway.FenceE, 0)
	for _, fe := range fences {
		if cmd.ServiceId > 0 && fe.ServiceId != cmd.ServiceId {
			continue
		}
		filtered = append(filtered, fe)
	}
	if len(filtered) == 0 {
		// Match Java BanRidingServiceImpl: new BanRidingCO() with null distance.
		return emptyBanRidingNearest(), nil
	}
	fe := filtered[0]
	co := toBanRidingCO(fe)
	loc := geo.Location{Lat: cmd.Location.Lat, Lng: cmd.Location.Lng}
	if inside, _ := geo.IsPointInPolygon(loc, fe.PointList); inside {
		co.Distance = 0
	} else {
		pointJSON := fmt.Sprintf(`{"type":"Point","coordinates":[%f,%f]}`, loc.Lng, loc.Lat)
		if dist, err := geo.PointToPolygonDistance(pointJSON, fe.PointList); err == nil {
			co.Distance = dist
		}
	}
	return co, nil
}

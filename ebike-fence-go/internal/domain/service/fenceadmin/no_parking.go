package fenceadmin

import (
	"context"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/fence"
	"ebike-fence-go/internal/domain/gateway"
	domainsvc "ebike-fence-go/internal/domain/service"
)

type NoParkingAdmin struct{}

func NewNoParkingAdmin() *NoParkingAdmin { return &NoParkingAdmin{} }

func (n *NoParkingAdmin) GetByID(ctx context.Context, tenantID string, id int64) (dto.NoParkingCO, error) {
	fe, err := gateway.GetFenceById(ctx, domainsvc.NoParkingPrefix, tenantID, id)
	if err != nil {
		return dto.NoParkingCO{}, err
	}
	if fe == nil {
		fe, err = getFenceOrErr(ctx, id)
		if err != nil {
			return dto.NoParkingCO{}, err
		}
	}
	return toNoParkingCO(*fe), nil
}

func (n *NoParkingAdmin) GetListByServiceID(ctx context.Context, tenantID string, serviceID int64) ([]dto.NoParkingCO, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return nil, err
	}
	list, err := r.ListByTypeAndServiceID(ctx, tenantID, fence.TypeNoParking, serviceID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.NoParkingCO, 0, len(list))
	for _, fe := range list {
		out = append(out, toNoParkingCO(fe))
	}
	return out, nil
}

func (n *NoParkingAdmin) Create(ctx context.Context, tenantID, pin string, cmd dto.NoParkingCmd) (int64, error) {
	fe := cmdToFenceE(fence.TypeNoParking, cmd.Name, cmd.ShapeType, cmd.CenterLat, cmd.CenterLng, cmd.PointList, cmd.ServiceId)
	sid := fe.ServiceId
	return saveFence(ctx, tenantID, pin, &fe, writeOpts{kind: gateway.CacheNoParking, serviceID: &sid})
}

func (n *NoParkingAdmin) Update(ctx context.Context, tenantID, pin string, cmd dto.NoParkingCmd) error {
	if cmd.Id == 0 {
		return newBizError("00004", "id must not be null")
	}
	r, err := requireFenceRepo()
	if err != nil {
		return err
	}
	old, err := r.GetByID(ctx, cmd.Id)
	if err != nil {
		return err
	}
	if old == nil {
		return newBizError("13031", "禁停区不存在，请确认后再试")
	}
	fe := cmdToFenceE(fence.TypeNoParking, cmd.Name, cmd.ShapeType, cmd.CenterLat, cmd.CenterLng, cmd.PointList, cmd.ServiceId)
	fe.Id = cmd.Id
	sid := fe.ServiceId
	return updateFence(ctx, tenantID, pin, &fe, writeOpts{kind: gateway.CacheNoParking, serviceID: &sid})
}

func (n *NoParkingAdmin) Delete(ctx context.Context, tenantID string, id int64) error {
	if err := deleteFence(ctx, tenantID, id, writeOpts{kind: gateway.CacheNoParking}); err != nil {
		return err
	}
	// Java NoParkingServiceImpl.deleteNoParking: clear t_parking.no_parking_id.
	return removeParkingBinding(ctx, tenantID, id, fence.TypeNoParking)
}

func (n *NoParkingAdmin) DeleteBatch(ctx context.Context, tenantID string, ids []int64) error {
	return deleteFences(ctx, tenantID, ids, writeOpts{kind: gateway.CacheNoParking})
}

func (n *NoParkingAdmin) GetNear(ctx context.Context, tenantID string, cmd dto.NearLocationCmd) ([]dto.NoParkingCO, error) {
	radius := 5000.0
	if cmd.Radius != nil {
		radius = *cmd.Radius
	}
	fences, err := gateway.GetFencesByLocation(ctx, domainsvc.NoParkingPrefix, domainsvc.NoParkingGeo+"_"+tenantID, tenantID, cmd.Location.Lng, cmd.Location.Lat, radius, 20, cmd.ServiceId)
	if err != nil {
		return nil, err
	}
	backcar, _ := gateway.NewConfigGateway().GetConfigByServiceId(ctx, tenantID, cmd.ServiceId)
	matched := make([]gateway.FenceE, 0, len(fences))
	for _, fe := range fences {
		if !includeNearNoParking(fe, cmd.ServiceId, backcar) {
			continue
		}
		matched = append(matched, fe)
	}
	out := make([]dto.NoParkingCO, 0, len(matched))
	for _, fe := range matched {
		out = append(out, toNearNoParkingCO(fe))
	}
	return out, nil
}

func (n *NoParkingAdmin) GetList(_ context.Context, _ dto.NoParkingCmd) ([]dto.NoParkingCO, error) {
	return nil, nil
}

func (n *NoParkingAdmin) GetNearest(ctx context.Context, tenantID string, cmd dto.NearLocationCmd) (*dto.NoParkingCO, error) {
	list, err := n.GetNear(ctx, tenantID, cmd)
	if err != nil || len(list) == 0 {
		return nil, err
	}
	return &list[0], nil
}

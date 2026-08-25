package fenceadmin

import (
	"context"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/fence"
	"ebike-fence-go/internal/domain/gateway"
	domainsvc "ebike-fence-go/internal/domain/service"
	"ebike-fence-go/internal/domain/geo"
	"ebike-fence-go/internal/infrastructure/rpc"
	"ebike-fence-go/internal/pkg/shadow"
)

type ServiceAreaAdmin struct{}

func NewServiceAreaAdmin() *ServiceAreaAdmin { return &ServiceAreaAdmin{} }

func (s *ServiceAreaAdmin) GetByID(ctx context.Context, tenantID string, id int64) (dto.ServiceAreaCO, error) {
	fe, err := domainsvc.GetServiceAreaById(ctx, tenantID, id)
	if err != nil {
		return dto.ServiceAreaCO{}, err
	}
	if fe == nil {
		fe, err = getFenceOrErr(ctx, id)
		if err != nil {
			return dto.ServiceAreaCO{}, err
		}
	}
	return toClientServiceAreaCO(*fe), nil
}

func (s *ServiceAreaAdmin) GetList(ctx context.Context, tenantID string) ([]dto.ServiceAreaCO, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return nil, err
	}
	list, err := r.ListByType(ctx, tenantID, fence.TypeServiceArea)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ServiceAreaCO, 0, len(list))
	for _, fe := range list {
		out = append(out, toServiceAreaCO(fe))
	}
	return out, nil
}

func (s *ServiceAreaAdmin) Create(ctx context.Context, tenantID, pin string, cmd dto.ServiceAreaCmd) (int64, error) {
	fe := cmdToFenceE(fence.TypeServiceArea, cmd.Name, cmd.ShapeType, cmd.CenterLat, cmd.CenterLng, cmd.PointList, 0)
	id, err := saveFence(ctx, tenantID, pin, &fe, writeOpts{kind: gateway.CacheServiceArea})
	if err != nil {
		return id, err
	}
	bindServiceAreaRoles(ctx, tenantID, pin, cmd.CommandContext, id)
	return id, nil
}

// bindServiceAreaRoles mirrors Java ServiceAreaServiceImpl.createServiceArea: bind the
// new service area to the headquarters admin (root) role and to the current operator's
// role. Java's batchSave swallows errors (best-effort), so failures here are non-fatal.
func bindServiceAreaRoles(ctx context.Context, tenantID, pin string, cmdCtx *dto.CommandContext, serviceID int64) {
	if shadow.IsShadowTest(ctx) {
		return
	}
	m := rpc.NewManagementRPC()
	cc := dto.EnsureCommandContext(cmdCtx, tenantID)
	serviceIDs := []int64{serviceID}
	var rootRoleID int64
	if root, err := m.GetRootRole(ctx, cc); err == nil && root != nil && root.Id != nil {
		_ = m.BatchSaveServiceRole(ctx, cc, *root.Id, serviceIDs)
		rootRoleID = *root.Id
	}
	if user, err := m.GetUserByPin(ctx, pin, cc); err == nil && user != nil && user.RoleId != nil && *user.RoleId != rootRoleID {
		_ = m.BatchSaveServiceRole(ctx, cc, *user.RoleId, serviceIDs)
	}
}

func (s *ServiceAreaAdmin) Update(ctx context.Context, tenantID, pin string, cmd dto.ServiceAreaCmd) error {
	if cmd.Id == nil {
		return newBizError("00004", "id must not be null")
	}
	fe := cmdToFenceE(fence.TypeServiceArea, cmd.Name, cmd.ShapeType, cmd.CenterLat, cmd.CenterLng, cmd.PointList, 0)
	fe.Id = *cmd.Id
	return updateFence(ctx, tenantID, pin, &fe, writeOpts{kind: gateway.CacheServiceArea})
}

// Delete mirrors Java ServiceAreaController.deleteServiceArea, whose actual delete
// call is commented out in production - the endpoint just returns success WITHOUT
// deleting anything. We intentionally keep the same no-op behavior.
//
// NOTE: the real Go delete implementation is preserved here for reference:
//
//	return deleteFence(ctx, tenantID, id, writeOpts{kind: gateway.CacheServiceArea})
func (s *ServiceAreaAdmin) Delete(ctx context.Context, tenantID string, id int64) error {
	return nil
}

func (s *ServiceAreaAdmin) GetByLocation(ctx context.Context, tenantID string, lat, lng float64) (*dto.ServiceAreaCO, error) {
	fe, err := domainsvc.GetServiceAreaByLocation(ctx, tenantID, geo.Location{Lat: lat, Lng: lng})
	if err != nil {
		return nil, err
	}
	if fe == nil {
		return nil, newBizError("13002", "数据查询失败，未查到数据")
	}
	co := toClientServiceAreaCO(*fe)
	return &co, nil
}

func (s *ServiceAreaAdmin) GetNearByLocation(ctx context.Context, tenantID string, lat, lng float64) (*dto.ServiceAreaCO, error) {
	fe, err := domainsvc.FindNearestServiceArea(ctx, tenantID, geo.Location{Lat: lat, Lng: lng})
	if err != nil {
		return nil, err
	}
	if fe == nil {
		return nil, newBizError("13002", "数据查询失败，未查到数据")
	}
	co := toClientServiceAreaCO(*fe)
	return &co, nil
}

func (s *ServiceAreaAdmin) GetByIDs(ctx context.Context, tenantID string, ids []int64) ([]dto.ServiceAreaCO, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return nil, err
	}
	m, err := r.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ServiceAreaCO, 0, len(m))
	for _, id := range ids {
		if fe, ok := m[id]; ok {
			out = append(out, toClientServiceAreaCO(fe))
		}
	}
	return out, nil
}

func (s *ServiceAreaAdmin) GetAllService(ctx context.Context) ([]dto.ServiceAreaIDCO, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return nil, err
	}
	list, err := r.ListAllByType(ctx, fence.TypeServiceArea)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ServiceAreaIDCO, 0, len(list))
	for _, fe := range list {
		out = append(out, dto.ServiceAreaIDCO{Id: fe.Id})
	}
	return out, nil
}

func (s *ServiceAreaAdmin) GetParkingPartStatistic(ctx context.Context, tenantID string, serviceID int64) (dto.ParkingPartCO, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return dto.ParkingPartCO{}, err
	}
	list, err := r.ListByTypeAndServiceID(ctx, tenantID, fence.TypeParking, serviceID)
	if err != nil {
		return dto.ParkingPartCO{}, err
	}
	var stat dto.ParkingPartCO
	for _, fe := range list {
		if fe.Rfid != nil && *fe.Rfid {
			stat.RfidNum++
		}
		if fe.Directional != nil && *fe.Directional {
			stat.DirectionNum++
		}
		if fe.Camera != nil && *fe.Camera {
			stat.CameraNum++
		}
		if fe.Kickstand != nil && *fe.Kickstand {
			stat.KickstandNum++
		}
		if fe.Tbeacon != nil && *fe.Tbeacon {
			stat.TbeaconNum++
		}
	}
	return stat, nil
}

func (s *ServiceAreaAdmin) GetListByCmd(_ context.Context, _ dto.ServiceAreaCmd) ([]dto.ServiceAreaCO, error) {
	return nil, nil
}

// ComputeOutServiceDistance delegates to domain service (Java ServiceAreaServiceImpl.computeOutServiceDistance).
func (s *ServiceAreaAdmin) ComputeOutServiceDistance(ctx context.Context, tenantID string, cmd dto.ComputeDistanceCmd) (dto.ComputeDistanceCO, error) {
	return domainsvc.ComputeOutServiceDistance(ctx, tenantID, cmd)
}

func (s *ServiceAreaAdmin) GetByRoleIDs(ctx context.Context, tenantID string, cmdCtx *dto.CommandContext, roleIDs []int64) ([]dto.ServiceAreaCO, error) {
	m := rpc.NewManagementRPC()
	cc := dto.EnsureCommandContext(cmdCtx, tenantID)
	serviceIDs, err := m.GetServiceIdsByRole(ctx, cc, roleIDs)
	if err != nil {
		return nil, err
	}
	return s.loadServiceAreasByIDs(ctx, tenantID, serviceIDs, "", false)
}

func (s *ServiceAreaAdmin) GetByRoleIDsAndSubTenant(ctx context.Context, tenantID, pin string, cmdCtx *dto.CommandContext, cmd dto.TenantServiceCmd) ([]dto.ServiceAreaCO, error) {
	m := rpc.NewManagementRPC()
	cc := dto.EnsureCommandContext(cmdCtx, tenantID)
	serviceIDs, err := m.GetServiceIdsByRole(ctx, cc, cmd.RoleIds)
	if err != nil {
		return nil, err
	}
	root, _ := m.IsRoot(ctx, cc, pin)
	rootTenantID, _ := m.GetRootTenantId(ctx, cc)
	isRoot := root && cmd.SubTenantId != "" && rootTenantID == cmd.SubTenantId
	list, err := s.loadServiceAreasByIDs(ctx, cmd.SubTenantId, serviceIDs, cmd.SubTenantId, isRoot)
	if err != nil {
		return nil, err
	}
	if !isRoot {
		return list, nil
	}
	ignore, _ := m.IgnoreTenant(ctx, cc)
	if len(ignore) == 0 {
		return list, nil
	}
	ignoreSet := make(map[string]struct{}, len(ignore))
	for _, t := range ignore {
		ignoreSet[t] = struct{}{}
	}
	out := make([]dto.ServiceAreaCO, 0, len(list))
	for _, co := range list {
		if _, skip := ignoreSet[co.TenantId]; skip {
			continue
		}
		out = append(out, co)
	}
	return out, nil
}

func (s *ServiceAreaAdmin) loadServiceAreasByIDs(ctx context.Context, lookupTenantID string, ids []int64, dbTenantID string, isRoot bool) ([]dto.ServiceAreaCO, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	found := make(map[int64]gateway.FenceE, len(ids))
	missing := make([]int64, 0)
	for _, id := range ids {
		fe, err := domainsvc.GetServiceAreaById(ctx, lookupTenantID, id)
		if err != nil {
			return nil, err
		}
		if fe != nil {
			found[id] = *fe
		} else {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		r, err := requireFenceRepo()
		if err != nil {
			return nil, err
		}
		tenantFilter := dbTenantID
		if isRoot {
			tenantFilter = ""
		}
		dbMap, err := r.GetByIDsAndType(ctx, missing, fence.TypeServiceArea, tenantFilter)
		if err != nil {
			return nil, err
		}
		for id, fe := range dbMap {
			found[id] = fe
		}
	}
	out := make([]dto.ServiceAreaCO, 0, len(found))
	for _, id := range ids {
		if fe, ok := found[id]; ok {
			out = append(out, toClientServiceAreaCO(fe))
		}
	}
	return out, nil
}

package fenceadmin

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"time"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/fence"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/geo"
	domainsvc "ebike-fence-go/internal/domain/service"
	"ebike-fence-go/internal/infrastructure/persistence"
	"ebike-fence-go/internal/infrastructure/persistence/convert"
	"ebike-fence-go/internal/infrastructure/persistence/repo"
)

// serviceAreaExists reports whether a service area fence exists for serviceID.
func serviceAreaExists(ctx context.Context, tenantID string, serviceID int64) (bool, error) {
	sa, err := domainsvc.GetServiceAreaById(ctx, tenantID, serviceID)
	if err != nil {
		return false, err
	}
	return sa != nil, nil
}

// javaListToString renders a string slice like Java ArrayList.toString(): "[a, b]" / "[]".
func javaListToString(items []string) string {
	return "[" + strings.Join(items, ", ") + "]"
}

type ParkingAdmin struct{}

func NewParkingAdmin() *ParkingAdmin { return &ParkingAdmin{} }

func parkingFromCmd(cmd dto.ParkingCmd) gateway.FenceE {
	fe := cmdToFenceE(fence.TypeParking, cmd.Name, cmd.ShapeType, cmd.CenterLat, cmd.CenterLng, cmd.PointList, cmd.ServiceId)
	fe.Id = cmd.Id
	fe.MaxParkingNumber = cmd.MaxParkingNumber
	fe.Tbeacon = cmd.Tbeacon
	fe.Directional = cmd.Directional
	fe.Direction = cmd.Direction
	fe.FormulateDirection = cmd.FormulateDirection
	fe.Rfid = cmd.Rfid
	fe.IzEnable = cmd.IzEnable
	if cmd.CoefficientOfDifficult != nil {
		fe.CoefficientOfDifficult = *cmd.CoefficientOfDifficult
	}
	fe.BufferDistance = 0
	if cmd.BufferDistance != nil {
		fe.BufferDistance = *cmd.BufferDistance
		fe.BufferDistanceSet = true
	}
	fe.Camera = cmd.Camera
	fe.Kickstand = cmd.Kickstand
	fe.IzFullPileNoStop = cmd.IzFullPileNoStop
	if cmd.OpeningHoursBegin != nil {
		fe.OpeningHoursBegin = *cmd.OpeningHoursBegin
	}
	if cmd.OpeningHoursEnd != nil {
		fe.OpeningHoursEnd = *cmd.OpeningHoursEnd
	}
	fe.IzOpenAllDay = cmd.IzOpenAllDay
	return fe
}

// GetByCarID mirrors Java ParkingServiceImpl.getParkingByCarId → getParkingById.
func (p *ParkingAdmin) GetByCarID(ctx context.Context, tenantID, carID string, detailGw *gateway.ParkingDetailGateway) (*dto.ParkingCO, error) {
	parkingID, err := resolveParkingIDByCar(ctx, tenantID, carID, detailGw)
	if err != nil {
		return nil, err
	}
	if parkingID == 0 {
		return nil, nil
	}
	co, err := p.GetByID(ctx, tenantID, parkingID)
	if err != nil {
		return nil, err
	}
	return &co, nil
}

func resolveParkingIDByCar(ctx context.Context, tenantID, carID string, detailGw *gateway.ParkingDetailGateway) (int64, error) {
	var detail *gateway.ParkingDetailE
	if detailGw != nil {
		d, err := detailGw.GetByCarIdFromRedis(ctx, tenantID, carID)
		if err != nil {
			return 0, err
		}
		detail = d
	}
	if detail == nil && persistence.Parking != nil {
		d, err := persistence.Parking.GetByCarID(ctx, tenantID, carID)
		if err != nil {
			return 0, err
		}
		detail = d
	}
	if detail == nil || detail.ParkingId == 0 {
		return 0, nil
	}
	return detail.ParkingId, nil
}

func (p *ParkingAdmin) GetByID(ctx context.Context, tenantID string, id int64) (dto.ParkingCO, error) {
	fe, err := gateway.GetFenceById(ctx, domainsvc.ParkingPrefix, tenantID, id)
	if err != nil {
		return dto.ParkingCO{}, err
	}
	if fe == nil {
		fe, err = getFenceOrErr(ctx, id)
		if err != nil {
			return dto.ParkingCO{}, err
		}
	}
	co := toParkingCO(*fe)
	if count, err := domainsvc.ParkingCarCount(ctx, tenantID, id); err == nil {
		setParkingCarCount(&co, count, false)
	}
	return co, nil
}

func (p *ParkingAdmin) GetByIDs(ctx context.Context, tenantID string, ids []int64) ([]dto.ParkingCO, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return nil, err
	}
	m, err := r.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ParkingCO, 0, len(m))
	for _, id := range ids {
		if fe, ok := m[id]; ok {
			co := toParkingCO(fe)
			if count, err := domainsvc.ParkingCarCount(ctx, tenantID, id); err == nil {
				setParkingCarCount(&co, count, false)
			}
			out = append(out, co)
		}
	}
	return out, nil
}

func (p *ParkingAdmin) GetListByServiceID(ctx context.Context, tenantID string, serviceID int64, withCount bool) ([]dto.ParkingCO, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return nil, err
	}
	list, err := r.ListByTypeAndServiceID(ctx, tenantID, fence.TypeParking, serviceID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ParkingCO, 0, len(list))
	for _, fe := range list {
		co := toParkingCO(fe)
		if withCount {
			if count, err := domainsvc.ParkingCarCount(ctx, tenantID, fe.Id); err == nil {
				setParkingCarCount(&co, count, false)
			}
		}
		out = append(out, co)
	}
	return out, nil
}

func (p *ParkingAdmin) GetPageListByServiceID(ctx context.Context, tenantID string, serviceID int64, name string, areaSize *float64, fieldList []string, pageNum, pageSize int) (dto.PageDTO[dto.ParkingCO], error) {
	r, err := requireFenceRepo()
	if err != nil {
		return dto.PageDTO[dto.ParkingCO]{}, err
	}
	list, total, err := r.PageByTypeAndServiceID(ctx, tenantID, fence.TypeParking, serviceID, name, areaSize, fieldList, pageNum, pageSize)
	if err != nil {
		return dto.PageDTO[dto.ParkingCO]{}, err
	}
	out := make([]dto.ParkingCO, 0, len(list))
	for _, fe := range list {
		co := toParkingCO(fe)
		if count, err := domainsvc.ParkingCarCount(ctx, tenantID, fe.Id); err == nil {
			setParkingCarCount(&co, count, false)
		}
		out = append(out, co)
	}
	return dto.PageDTO[dto.ParkingCO]{List: out, Total: total, PageNum: pageNum, PageSize: pageSize}, nil
}

func (p *ParkingAdmin) Create(ctx context.Context, tenantID, pin string, cmd dto.ParkingCmd) (int64, error) {
	fe := parkingFromCmd(cmd)
	sid := fe.ServiceId
	// Java ParkingServiceImpl.createParking: reject when the service area does not exist.
	if exists, err := serviceAreaExists(ctx, tenantID, sid); err != nil {
		return 0, err
	} else if !exists {
		return 0, newBizError("13019", "服务区不存在")
	}
	return saveFence(ctx, tenantID, pin, &fe, writeOpts{kind: gateway.CacheParking, serviceID: &sid})
}

func (p *ParkingAdmin) Update(ctx context.Context, tenantID, pin string, cmd dto.ParkingCmd) error {
	if cmd.Id == 0 {
		return newBizError("00004", "id must not be null")
	}
	fe := parkingFromCmd(cmd)
	sid := fe.ServiceId
	// Java ParkingServiceImpl.updateParking: reject when the service area does not exist.
	if exists, err := serviceAreaExists(ctx, tenantID, sid); err != nil {
		return err
	} else if !exists {
		return newBizError("13019", "服务区不存在")
	}
	return updateFence(ctx, tenantID, pin, &fe, writeOpts{kind: gateway.CacheParking, serviceID: &sid})
}

func (p *ParkingAdmin) Delete(ctx context.Context, tenantID string, id int64) error {
	if err := deleteFence(ctx, tenantID, id, writeOpts{kind: gateway.CacheParking}); err != nil {
		return err
	}
	// Java ParkingServiceImpl.deleteParking: clear t_parking.parking_id for this station.
	return removeParkingBinding(ctx, tenantID, id, fence.TypeParking)
}

// DeleteBatch mirrors Java ParkingGatewayImpl.deleteByIds: unlike noParking/banRiding,
// the parking batch delete DOES clear t_parking.parking_id for each station (per-id
// parkingDetailGateway.removeParkingId(id, FORPARK)).
func (p *ParkingAdmin) DeleteBatch(ctx context.Context, tenantID string, ids []int64) error {
	if err := deleteFences(ctx, tenantID, ids, writeOpts{kind: gateway.CacheParking}); err != nil {
		return err
	}
	for _, id := range ids {
		if err := removeParkingBinding(ctx, tenantID, id, fence.TypeParking); err != nil {
			return err
		}
	}
	return nil
}

func (p *ParkingAdmin) SetEnable(ctx context.Context, tenantID, pin string, id int64, enable bool) error {
	fe, err := getFenceOrErr(ctx, id)
	if err != nil {
		return err
	}
	fe.IzEnable = enable
	sid := fe.ServiceId
	return updateFence(ctx, tenantID, pin, fe, writeOpts{kind: gateway.CacheParking, serviceID: &sid})
}

func (p *ParkingAdmin) SetEnableBatch(ctx context.Context, tenantID, pin string, ids []int64, enable bool) error {
	r, err := requireFenceRepo()
	if err != nil {
		return err
	}
	if err := r.UpdateByIDs(ctx, ids, map[string]interface{}{"iz_enable": enable}); err != nil {
		return err
	}
	for _, id := range ids {
		fe, _ := r.GetByID(ctx, id)
		if fe != nil {
			fe.IzEnable = enable
			sid := fe.ServiceId
			_ = updateFence(ctx, tenantID, pin, fe, writeOpts{kind: gateway.CacheParking, serviceID: &sid})
		}
	}
	return nil
}

func (p *ParkingAdmin) GetNearParking(ctx context.Context, tenantID string, cmd dto.NearLocationCmd) ([]dto.ParkingCO, error) {
	radius := 5000.0
	if cmd.Radius != nil {
		radius = *cmd.Radius
	}
	fences, err := gateway.GetFencesByLocation(ctx, domainsvc.ParkingPrefix, domainsvc.ParkingGeo+"_"+tenantID, tenantID, cmd.Location.Lng, cmd.Location.Lat, radius, 20, cmd.ServiceId)
	if err != nil {
		return nil, err
	}
	backcar, _ := gateway.NewConfigGateway().GetConfigByServiceId(ctx, tenantID, cmd.ServiceId)
	out := make([]dto.ParkingCO, 0)
	for _, fe := range fences {
		if !includeNearParking(fe, cmd.ServiceId, backcar) {
			continue
		}
		co := toNearParkingCO(fe)
		co.Distance = geo.PointToPointDistanceJava(cmd.Location.Lng, cmd.Location.Lat, fe.CenterLng, fe.CenterLat)
		if count, err := domainsvc.ParkingCarCount(ctx, tenantID, fe.Id); err == nil {
			setParkingCarCount(&co, count, true)
		} else {
			setParkingCarCount(&co, 0, true)
		}
		carCount := int64(0)
		if co.CarCount != nil {
			carCount = *co.CarCount
		}
		fullCar := parkingFullCar(fe, carCount, backcar)
		co.FullCar = &fullCar
		out = append(out, co)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Distance != out[j].Distance {
			return out[i].Distance < out[j].Distance
		}
		return out[i].Id < out[j].Id
	})
	return out, nil
}

func (p *ParkingAdmin) GetByLocations(ctx context.Context, tenantID string, cmd dto.LocationsCmd) ([]dto.ParkingCO, error) {
	var out []dto.ParkingCO
	for _, loc := range cmd.Locations {
		near, err := p.GetNearParking(ctx, tenantID, dto.NearLocationCmd{
			ServiceId: cmd.ServiceId,
			Location:  loc,
		})
		if err != nil {
			return nil, err
		}
		if len(near) > 0 {
			out = append(out, near[0])
		} else {
			out = append(out, dto.ParkingCO{})
		}
	}
	return out, nil
}

func (p *ParkingAdmin) ParkingCarCount(ctx context.Context, tenantID string, parkingID int64) (int64, error) {
	return domainsvc.ParkingCarCount(ctx, tenantID, parkingID)
}

func (p *ParkingAdmin) GetListByCmd(ctx context.Context, tenantID string, cmd dto.ParkingCmd) ([]dto.ParkingCO, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return nil, err
	}
	list, err := r.ListByTypeAndServiceID(ctx, tenantID, fence.TypeParking, cmd.ServiceId)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ParkingCO, 0)
	for _, fe := range list {
		if cmd.Name != "" && fe.Name != cmd.Name {
			continue
		}
		co := toParkingCO(fe)
		if count, err := domainsvc.ParkingCarCount(ctx, tenantID, fe.Id); err == nil {
			setParkingCarCount(&co, count, false)
		}
		out = append(out, co)
	}
	return out, nil
}

// Copy mirrors Java ParkingServiceImpl.copy(IdCmd): the incoming id is a
// SERVICE-AREA id. It duplicates the parking station fence(s) belonging to that
// service area (keeping the original name) and returns no data.
//
// NOTE: the original Go behavior treated the id as a single station fence id and
// copied just that one record with a "_copy" suffix. Preserved here for reference:
//
//	fe, err := getFenceOrErr(ctx, sourceID)
//	if err != nil {
//		return 0, err
//	}
//	fe.Id = 0
//	fe.Name = fe.Name + "_copy"
//	return saveFence(ctx, tenantID, pin, fe, writeOpts{kind: gateway.CacheParking, serviceID: &fe.ServiceId})
func (p *ParkingAdmin) Copy(ctx context.Context, tenantID, pin string, serviceID int64) error {
	r, err := requireFenceRepo()
	if err != nil {
		return err
	}
	// Java getListByBikeTypeAndServiceId(FORPARK, serviceId, 1): the bikeType filter
	// is commented out in Java, so it is equivalent to listing by type + serviceId.
	list, err := r.ListByTypeAndServiceID(ctx, tenantID, fence.TypeParking, serviceID)
	if err != nil {
		return err
	}
	if len(list) > 1 {
		return newBizError("13016", "失败！请确保单车无站点围栏")
	}
	if len(list) < 1 {
		return newBizError("13016", "失败！请确保电单车有站点围栏")
	}
	for i := range list {
		copyFe := list[i]
		copyFe.Id = 0
		sid := copyFe.ServiceId
		if _, err := saveFenceNoDedupe(ctx, tenantID, pin, &copyFe, writeOpts{kind: gateway.CacheParking, serviceID: &sid}); err != nil {
			return err
		}
	}
	return nil
}

func (p *ParkingAdmin) CopyByService(ctx context.Context, tenantID, pin string, cmd dto.CopyByServiceCmd) (bool, error) {
	var sources []gateway.FenceE
	r, err := requireFenceRepo()
	if err != nil {
		return false, err
	}
	if len(cmd.OriginalParkingIds) > 0 {
		m, err := r.GetByIDs(ctx, cmd.OriginalParkingIds)
		if err != nil {
			return false, err
		}
		for _, fe := range m {
			sources = append(sources, fe)
		}
	} else {
		sources, err = r.ListByTypeAndServiceID(ctx, tenantID, fence.TypeParking, cmd.OriginalServiceId)
		if err != nil {
			return false, err
		}
	}
	// Java ParkingServiceImpl.copyByServiceId: reject when the source has no parking fences.
	if len(sources) < 1 {
		return false, newBizError("13016", "失败！请确保电单车有站点围栏")
	}
	for _, src := range sources {
		copyFe := src
		copyFe.Id = 0
		copyFe.ServiceId = cmd.TargetServiceId
		sid := copyFe.ServiceId
		if _, err := saveFence(ctx, tenantID, pin, &copyFe, writeOpts{kind: gateway.CacheParking, serviceID: &sid}); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (p *ParkingAdmin) InParkInfoByLocation(ctx context.Context, tenantID string, cmd dto.NearLocationCmd) (*dto.ParkingCO, error) {
	loc := geo.Location{Lat: cmd.Location.Lat, Lng: cmd.Location.Lng}
	backcarConfig, _ := gateway.NewConfigGateway().GetConfigByServiceId(ctx, tenantID, cmd.ServiceId)
	fe, err := domainsvc.ParkingReturnCar(ctx, tenantID, loc, loc, cmd.ServiceId, backcarConfig)
	if err != nil || fe == nil {
		return nil, err
	}
	// Java returns null when the matched parking is disabled.
	if !fe.IzEnable {
		return nil, nil
	}

	// Mirror Java getInParkInfoByLocation: resolve bufferDistance onto the entity
	// before mapping to CO (site value → backcar config → 5).
	bufferDistance := fe.BufferDistance
	if bufferDistance <= 0 && backcarConfig != nil && backcarConfig.BufferDistance != nil {
		bufferDistance = *backcarConfig.BufferDistance
	}
	if bufferDistance <= 0 {
		bufferDistance = 5
	}

	co := toParkingCO(*fe)
	co.BufferDistance = &bufferDistance
	// Redis/location query path does not surface DB audit fields (Java ConvertorHelper nulls).
	stripFenceAudit(&co.FenceCO)
	// Java ParkingCO uses nullable wrappers for these — keep them null on this endpoint.
	co.CarCount = nil
	co.FullCar = nil
	co.RefTags = nil

	// Mirror Java inParkInfoByLocation: expand pointList by buffer polygon
	// (createPolygonBufferArea with coefficientOfDifficult).
	if fe.PointList != "" && bufferDistance >= 0 {
		coef := 0.0
		if fe.CoefficientOfDifficultSet {
			coef = fe.CoefficientOfDifficult
		}
		if buffered, err := geo.CreatePolygonBufferArea(fe.PointList, bufferDistance, coef); err == nil && buffered != "" {
			co.PointList = buffered
		}
	}
	return &co, nil
}

func (p *ParkingAdmin) GetNearParkingNum(ctx context.Context, tenantID string, cmd dto.NearLocationCmd) (dto.NearParkingCO, error) {
	const defaultRadius = 500.0
	loc := geo.Location{Lat: cmd.Location.Lat, Lng: cmd.Location.Lng}
	serviceArea, err := domainsvc.GetServiceAreaById(ctx, tenantID, cmd.ServiceId)
	if err != nil {
		return dto.NearParkingCO{}, err
	}
	izInService := false
	if serviceArea != nil {
		izInService, _ = geo.IsPointInPolygon(loc, serviceArea.PointList)
	}
	if !izInService {
		return dto.NearParkingCO{IzOutService: 1, ParkingNum: 0, Distance: defaultRadius}, nil
	}
	list, err := p.GetNearParking(ctx, tenantID, cmd)
	if err != nil {
		return dto.NearParkingCO{}, err
	}
	if len(list) == 0 {
		return dto.NearParkingCO{IzOutService: 0, ParkingNum: 0, Distance: defaultRadius}, nil
	}
	dist := geo.PointToPointDistanceJava(cmd.Location.Lng, cmd.Location.Lat, list[0].CenterLng, list[0].CenterLat)
	return dto.NearParkingCO{IzOutService: 0, ParkingNum: len(list), Distance: dist}, nil
}

func (p *ParkingAdmin) ParkingCarCountByLocation(ctx context.Context, tenantID string, cmd dto.NearLocationCmd) (*dto.ParkingCarCount, error) {
	co, err := p.InParkInfoByLocation(ctx, tenantID, cmd)
	if err != nil {
		return nil, err
	}
	// Java parkingCarCountByLocation calls getInParkInfoByLocation which throws
	// PARKING_NOT_FOUND (13015) when no station matches; inParkInfoByLocation
	// itself catches and returns null, so only this path must surface 13015.
	if co == nil {
		return nil, newBizError("13015", "parking not found")
	}
	count, _ := domainsvc.ParkingCarCount(ctx, tenantID, co.Id)
	return &dto.ParkingCarCount{ParkingId: co.Id, CarCount: count, MaxParkingNumber: co.MaxParkingNumber}, nil
}

func (p *ParkingAdmin) Page(ctx context.Context, tenantID string, q dto.ParkingPageQuery) (dto.PageDTO[dto.ParkingCO], error) {
	return p.GetPageListByServiceID(ctx, tenantID, q.ServiceId, q.Name, nil, nil, q.PageNum, q.PageSize)
}

// CreateBatch mirrors Java ParkingServiceImpl.createParkingBatch: it validates the
// service area exists, then saves each station with saveNoRepetition (duplicate name
// -> FENCE_EXIST). Per-item failures are collected and the method returns success
// with the error list rendered as a string (Java errorList.toString()), e.g. "[]".
//
// NOTE: the original Go behavior was fail-fast + auto-rename (via Create/saveFence).
// Preserved here for reference:
//
//	for _, pc := range cmd.ParkingCmdList {
//		pc.ServiceId = cmd.ServiceId
//		if _, err := p.Create(ctx, tenantID, pin, pc); err != nil {
//			return err
//		}
//	}
//	return nil
func (p *ParkingAdmin) CreateBatch(ctx context.Context, tenantID, pin string, cmd dto.CreateParkingBatchCmd) (string, error) {
	// serviceArea existence check (Java throws SERVICE_AREA_NOT_EXISTS / 13019)
	if exists, err := serviceAreaExists(ctx, tenantID, cmd.ServiceId); err != nil {
		return "", err
	} else if !exists {
		return "", newBizError("13019", "服务区不存在")
	}
	errorList := make([]string, 0)
	for _, pc := range cmd.ParkingCmdList {
		pc.ServiceId = cmd.ServiceId
		if _, err := p.saveNoRepetition(ctx, tenantID, pin, pc); err != nil {
			msg := err.Error()
			if biz, ok := err.(*bizError); ok {
				msg = biz.msg
			}
			errorList = append(errorList, pc.Name+" error msg:"+msg+"\n")
		}
	}
	return javaListToString(errorList), nil
}

// saveNoRepetition mirrors Java ParkingGatewayImpl.saveNoRepetition: reject duplicate
// names (FENCE_EXIST / 13040) instead of auto-renaming, then plain save + cache.
func (p *ParkingAdmin) saveNoRepetition(ctx context.Context, tenantID, pin string, cmd dto.ParkingCmd) (int64, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return 0, err
	}
	fe := parkingFromCmd(cmd)
	sid := fe.ServiceId
	exists, err := r.ExistsByName(ctx, tenantID, fe.Name, fence.TypeParking, &sid)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, newBizError("13040", "站点名称重复")
	}
	return saveFenceNoDedupe(ctx, tenantID, pin, &fe, writeOpts{kind: gateway.CacheParking, serviceID: &sid})
}

// UpdateBatch mirrors Java ParkingServiceImpl.updateParkingBatch: only the 9 attribute
// columns below are updated (via updateByIds, which skips null fields). It must NOT
// touch name/geometry/pointList/serviceId.
//
// NOTE: the original Go behavior re-ran the full per-record Update (overwriting all
// fields incl. name/geometry). Preserved here for reference:
//
//	for _, id := range cmd.Ids {
//		c := cmd.ParkingCmd
//		c.Id = id
//		if err := p.Update(ctx, tenantID, pin, c); err != nil {
//			return err
//		}
//	}
//	return nil
func (p *ParkingAdmin) UpdateBatch(ctx context.Context, tenantID, pin string, cmd dto.UpdateParkingBatchCMD) error {
	r, err := requireFenceRepo()
	if err != nil {
		return err
	}
	if len(cmd.Ids) == 0 {
		return nil
	}
	// max_parking_number / iz_enable are value types in the Go DTO so nullability
	// cannot be detected; they are always written (matches the common batch call).
	updates := map[string]interface{}{
		"max_parking_number": cmd.MaxParkingNumber,
		"iz_enable":          cmd.IzEnable,
	}
	if cmd.CoefficientOfDifficult != nil {
		updates["coefficient_of_difficult"] = *cmd.CoefficientOfDifficult
	}
	if cmd.BufferDistance != nil {
		updates["buffer_distance"] = *cmd.BufferDistance
	}
	if cmd.Tbeacon != nil {
		updates["tbeacon"] = *cmd.Tbeacon
	}
	if cmd.Directional != nil {
		updates["directional"] = *cmd.Directional
	}
	if cmd.Kickstand != nil {
		updates["kickstand"] = *cmd.Kickstand
	}
	if cmd.Rfid != nil {
		updates["rfid"] = *cmd.Rfid
	}
	if cmd.Camera != nil {
		updates["camera"] = *cmd.Camera
	}
	if err := r.UpdateByIDs(ctx, cmd.Ids, updates); err != nil {
		return err
	}
	for _, id := range cmd.Ids {
		_ = gateway.InvalidateFenceCache(ctx, gateway.CacheParking, tenantID, id)
	}
	return nil
}

func (p *ParkingAdmin) FilterParkingList(ctx context.Context, tenantID string, q dto.ParkingStationFilterCmd) (dto.JavaPageDTO[dto.ParkingCO], error) {
	pageNum := q.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	r, err := requireFenceRepo()
	if err != nil {
		return dto.JavaPageDTO[dto.ParkingCO]{}, err
	}
	base, err := r.ListByTypeAndServiceID(ctx, tenantID, fence.TypeParking, q.Id)
	if err != nil {
		return dto.JavaPageDTO[dto.ParkingCO]{}, err
	}
	filledIDs := make([]int64, 0, len(base))
	for _, fe := range base {
		filledIDs = append(filledIDs, fe.Id)
	}
	if len(filledIDs) == 0 {
		return dto.JavaPageDTO[dto.ParkingCO]{
			PageNum:     pageNum,
			PageSize:    pageSize,
			Orders:      nil,
			SearchCount: true,
			List:        []dto.ParkingCO{},
		}, nil
	}
	if len(q.Tags) > 0 {
		if persistence.Config == nil {
			filledIDs = nil
		} else {
			filledIDs, err = persistence.Config.FilterFenceIDsByTags(ctx, q.Tags, filledIDs, fence.TypeParking)
			if err != nil {
				return dto.JavaPageDTO[dto.ParkingCO]{}, err
			}
		}
	}
	if len(filledIDs) == 0 {
		return dto.JavaPageDTO[dto.ParkingCO]{
			PageNum:     pageNum,
			PageSize:    pageSize,
			Orders:      nil,
			SearchCount: true,
			List:        []dto.ParkingCO{},
		}, nil
	}
	fieldList := q.FieldList
	if fieldList == nil {
		fieldList = []string{}
	}
	list, total, err := r.PageParkingByFilter(ctx, repo.ParkingPageFilter{
		TenantID:  tenantID,
		ServiceID: q.Id,
		FenceIDs:  filledIDs,
		Name:      q.Name,
		AreaSize:  q.AreaSize,
		FieldList: fieldList,
		IzEnable:  q.IzEnable,
		PageNum:   pageNum,
		PageSize:  pageSize,
	})
	if err != nil {
		return dto.JavaPageDTO[dto.ParkingCO]{}, err
	}
	var carCountMap map[int64]int64
	if persistence.Parking != nil {
		carCountMap, _ = persistence.Parking.CountGroupedByParkingIDByService(ctx, q.Id)
	}
	out := make([]dto.ParkingCO, 0, len(list))
	for _, fe := range list {
		co := toParkingCO(fe)
		if carCountMap != nil {
			if count, ok := carCountMap[fe.Id]; ok {
				setParkingCarCount(&co, count, false)
			}
		}
		out = append(out, co)
	}
	return dto.JavaPageDTO[dto.ParkingCO]{
		Count:       total,
		PageNum:     pageNum,
		PageSize:    pageSize,
		Orders:      nil,
		SearchCount: true,
		List:        out,
	}, nil
}

func (p *ParkingAdmin) GetPageOrderBy(ctx context.Context, tenantID string, q dto.ParkingPageOrderByQry) (dto.PageDTO[dto.ParkingCO], error) {
	return p.GetPageListByServiceID(ctx, tenantID, q.ServiceId, "", nil, nil, q.PageNum, q.PageSize)
}

func (p *ParkingAdmin) GetParkingMonitorPage(ctx context.Context, tenantID string, q dto.ParkingMonitorPageQry) (dto.PageDTO[dto.ParkingMonitorCO], error) {
	page, err := p.GetPageListByServiceID(ctx, tenantID, q.ServiceId, "", nil, nil, q.PageNum, q.PageSize)
	if err != nil {
		return dto.PageDTO[dto.ParkingMonitorCO]{}, err
	}
	out := make([]dto.ParkingMonitorCO, 0, len(page.List))
	for _, pco := range page.List {
		out = append(out, dto.ParkingMonitorCO{ParkingCO: pco})
	}
	return dto.PageDTO[dto.ParkingMonitorCO]{List: out, Total: page.Total, PageNum: page.PageNum, PageSize: page.PageSize}, nil
}

func (p *ParkingAdmin) GetParkingMonitorList(ctx context.Context, tenantID string, q dto.ParkingMonitorPageQry) ([]dto.ParkingMonitorCO, error) {
	page, err := p.GetParkingMonitorPage(ctx, tenantID, q)
	if err != nil {
		return nil, err
	}
	return page.List, nil
}

func (p *ParkingAdmin) GetParkingMonitorDetail(ctx context.Context, tenantID string, id int64) (dto.ParkingMonitorDetailCO, error) {
	co, err := p.GetByID(ctx, tenantID, id)
	if err != nil {
		return dto.ParkingMonitorDetailCO{}, err
	}
	carCount := int64(0)
	if co.CarCount != nil {
		carCount = *co.CarCount
	}
	return dto.ParkingMonitorDetailCO{ParkingCO: co, CarCount: carCount}, nil
}

func (p *ParkingAdmin) CreateRadPacketParking(ctx context.Context, tenantID, pin string, cmd dto.CreateRadPacketParkingCmd) (bool, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return false, err
	}
	for _, item := range cmd.List {
		var exp sql.NullTime
		if item.ExpirationTime != "" {
			if t, parseErr := time.Parse("2006-01-02 15:04:05", item.ExpirationTime); parseErr == nil {
				exp = sql.NullTime{Time: t, Valid: true}
			} else if t, parseErr := time.Parse(time.RFC3339, item.ExpirationTime); parseErr == nil {
				exp = sql.NullTime{Time: t, Valid: true}
			}
		}
		if err := r.UpdateRadPacketParking(ctx, tenantID, item.Id, cmd.MinAmount, cmd.MaxAmount, cmd.ActivityId, exp); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (p *ParkingAdmin) RefTags(ctx context.Context, tenantID string, fenceID int64) ([]dto.FenceTagCO, error) {
	if persistence.Config == nil {
		return nil, nil
	}
	rows, err := persistence.Config.ListFenceRefTagsByFenceID(ctx, fenceID, fence.TypeParking)
	if err != nil {
		return nil, err
	}
	out := make([]dto.FenceTagCO, 0, len(rows))
	for _, row := range rows {
		out = append(out, convert.FenceTagToCO(row))
	}
	return out, nil
}

func (p *ParkingAdmin) StationCars(ctx context.Context, tenantID string, q dto.ParkStationCarsPageQuery) (dto.PageDTO[dto.ParkingDetailCO], error) {
	if persistence.Parking == nil {
		return dto.PageDTO[dto.ParkingDetailCO]{}, nil
	}
	stationID := q.ParkingId
	if q.Id != nil && *q.Id > 0 {
		stationID = *q.Id
	}
	page, err := persistence.Parking.PageByParkingID(ctx, tenantID, stationID, q.PageNum, q.PageSize)
	if err != nil {
		return dto.PageDTO[dto.ParkingDetailCO]{}, err
	}
	out := make([]dto.ParkingDetailCO, 0, len(page.List))
	for _, row := range page.List {
		out = append(out, dto.ParkingDetailCO{CarId: row.CarId, ParkingId: row.ParkingId})
	}
	return dto.PageDTO[dto.ParkingDetailCO]{
		List:     out,
		Total:    page.Total,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
	}, nil
}

func (p *ParkingAdmin) ParkingDetailByMaintainAreaId(ctx context.Context, tenantID string, maintainAreaID int64) ([]dto.ParkingDetailCO, error) {
	if persistence.Parking == nil {
		return nil, nil
	}
	rows, err := persistence.Parking.ListByMaintainAreaID(ctx, tenantID, maintainAreaID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ParkingDetailCO, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.ParkingDetailCO{CarId: row.CarId, ParkingId: row.ParkingId})
	}
	return out, nil
}

func (p *ParkingAdmin) BindParking(ctx context.Context, tenantID string, cmd dto.ParkingBindCarCmd) (*dto.BindParkingCO, error) {
	return domainsvc.BindParking(ctx, tenantID, cmd)
}

func (p *ParkingAdmin) UnBindParking(ctx context.Context, tenantID string, cmd dto.ParkingUnBindCarCmd) (bool, error) {
	return domainsvc.UnBindParking(ctx, tenantID, cmd)
}

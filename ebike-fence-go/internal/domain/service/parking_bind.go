package service

import (
	"context"
	"log"
	"sync"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/config"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/geo"
	"ebike-fence-go/internal/domain/rediskeys"
	"ebike-fence-go/internal/infrastructure/persistence"
	"ebike-fence-go/internal/infrastructure/rpc"
)

const codeCarNotFoundBind = "13013"

// BindParking mirrors Java ParkingServiceImpl.bindParking.
func BindParking(ctx context.Context, tenantID string, cmd dto.ParkingBindCarCmd) (*dto.BindParkingCO, error) {
	if persistence.Parking == nil {
		return nil, newBizError("00004", "mysql not initialized")
	}

	mgmt := rpc.NewManagementRPC()
	car, err := mgmt.GetCarInfoByImei(ctx, cmd.Imei, dto.EnsureCommandContext(cmd.CommandContext, tenantID))
	if err != nil {
		return nil, err
	}
	if car == nil || car.ServiceId == 0 {
		return nil, newBizError(codeCarNotFoundBind, "车辆未找到")
	}

	cfgGw := gateway.NewConfigGateway()
	backcarConfig, err := cfgGw.GetConfigByServiceId(ctx, tenantID, car.ServiceId)
	if err != nil {
		return nil, err
	}
	if backcarConfig == nil {
		defaultBuffer := 10.0
		backcarConfig = &config.ConfigBackcarCO{ServiceId: &car.ServiceId, BufferDistance: &defaultBuffer}
	}

	loc := geo.Location{Lat: *cmd.Lat, Lng: *cmd.Lng}
	detail, err := loadParkingDetail(ctx, tenantID, car.CarId)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		detail = &gateway.ParkingDetailE{}
	}

	// Parallel fence matching:
	// Group A (main goroutine): parking → noParking → banRiding (sequential, has dependencies)
	// Group B (background):     customFence + maintainArea (independent, no cross-dependencies)
	var fenceCustom *gateway.FenceE
	var maintainArea *gateway.FenceE
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		fenceCustom = findCustomFenceAtLocation(ctx, tenantID, loc, car.ServiceId)
	}()
	go func() {
		defer wg.Done()
		maintainArea = findMaintainAreaAtLocation(ctx, tenantID, loc, car.ServiceId)
	}()

	izInParkingBind(ctx, tenantID, loc, car.ServiceId, backcarConfig, detail)

	var noParking *gateway.FenceE
	var banRiding *gateway.FenceE
	if detail.ParkingId == 0 {
		noParking, _ = izInNoParking(ctx, tenantID, loc, car.ServiceId, backcarConfig)
		if noParking == nil {
			banRiding, _ = izInBanRiding(ctx, tenantID, loc)
		}
	}

	wg.Wait()

	oldFenceCustomID := detail.FenceCustomId
	syncCarTagOnCustomFenceChange(ctx, cmd.CommandContext, tenantID, car.CarId, car.ServiceId, oldFenceCustomID, fenceCustom)

	returnParkingID := detail.ParkingId
	if detail.ServiceId != 0 && detail.ServiceId != car.ServiceId {
		detail.ParkingId = 0
	}
	detail.Imei = cmd.Imei
	detail.ServiceId = car.ServiceId
	detail.CarId = car.CarId
	if noParking != nil {
		detail.NoParkingId = noParking.Id
	} else {
		detail.NoParkingId = 0
	}
	if banRiding != nil {
		detail.BanRidingId = banRiding.Id
	} else {
		detail.BanRidingId = 0
	}
	if maintainArea != nil {
		detail.MaintainAreaId = maintainArea.Id
	} else {
		detail.MaintainAreaId = 0
	}
	if fenceCustom != nil {
		detail.FenceCustomId = fenceCustom.Id
	} else {
		detail.FenceCustomId = 0
	}

	nearParkingID := detail.NearParkingId

	// Async persistence: MySQL write + Redis detail cache run in the background
	// so the HTTP response is not blocked by DB I/O. The detail struct is copied
	// to avoid data races with the caller.
	detailCopy := *detail

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[ERROR] async bindParking panic recovered: %v", r)
			}
		}()
		bgCtx := context.Background()
		if err := persistence.Parking.SaveOrUpdateV1(bgCtx, tenantID, "", &detailCopy); err != nil {
			log.Printf("[WARN] async bindParking SaveOrUpdateV1 failed: tenant=%s car=%s err=%v", tenantID, detailCopy.CarId, err)
		}
		_ = gateway.SetDetailCache(bgCtx, tenantID, &detailCopy)
	}()

	return &dto.BindParkingCO{
		ParkingId:      zeroToNil(returnParkingID),
		NoParkingId:    fenceIDPtr(noParking),
		MaintainAreaId: fenceIDPtr(maintainArea),
		NearParkingId:  zeroToNil(nearParkingID),
		FenceCustomId:  fenceIDPtr(fenceCustom),
	}, nil
}

// UnBindParking mirrors Java ParkingServiceImpl.unBindParking.
func UnBindParking(ctx context.Context, tenantID string, cmd dto.ParkingUnBindCarCmd) (bool, error) {
	if persistence.Parking == nil {
		return false, newBizError("00004", "mysql not initialized")
	}

	// Java: if carId provided, use directly; otherwise resolve from imei.
	carID := ""
	if cmd.CarId != nil && *cmd.CarId != "" {
		carID = *cmd.CarId
	} else {
		mgmt := rpc.NewManagementRPC()
		car, err := mgmt.GetCarInfoByImei(ctx, cmd.Imei, dto.EnsureCommandContext(cmd.CommandContext, tenantID))
		if err != nil {
			return false, err
		}
		if car == nil || car.ServiceId == 0 {
			return false, newBizError(codeCarNotFoundBind, "车辆未找到")
		}
		carID = car.CarId
	}

	parkingID := int64(0)
	if cmd.ParkingId != nil {
		parkingID = *cmd.ParkingId
	}
	return persistence.Parking.UnBindParking(ctx, tenantID, carID, parkingID)
}

func loadParkingDetail(ctx context.Context, tenantID, carID string) (*gateway.ParkingDetailE, error) {
	gw := gateway.NewParkingDetailGateway()
	if cached, err := gw.GetByCarIdFromRedis(ctx, tenantID, carID); err != nil {
		return nil, err
	} else if cached != nil {
		return cached, nil
	}
	return persistence.Parking.GetByCarID(ctx, tenantID, carID)
}

func izInParkingBind(ctx context.Context, tenantID string, loc geo.Location, serviceID int64, backcarConfig *config.ConfigBackcarCO, detail *gateway.ParkingDetailE) {
	detail.NearParkingId = 0
	detail.ParkingId = 0
	if loc.Lng == 0 && loc.Lat == 0 {
		return
	}
	fences, err := gateway.GetFencesByLocation(ctx, ParkingPrefix, ParkingGeo+"_"+tenantID, tenantID, loc.Lng, loc.Lat, 1000, 20, 0)
	if err != nil {
		return
	}
	useOther := backcarConfig != nil && backcarConfig.GetIzUseOtherParking()
	for _, fenceItem := range fences {
		if !useOther && fenceItem.ServiceId != serviceID {
			continue
		}
		if !fenceItem.IzEnable || !isInOpeningHours(&fenceItem) {
			continue
		}
		bufferDistance := fenceItem.BufferDistance
		if bufferDistance <= 0 && backcarConfig != nil && backcarConfig.BufferDistance != nil {
			bufferDistance = *backcarConfig.BufferDistance
		}
		if bufferDistance <= 0 {
			bufferDistance = 5
		}
		if geo.Intersect(loc, fenceItem.ParsedPolygon, bufferDistance) {
			detail.ParkingId = fenceItem.Id
			detail.ServiceId = fenceItem.ServiceId
			break
		}
	}
	if detail.ParkingId != 0 {
		return
	}
	for _, fenceItem := range fences {
		if fenceItem.ServiceId != serviceID {
			continue
		}
		if !fenceItem.IzEnable || !isInOpeningHours(&fenceItem) {
			continue
		}
		bufferDistance := fenceItem.BufferDistance
		if bufferDistance <= 0 && backcarConfig != nil && backcarConfig.BufferDistance != nil {
			bufferDistance = *backcarConfig.BufferDistance
		}
		if bufferDistance <= 0 {
			bufferDistance = 5
		}
		bufferDistance += 10
		if geo.Intersect(loc, fenceItem.ParsedPolygon, bufferDistance) {
			detail.NearParkingId = fenceItem.Id
			break
		}
	}
}

func findMaintainAreaAtLocation(ctx context.Context, tenantID string, loc geo.Location, serviceID int64) *gateway.FenceE {
	fences, err := gateway.GetFencesByLocation(ctx, MaintainAreaPrefix, MaintainAreaGeo+"_"+tenantID, tenantID, loc.Lng, loc.Lat, 20000, 20, 0)
	if err != nil {
		return nil
	}
	for _, fe := range fences {
		if fe.ServiceId != serviceID {
			continue
		}
		if geo.IsPointInParsedPolygon(loc, fe.ParsedPolygon) {
			f := fe
			return &f
		}
	}
	return nil
}

func findCustomFenceAtLocation(ctx context.Context, tenantID string, loc geo.Location, serviceID int64) *gateway.FenceE {
	fences, err := gateway.GetFencesByLocation(ctx, FenceCustomPrefix, rediskeys.FenceCustomGeoSearchKey(tenantID), tenantID, loc.Lng, loc.Lat, 20000, 20, 0)
	if err != nil {
		return nil
	}
	for _, fe := range fences {
		if fe.ServiceId != serviceID {
			continue
		}
		if geo.IsPointInParsedPolygon(loc, fe.ParsedPolygon) {
			f := fe
			return &f
		}
	}
	return nil
}

func zeroToNil(id int64) *int64 {
	if id == 0 {
		return nil
	}
	v := id
	return &v
}

func fenceIDPtr(fe *gateway.FenceE) *int64 {
	if fe == nil {
		return nil
	}
	v := fe.Id
	return &v
}

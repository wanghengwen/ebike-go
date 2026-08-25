package service

import (
	"context"
	"strings"
	"time"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/config"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/geo"
	"ebike-fence-go/internal/infrastructure/rpc"
)

type FenceRelation string

const (
	IN_PARKING_LAT                FenceRelation = "IN_PARKING_LAT"
	IN_PARKING_PART               FenceRelation = "IN_PARKING_PART"
	OUT_PARKING_LAT               FenceRelation = "OUT_PARKING_LAT"
	IN_PARKING_DIRECTION          FenceRelation = "IN_PARKING_DIRECTION"
	IN_PARKING_RFID               FenceRelation = "IN_PARKING_RFID"
	IN_PARKING_KICK               FenceRelation = "IN_PARKING_KICK"
	IN_PARKING_CAMERA_DIRECTIONAL FenceRelation = "IN_PARKING_CAMERA_DIRECTIONAL"
	IN_PARKING_CAMERA_POINT       FenceRelation = "IN_PARKING_CAMERA_POINT"
	IN_PARKING_CAMERA             FenceRelation = "IN_PARKING_CAMERA"
	IN_PARKING_HELMET             FenceRelation = "IN_PARKING_HELMET"
	OUT_PARKING_HELMET            FenceRelation = "OUT_PARKING_HELMET"
	IN_NO_PARKING                 FenceRelation = "IN_NO_PARKING"
	IN_NO_PARKING_HELMET          FenceRelation = "IN_NO_PARKING_HELMET"
	OUT_SERVICE_AREA              FenceRelation = "OUT_SERVICE_AREA"
	OUT_SERVICE_HELMET            FenceRelation = "OUT_SERVICE_HELMET"
	IN_BAN_RIDING                 FenceRelation = "IN_BAN_RIDING"
	FULL_PILE_NO_PARKING          FenceRelation = "FULL_PILE_NO_PARKING"
)

type ReturnCarService struct {
	configGateway *gateway.ConfigGateway
	partService   *PartService
	deviceRpc     *rpc.DeviceRPC
	fenceRfidGw   *gateway.FenceRfidGateway
}

func NewReturnCarService(
	configGateway *gateway.ConfigGateway,
	partService *PartService,
	deviceRpc *rpc.DeviceRPC,
	fenceRfidGw *gateway.FenceRfidGateway,
) *ReturnCarService {
	return &ReturnCarService{
		configGateway: configGateway,
		partService:   partService,
		deviceRpc:     deviceRpc,
		fenceRfidGw:   fenceRfidGw,
	}
}

// ReturnCar mirrors Java ServiceAreaServiceImpl.returnCar.
func (s *ReturnCarService) ReturnCar(ctx context.Context, tenantId string, req dto.ReturnCarCmd) (dto.ReturnCarCO, error) {
	car := req.CarCmd
	result := dto.ReturnCarCO{}
	trueVal := true
	result.IzCanReturn = &trueVal

	serviceArea, err := GetServiceAreaById(ctx, tenantId, *req.ServiceAreaId)
	if err != nil {
		return result, err
	}
	if serviceArea == nil {
		falseVal := false
		result.ReturnType = string(OUT_SERVICE_AREA)
		result.IzCanReturn = &falseVal
		return result, nil
	}

	carLoc := geo.Location{}
	if car.CarLocation != nil {
		carLoc = geo.Location{Lng: car.CarLocation.Lng, Lat: car.CarLocation.Lat}
	}

	checkType := PartCheckReturnCar
	if req.CheckType != nil {
		checkType = *req.CheckType
	}

	cmdCtx := dto.EnsureCommandContext(req.CommandContext, tenantId)
	if err := s.refreshCarLocationIfNeeded(ctx, req, car, &carLoc, &result, cmdCtx); err != nil {
		return result, err
	}

	rfidCard := ""
	if containsPart(car.Parts, "rfid") {
		rfidCard = s.resolveRfidCard(ctx, cmdCtx, car.Imei)
	}

	userLoc := geo.Location{}
	if req.UserCmd != nil && req.UserCmd.UserLocation != nil {
		userLoc = geo.Location{Lng: req.UserCmd.UserLocation.Lng, Lat: req.UserCmd.UserLocation.Lat}
	}

	backcarConfig, err := s.configGateway.GetConfigByServiceId(ctx, tenantId, serviceArea.Id)
	if err != nil {
		return result, err
	}
	if backcarConfig == nil {
		defaultBuffer := 10.0
		backcarConfig = &config.ConfigBackcarCO{ServiceId: req.ServiceAreaId, BufferDistance: &defaultBuffer}
	}

	inService := izInService(carLoc, userLoc, serviceArea)
	recordHelmetUnLock, _ := s.partService.helmetGateway.GetRecordHelmetLock(ctx, tenantId, car.CarId)
	izHelmetReign := backcarConfig.GetIzHelmetReign()
	result.IzHelmetReign = &izHelmetReign

	preciseParts := s.partService.GetPreciseParts(car.Parts, backcarConfig)
	assertParts := s.partService.GetAssertParts(car.Parts, recordHelmetUnLock, izHelmetReign, backcarConfig)
	locationAudit, _ := s.partService.GetLocationAudit(ctx, tenantId, car.CarId, req.OrderId)

	partCtx := carPartContext{
		tenantId: tenantId, car: car, checkType: checkType, orderId: req.OrderId, blueResults: req.Result, backcarConfig: backcarConfig, cmdCtx: cmdCtx,
	}

	if rfidCard != "" {
		if co, ok, err := s.rfidCardMatch(ctx, tenantId, rfidCard, req, &result, assertParts, partCtx, locationAudit, backcarConfig); err != nil {
			return result, err
		} else if ok {
			return *co, nil
		}
	}

	if !inService {
		result.ReturnType = string(OUT_SERVICE_AREA)
		near := true
		result.IzNearService = &near
		if err := s.checkAssertPart(ctx, partCtx, &result, assertParts, nil); err != nil {
			return result, err
		}
	} else {
		izNear := geo.Intersect(carLoc, serviceArea.ParsedPolygon, 200)
		result.IzNearService = &izNear

		ban, _ := BanRidingReturnCar(ctx, tenantId, carLoc, userLoc, backcarConfig)
		if ban != nil {
			result.ReturnType = string(IN_BAN_RIDING)
		} else {
			parking, _ := ParkingReturnCar(ctx, tenantId, carLoc, userLoc, serviceArea.Id, backcarConfig)
			if parking != nil {
				early, err := s.handleInParking(ctx, tenantId, req, parking, &result, assertParts, partCtx, backcarConfig)
				if err != nil {
					return result, err
				}
				if result.ReturnType == string(FULL_PILE_NO_PARKING) {
					return result, nil
				}
				if early {
					return *getReturnCarCO(&result, backcarConfig, locationAudit), nil
				}
			} else {
				noParking, _ := NoParkingReturnCar(ctx, tenantId, carLoc, userLoc, serviceArea.Id, backcarConfig)
				if noParking != nil {
					result.NoParkingCmd = fenceToMap(noParking)
					result.ReturnType = string(IN_NO_PARKING)
					if err := s.checkAssertPart(ctx, partCtx, &result, assertParts, nil); err != nil {
						return result, err
					}
					if result.IzCanReturn != nil && !*result.IzCanReturn {
						return *getReturnCarCO(&result, backcarConfig, locationAudit), nil
					}
				} else {
					result.ReturnType = string(OUT_PARKING_LAT)
					if err := s.checkAssertPart(ctx, partCtx, &result, assertParts, nil); err != nil {
						return result, err
					}
					if result.IzCanReturn != nil && !*result.IzCanReturn {
						return *getReturnCarCO(&result, backcarConfig, locationAudit), nil
					}
					if err := s.checkPrecisePart(ctx, partCtx, &result, preciseParts); err != nil {
						return result, err
					}
				}
			}
		}

		if maintain, _ := FindMaintainAreaNear(ctx, tenantId, carLoc, 100000, 1); maintain != nil {
			result.MaintainAreaCmd = fenceToMap(maintain)
		}
	}

	return *getReturnCarCO(&result, backcarConfig, locationAudit), nil
}

// FenceRelation mirrors Java ServiceAreaServiceImpl.fenceRelation.
func (s *ReturnCarService) FenceRelation(ctx context.Context, tenantId string, req dto.ReturnCarCmd) (dto.ReturnCarCO, error) {
	return s.fenceRelationInternal(ctx, tenantId, req, true)
}

// BlueFenceRelation mirrors Java ServiceAreaServiceImpl.blueFenceRelation (no izHelmetReign in response).
func (s *ReturnCarService) BlueFenceRelation(ctx context.Context, tenantId string, req dto.ReturnCarCmd) (dto.ReturnCarCO, error) {
	return s.fenceRelationInternal(ctx, tenantId, req, false)
}

func (s *ReturnCarService) fenceRelationInternal(ctx context.Context, tenantId string, req dto.ReturnCarCmd, setHelmetReign bool) (dto.ReturnCarCO, error) {
	var result dto.ReturnCarCO

	serviceArea, err := GetServiceAreaById(ctx, tenantId, *req.ServiceAreaId)
	if err != nil {
		return result, err
	}
	if serviceArea == nil {
		falseVal := false
		result.ReturnType = string(OUT_SERVICE_AREA)
		result.IzCanReturn = &falseVal
		return result, nil
	}

	carLoc, userLoc := carAndUserLocations(req)
	backcarConfig, err := s.configGateway.GetConfigByServiceId(ctx, tenantId, serviceArea.Id)
	if err != nil || backcarConfig == nil {
		defaultBuffer := 10.0
		backcarConfig = &config.ConfigBackcarCO{ServiceId: req.ServiceAreaId, BufferDistance: &defaultBuffer}
	}

	inService := izInService(carLoc, userLoc, serviceArea)
	recordHelmetUnLock, _ := s.partService.helmetGateway.GetRecordHelmetLock(ctx, tenantId, req.CarCmd.CarId)
	izHelmetReign := backcarConfig.GetIzHelmetReign()
	if setHelmetReign {
		result.IzHelmetReign = &izHelmetReign
	}
	assertParts := s.partService.GetAssertParts(req.CarCmd.Parts, recordHelmetUnLock, izHelmetReign, backcarConfig)

	checkType := PartCheckReturnCar
	if req.CheckType != nil {
		checkType = *req.CheckType
	}
	partCtx := carPartContext{
		tenantId: tenantId, car: req.CarCmd, checkType: checkType, orderId: req.OrderId, blueResults: req.Result, backcarConfig: backcarConfig,
		cmdCtx: dto.EnsureCommandContext(req.CommandContext, tenantId),
	}

	if !inService {
		result.ReturnType = string(OUT_SERVICE_AREA)
		near := true
		result.IzNearService = &near
		_ = s.checkAssertPart(ctx, partCtx, &result, assertParts, nil)
	} else {
		izNear := geo.Intersect(carLoc, serviceArea.ParsedPolygon, 200)
		result.IzNearService = &izNear
		ban, _ := BanRidingReturnCar(ctx, tenantId, carLoc, userLoc, backcarConfig)
		if ban != nil {
			result.ReturnType = string(IN_BAN_RIDING)
		} else {
			parking, _ := ParkingReturnCar(ctx, tenantId, carLoc, userLoc, serviceArea.Id, backcarConfig)
			if parking != nil {
				result.ReturnType = string(IN_PARKING_LAT)
				result.Parking = parkingRelationToMap(parking)
				_ = s.checkAssertPart(ctx, partCtx, &result, assertParts, nil)
			} else {
				noParking, _ := NoParkingReturnCar(ctx, tenantId, carLoc, userLoc, serviceArea.Id, backcarConfig)
				if noParking != nil {
					result.NoParkingCmd = fenceToMap(noParking)
					result.ReturnType = string(IN_NO_PARKING)
					_ = s.checkAssertPart(ctx, partCtx, &result, assertParts, nil)
				} else {
					result.ReturnType = string(OUT_PARKING_LAT)
					_ = s.checkAssertPart(ctx, partCtx, &result, assertParts, nil)
				}
			}
		}
	}

	rt := fenceRelationFromCO(result.ReturnType)
	result.IzCanReturn = izCanReturn(rt, backcarConfig)
	return result, nil
}

func (s *ReturnCarService) handleInParking(
	ctx context.Context,
	tenantId string,
	req dto.ReturnCarCmd,
	parking *gateway.FenceE,
	result *dto.ReturnCarCO,
	assertParts []string,
	partCtx carPartContext,
	backcarConfig *config.ConfigBackcarCO,
) (bool, error) {
	result.ReturnType = string(IN_PARKING_LAT)
	result.Parking = parkingRelationToMap(parking)

	if req.IzFrontSuppotFullPile != nil && *req.IzFrontSuppotFullPile && parking.IzFullPileNoStopEnabled() {
		carCount, err := ParkingCarCount(ctx, tenantId, parking.Id)
		if err != nil {
			return false, newBizError(codeInvokeFeignFail, "调用FEIGN失败")
		}
		if parking.MaxParkingNumber > 0 && int64(parking.MaxParkingNumber) <= carCount {
			falseVal := false
			result.IzCanReturn = &falseVal
			result.ReturnType = string(FULL_PILE_NO_PARKING)
			return false, nil
		}
	}

	if err := s.checkAssertPart(ctx, partCtx, result, assertParts, nil); err != nil {
		return false, err
	}
	if result.IzCanReturn != nil && !*result.IzCanReturn {
		return true, nil
	}

	parkParts := s.partService.GetParkingPart(parking.GetParkingParts(), req.CarCmd.Parts, backcarConfig, tenantId)
	if len(parkParts) > 0 {
		if err := s.checkParkingPart(ctx, partCtx, result, parkParts, parking.Id); err != nil {
			return false, err
		}
	}
	return false, nil
}

func (s *ReturnCarService) rfidCardMatch(
	ctx context.Context,
	tenantId, cardID string,
	req dto.ReturnCarCmd,
	result *dto.ReturnCarCO,
	assertParts []string,
	partCtx carPartContext,
	locationAudit bool,
	backcarConfig *config.ConfigBackcarCO,
) (*dto.ReturnCarCO, bool, error) {
	fenceId, err := s.fenceRfidGw.GetFenceIdByRfid(ctx, tenantId, cardID)
	if err != nil || fenceId == 0 {
		return nil, false, err
	}
	parking, err := gateway.GetFenceById(ctx, ParkingPrefix, tenantId, fenceId)
	if err != nil {
		return nil, false, err
	}
	if parking == nil {
		return nil, false, nil
	}

	if early, err := s.handleInParking(ctx, tenantId, req, parking, result, assertParts, partCtx, backcarConfig); err != nil {
		return nil, true, err
	} else if early {
		return getReturnCarCO(result, backcarConfig, locationAudit), true, nil
	}
	if result.ReturnType == string(FULL_PILE_NO_PARKING) {
		return result, true, nil
	}
	return getReturnCarCO(result, backcarConfig, locationAudit), true, nil
}

func (s *ReturnCarService) refreshCarLocationIfNeeded(ctx context.Context, req dto.ReturnCarCmd, car *dto.CarCmd, carLoc *geo.Location, result *dto.ReturnCarCO, cmdCtx *dto.CommandContext) error {
	now := time.Now().Unix()
	reportTime := int64(0)
	if car.ReportTime != nil {
		reportTime = *car.ReportTime
	}
	version := car.Version
	checkType := PartCheckReturnCar
	if req.CheckType != nil {
		checkType = *req.CheckType
	}
	izReturn := req.IzReturn != nil && *req.IzReturn

	if !izReturn || now-reportTime <= 5 || checkType == PartCheckBlueReturnCar {
		return nil
	}
	if strings.HasPrefix(version, "32.") || strings.HasPrefix(version, "8.") {
		return nil
	}

	deviceInfo, err := s.deviceRpc.GetDeviceInfo(ctx, cmdCtx, car.Imei)
	if err != nil {
		if checkType == PartCheckAutoReturnCar || checkType == PartCheckFocusReturnCar || checkType == PartCheckPaybackReturn {
			weak := true
			result.IzDeviceWeakNet = &weak
			return nil
		}
		if biz, ok := err.(*BizError); ok {
			return biz
		}
		return newBizError(codeInvokeFeignFail, "调用FEIGN失败")
	}
	falseVal := false
	result.IzDeviceWeakNet = &falseVal
	if deviceInfo != nil && deviceInfo.Lat != 0 && deviceInfo.Lng != 0 {
		carLoc.Lat = deviceInfo.Lat
		carLoc.Lng = deviceInfo.Lng
	}
	return nil
}

func (s *ReturnCarService) resolveRfidCard(ctx context.Context, cmdCtx *dto.CommandContext, imei string) string {
	detail, err := s.deviceRpc.GetDeviceDetail(ctx, cmdCtx, imei)
	if err != nil || detail == nil {
		return ""
	}
	if detail.RfidCarId == "" {
		return ""
	}
	if time.Now().UnixMilli()-detail.RfidTimestamp < 10*1000 {
		return detail.RfidCarId
	}
	return ""
}

func izInService(carLoc, userLoc geo.Location, serviceArea *gateway.FenceE) bool {
	// Java hardcoded switch(1): car location only
	return geo.IsPointInParsedPolygon(carLoc, serviceArea.ParsedPolygon)
}

func carAndUserLocations(req dto.ReturnCarCmd) (geo.Location, geo.Location) {
	carLoc, userLoc := geo.Location{}, geo.Location{}
	if req.CarCmd != nil && req.CarCmd.CarLocation != nil {
		carLoc = geo.Location{Lng: req.CarCmd.CarLocation.Lng, Lat: req.CarCmd.CarLocation.Lat}
	}
	if req.UserCmd != nil && req.UserCmd.UserLocation != nil {
		userLoc = geo.Location{Lng: req.UserCmd.UserLocation.Lng, Lat: req.UserCmd.UserLocation.Lat}
	}
	return carLoc, userLoc
}

func parkingRelationToMap(f *gateway.FenceE) map[string]interface{} {
	if f == nil {
		return nil
	}
	openBegin, openEnd := f.OpeningHoursBegin, f.OpeningHoursEnd
	if openBegin != "" && len(openBegin) == 5 {
		openBegin += ":00"
	}
	if openEnd != "" && len(openEnd) == 5 {
		openEnd += ":00"
	}
	m := map[string]interface{}{
		"commandContext":             nil,
		"id":                         f.Id,
		"name":                       f.Name,
		"shapeType":                  f.ShapeType,
		"maxParkingNumber":           f.MaxParkingNumber,
		"centerLat":                  f.CenterLat,
		"centerLng":                  f.CenterLng,
		"pointList":                  f.PointList,
		"serviceId":                  f.ServiceId,
		"tbeacon":                    boolDefault(f.Tbeacon),
		"directional":                boolDefault(f.Directional),
		"rfid":                       boolDefault(f.Rfid),
		"izEnable":                   f.IzEnable,
		"bufferDistance":             f.BufferDistance,
		"camera":                     boolDefault(f.Camera),
		"izCameraDirectionalBackcar": boolDefault(f.IzCameraDirectionalBackcar),
		"izCameraPointBackcar":       boolDefault(f.IzCameraPointBackcar),
		"kickstand":                  boolDefault(f.Kickstand),
		"pics":                       nil,
		"refTags":                    nil,
	}
	if f.AreaSizeSet {
		m["areaSize"] = f.AreaSize
	} else {
		m["areaSize"] = nil
	}
	if f.CoefficientOfDifficultSet {
		m["coefficientOfDifficult"] = f.CoefficientOfDifficult
	} else {
		m["coefficientOfDifficult"] = nil
	}
	if f.Direction != nil {
		m["direction"] = *f.Direction
	} else {
		m["direction"] = 0.0
	}
	if f.IzFullPileNoStop != nil {
		m["izFullPileNoStop"] = *f.IzFullPileNoStop
	} else {
		m["izFullPileNoStop"] = 0
	}
	if openBegin != "" {
		m["openingHoursBegin"] = openBegin
	}
	if openEnd != "" {
		m["openingHoursEnd"] = openEnd
	}
	return m
}

func boolDefault(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

func fenceToMap(f *gateway.FenceE) map[string]interface{} {
	if f == nil {
		return nil
	}
	m := map[string]interface{}{
		"id":        f.Id,
		"name":      f.Name,
		"serviceId": f.ServiceId,
		"izEnable":  f.IzEnable,
	}
	if f.BufferDistance > 0 {
		m["bufferDistance"] = f.BufferDistance
	}
	if f.MaxParkingNumber > 0 {
		m["maxParkingNumber"] = f.MaxParkingNumber
	}
	if f.IzFullPileNoStop != nil {
		m["izFullPileNoStop"] = *f.IzFullPileNoStop
	}
	if f.CoefficientOfDifficult > 0 {
		m["coefficientOfDifficult"] = f.CoefficientOfDifficult
	}
	if f.OpeningHoursBegin != "" {
		m["openingHoursBegin"] = f.OpeningHoursBegin
	}
	if f.OpeningHoursEnd != "" {
		m["openingHoursEnd"] = f.OpeningHoursEnd
	}
	if f.IzOpenAllDay != nil {
		m["izOpenAllDay"] = *f.IzOpenAllDay
	}
	if f.Tbeacon != nil {
		m["tbeacon"] = *f.Tbeacon
	}
	if f.Directional != nil {
		m["directional"] = *f.Directional
	}
	if f.Rfid != nil {
		m["rfid"] = *f.Rfid
	}
	if f.Camera != nil {
		m["camera"] = *f.Camera
	}
	if f.Kickstand != nil {
		m["kickstand"] = *f.Kickstand
	}
	if f.Direction != nil {
		m["direction"] = *f.Direction
	}
	if f.FormulateDirection != nil {
		m["formulateDirection"] = *f.FormulateDirection
	}
	return m
}

func convertPartRes(in []PartAnalysisResultCO) []dto.PartAnalysisResultCO {
	out := make([]dto.PartAnalysisResultCO, len(in))
	for i, v := range in {
		out[i] = dto.PartAnalysisResultCO{Name: dto.PartNamePtr(v.Name), Result: v.Result}
	}
	return out
}

func izCanReturn(returnType FenceRelation, cfg *config.ConfigBackcarCO) *bool {
	var res bool
	switch returnType {
	case IN_PARKING_LAT, IN_PARKING_PART:
		res = true
	case OUT_PARKING_LAT:
		res = cfg.AllowOutofParking != nil && *cfg.AllowOutofParking
	case IN_PARKING_DIRECTION, IN_PARKING_RFID, IN_PARKING_KICK, IN_PARKING_CAMERA_DIRECTIONAL, IN_PARKING_CAMERA_POINT, IN_PARKING_CAMERA:
		res = false
	case IN_PARKING_HELMET, OUT_PARKING_HELMET, IN_NO_PARKING_HELMET, OUT_SERVICE_HELMET:
		res = !cfg.GetIzHelmetReign()
	case IN_NO_PARKING:
		res = cfg.AllowInNostop != nil && *cfg.AllowInNostop
	case OUT_SERVICE_AREA:
		res = cfg.AllowOutofService != nil && *cfg.AllowOutofService
	case IN_BAN_RIDING:
		res = cfg.AllowInBanRiding != nil && *cfg.AllowInBanRiding
	default:
		res = false
	}
	return &res
}

func applyPenaltyConfigToReturnType(returnType FenceRelation, cfg *config.ConfigBackcarCO, locationAudit bool) FenceRelation {
	allowOutOfParking := cfg.AllowOutofParking != nil && *cfg.AllowOutofParking
	allowInNoStop := cfg.AllowInNostop != nil && *cfg.AllowInNostop
	allowInBanRiding := cfg.AllowInBanRiding != nil && *cfg.AllowInBanRiding
	allowOutOfService := cfg.AllowOutofService != nil && *cfg.AllowOutofService

	dispatchCost := 0
	if cfg.DispatchCost != nil {
		dispatchCost = *cfg.DispatchCost
	}
	penInNoStop := 0
	if cfg.PenaltyInNostop != nil {
		penInNoStop = *cfg.PenaltyInNostop
	}
	penInOutOfService := 0
	if cfg.PenaltyOutofService != nil {
		penInOutOfService = *cfg.PenaltyOutofService
	}
	penaInBanRiding := 0
	if cfg.PenaltyInBanRiding != nil {
		penaInBanRiding = *cfg.PenaltyInBanRiding
	}

	switch returnType {
	case IN_PARKING_LAT, IN_PARKING_PART, OUT_PARKING_LAT:
		if locationAudit || (allowOutOfParking && dispatchCost == 0) {
			return IN_PARKING_LAT
		}
	case IN_PARKING_DIRECTION, IN_PARKING_RFID, IN_PARKING_KICK, IN_PARKING_HELMET, OUT_PARKING_HELMET:
		if locationAudit {
			return IN_PARKING_HELMET
		}
	case IN_NO_PARKING:
		if locationAudit || (allowInNoStop && penInNoStop == 0) {
			return IN_PARKING_LAT
		}
	case IN_NO_PARKING_HELMET:
		if locationAudit {
			return IN_PARKING_HELMET
		}
	case OUT_SERVICE_HELMET:
		// Java case 4102: no change
	case OUT_SERVICE_AREA:
		if locationAudit || (allowOutOfService && penInOutOfService == 0) {
			return IN_PARKING_LAT
		}
	case IN_BAN_RIDING:
		if locationAudit || (allowInBanRiding && penaInBanRiding == 0) {
			return IN_PARKING_LAT
		}
	}
	return returnType
}

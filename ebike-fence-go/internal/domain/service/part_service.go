package service

import (
	"context"

	"ebike-fence-go/internal/api/dto"
	domainconfig "ebike-fence-go/internal/domain/config"
	"ebike-fence-go/internal/domain/gateway"
	domainpart "ebike-fence-go/internal/domain/part"
	"ebike-fence-go/internal/infrastructure/rpc"
	appconfig "ebike-fence-go/internal/pkg/config"
	"go.uber.org/zap"
)

type PartService struct {
	helmetGateway *gateway.HelmetGateway
	partGateway   *gateway.PartGateway
	fenceRfidGw   *gateway.FenceRfidGateway
	configGateway *gateway.ConfigGateway
	deviceRpc     *rpc.DeviceRPC
	managementRpc *rpc.ManagementRPC
	ticketRpc     *rpc.TicketRPC
}

func NewPartService(helmetGateway *gateway.HelmetGateway, partGateway *gateway.PartGateway, fenceRfidGw *gateway.FenceRfidGateway, configGateway *gateway.ConfigGateway, deviceRpc *rpc.DeviceRPC, managementRpc *rpc.ManagementRPC, ticketRpc *rpc.TicketRPC) *PartService {
	return &PartService{
		helmetGateway: helmetGateway,
		partGateway:   partGateway,
		fenceRfidGw:   fenceRfidGw,
		configGateway: configGateway,
		deviceRpc:     deviceRpc,
		managementRpc: managementRpc,
		ticketRpc:     ticketRpc,
	}
}

type PartAnalysisResultCO struct {
	Name   string `json:"name"`
	Result *bool  `json:"result,omitempty"`
}

func partResultBool(v bool) *bool {
	b := v
	return &b
}

func partResultIsFalse(r *bool) bool {
	return r != nil && !*r
}

func partResultIsTrue(r *bool) bool {
	return r != nil && *r
}

func (s *PartService) GetPreciseParts(parts []string, backcarConfig *domainconfig.ConfigBackcarCO) []string {
	return getPreciseParts(parts)
}

func (s *PartService) GetAssertParts(parts []string, recordHelmetUnLock string, izHelmetReign bool, backcarConfig *domainconfig.ConfigBackcarCO) []string {
	assertPartsList := map[string]bool{"helmet": true}
	var assertParts []string
	for _, p := range parts {
		if assertPartsList[p] {
			if p == "helmet" {
				if recordHelmetUnLock != "" {
					assertParts = append(assertParts, p)
				}
			} else {
				assertParts = append(assertParts, p)
			}
		}
	}
	return assertParts
}

func (s *PartService) GetParkingPart(parkingParts []string, carParts []string, backcarConfig *domainconfig.ConfigBackcarCO, tenantId string) []string {
	out := append([]string(nil), parkingParts...)
	if containsPart(out, "bluetooth_beacon") {
		if !containsPart(carParts, "bluetooth_beacon") && backcarConfig.BeaconEffect != nil && *backcarConfig.BeaconEffect == effectHasPart {
			out = filterPart(out, "bluetooth_beacon")
		}
	}
	if containsPart(out, "rfid_beacon") {
		if !containsPart(carParts, "rfid_beacon") && backcarConfig.RfidEffect != nil && *backcarConfig.RfidEffect == effectHasPart {
			out = filterPart(out, "rfid_beacon")
		}
	}
	if containsPart(out, "direction") {
		if !containsPart(carParts, "direction") && backcarConfig.DirectionEffect != nil && *backcarConfig.DirectionEffect == effectHasPart {
			out = filterPart(out, "direction")
		}
	}
	if containsPart(out, "kickstand") {
		if !containsPart(carParts, "kickstand") && backcarConfig.FootSupportEffect != nil && *backcarConfig.FootSupportEffect == effectHasPart {
			out = filterPart(out, "kickstand")
		}
	}
	if (!containsPart(carParts, "camera") && backcarConfig.CameraEffect != nil && *backcarConfig.CameraEffect == effectHasPart) || !backcarConfig.GetIzCameraBackcar() {
		out = filterParts(out, "camera", "camera_directional", "camera_point")
	}
	for _, ignoreTenant := range appconfig.GlobalConfig.Xyy.DirectionIgnoreTenantIds {
		if ignoreTenant == tenantId && !containsPart(carParts, "direction") {
			out = filterPart(out, "direction")
			break
		}
	}
	return out
}

func (s *PartService) GetLocationAudit(ctx context.Context, tenantId, carId string, orderId *int64) (bool, error) {
	return s.partGateway.GetLocationAudit(ctx, tenantId, carId, orderId)
}

func filterPart(parts []string, name string) []string {
	var out []string
	for _, p := range parts {
		if p != name {
			out = append(out, p)
		}
	}
	return out
}

func filterParts(parts []string, names ...string) []string {
	skip := map[string]bool{}
	for _, n := range names {
		skip[n] = true
	}
	var out []string
	for _, p := range parts {
		if !skip[p] {
			out = append(out, p)
		}
	}
	return out
}

type partAnalysisParams struct {
	tenantId      string
	imei          string
	carId         string
	parts         []string
	checkType     int
	orderId       *int64
	fenceId       *int64
	backcarConfig *domainconfig.ConfigBackcarCO
	cmdCtx        *dto.CommandContext
}

func (s *PartService) GetPartResult(ctx context.Context, tenantId string, imei string, carId string, parts []string, checkType int, parkingId *int64, orderId *int64, backcarConfig *domainconfig.ConfigBackcarCO, cmdCtx *dto.CommandContext) ([]PartAnalysisResultCO, error) {
	var fenceId *int64
	if parkingId != nil {
		fenceId = parkingId
	}
	return s.partAnalysis(ctx, partAnalysisParams{
		tenantId:      tenantId,
		imei:          imei,
		carId:         carId,
		parts:         parts,
		checkType:     checkType,
		orderId:       orderId,
		fenceId:       fenceId,
		backcarConfig: backcarConfig,
		cmdCtx:        dto.EnsureCommandContext(cmdCtx, tenantId),
	})
}

func (s *PartService) PartAnalysis(ctx context.Context, tenantId string, imei string, carId string, parts []string, checkType int) ([]PartAnalysisResultCO, error) {
	return s.partAnalysis(ctx, partAnalysisParams{
		tenantId:  tenantId,
		imei:      imei,
		carId:     carId,
		parts:     parts,
		checkType: checkType,
	})
}

func (s *PartService) partAnalysis(ctx context.Context, p partAnalysisParams) ([]PartAnalysisResultCO, error) {
	var results []PartAnalysisResultCO
	if len(p.parts) == 0 {
		return results, nil
	}

	carInfo, err := s.managementRpc.GetCarInfoByImei(ctx, p.imei, dto.EnsureCommandContext(p.cmdCtx, p.tenantId))
	if err != nil {
		zap.L().Warn("GetCarInfoByImei failed", zap.String("imei", p.imei), zap.Error(err))
	}
	if carInfo == nil {
		carInfo = &rpc.CarInfo{Imei: p.imei, CarId: p.carId}
	}

	hasCameraTicket := s.shouldSkipCameraQuery(ctx, p, carInfo)
	cmdCtx := dto.EnsureCommandContext(p.cmdCtx, p.tenantId)
	var cachedCamera *rpc.CameraEntity
	if !hasCameraTicket && s.hasCameraPart(p.parts) && p.backcarConfig != nil && p.backcarConfig.CameraAngle == nil {
		if cache, err := s.deviceRpc.QueryCameraState(ctx, cmdCtx, p.imei); err == nil && rpc.CameraCacheValid(cache) {
			cachedCamera = &rpc.CameraEntity{Event: cache.Event, AngleDet: cache.AngleDet}
		}
	}
	deviceInfo, err := s.deviceRpc.GetDeviceInfo(ctx, cmdCtx, p.imei)
	if err != nil {
		return nil, deviceInfoBizError(err)
	}
	if deviceInfo != nil && cachedCamera != nil {
		deviceInfo.Camera = cachedCamera
	}

	recordHelmetUnLock, err := s.helmetGateway.GetRecordHelmetLock(ctx, p.tenantId, p.carId)
	if err != nil {
		zap.L().Warn("GetRecordHelmetLock failed", zap.String("tenantId", p.tenantId), zap.String("carId", p.carId), zap.Error(err))
	}

	helmetAuditWhite := appconfig.GlobalConfig.Xyy.HelmetAuditWhite

	for _, partName := range p.parts {
		res, err := s.analyzePart(ctx, p, partName, carInfo, deviceInfo, recordHelmetUnLock, helmetAuditWhite, hasCameraTicket)
		if err != nil {
			return nil, err
		}
		results = append(results, PartAnalysisResultCO{Name: partName, Result: res})
	}
	return results, nil
}

func (s *PartService) hasCameraPart(parts []string) bool {
	for _, partName := range parts {
		switch partName {
		case "camera", "camera_directional", "camera_point":
			return true
		}
	}
	return false
}

func (s *PartService) shouldSkipCameraQuery(ctx context.Context, p partAnalysisParams, carInfo *rpc.CarInfo) bool {
	if !s.hasCameraPart(p.parts) {
		return true
	}
	helmetAudit, _ := s.partGateway.GetHelmetAudit(ctx, p.tenantId, p.carId, p.orderId)
	if helmetAudit {
		return true
	}
	if carInfo.Model == "" {
		return false
	}
	repair, err := s.managementRpc.GetCarReapirsByModel(ctx, dto.EnsureCommandContext(p.cmdCtx, p.tenantId), carInfo.Model, 3)
	if err != nil || repair == nil {
		return false
	}
	hasTicket, err := s.ticketRpc.QueryTicket(ctx, dto.EnsureCommandContext(p.cmdCtx, p.tenantId), p.carId, repair.Id)
	return err == nil && hasTicket
}

func (s *PartService) analyzePart(ctx context.Context, p partAnalysisParams, partName string, carInfo *rpc.CarInfo, deviceInfo *rpc.DeviceInfoEntity, recordHelmetUnLock string, helmetAuditWhite []string, hasCameraTicket bool) (*bool, error) {
	switch partName {
	case "helmet":
		helmetAudit, _ := s.partGateway.GetHelmetAudit(ctx, p.tenantId, p.carId, p.orderId)
		var helmetLock, helmet6React, helmet6Lock *int
		if deviceInfo != nil {
			helmetLock = deviceInfo.HelmetLock
			helmet6React = deviceInfo.Helmet6React
			helmet6Lock = deviceInfo.Helmet6Lock
		}
		ok := domainpart.HelmetReturnResult(recordHelmetUnLock, helmetAudit, helmetLock, helmet6React, helmet6Lock, helmetAuditWhite, p.tenantId)
		if !ok && helmetAudit {
			ok = true
		}
		return partResultBool(ok), nil

	case "rfid_beacon":
		pointAudit, _ := s.partGateway.GetPointAudit(ctx, p.tenantId, p.carId, p.orderId)
		if deviceInfo == nil || deviceInfo.Rfid == nil {
			if p.backcarConfig != nil && p.backcarConfig.RfidEffect != nil && *p.backcarConfig.RfidEffect == 1 && !pointAudit {
				return partResultBool(false), nil
			}
			return partResultBool(true), nil
		}
		ok := domainpart.RFIDReturnResult(deviceInfo.Rfid.Event)
		if !ok && pointAudit {
			ok = true
		}
		return partResultBool(ok), nil

	case "camera", "camera_directional", "camera_point":
		return s.analyzeCamera(ctx, p, deviceInfo, carInfo, hasCameraTicket)

	case "kickstand":
		kickAudit, _ := s.partGateway.GetKickstandAudit(ctx, p.tenantId, p.carId, p.orderId)
		if kickAudit {
			return partResultBool(true), nil
		}
		if deviceInfo == nil || deviceInfo.KickStand == nil {
			if p.backcarConfig != nil && p.backcarConfig.FootSupportEffect != nil && *p.backcarConfig.FootSupportEffect == 1 && !kickAudit {
				return partResultBool(false), nil
			}
			return partResultBool(true), nil
		}
		ok := domainpart.RFIDReturnResult(deviceInfo.KickStand.Event)
		if !ok && kickAudit {
			ok = true
		}
		return partResultBool(ok), nil

	case "bluetooth_beacon":
		beacon, err := s.deviceRpc.GetBlueTBeacon(ctx, dto.EnsureCommandContext(p.cmdCtx, p.tenantId), p.tenantId, p.imei)
		if err != nil {
			return nil, err
		}
		if beacon == nil {
			beacon = &rpc.BluetoothBeaconEntity{}
		}
		ok := domainpart.BeaconReturnResult(beacon.Event, beacon.TBeaconAddr)
		beaconAudit, _ := s.partGateway.GetTBeaconAudit(ctx, p.tenantId, p.carId, p.orderId)
		if !ok && beaconAudit {
			ok = true
		}
		return partResultBool(ok), nil

	case "direction":
		directionAudit, _ := s.partGateway.GetDirectionAudit(ctx, p.tenantId, p.carId, p.orderId)
		ignoreTenant := false
		for _, t := range appconfig.GlobalConfig.Xyy.DirectionIgnoreTenantIds {
			if t == p.tenantId {
				ignoreTenant = true
				break
			}
		}
		if deviceInfo == nil || deviceInfo.Heading == nil {
			if p.backcarConfig != nil && p.backcarConfig.DirectionEffect != nil && *p.backcarConfig.DirectionEffect == 1 && !directionAudit {
				return partResultBool(false), nil
			}
			return partResultBool(true), nil
		}
		var headingAngle *float64
		if deviceInfo.Heading.HeadingAngle != nil {
			normalized := domainpart.NormalizeHeadingAngle(*deviceInfo.Heading.HeadingAngle)
			headingAngle = &normalized
		}
		fenceDir, formDir := s.resolveParkingDirections(ctx, p, deviceInfo)
		ok := domainpart.DirectionReturnResult(headingAngle, fenceDir, formDir, ignoreTenant)
		if !ok && directionAudit {
			ok = true
		}
		return partResultBool(ok), nil

	default:
		return partResultBool(true), nil
	}
}

func (s *PartService) analyzeCamera(ctx context.Context, p partAnalysisParams, deviceInfo *rpc.DeviceInfoEntity, carInfo *rpc.CarInfo, hasCameraTicket bool) (*bool, error) {
	cameraAudit, _ := s.partGateway.GetCameraAudit(ctx, p.tenantId, p.carId, p.orderId)
	if cameraAudit || hasCameraTicket {
		return partResultBool(true), nil
	}
	cam := &rpc.CameraEntity{}
	if deviceInfo != nil && deviceInfo.Camera != nil {
		cam = deviceInfo.Camera
	}
	if cam.Event != nil {
		if *cam.Event == 2 {
			errNum, _ := s.partGateway.IncCameraErrorCount(ctx, p.tenantId, p.carId)
			if errNum >= 2 && carInfo != nil && carInfo.Model != "" {
				cmdCtx := dto.EnsureCommandContext(p.cmdCtx, p.tenantId)
				repair, err := s.managementRpc.GetCarReapirsByModel(ctx, cmdCtx, carInfo.Model, 3)
				if err == nil && repair != nil {
					if ticketErr := s.ticketRpc.CreateTicket(ctx, cmdCtx, p.carId, int64(repair.Type)); ticketErr != nil {
						zap.L().Warn("CreateTicket failed", zap.String("carId", p.carId), zap.Error(ticketErr))
					}
					_ = s.partGateway.DelCameraErrorCount(ctx, p.tenantId, p.carId)
				}
			}
		} else {
			_ = s.partGateway.DelCameraErrorCount(ctx, p.tenantId, p.carId)
		}
	}
	cameraAngle := cam.CameraAngle
	if p.backcarConfig != nil && p.backcarConfig.CameraAngle != nil {
		cameraAngle = p.backcarConfig.CameraAngle
	}
	ok := domainpart.CameraReturnResult(cam.Event, cam.AngleDet, cameraAngle)
	if !ok {
		_, _ = s.partGateway.IncCameraFailCount(ctx, p.tenantId, p.carId, p.orderId)
		return partResultBool(false), nil
	}
	return partResultBool(true), nil
}

func (s *PartService) resolveParkingDirections(ctx context.Context, p partAnalysisParams, deviceInfo *rpc.DeviceInfoEntity) (*float64, *float64) {
	cardID := ""
	if deviceInfo != nil && deviceInfo.Rfid != nil {
		cardID = deviceInfo.Rfid.CardID
	}
	var fenceID int64
	if cardID != "" && s.fenceRfidGw != nil {
		if id, err := s.fenceRfidGw.GetFenceIdByRfid(ctx, p.tenantId, cardID); err == nil && id > 0 {
			fenceID = id
		}
	} else if p.fenceId != nil {
		fenceID = *p.fenceId
	}
	if fenceID == 0 {
		return nil, nil
	}
	parking, err := gateway.GetFenceById(ctx, ParkingPrefix, p.tenantId, fenceID)
	if err != nil || parking == nil {
		return nil, nil
	}
	return parking.Direction, parking.FormulateDirection
}

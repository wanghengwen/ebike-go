package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ebike-fence-go/internal/api/dto"
	domainconfig "ebike-fence-go/internal/domain/config"
	domainpart "ebike-fence-go/internal/domain/part"
	"ebike-fence-go/internal/infrastructure/rpc"
	appconfig "ebike-fence-go/internal/pkg/config"
)

const codeFenceSearchFail = "13002"

func (s *PartService) resolveBackcarConfig(ctx context.Context, tenantId string, cmd dto.CheckingPartCmd) (*domainconfig.ConfigBackcarCO, int64) {
	var serviceId int64
	if cmd.ServiceId != nil {
		serviceId = *cmd.ServiceId
	}
	carInfo, _ := s.managementRpc.GetCarInfoByImei(ctx, cmd.Imei, dto.EnsureCommandContext(cmd.CommandContext, tenantId))
	if carInfo != nil && serviceId == 0 {
		serviceId = carInfo.ServiceId
	}
	if serviceId == 0 || s.configGateway == nil {
		return nil, serviceId
	}
	cfg, _ := s.configGateway.GetConfigByServiceId(ctx, tenantId, serviceId)
	return cfg, serviceId
}

func (s *PartService) PartAnalysisList(ctx context.Context, tenantId string, cmd dto.CheckingPartCmd) ([]dto.PartAnalysisResultCO, error) {
	checkType := 0
	if cmd.CheckType != nil {
		checkType = *cmd.CheckType
	}
	backcar, _ := s.resolveBackcarConfig(ctx, tenantId, cmd)
	results, err := s.GetPartResult(ctx, tenantId, cmd.Imei, cmd.CarId, cmd.Parts, checkType, cmd.FenceId, cmd.OrderId, backcar, cmd.CommandContext)
	if err != nil {
		return nil, err
	}
	out := make([]dto.PartAnalysisResultCO, len(results))
	for i, r := range results {
		out[i] = dto.PartAnalysisResultCO{Name: dto.PartNamePtr(r.Name), Result: r.Result}
	}
	return out, nil
}

func partAnalysisByAppSuccess() dto.PartAnalysisResultCO {
	// Java partAnalysisByReturn: return new PartAnalysisResultCO(null, true);
	ok := true
	canUse := true
	return dto.PartAnalysisResultCO{Result: &ok, CanUse: &canUse}
}

func partAnalysisByAppFailure(name string, result *bool) dto.PartAnalysisResultCO {
	canUse := true
	return dto.PartAnalysisResultCO{Name: dto.PartNamePtr(name), Result: result, CanUse: &canUse}
}

// partHelmetTempParkingResult mirrors Java PartAnalysisResultCO(name, result, izExist) where canUse defaults to true.
func partHelmetTempParkingResult(result, izExist bool) dto.PartAnalysisResultCO {
	canUse := true
	return dto.PartAnalysisResultCO{
		Name:    dto.PartNamePtr("helmet"),
		Result:  &result,
		IzExist: partBoolPtr(izExist),
		CanUse:  &canUse,
	}
}

func (s *PartService) PartAnalysisByApp(ctx context.Context, tenantId string, cmd dto.CheckingPartCmd) (dto.PartAnalysisResultCO, error) {
	parts := cmd.Parts
	if len(parts) == 0 {
		parts = []string{"bluetooth_beacon", "rfid_beacon", "direction", "kickstand", "camera", "helmet"}
	}
	cmd.Parts = parts
	results, err := s.PartAnalysisList(ctx, tenantId, cmd)
	if err != nil {
		return dto.PartAnalysisResultCO{}, err
	}
	for _, r := range results {
		if r.Result != nil && !*r.Result {
			return partAnalysisByAppFailure(dtoPartNameStr(r.Name), r.Result), nil
		}
	}
	return partAnalysisByAppSuccess(), nil
}

func dtoPartNameStr(name *string) string {
	if name == nil {
		return ""
	}
	return *name
}

func (s *PartService) CheckHelmet(ctx context.Context, tenantId string, cmd dto.CheckingPartCmd) (dto.PartAnalysisResultCO, error) {
	cmdCtx := dto.EnsureCommandContext(cmd.CommandContext, tenantId)
	carInfo, _ := s.managementRpc.GetCarInfoByImei(ctx, cmd.Imei, cmdCtx)
	deviceInfo, err := s.deviceRpc.GetDeviceInfo(ctx, cmdCtx, cmd.Imei)
	if err != nil {
		return dto.PartAnalysisResultCO{}, deviceInfoBizError(err)
	}
	izExist := carInfo != nil && carInfo.Helmet != ""
	var helmetLock, helmet6React, helmet6Lock *int
	if deviceInfo != nil {
		helmetLock = deviceInfo.HelmetLock
		helmet6React = deviceInfo.Helmet6React
		helmet6Lock = deviceInfo.Helmet6Lock
	}
	out := domainpart.HelmetReturnAnalysis(domainpart.HelmetAnalysisInput{
		HelmetLock:       helmetLock,
		Helmet6React:     helmet6React,
		Helmet6Lock:      helmet6Lock,
		IzExist:          izExist,
		HelmetAuditWhite: appconfig.GlobalConfig.Xyy.HelmetAuditWhite,
		TenantId:         tenantId,
	})
	return helmetResultToDTO(out), nil
}

func (s *PartService) CheckHelmetByTempParking(ctx context.Context, tenantId string, cmd dto.CheckingPartCmd) (dto.PartAnalysisResultCO, error) {
	cmdCtx := dto.EnsureCommandContext(cmd.CommandContext, tenantId)
	carInfo, _ := s.managementRpc.GetCarInfoByCarId(ctx, cmdCtx, cmd.CarId)
	izExist := carInfo != nil && carInfo.Helmet != ""
	if carInfo == nil {
		return partHelmetTempParkingResult(true, izExist), nil
	}
	useCarCfg, _ := s.configGateway.GetUseCarConfig(ctx, tenantId, carInfo.ServiceId)
	var helmetCfg *domainconfig.HelmetConfig
	if useCarCfg != nil {
		helmetCfg = useCarCfg.HelmetConfig
	}
	if helmetCfg != nil && helmetCfg.GetIzTempParkingReturnHelmet() && izExist {
		deviceInfo, err := s.deviceRpc.GetDeviceInfo(ctx, cmdCtx, carInfo.Imei)
		if err != nil {
			return dto.PartAnalysisResultCO{}, deviceInfoBizError(err)
		}
		recordHelmetUnLock, _ := s.helmetGateway.GetRecordHelmetLock(ctx, tenantId, cmd.CarId)
		var helmetLock, helmet6React, helmet6Lock *int
		if deviceInfo != nil {
			helmetLock = deviceInfo.HelmetLock
			helmet6React = deviceInfo.Helmet6React
			helmet6Lock = deviceInfo.Helmet6Lock
		}
		out := domainpart.HelmetReturnAnalysis(domainpart.HelmetAnalysisInput{
			OpenHelmet:       recordHelmetUnLock,
			HelmetLock:       helmetLock,
			Helmet6React:     helmet6React,
			Helmet6Lock:      helmet6Lock,
			IzExist:          true,
			HelmetAuditWhite: appconfig.GlobalConfig.Xyy.HelmetAuditWhite,
			TenantId:         tenantId,
		})
		helmetAudit, _ := s.partGateway.GetHelmetAudit(ctx, tenantId, cmd.CarId, cmd.OrderId)
		if out.Result != nil && !*out.Result && !helmetAudit {
			return partHelmetTempParkingResult(false, true), nil
		}
	}
	return partHelmetTempParkingResult(true, izExist), nil
}

func (s *PartService) CheckKickstand(ctx context.Context, cmd dto.CheckingPartCmd) (dto.PartAnalysisResultCO, error) {
	tenantID := ""
	if cmd.CommandContext != nil {
		tenantID = cmd.CommandContext.TenantId
	}
	cmdCtx := dto.EnsureCommandContext(cmd.CommandContext, tenantID)
	carInfo, _ := s.managementRpc.GetCarInfoByImei(ctx, cmd.Imei, cmdCtx)
	deviceInfo, err := s.deviceRpc.GetDeviceInfo(ctx, cmdCtx, cmd.Imei)
	if err != nil {
		return dto.PartAnalysisResultCO{}, deviceInfoBizError(err)
	}
	izUsed := carInfo != nil && carInfo.Kickstand != ""
	var kick *rpc.KickstandEntity
	if deviceInfo != nil {
		kick = deviceInfo.KickStand
	}
	name, izExist, result := domainpart.KickstandReturnAnalysis(kick, izUsed)
	return dto.PartAnalysisResultCO{Name: dto.PartNamePtr(name), IzExist: izExist, Result: result}, nil
}

func (s *PartService) CheckHelmetState(ctx context.Context, tenantId string, cmd dto.HelmetCheckCmd) (dto.PartAnalysisResultCO, error) {
	cmdCtx := dto.EnsureCommandContext(cmd.CommandContext, tenantId)
	carInfo, _ := s.managementRpc.GetCarInfoByImei(ctx, cmd.Imei, cmdCtx)
	izExist := carInfo != nil && carInfo.Helmet != ""
	var backcar *domainconfig.ConfigBackcarCO
	if carInfo != nil && s.configGateway != nil {
		backcar, _ = s.configGateway.GetConfigByServiceId(ctx, tenantId, carInfo.ServiceId)
	}
	var fixTicketNo *int64
	if carInfo != nil {
		tickets, _ := s.ticketRpc.GetHelmetTicket(ctx, cmdCtx, carInfo.Imei, false, carInfo.ServiceId, []int{13, 14})
		if len(tickets) > 0 && tickets[0].TicketNo != "" {
			if n, err := parseTicketNo(tickets[0].TicketNo); err == nil {
				fixTicketNo = &n
			}
		}
	}
	out := domainpart.HelmetReturnAnalysis(domainpart.HelmetAnalysisInput{
		IzExist:          izExist,
		FixTicketNo:      fixTicketNo,
		BackcarConfig:    backcar,
		HelmetAuditWhite: appconfig.GlobalConfig.Xyy.HelmetAuditWhite,
		TenantId:         tenantId,
	})
	return helmetResultToDTO(out), nil
}

func (s *PartService) SetHelmetAudit(ctx context.Context, tenantId, carId string, orderId int64) error {
	return s.partGateway.SetHelmetAudit(ctx, tenantId, carId, orderId)
}

func (s *PartService) SetPointAudit(ctx context.Context, tenantId, carId string, orderId int64) error {
	return s.partGateway.SetPointAudit(ctx, tenantId, carId, orderId)
}

func (s *PartService) SetTBeaconAudit(ctx context.Context, tenantId, carId string, orderId int64) error {
	return s.partGateway.SetTBeaconAudit(ctx, tenantId, carId, orderId)
}

func (s *PartService) SetDirectionAudit(ctx context.Context, tenantId, carId string, orderId int64) error {
	return s.partGateway.SetDirectionAudit(ctx, tenantId, carId, orderId)
}

func (s *PartService) SetKickstandAudit(ctx context.Context, tenantId, carId string, orderId int64) error {
	return s.partGateway.SetKickstandAudit(ctx, tenantId, carId, orderId)
}

func (s *PartService) SetCameraAudit(ctx context.Context, tenantId, carId string, orderId int64) error {
	return s.partGateway.SetCameraAudit(ctx, tenantId, carId, orderId)
}

func (s *PartService) SetLocationAudit(ctx context.Context, tenantId, carId string, orderId int64) error {
	return s.partGateway.SetLocationAudit(ctx, tenantId, carId, orderId)
}

func (s *PartService) IzCanCameraAudit(ctx context.Context, tenantId, carId string, orderId, serviceId *int64) (bool, error) {
	if orderId == nil || serviceId == nil {
		return false, nil
	}
	failCount, _ := s.partGateway.GetCameraFailCount(ctx, tenantId, carId, orderId)
	cfg, _ := s.configGateway.GetConfigByServiceId(ctx, tenantId, *serviceId)
	threshold := 4
	if cfg != nil {
		threshold = cfg.GetCameraImpunityCount()
	}
	return failCount >= int64(threshold), nil
}

func (s *PartService) GetBlueTBeaconInfoFromCache(ctx context.Context, tenantId, imei string) (dto.BlueTBeaconInfoCo, error) {
	raw, err := s.deviceRpc.GetBlueTBeaconInfoFromCache(ctx, tenantId, imei)
	if err != nil {
		return dto.BlueTBeaconInfoCo{}, err
	}
	if strings.TrimSpace(raw) == "" {
		return dto.BlueTBeaconInfoCo{}, newBizError(codeFenceSearchFail, "数据查询失败，未查到数据")
	}
	var out dto.BlueTBeaconInfoCo
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return dto.BlueTBeaconInfoCo{}, err
	}
	return out, nil
}

func helmetResultToDTO(out domainpart.HelmetAnalysisOutput) dto.PartAnalysisResultCO {
	return dto.PartAnalysisResultCO{
		Name:    dto.PartNamePtr(out.Name),
		IzExist: out.IzExist,
		Result:  out.Result,
		CanUse:  out.CanUse,
		UseType: out.UseType,
	}
}

func partBoolPtr(v bool) *bool {
	b := v
	return &b
}

func parseTicketNo(s string) (int64, error) {
	var n int64
	_, err := fmt.Sscan(s, &n)
	return n, err
}

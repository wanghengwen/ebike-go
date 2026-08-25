package service

import (
	"context"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/infrastructure/rpc"
)

type HelmetConfigQuery interface {
	GetUseCarConfig(ctx context.Context, tenantID string, serviceID int64) (*dto.ConfigUseCarCO, error)
	GetBackcarConfig(ctx context.Context, tenantID string, serviceID int64) (*dto.ConfigBackcarCO, error)
	GetBaseItemConfig(ctx context.Context, tenantID string, serviceID int64) (*dto.ConfigBaseItemCO, error)
}

type HelmetService struct {
	helmetGateway *gateway.HelmetGateway
	managementRpc *rpc.ManagementRPC
	configGateway *gateway.ConfigGateway
	deviceRpc     *rpc.DeviceRPC
	configQuery   HelmetConfigQuery
}

func NewHelmetService(helmetGateway *gateway.HelmetGateway, managementRpc *rpc.ManagementRPC, configGateway *gateway.ConfigGateway, deviceRpc *rpc.DeviceRPC) *HelmetService {
	return &HelmetService{
		helmetGateway: helmetGateway,
		managementRpc: managementRpc,
		configGateway: configGateway,
		deviceRpc:     deviceRpc,
	}
}

func (s *HelmetService) SetConfigQuery(q HelmetConfigQuery) {
	s.configQuery = q
}

func (s *HelmetService) Unlock(ctx context.Context, tenantId string, cmd dto.HelmetCmd) (bool, error) {
	cmdCtx := dto.EnsureCommandContext(cmd.CommandContext, tenantId)
	car, err := s.requireCar(ctx, cmdCtx, cmd.CarId)
	if err != nil {
		return false, err
	}
	return s.helmetGateway.Unlock(ctx, tenantId, cmdCtx, car.Imei, car.CarId)
}

func (s *HelmetService) Lock(ctx context.Context, tenantId string, cmd dto.HelmetCmd) (bool, error) {
	cmdCtx := dto.EnsureCommandContext(cmd.CommandContext, tenantId)
	car, err := s.requireCar(ctx, cmdCtx, cmd.CarId)
	if err != nil {
		return false, err
	}
	return s.helmetGateway.Lock(ctx, tenantId, cmdCtx, car.Imei, car.CarId)
}

func (s *HelmetService) GetRecordHelmetLock(ctx context.Context, tenantId, carId string) (string, error) {
	return s.helmetGateway.GetRecordHelmetLock(ctx, tenantId, carId)
}

func (s *HelmetService) GetRecordHelmetReact(ctx context.Context, tenantId, carId string) (string, error) {
	return s.helmetGateway.GetRecordHelmetReact(ctx, tenantId, carId)
}

func (s *HelmetService) DelHelmet(ctx context.Context, tenantId, carId string) error {
	return s.helmetGateway.DelHelmet(ctx, tenantId, carId)
}

func (s *HelmetService) AccUnlockByHelmetState(ctx context.Context, tenantId string, cmdCtx *dto.CommandContext, imei string, serviceId int64) (bool, error) {
	cmdCtx = dto.EnsureCommandContext(cmdCtx, tenantId)
	useCarCfg, err := s.configGateway.GetUseCarConfig(ctx, tenantId, serviceId)
	if err != nil {
		return true, nil
	}
	helmetCfg := useCarCfg.HelmetConfig
	if helmetCfg == nil {
		return true, nil
	}
	enabled := cfgBool(helmetCfg.IzHelmetUnlock) || cfgBool(helmetCfg.IzHelmetRemovalDetection) ||
		cfgBool(helmetCfg.IzHelmetWearDetection) || cfgBool(helmetCfg.IzRidingHelmetWear)
	if !enabled {
		return true, nil
	}
	deviceInfo, _ := s.deviceRpc.GetDeviceInfo(ctx, cmdCtx, imei)
	if cfgBool(helmetCfg.IzHelmetWearDetection) {
		if deviceInfo != nil && deviceInfo.BleHelmetState != nil && *deviceInfo.BleHelmetState == 1 {
			return true, nil
		}
		return false, nil
	}
	if cfgBool(helmetCfg.IzHelmetRemovalDetection) {
		if deviceInfo != nil && deviceInfo.Helmet6React != nil && *deviceInfo.Helmet6React == 0 {
			return true, nil
		}
		return false, nil
	}
	return true, nil
}

func (s *HelmetService) HelmetStateChange(ctx context.Context, tenantId string, serviceId int64) (dto.ConfigDto, error) {
	if s.configQuery == nil {
		return dto.ConfigDto{}, newBizError(dto.CodeException, "config services not initialized")
	}
	useCar, err := s.configQuery.GetUseCarConfig(ctx, tenantId, serviceId)
	if err != nil {
		return dto.ConfigDto{}, err
	}
	backcar, err := s.configQuery.GetBackcarConfig(ctx, tenantId, serviceId)
	if err != nil {
		return dto.ConfigDto{}, err
	}
	baseItem, err := s.configQuery.GetBaseItemConfig(ctx, tenantId, serviceId)
	if err != nil {
		return dto.ConfigDto{}, err
	}
	return dto.ConfigDto{
		ConfigBaseItemCO: baseItem,
		ConfigBackcarCO:  backcar,
		ConfigUseCarCO:   useCar,
	}, nil
}

func (s *HelmetService) requireCar(ctx context.Context, cmdCtx *dto.CommandContext, carId string) (*rpc.CarInfo, error) {
	car, err := s.managementRpc.GetCarInfoByCarId(ctx, cmdCtx, carId)
	if err != nil {
		return nil, err
	}
	if car == nil || car.Imei == "" {
		return nil, newBizError(dto.CodeException, "车辆信息未查到")
	}
	return car, nil
}

func cfgBool(v *bool) bool {
	return v != nil && *v
}

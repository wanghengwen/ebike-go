package service

import (
	"context"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/geo"
	appconfig "ebike-fence-go/internal/pkg/config"
)

const (
	codeRideCarFail        = "13012"
	codeCanNotRemoteUnlock = "13020"
	rideTypeSuccess        = "RIDE_SUCCESS"
	rideTypePartCheckFail  = "PART_CHECK_FAIL"
)

type RidingCarService struct {
	configGateway        *gateway.ConfigGateway
	partService          *PartService
	parkingDetailGateway *gateway.ParkingDetailGateway
}

func NewRidingCarService(configGateway *gateway.ConfigGateway, partService *PartService, parkingDetailGateway *gateway.ParkingDetailGateway) *RidingCarService {
	return &RidingCarService{
		configGateway:        configGateway,
		partService:          partService,
		parkingDetailGateway: parkingDetailGateway,
	}
}

func (s *RidingCarService) RidingCar(ctx context.Context, tenantId string, req dto.RideCarCmd) (dto.RideCarCO, error) {
	carLng, carLat := resolveRideCarLocation(req)
	userLng, userLat := 0.0, 0.0
	if req.UserLocation != nil {
		userLng = req.UserLocation.Lng
		userLat = req.UserLocation.Lat
	}

	if req.IzRemoteUnlock != nil && !*req.IzRemoteUnlock && req.UserLocation != nil && req.CarLocation != nil {
		dist := geo.PointToPointDistance(userLng, userLat, carLng, carLat)
		remoteLockDistance := 100
		if appconfig.GlobalConfig.Xyy.RemoteLockDistance != nil {
			remoteLockDistance = *appconfig.GlobalConfig.Xyy.RemoteLockDistance
		}
		if dist > float64(remoteLockDistance) {
			return dto.RideCarCO{}, newBizError(codeCanNotRemoteUnlock, "无法远程开启车辆")
		}
	}

	if req.ServiceAreaId == nil {
		return dto.RideCarCO{}, newBizError(codeRideCarFail, "服务区外不可用车")
	}

	serviceArea, err := GetServiceAreaById(ctx, tenantId, *req.ServiceAreaId)
	if err != nil {
		return dto.RideCarCO{}, err
	}
	if serviceArea == nil {
		return dto.RideCarCO{}, newBizError(codeRideCarFail, "服务区外不可用车")
	}

	carLoc := geo.Location{Lng: carLng, Lat: carLat}
	if !geo.IsPointInParsedPolygon(carLoc, serviceArea.ParsedPolygon) {
		return dto.RideCarCO{}, newBizError(codeRideCarFail, "服务区外不可用车")
	}

	banRiding, _ := izInBanRiding(ctx, tenantId, carLoc)
	if banRiding != nil {
		return dto.RideCarCO{}, newBizError(codeRideCarFail, "禁行区内不可用车")
	}

	rideResult := true
	reason := rideTypeSuccess
	var parkingCO map[string]interface{}
	var partResult []dto.PartAnalysisResultCO

	car := req.CarCmd
	if car != nil {
		parkingCO, _ = GetParkingByCarId(ctx, tenantId, car.CarId, s.parkingDetailGateway)

		useCarCfg, err := s.configGateway.GetUseCarConfig(ctx, tenantId, *req.ServiceAreaId)
		if err != nil {
			return dto.RideCarCO{}, err
		}
		izBeacon := useCarCfg.GetIzBeacon()

		if parkingCO == nil && izBeacon {
			backcarConfig, err := s.configGateway.GetConfigByServiceId(ctx, tenantId, serviceArea.Id)
			if err != nil {
				return dto.RideCarCO{}, err
			}
			parking, _ := ParkingReturnCar(ctx, tenantId, carLoc, geo.Location{}, serviceArea.Id, backcarConfig)
			if parking == nil {
				return dto.RideCarCO{}, newBizError(codeRideCarFail, "车辆不在租还区")
			}
			parkingCO = fenceToMap(parking)
			parkingParts := parking.GetParkingParts()

			if containsPart(car.Parts, "bluetooth_beacon") && containsPart(parkingParts, "bluetooth_beacon") {
				beaconParts := []string{"bluetooth_beacon"}
				results, err := s.partService.GetPartResult(ctx, tenantId, car.Imei, car.CarId, beaconParts, PartCheckUseCar, nil, nil, backcarConfig, req.CommandContext)
				if err != nil {
					return dto.RideCarCO{}, err
				}
				partResult = convertPartRes(results)
				izCanUseBeacon := false
				for _, pr := range results {
					if pr.Name == "bluetooth_beacon" && partResultIsTrue(pr.Result) {
						izCanUseBeacon = true
						break
					}
				}
				if !izCanUseBeacon {
					rideResult = false
					reason = rideTypePartCheckFail
				}
			}
		}
	}

	res := dto.RideCarCO{
		RideResult: &rideResult,
		Reason:     reason,
		ParkingCO:  parkingCO,
		PartResult: partResult,
	}
	return res, nil
}

func resolveRideCarLocation(req dto.RideCarCmd) (float64, float64) {
	if req.CarLocation != nil {
		return req.CarLocation.Lng, req.CarLocation.Lat
	}
	if req.CarCmd != nil && req.CarCmd.CarLocation != nil {
		return req.CarCmd.CarLocation.Lng, req.CarCmd.CarLocation.Lat
	}
	return 0, 0
}

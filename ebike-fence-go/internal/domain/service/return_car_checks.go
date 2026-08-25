package service

import (
	"context"
	"strings"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/config"
)

type carPartContext struct {
	tenantId      string
	car           *dto.CarCmd
	checkType     int
	orderId       *int64
	blueResults   []dto.BlueResult
	backcarConfig *config.ConfigBackcarCO
	cmdCtx        *dto.CommandContext
}

func (s *ReturnCarService) checkAssertPart(ctx context.Context, p carPartContext, result *dto.ReturnCarCO, parts []string, parkingId *int64) error {
	if result.IzDeviceWeakNet != nil && *result.IzDeviceWeakNet {
		return nil
	}
	partResults, err := s.resolvePartResults(ctx, p, parts, parkingId)
	if err != nil {
		return err
	}
	if !isReturnCarPartCheck(p.checkType) {
		return nil
	}
	for _, pr := range partResults {
		if partResultIsFalse(pr.Result) {
			rt := getReturnTypeByParts([]PartAnalysisResultCO{pr}, fenceRelationFromCO(result.ReturnType))
			result.ReturnType = string(rt)
			falseVal := false
			result.IzCanReturn = &falseVal
			break
		}
	}
	return nil
}

func (s *ReturnCarService) checkParkingPart(ctx context.Context, p carPartContext, result *dto.ReturnCarCO, parts []string, parkingId int64) error {
	if result.IzDeviceWeakNet != nil && *result.IzDeviceWeakNet {
		return nil
	}
	pid := parkingId
	partResults, err := s.resolvePartResults(ctx, p, parts, &pid)
	if err != nil {
		return err
	}
	result.PartResult = convertPartRes(partResults)
	if !isReturnCarPartCheck(p.checkType) {
		return nil
	}
	for _, pr := range partResults {
		if partResultIsFalse(pr.Result) {
			rt := getReturnTypeByParts([]PartAnalysisResultCO{pr}, fenceRelationFromCO(result.ReturnType))
			result.ReturnType = string(rt)
			falseVal := false
			result.IzCanReturn = &falseVal
			break
		}
	}
	return nil
}

func (s *ReturnCarService) checkPrecisePart(ctx context.Context, p carPartContext, result *dto.ReturnCarCO, parts []string) error {
	if result.IzDeviceWeakNet != nil && *result.IzDeviceWeakNet {
		return nil
	}
	partResults, err := s.resolvePartResults(ctx, p, parts, nil)
	if err != nil {
		return err
	}
	result.PartResult = convertPartRes(partResults)
	if !isReturnCarPartCheck(p.checkType) {
		return nil
	}
	for _, pr := range partResults {
		if partResultIsTrue(pr.Result) {
			result.ReturnType = string(IN_PARKING_PART)
			trueVal := true
			result.IzCanReturn = &trueVal
			break
		}
	}
	return nil
}

func (s *ReturnCarService) resolvePartResults(ctx context.Context, p carPartContext, parts []string, parkingId *int64) ([]PartAnalysisResultCO, error) {
	if p.checkType == PartCheckGetPart {
		var out []PartAnalysisResultCO
		for _, name := range parts {
			out = append(out, PartAnalysisResultCO{Name: name})
		}
		return out, nil
	}
	if isBluePartCheck(p.checkType) {
		return handleBlueResult(parts, p.blueResults), nil
	}
	return s.getPartResultSafe(ctx, p.tenantId, p.car, parts, p.checkType, parkingId, p.orderId, p.backcarConfig, p.cmdCtx)
}

func (s *ReturnCarService) getPartResultSafe(ctx context.Context, tenantId string, car *dto.CarCmd, parts []string, checkType int, parkingId *int64, orderId *int64, backcarConfig *config.ConfigBackcarCO, cmdCtx *dto.CommandContext) ([]PartAnalysisResultCO, error) {
	res, err := s.partService.GetPartResult(ctx, tenantId, car.Imei, car.CarId, parts, checkType, parkingId, orderId, backcarConfig, cmdCtx)
	if err != nil {
		if biz, ok := err.(*BizError); ok && biz.Code == codeDeviceTimeout {
			if checkType == PartCheckFocusReturnCar || checkType == PartCheckAutoReturnCar || checkType == PartCheckPaybackReturn {
				return nil, nil
			}
		}
		return nil, err
	}
	return res, nil
}

func handleBlueResult(parts []string, blue []dto.BlueResult) []PartAnalysisResultCO {
	if len(blue) == 0 {
		return nil
	}
	var out []PartAnalysisResultCO
	for _, item := range blue {
		out = append(out, PartAnalysisResultCO{Name: item.Name, Result: partResultBool(handlePartState(item))})
	}
	return out
}

func handlePartState(blue dto.BlueResult) bool {
	if blue.Name != "bluetooth_beacon" {
		return false
	}
	if blue.State == nil {
		return false
	}
	if ev, ok := blue.State["event"].(float64); ok {
		return int(ev) != 0
	}
	return false
}

func getReturnTypeByParts(partResult []PartAnalysisResultCO, returnType FenceRelation) FenceRelation {
	for _, pr := range partResult {
		if pr.Name == "helmet" {
			switch returnType {
			case IN_PARKING_LAT, IN_PARKING_PART:
				return IN_PARKING_HELMET
			case IN_NO_PARKING:
				return IN_NO_PARKING_HELMET
			case OUT_PARKING_LAT:
				return OUT_PARKING_HELMET
			case OUT_SERVICE_AREA:
				return OUT_SERVICE_HELMET
			}
		} else {
			return relationByPartName(pr.Name)
		}
	}
	return returnType
}

func relationByPartName(partName string) FenceRelation {
	switch partName {
	case "rfid_beacon":
		return IN_PARKING_RFID
	case "direction":
		return IN_PARKING_DIRECTION
	case "kickstand":
		return IN_PARKING_KICK
	case "camera":
		return IN_PARKING_CAMERA
	case "camera_directional":
		return IN_PARKING_CAMERA_DIRECTIONAL
	case "camera_point":
		return IN_PARKING_CAMERA_POINT
	default:
		return OUT_PARKING_LAT
	}
}

func fenceRelationFromCO(v interface{}) FenceRelation {
	if s, ok := v.(string); ok {
		return FenceRelation(s)
	}
	return OUT_PARKING_LAT
}

func getReturnCarCO(result *dto.ReturnCarCO, backcarConfig *config.ConfigBackcarCO, locationAudit bool) *dto.ReturnCarCO {
	rt := fenceRelationFromCO(result.ReturnType)
	can := izCanReturn(rt, backcarConfig)
	result.IzCanReturn = can
	adjusted := applyPenaltyConfigToReturnType(rt, backcarConfig, locationAudit)
	result.ReturnType = string(adjusted)
	return result
}

func getPreciseParts(parts []string) []string {
	var out []string
	for _, p := range parts {
		if strings.Contains(precisePartsCSV, p) {
			out = append(out, p)
		}
	}
	return out
}

func processPartResults(partRes []PartAnalysisResultCO, currentType FenceRelation) FenceRelation {
	for _, p := range partRes {
		if partResultIsFalse(p.Result) {
			return getReturnTypeByParts([]PartAnalysisResultCO{p}, currentType)
		}
	}
	return currentType
}

func containsPart(parts []string, name string) bool {
	for _, p := range parts {
		if p == name {
			return true
		}
	}
	return false
}

package fenceadmin

import (
	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/config"
	"ebike-fence-go/internal/domain/gateway"
)

func stripFenceAudit(co *dto.FenceCO) {
	co.Pics = ""
	co.TenantId = ""
	co.CreatedPin = ""
	co.CreatedAt = ""
	co.UpdatedPin = ""
	co.UpdatedAt = ""
	co.OpeningHoursBegin = ""
	co.OpeningHoursEnd = ""
}

func toNearFenceCO(fe gateway.FenceE) dto.FenceCO {
	co := toFenceCO(fe)
	stripFenceAudit(&co)
	return co
}

func toNearParkingCO(fe gateway.FenceE) dto.ParkingCO {
	co := toParkingCO(fe)
	normalizeNearParkingCO(&co, fe)
	stripFenceAudit(&co.FenceCO)
	return co
}

func toNearNoParkingCO(fe gateway.FenceE) dto.NoParkingCO {
	co := toNoParkingCO(fe)
	stripFenceAudit(&co.FenceCO)
	return co
}

func toNearBanRidingCO(fe gateway.FenceE) dto.BanRidingCO {
	co := toBanRidingCO(fe)
	stripFenceAudit(&co.FenceCO)
	return co
}

func emptyBanRidingNearest() map[string]interface{} {
	return map[string]interface{}{
		"type":              9,
		"id":                nil,
		"name":              nil,
		"shapeType":         nil,
		"centerLat":         nil,
		"centerLng":         nil,
		"pointList":         nil,
		"pics":              nil,
		"createdPin":        nil,
		"createdAt":         nil,
		"updatedPin":        nil,
		"updatedAt":         nil,
		"openingHoursBegin": nil,
		"openingHoursEnd":   nil,
		"tenantId":          nil,
		"serviceId":         nil,
		"distance":          nil,
	}
}

func toClientServiceAreaCO(fe gateway.FenceE) dto.ServiceAreaCO {
	co := toServiceAreaCO(fe)
	co.Pics = ""
	co.CreatedPin = ""
	co.CreatedAt = ""
	co.UpdatedPin = ""
	co.UpdatedAt = ""
	co.OpeningHoursBegin = ""
	co.OpeningHoursEnd = ""
	return co
}

func normalizeOpeningHours(s string) string {
	if len(s) == 5 {
		return s + ":00"
	}
	return s
}

// normalizeOpeningHoursPtr returns a *string for ParkingCO (nullable), nil if empty.
func normalizeOpeningHoursPtr(s string) *string {
	if s == "" {
		return nil
	}
	v := normalizeOpeningHours(s)
	return &v
}

func toFenceCO(fe gateway.FenceE) dto.FenceCO {
	co := dto.FenceCO{
		Id:                fe.Id,
		Type:              fe.Type,
		Name:              fe.Name,
		ShapeType:         fe.ShapeType,
		CenterLat:         fe.CenterLat,
		CenterLng:         fe.CenterLng,
		PointList:         fe.PointList,
		OpeningHoursBegin: normalizeOpeningHours(fe.OpeningHoursBegin),
		OpeningHoursEnd:   normalizeOpeningHours(fe.OpeningHoursEnd),
		TenantId:          fe.TenantId,
		CreatedPin:        fe.CreatedPin,
		CreatedAt:         fe.CreatedAt,
		UpdatedPin:        fe.UpdatedPin,
		UpdatedAt:         fe.UpdatedAt,
	}
	return co
}

func toServiceAreaCO(fe gateway.FenceE) dto.ServiceAreaCO {
	return dto.ServiceAreaCO{FenceCO: toFenceCO(fe), DataVersion: fe.DataVersion}
}

func toParkingCO(fe gateway.FenceE) dto.ParkingCO {
	co := dto.ParkingCO{
		FenceCO:                toFenceCO(fe),
		MaxParkingNumber:       fe.MaxParkingNumber,
		ServiceId:              fe.ServiceId,
		Tbeacon:                fe.Tbeacon,
		Directional:            fe.Directional,
		Direction:              fe.Direction,
		Rfid:                   fe.Rfid,
		IzEnable:                   fe.IzEnable,
		Camera:                     fe.Camera,
		Kickstand:              fe.Kickstand,
		IzFullPileNoStop:       fe.IzFullPileNoStop,
		OpeningHoursBegin:      normalizeOpeningHoursPtr(fe.OpeningHoursBegin),
		OpeningHoursEnd:        normalizeOpeningHoursPtr(fe.OpeningHoursEnd),
		IzCameraDirectionalBackcar: fe.IzCameraDirectionalBackcar,
		IzCameraPointBackcar:       fe.IzCameraPointBackcar,
	}
	applyParkingNullableFields(&co, fe)
	if fe.CoefficientOfDifficultSet {
		cf := fe.CoefficientOfDifficult
		co.CoefficientOfDifficult = &cf
	}
	if fe.IzCameraDirectionalBackcar == nil {
		f := false
		co.IzCameraDirectionalBackcar = &f
	}
	if fe.IzCameraPointBackcar == nil {
		f := false
		co.IzCameraPointBackcar = &f
	}
	applyParkingStats(&co)
	return co
}

func applyParkingNullableFields(co *dto.ParkingCO, fe gateway.FenceE) {
	if fe.BufferDistanceSet {
		v := fe.BufferDistance
		co.BufferDistance = &v
	}
	if fe.AreaSizeSet {
		v := fe.AreaSize
		co.AreaSize = &v
	}
	if fe.MinAmountSet {
		co.MinAmount = fe.MinAmount
	}
	if fe.MaxAmountSet {
		co.MaxAmount = fe.MaxAmount
	}
}

func normalizeNearParkingCO(co *dto.ParkingCO, fe gateway.FenceE) {
	applyParkingNullableFields(co, fe)
	if !fe.CoefficientOfDifficultSet {
		co.CoefficientOfDifficult = nil
	}
	if co.OpeningHoursBegin != nil && len(*co.OpeningHoursBegin) == 5 {
		v := *co.OpeningHoursBegin + ":00"
		co.OpeningHoursBegin = &v
	}
	if co.OpeningHoursEnd != nil && len(*co.OpeningHoursEnd) == 5 {
		v := *co.OpeningHoursEnd + ":00"
		co.OpeningHoursEnd = &v
	}
	f := false
	if co.Tbeacon == nil {
		co.Tbeacon = &f
	}
	if co.Directional == nil {
		co.Directional = &f
	}
	if co.Rfid == nil {
		co.Rfid = &f
	}
	if co.Camera == nil {
		co.Camera = &f
	}
	if co.Kickstand == nil {
		co.Kickstand = &f
	}
	co.FullCar = &f
}

func includeNearParking(fe gateway.FenceE, serviceID int64, backcar *config.ConfigBackcarCO) bool {
	if backcar != nil && backcar.GetIzUseOtherParking() {
		return true
	}
	return fe.ServiceId == serviceID && fe.IzEnable
}

func includeNearNoParking(fe gateway.FenceE, serviceID int64, backcar *config.ConfigBackcarCO) bool {
	if backcar != nil && backcar.GetIzUseOtherParking() {
		return true
	}
	return fe.ServiceId == serviceID
}

func parkingFullCar(fe gateway.FenceE, carCount int64, backcar *config.ConfigBackcarCO) bool {
	izFullPileNoStop := 0
	if fe.IzFullPileNoStop != nil {
		izFullPileNoStop = *fe.IzFullPileNoStop
	} else if backcar != nil && backcar.IzFullPileNoStop != nil {
		izFullPileNoStop = *backcar.IzFullPileNoStop
	}
	maxParkingNumber := fe.MaxParkingNumber
	if maxParkingNumber == 0 {
		maxParkingNumber = 10
	}
	return izFullPileNoStop == 1 && carCount >= int64(maxParkingNumber)
}

func applyParkingStats(co *dto.ParkingCO) {
	co.UseCount = 0
	co.OfflineCount = 0
	co.RepairCount = 0
	co.OrderCount = 0
	co.OrderCost = 0
}

func setParkingCarCount(co *dto.ParkingCO, count int64, always bool) {
	co.CarCount = &count
}

func toNoParkingCO(fe gateway.FenceE) dto.NoParkingCO {
	return dto.NoParkingCO{FenceCO: toFenceCO(fe), ServiceId: fe.ServiceId}
}

func toBanRidingCO(fe gateway.FenceE) dto.BanRidingCO {
	return dto.BanRidingCO{FenceCO: toFenceCO(fe), ServiceId: fe.ServiceId}
}

func toMaintainAreaCO(fe gateway.FenceE) dto.MaintainAreaCO {
	return dto.MaintainAreaCO{
		FenceCO:          toFenceCO(fe),
		ServiceId:        fe.ServiceId,
		AreaSize:         fe.AreaSize,
		MaxParkingNumber: fe.MaxParkingNumber,
	}
}

func toFenceCustomCO(fe gateway.FenceE) dto.FenceCustomCO {
	return dto.FenceCustomCO{FenceCO: toFenceCO(fe), ServiceId: fe.ServiceId, CustomTypeId: fe.CustomTypeId}
}

func cmdToFenceE(fenceType int, name, shapeType string, centerLat, centerLng float64, pointList string, serviceID int64) gateway.FenceE {
	izEnable := true
	return gateway.FenceE{
		Type:      fenceType,
		Name:      name,
		ShapeType: shapeType,
		CenterLat: centerLat,
		CenterLng: centerLng,
		PointList: pointList,
		ServiceId: serviceID,
		IzEnable:  izEnable,
	}
}

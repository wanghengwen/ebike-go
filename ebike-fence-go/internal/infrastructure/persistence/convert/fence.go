package convert

import (
	"database/sql"
	"time"

	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/geo"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/pkg/timefmt"
)

func formatFenceTime(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return timefmt.FormatJavaLocal(t.Time)
}

func FenceToEntity(row model.TFence) gateway.FenceE {
	fe := gateway.FenceE{
		Id:                row.ID,
		Name:              row.Name,
		ShapeType:         row.ShapeType,
		CenterLat:         row.CenterLat,
		CenterLng:         row.CenterLng,
		PointList:         row.PointList,
		Type:              row.Type,
		ServiceId:         row.ServiceID,
		TenantId:          row.TenantID,
		MaxParkingNumber:  row.MaxParkingNumber,
		OpeningHoursBegin: nullString(row.OpeningHoursBegin),
		OpeningHoursEnd:   nullString(row.OpeningHoursEnd),
		CreatedPin:        nullString(row.CreatedPin),
		CreatedAt:         formatFenceTime(row.CreatedAt),
		UpdatedPin:        nullString(row.UpdatedPin),
		UpdatedAt:         formatFenceTime(row.UpdatedAt),
	}
	if row.DataVersion.Valid {
		fe.DataVersion = row.DataVersion.Int64
	}
	if row.BufferDistance.Valid {
		fe.BufferDistance = row.BufferDistance.Float64
		fe.BufferDistanceSet = true
	}
	if row.CoefficientOfDifficult.Valid {
		fe.CoefficientOfDifficult = row.CoefficientOfDifficult.Float64
		fe.CoefficientOfDifficultSet = true
	}
	if row.IzFullPileNoStop.Valid {
		v := int(row.IzFullPileNoStop.Int32)
		fe.IzFullPileNoStop = &v
	}
	if row.IzEnable.Valid {
		fe.IzEnable = row.IzEnable.Bool
	} else {
		fe.IzEnable = true
	}
	if row.IzOpenAllDay.Valid {
		v := row.IzOpenAllDay.Bool
		fe.IzOpenAllDay = &v
	}
	fe.Tbeacon = nullBoolPtr(row.Tbeacon)
	fe.Directional = nullBoolPtr(row.Directional)
	fe.Rfid = nullBoolPtr(row.Rfid)
	fe.Camera = nullBoolPtr(row.Camera)
	fe.Kickstand = nullBoolPtr(row.Kickstand)
	if row.Direction.Valid {
		v := row.Direction.Float64
		fe.Direction = &v
	}
	if row.FormulateDirection.Valid {
		v := row.FormulateDirection.Float64
		fe.FormulateDirection = &v
	}
	if row.CustomTypeID.Valid {
		fe.CustomTypeId = row.CustomTypeID.Int64
	}
	if row.AreaSize.Valid {
		fe.AreaSize = row.AreaSize.Float64
		fe.AreaSizeSet = true
	}
	if row.MinAmount.Valid {
		v := int(row.MinAmount.Int32)
		fe.MinAmount = &v
		fe.MinAmountSet = true
	}
	if row.MaxAmount.Valid {
		v := int(row.MaxAmount.Int32)
		fe.MaxAmount = &v
		fe.MaxAmountSet = true
	}
	fe.IzCameraDirectionalBackcar = nullBoolPtr(row.IzCameraDirectionalBackcar)
	fe.IzCameraPointBackcar = nullBoolPtr(row.IzCameraPointBackcar)
	if parsed, err := geo.ParsePolygon(fe.PointList); err == nil {
		fe.ParsedPolygon = parsed
	}
	return fe
}

func EntityToModel(tenantID, pin string, fe gateway.FenceE) model.TFence {
	row := model.TFence{
		ID:                     fe.Id,
		TenantID:               tenantID,
		Type:                   fe.Type,
		Name:                   fe.Name,
		ShapeType:              fe.ShapeType,
		MaxParkingNumber:       fe.MaxParkingNumber,
		CenterLat:              fe.CenterLat,
		CenterLng:              fe.CenterLng,
		PointList:              fe.PointList,
		ServiceID:              fe.ServiceId,
		OpeningHoursBegin:      sql.NullString{String: fe.OpeningHoursBegin, Valid: fe.OpeningHoursBegin != ""},
		OpeningHoursEnd:        sql.NullString{String: fe.OpeningHoursEnd, Valid: fe.OpeningHoursEnd != ""},
		BufferDistance:         sql.NullFloat64{Float64: fe.BufferDistance, Valid: fe.BufferDistance > 0},
		CoefficientOfDifficult: sql.NullFloat64{Float64: fe.CoefficientOfDifficult, Valid: fe.CoefficientOfDifficult > 0},
		IzEnable:               sql.NullBool{Bool: fe.IzEnable, Valid: true},
	}
	if fe.IzFullPileNoStop != nil {
		row.IzFullPileNoStop = sql.NullInt32{Int32: int32(*fe.IzFullPileNoStop), Valid: true}
	}
	if fe.IzOpenAllDay != nil {
		row.IzOpenAllDay = sql.NullBool{Bool: *fe.IzOpenAllDay, Valid: true}
	}
	row.Tbeacon = boolToNull(fe.Tbeacon)
	row.Directional = boolToNull(fe.Directional)
	row.Rfid = boolToNull(fe.Rfid)
	row.Camera = boolToNull(fe.Camera)
	row.Kickstand = boolToNull(fe.Kickstand)
	if fe.Direction != nil {
		row.Direction = sql.NullFloat64{Float64: *fe.Direction, Valid: true}
	}
	if fe.FormulateDirection != nil {
		row.FormulateDirection = sql.NullFloat64{Float64: *fe.FormulateDirection, Valid: true}
	}
	if fe.CustomTypeId > 0 {
		row.CustomTypeID = sql.NullInt64{Int64: fe.CustomTypeId, Valid: true}
	}
	if fe.AreaSize > 0 {
		row.AreaSize = sql.NullFloat64{Float64: fe.AreaSize, Valid: true}
	}
	if pin != "" {
		row.UpdatedPin = sql.NullString{String: pin, Valid: true}
	}
	row.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	return row
}

func boolToNull(v *bool) sql.NullBool {
	if v == nil {
		return sql.NullBool{}
	}
	return sql.NullBool{Bool: *v, Valid: true}
}

func nullString(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func nullBoolPtr(v sql.NullBool) *bool {
	if !v.Valid {
		return nil
	}
	b := v.Bool
	return &b
}

package repo

import (
	"database/sql"

	"ebike-fence-go/internal/infrastructure/persistence/model"
)

// NewDefaultUseCarRow mirrors Java/MyBatis insert with only serviceId set — DB NOT NULL
// columns must be populated explicitly because GORM writes NULL for unset sql.Null* fields.
func NewDefaultUseCarRow(id, serviceID int64) *model.ConfigUseCar {
	return &model.ConfigUseCar{
		ID:                          id,
		ServiceID:                   serviceID,
		RechargeBeforeUse:           sql.NullBool{Bool: false, Valid: true},
		RechargeVisibleRange:        sql.NullInt32{Int32: 0, Valid: true},
		RechargeByRegister:          sql.NullBool{Bool: true, Valid: true},
		RechargeByTags:              sql.NullBool{Bool: false, Valid: true},
		RechargeCost:                sql.NullInt32{Int32: 0, Valid: true},
		IzStopService:               sql.NullBool{Bool: false, Valid: true},
		IzAutoRecovery:              sql.NullBool{Bool: false, Valid: true},
		IzOnCertification:           sql.NullBool{Bool: false, Valid: true},
		IzOnUseCar:                  sql.NullBool{Bool: false, Valid: true},
		RecognitionDegree:           sql.NullInt32{Int32: 80, Valid: true},
		EffectiveTime:               sql.NullInt32{Int32: 1, Valid: true},
		IzRidingStopTrigger:         sql.NullBool{Bool: false, Valid: true},
		IzParkingTriggerReturnBike:  sql.NullBool{Bool: false, Valid: true},
		IzBeacon:                    sql.NullBool{Bool: false, Valid: true},
		NearLine:                    sql.NullInt32{Int32: 200, Valid: true},
		OutServiceAreaAutoLock:      sql.NullInt32{Int32: 10, Valid: true},
		IzRemoteUnlock:              sql.NullBool{Bool: true, Valid: true},
		IzOpenSaddleOverloadMonitor: sql.NullBool{Bool: false, Valid: true},
		OverloadRemind:              sql.NullInt32{Int32: 0, Valid: true},
		IzOrderNotice:               sql.NullBool{Bool: false, Valid: true},
		NoticeRidingTime:            sql.NullInt32{Int32: 60, Valid: true},
		MinAge:                      sql.NullInt32{Int32: 16, Valid: true},
		MaxAge:                      sql.NullInt32{Int32: 65, Valid: true},
		IzAuth:                      sql.NullBool{Bool: true, Valid: true},
	}
}

func NewDefaultPayRow(id, serviceID int64) *model.ConfigPay {
	return &model.ConfigPay{
		ID:                        id,
		ServiceID:                 serviceID,
		IzBalanceEnoughReturnBike: sql.NullBool{Bool: false, Valid: true},
		IzMakeupBalancePay:        sql.NullBool{Bool: false, Valid: true},
		IzNotifyUnpaidOrder:       sql.NullBool{Bool: false, Valid: true},
	}
}

func NewDefaultParkApplyRow(id, serviceID int64) *model.ParkApplyConfig {
	return &model.ParkApplyConfig{
		ID:          id,
		ServiceID:   serviceID,
		IzApplyPark: sql.NullInt32{Int32: 1, Valid: true},
	}
}

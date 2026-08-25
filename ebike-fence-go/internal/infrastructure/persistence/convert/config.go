package convert

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/pkg/timefmt"
)

func nullIntPtr(v sql.NullInt32) *int {
	if !v.Valid {
		return nil
	}
	i := int(v.Int32)
	return &i
}

func nullInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}

func nullFloatPtr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	f := v.Float64
	return &f
}

func nullStrPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func normalizeTimeStrPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	if strings.Count(s, ":") == 1 {
		s = s + ":00"
	}
	return &s
}

func BackcarToCO(m *model.ConfigBackcar) *dto.ConfigBackcarCO {
	if m == nil {
		return nil
	}
	id, sid := m.ID, m.ServiceID
	return &dto.ConfigBackcarCO{
		Id:                         &id,
		ServiceId:                  &sid,
		AllowOutofService:          nullBoolPtr(m.AllowOutofService),
		AllowInNostop:              nullBoolPtr(m.AllowInNostop),
		AllowOutofParking:          nullBoolPtr(m.AllowOutofParking),
		AllowInBanRiding:           nullBoolPtr(m.AllowInBanRiding),
		DispatchCost:               nullIntPtr(m.DispatchCost),
		PenaltyInNostop:            nullIntPtr(m.PenaltyInNostop),
		PenaltyOutofService:        nullIntPtr(m.PenaltyOutofService),
		PenaltyInBanRiding:         nullIntPtr(m.PenaltyInBanRiding),
		IzCivilizationRemind:       nullBoolPtr(m.IzCivilizationRemind),
		CoefficientOfDifficult:     nullFloatPtr(m.CoefficientOfDifficult),
		BufferDistance:             nullFloatPtr(m.BufferDistance),
		ParkingBackcar:             nullIntPtr(m.ParkingBackcar),
		NoParkingBackcar:           nullIntPtr(m.NoParkingBackcar),
		BanRidingBackcar:           nullIntPtr(m.BanRidingBackcar),
		IzHelmetLock:               nullBoolPtr(m.IzHelmetLock),
		IzHelmetPhoto:              nullBoolPtr(m.IzHelmetPhoto),
		IzBeacon:                   nullBoolPtr(m.IzBeacon),
		IzHelmetRemovalDetection:   nullBoolPtr(m.IzHelmetRemovalDetection),
		IzHelmetWearDetection:      nullBoolPtr(m.IzHelmetWearDetection),
		IzHelmetAutoRepair:         nullBoolPtr(m.IzHelmetAutoRepair),
		PowerOff:                   nullIntPtr(m.PowerOff),
		HelmetPenalty:              nullIntPtr(m.HelmetPenalty),
		IzHelmetReign:              nullBoolPtr(m.IzHelmetReign),
		IzCameraBackcar:            nullBoolPtr(m.IzCameraBackcar),
		IzCameraDirectionalBackcar: nullBoolPtr(m.IzCameraDirectionalBackcar),
		IzCameraPointBackcar:       nullBoolPtr(m.IzCameraPointBackcar),
		CameraImpunityCount:        nullIntPtr(m.CameraImpunityCount),
		IzFullPileNoStop:           nullIntPtr(m.IzFullPileNoStop),
		IzUseOtherParking:          nullBoolPtr(m.IzUseOtherParking),
		BeaconEffect:               nullIntPtr(m.BeaconEffect),
		RfidEffect:                 nullIntPtr(m.RfidEffect),
		DirectionEffect:            nullIntPtr(m.DirectionEffect),
		FootSupportEffect:          nullIntPtr(m.FootSupportEffect),
		CameraEffect:               nullIntPtr(m.CameraEffect),
		CameraAngle:                nullIntPtr(m.CameraAngle),
		IzStandardReturn:           nullIntPtr(m.IzStandardReturn),
	}
}

func ApplyBackcarDefaults(co *dto.ConfigBackcarCO) {
	if co == nil {
		return
	}
	if co.IzHelmetAutoRepair == nil {
		f := false
		co.IzHelmetAutoRepair = &f
	}
	zero := 0
	if co.BeaconEffect == nil {
		co.BeaconEffect = &zero
	}
	if co.RfidEffect == nil {
		co.RfidEffect = &zero
	}
	if co.DirectionEffect == nil {
		co.DirectionEffect = &zero
	}
	if co.FootSupportEffect == nil {
		co.FootSupportEffect = &zero
	}
	if co.CameraEffect == nil {
		co.CameraEffect = &zero
	}
	if co.IzStandardReturn == nil {
		co.IzStandardReturn = &zero
	}
}

func PayToCO(m *model.ConfigPay) *dto.ConfigPayCO {
	if m == nil {
		return nil
	}
	id, sid := m.ID, m.ServiceID
	return &dto.ConfigPayCO{
		Id:                        &id,
		ServiceId:                 &sid,
		IzBalanceEnoughReturnBike: nullBoolPtr(m.IzBalanceEnoughReturnBike),
		IzMakeupBalancePay:        nullBoolPtr(m.IzMakeupBalancePay),
		IzNotifyUnpaidOrder:       nullBoolPtr(m.IzNotifyUnpaidOrder),
		NotifyInterval:            nullIntPtr(m.NotifyInterval),
		TimesUpperBound:           nullIntPtr(m.TimesUpperBound),
		RemindWay:                 nullStrPtr(m.RemindWay),
	}
}

func UseCarToCO(m *model.ConfigUseCar) *dto.ConfigUseCarCO {
	if m == nil {
		return nil
	}
	id, sid := m.ID, m.ServiceID
	co := &dto.ConfigUseCarCO{
		Id:                         &id,
		ServiceId:                  &sid,
		RechargeBeforeUse:          nullBoolPtr(m.RechargeBeforeUse),
		RechargeVisibleRange:       nullIntPtr(m.RechargeVisibleRange),
		RechargeByRegister:         nullBoolPtr(m.RechargeByRegister),
		RechargeByTags:             nullBoolPtr(m.RechargeByTags),
		RechargeTagIds:             nullStrPtr(m.RechargeTagIds),
		RechargeCost:               nullIntPtr(m.RechargeCost),
		IzStopService:              nullBoolPtr(m.IzStopService),
		StopTimeStart:              normalizeTimeStrPtr(m.StopTimeStart),
		StopTimeEnd:                normalizeTimeStrPtr(m.StopTimeEnd),
		IzAutoRecovery:             nullBoolPtr(m.IzAutoRecovery),
		StopServiceNotice:          nullStrPtr(m.StopServiceNotice),
		IzOnCertification:          nullBoolPtr(m.IzOnCertification),
		IzOnUseCar:                 nullBoolPtr(m.IzOnUseCar),
		RecognitionDegree:          nullIntPtr(m.RecognitionDegree),
		EffectiveTime:              nullIntPtr(m.EffectiveTime),
		IzRidingStopTrigger:        nullBoolPtr(m.IzRidingStopTrigger),
		IzParkingTriggerReturnBike: nullBoolPtr(m.IzParkingTriggerReturnBike),
		RidingStopTime:             nullIntPtr(m.RidingStopTime),
		RidingStopEvent:            nullIntPtr(m.RidingStopEvent),
		ParkingTime:                nullIntPtr(m.ParkingTime),
		RemindWay:                  nullStrPtr(m.RemindWay),
		IzBeacon:                   nullBoolPtr(m.IzBeacon),
		OutServiceAreaAutoLock:     nullIntPtr(m.OutServiceAreaAutoLock),
		OutServiceAreaAutoLockRemindWay: nullStrPtr(m.OutServiceAreaAutoLockRemindWay),
		IzOpenSaddleOverloadMonitor: nullBoolPtr(m.IzOpenSaddleOverloadMonitor),
		OverloadRemind:             nullIntPtr(m.OverloadRemind),
		IzRemoteUnlock:             nullBoolPtr(m.IzRemoteUnlock),
		HelmetConfig:               nullStrPtr(m.HelmetConfig),
		IzOrderNotice:              nullBoolPtr(m.IzOrderNotice),
		NoticeRidingTime:           nullIntPtr(m.NoticeRidingTime),
		OrderRemindWay:             nullStrPtr(m.OrderRemindWay),
		MinAge:                     nullIntPtr(m.MinAge),
		MaxAge:                     nullIntPtr(m.MaxAge),
		HideCarConfig:              nullStrPtr(m.HideCarConfig),
		IzAuth:                     nullBoolPtr(m.IzAuth),
	}
	if m.RecoveryData.Valid {
		s := m.RecoveryData.Time.UTC().Format("2006-01-02T15:04:05")
		co.RecoveryData = &s
	}
	nearLine := 200
	if m.NearLine.Valid {
		nearLine = int(m.NearLine.Int32)
	}
	co.NearLine = &nearLine
	return co
}

func BaseItemToCO(m *model.ConfigBaseItem) *dto.ConfigBaseItemCO {
	if m == nil {
		return nil
	}
	id, sid := m.ID, m.ServiceID
	co := &dto.ConfigBaseItemCO{
		Id:                     &id,
		ServiceId:              &sid,
		OfflineTicketJudgeTime: nullIntPtr(m.OfflineTicketJudgeTime),
		MsgSign:                nullStrPtr(m.MsgSign),
		MsgCode:                nullStrPtr(m.MsgCode),
		IzOnAbnormalMovement:   nullBoolPtr(m.IzOnAbnormalMovement),
		IzOnBatteryRemoval:     nullBoolPtr(m.IzOnBatteryRemoval),
		IzOfflineTicket:        nullBoolPtr(m.IzOfflineTicket),
		IzAutoRepairTicket:     nullBoolPtr(m.IzAutoRepairTicket),
		SwapBatteryThreshold:   nullIntPtr(m.SwapBatteryThreshold),
		IzAutoSwapBattery:      nullBoolPtr(m.IzAutoSwapBattery),
		IzOpenInvoice:          nullBoolPtr(m.IzOpenInvoice),
		IzTempUnlock:           nullBoolPtr(m.IzTempUnlock),
		TempUnlockTime:         nullIntPtr(m.TempUnlockTime),
		DashboardSwitch:        nullBoolPtr(m.DashboardSwitch),
		DashboardStartTime:     nullStrPtr(m.DashboardStartTime),
		DashboardEndTime:       nullStrPtr(m.DashboardEndTime),
		IzCanInitiativeRepair:  nullBoolPtr(m.IzCanInitiativeRepair),
		IzCanAfterRidingRepair: nullBoolPtr(m.IzCanAfterRidingRepair),
		CanRepairCon:           nullIntPtr(m.CanRepairCon),
		IzHelmetAbnormalRiding: nullBoolPtr(m.IzHelmetAbnormalRiding),
		IzEnableMoveAlarm:      nullBoolPtr(m.IzEnableMoveAlarm),
		MoveAlarmTime:          nullIntPtr(m.MoveAlarmTime),
		MoveAlarmCount:         nullIntPtr(m.MoveAlarmCount),
		MoveAlarmDistance:      nullIntPtr(m.MoveAlarmDistance),
		IzCreateShortOrder:     nullBoolPtr(m.IzCreateShortOrder),
		ShortOrderDuration:     nullIntPtr(m.ShortOrderDuration),
		ShortOrderCon:          nullIntPtr(m.ShortOrderCon),
		IzWithdraw:             nullBoolPtr(m.IzWithdraw),
	}
	if m.UserTicketPhotoWays.Valid && m.UserTicketPhotoWays.String != "" {
		parts := strings.Split(m.UserTicketPhotoWays.String, ",")
		ways := make([]int, 0, len(parts))
		for _, p := range parts {
			if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
				ways = append(ways, n)
			}
		}
		co.UserTicketPhotoWays = ways
	} else {
		co.UserTicketPhotoWays = []int{0}
	}
	return co
}

// baseItemRedisDO mirrors Java ConfigBaseItemDO JSON in Redis (Fastjson camelCase).
type baseItemRedisDO struct {
	ID                     int64   `json:"id"`
	TenantID               string  `json:"tenantId,omitempty"`
	ServiceID              int64   `json:"serviceId"`
	OfflineTicketJudgeTime *int    `json:"offlineTicketJudgeTime,omitempty"`
	MsgSign                *string `json:"msgSign,omitempty"`
	MsgCode                *string `json:"msgCode,omitempty"`
	IzOnAbnormalMovement   *bool   `json:"izOnAbnormalMovement,omitempty"`
	IzOnBatteryRemoval     *bool   `json:"izOnBatteryRemoval,omitempty"`
	IzOfflineTicket        *bool   `json:"izOfflineTicket,omitempty"`
	IzAutoRepairTicket     *bool   `json:"izAutoRepairTicket,omitempty"`
	IzAutoSwapBattery      *bool   `json:"izAutoSwapBattery,omitempty"`
	SwapBatteryThreshold   *int    `json:"swapBatteryThreshold,omitempty"`
	IzOpenInvoice          *bool   `json:"izOpenInvoice,omitempty"`
	IzTempUnlock           *bool   `json:"izTempUnlock,omitempty"`
	TempUnlockTime         *int    `json:"tempUnlockTime,omitempty"`
	DashboardSwitch        *bool   `json:"dashboardSwitch,omitempty"`
	DashboardStartTime     *string `json:"dashboardStartTime,omitempty"`
	DashboardEndTime       *string `json:"dashboardEndTime,omitempty"`
	IzCanInitiativeRepair  *bool   `json:"izCanInitiativeRepair,omitempty"`
	IzCanAfterRidingRepair *bool   `json:"izCanAfterRidingRepair,omitempty"`
	CanRepairCon           *int    `json:"canRepairCon,omitempty"`
	IzHelmetAbnormalRiding *bool   `json:"izHelmetAbnormalRiding,omitempty"`
	IzEnableMoveAlarm      *bool   `json:"izEnableMoveAlarm,omitempty"`
	MoveAlarmTime          *int    `json:"moveAlarmTime,omitempty"`
	MoveAlarmCount         *int    `json:"moveAlarmCount,omitempty"`
	MoveAlarmDistance      *int    `json:"moveAlarmDistance,omitempty"`
	IzCreateShortOrder     *bool   `json:"izCreateShortOrder,omitempty"`
	ShortOrderDuration     *int    `json:"shortOrderDuration,omitempty"`
	ShortOrderCon          *int    `json:"shortOrderCon,omitempty"`
	UserTicketPhotoWays    *string `json:"userTicketPhotoWays,omitempty"`
	IzWithdraw             *bool   `json:"izWithdraw,omitempty"`
	CreatedPin             string  `json:"createdPin,omitempty"`
	UpdatedPin             string  `json:"updatedPin,omitempty"`
	CreatedAt              *string `json:"createdAt,omitempty"`
	UpdatedAt              *string `json:"updatedAt,omitempty"`
	Version                *int    `json:"version,omitempty"`
	IzDel                  *bool   `json:"izDel,omitempty"`
}

// MarshalBaseItemForRedis serializes like Java ConfigBaseItemQueryImpl DB backfill.
func MarshalBaseItemForRedis(m *model.ConfigBaseItem) (string, error) {
	if m == nil {
		return "", nil
	}
	co := BaseItemToCO(m)
	ways := "0"
	if m.UserTicketPhotoWays.Valid && m.UserTicketPhotoWays.String != "" {
		ways = m.UserTicketPhotoWays.String
	}
	do := baseItemRedisDO{
		ID:                     m.ID,
		TenantID:               m.TenantID,
		ServiceID:              m.ServiceID,
		OfflineTicketJudgeTime: co.OfflineTicketJudgeTime,
		MsgSign:                co.MsgSign,
		MsgCode:                co.MsgCode,
		IzOnAbnormalMovement:   co.IzOnAbnormalMovement,
		IzOnBatteryRemoval:     co.IzOnBatteryRemoval,
		IzOfflineTicket:        co.IzOfflineTicket,
		IzAutoRepairTicket:     co.IzAutoRepairTicket,
		IzAutoSwapBattery:      co.IzAutoSwapBattery,
		SwapBatteryThreshold:   co.SwapBatteryThreshold,
		IzOpenInvoice:          co.IzOpenInvoice,
		IzTempUnlock:           co.IzTempUnlock,
		TempUnlockTime:         co.TempUnlockTime,
		DashboardSwitch:        co.DashboardSwitch,
		DashboardStartTime:     co.DashboardStartTime,
		DashboardEndTime:       co.DashboardEndTime,
		IzCanInitiativeRepair:  co.IzCanInitiativeRepair,
		IzCanAfterRidingRepair: co.IzCanAfterRidingRepair,
		CanRepairCon:           co.CanRepairCon,
		IzHelmetAbnormalRiding: co.IzHelmetAbnormalRiding,
		IzEnableMoveAlarm:      co.IzEnableMoveAlarm,
		MoveAlarmTime:          co.MoveAlarmTime,
		MoveAlarmCount:         co.MoveAlarmCount,
		MoveAlarmDistance:      co.MoveAlarmDistance,
		IzCreateShortOrder:     co.IzCreateShortOrder,
		ShortOrderDuration:     co.ShortOrderDuration,
		ShortOrderCon:          co.ShortOrderCon,
		UserTicketPhotoWays:    &ways,
		IzWithdraw:             co.IzWithdraw,
		CreatedPin:             m.CreatedPin,
		UpdatedPin:             m.UpdatedPin,
	}
	if !m.CreatedAt.IsZero() {
		s := timefmt.FormatJavaLocalSpace(m.CreatedAt)
		do.CreatedAt = &s
	}
	if !m.UpdatedAt.IsZero() {
		s := timefmt.FormatJavaLocalSpace(m.UpdatedAt)
		do.UpdatedAt = &s
	}
	if m.Version.Valid {
		v := int(m.Version.Int32)
		do.Version = &v
	}
	if m.IzDel.Valid {
		do.IzDel = &m.IzDel.Bool
	}
	return MarshalRow(do)
}

// UnmarshalBaseItemCO parses Java DO or Go CO JSON from shared Redis.
func UnmarshalBaseItemCO(raw string) (*dto.ConfigBaseItemCO, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return &dto.ConfigBaseItemCO{}, nil
	}
	type baseItemAlias dto.ConfigBaseItemCO
	aux := &struct {
		UserTicketPhotoWays  json.RawMessage `json:"userTicketPhotoWays"`
		DashboardStartTime   json.RawMessage `json:"dashboardStartTime"`
		DashboardEndTime     json.RawMessage `json:"dashboardEndTime"`
		*baseItemAlias
	}{
		baseItemAlias: (*baseItemAlias)(&dto.ConfigBaseItemCO{}),
	}
	if err := json.Unmarshal([]byte(raw), aux); err != nil {
		return nil, err
	}
	co := (*dto.ConfigBaseItemCO)(aux.baseItemAlias)
	co.UserTicketPhotoWays = parseFlexibleIntList(aux.UserTicketPhotoWays)
	if s := parseFlexibleLocalTime(aux.DashboardStartTime); s != "" {
		co.DashboardStartTime = &s
	}
	if s := parseFlexibleLocalTime(aux.DashboardEndTime); s != "" {
		co.DashboardEndTime = &s
	}
	return co, nil
}

func parseFlexibleIntList(raw json.RawMessage) []int {
	if len(raw) == 0 || string(raw) == "null" {
		return []int{0}
	}
	var arr []int
	if json.Unmarshal(raw, &arr) == nil {
		return arr
	}
	var s string
	if json.Unmarshal(raw, &s) == nil && s != "" {
		parts := strings.Split(s, ",")
		ways := make([]int, 0, len(parts))
		for _, p := range parts {
			if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
				ways = append(ways, n)
			}
		}
		if len(ways) > 0 {
			return ways
		}
	}
	return []int{0}
}

func parseFlexibleLocalTime(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var parts struct {
		Hour   int `json:"hour"`
		Minute int `json:"minute"`
		Second int `json:"second"`
	}
	if json.Unmarshal(raw, &parts) == nil {
		return fmt.Sprintf("%02d:%02d:%02d", parts.Hour, parts.Minute, parts.Second)
	}
	return ""
}

func ParkApplyToCO(m *model.ParkApplyConfig) *dto.ConfigParkApplyCO {
	if m == nil {
		return nil
	}
	id, sid := m.ID, m.ServiceID
	return &dto.ConfigParkApplyCO{
		Id:          &id,
		ServiceId:   &sid,
		IzApplyPark: nullIntPtr(m.IzApplyPark),
	}
}

func PushRidingCardToCO(m *model.PushRidingCardConfig) *dto.ConfigPushRidingCardCO {
	if m == nil {
		return nil
	}
	id, sid := m.ID, m.ServiceID
	co := &dto.ConfigPushRidingCardCO{
		Id:         &id,
		ServiceId:  &sid,
		IzOpen:     nullBoolPtr(m.IzOpen),
		UpdatedPin: m.UpdatedPin,
		UpdateName: nil,
	}
	if !m.UpdatedAt.IsZero() {
		co.UpdatedAt = timefmt.FormatJavaLocal(m.UpdatedAt)
	}
	return co
}

func CreditScoreToCO(m *model.CreditScoreConfig) *dto.CreditScoreConfigCO {
	if m == nil {
		f := false
		return &dto.CreditScoreConfigCO{IzCreditScore: &f}
	}
	id := m.ID
	co := &dto.CreditScoreConfigCO{
		Id:                 &id,
		IzCreditScore:      nullBoolPtr(m.IzCreditScore),
		Score:              nullFloatPtr(m.Score),
		WarnScore:          nullFloatPtr(m.WarnScore),
		NoRiddingScore:     nullFloatPtr(m.NoRiddingScore),
		FirstNoRiddingDays: nullIntPtr(m.FirstNoRiddingDays),
		SecondNoRiddingDays: nullIntPtr(m.SecondNoRiddingDays),
		MoreNoRiddingDays:  nullIntPtr(m.MoreNoRiddingDays),
		AddScore:           nullFloatPtr(m.AddScore),
		AddScoreUpperLimit: nullFloatPtr(m.AddScoreUpperLimit),
		RemindWay:          nullStrPtr(m.RemindWay),
	}
	return co
}

func BigScreenToCO(m *model.BigScreen) *dto.ConfigBigScreenCO {
	if m == nil {
		return &dto.ConfigBigScreenCO{}
	}
	id := m.ID
	return &dto.ConfigBigScreenCO{
		Id:                 &id,
		DisplayCoefficient: nullFloatPtr(m.DisplayCoefficient),
	}
}

func RidingPermissionToCO(m *model.RidingPermission) *dto.RidingPermissionCO {
	if m == nil {
		return nil
	}
	sid := m.ServiceID
	co := &dto.RidingPermissionCO{
		ServiceId:                    &sid,
		IzDefaultPermit:              nullBoolPtr(m.IzDefaultPermit),
		IzDeposit:                    nullBoolPtr(m.IzDeposit),
		Deposit:                      nullIntPtr(m.Deposit),
		IzCareer:                     nullBoolPtr(m.IzCareer),
		IzDepositCard:                nullBoolPtr(m.IzDepositCard),
		IzWxScorePayDeposited:        nullBoolPtr(m.IzWxScorePayDeposited),
		IzWxScorePayNoPassword:       nullBoolPtr(m.IzWxScorePayNoPassword),
		IzFreeDeposit:                nullBoolPtr(m.IzFreeDeposit),
		IzZhimaPayAfterUseDeposited:  nullBoolPtr(m.IzZhimaPayAfterUseDeposited),
		IzZhimaPayAfterUseNoPassword: nullBoolPtr(m.IzZhimaPayAfterUseNoPassword),
	}
	var emptyCareer *string
	co.Career = emptyCareer
	if m.Career.Valid {
		co.Career = nullStrPtr(m.Career)
	}
	if !m.UpdatedAt.IsZero() {
		s := timefmt.FormatJavaLocal(m.UpdatedAt)
		co.UpdatedAt = &s
	}
	if m.UpdatedPin != "" {
		p := m.UpdatedPin
		co.UpdatedPin = &p
	}
	return co
}

func AlarmContactToCO(m model.AlarmContact) dto.AlarmContactCO {
	id, sid := m.ID, m.ServiceID
	return dto.AlarmContactCO{
		Id:          &id,
		ServiceId:   &sid,
		Type:        nullIntPtr(m.Type),
		Name:        nullStrPtr(m.Name),
		Phone:       nullStrPtr(m.Phone),
		StartTime:   nullStrPtr(m.StartTime),
		EndTime:     nullStrPtr(m.EndTime),
		NotifyType:  nullStrPtr(m.NotifyType),
		TriggerArea: nullIntPtr(m.TriggerArea),
	}
}

func ProtocolToCO(m *model.ConfigProtocol) *dto.ConfigProtocolCO {
	if m == nil {
		return nil
	}
	typ := m.Type
	title, content := m.Title, m.Content
	co := &dto.ConfigProtocolCO{
		Type:    &typ,
		Title:   &title,
		Content: &content,
	}
	if m.ID != 0 {
		id := m.ID
		co.Id = &id
	}
	if m.ServiceID != 0 {
		sid := m.ServiceID
		co.ServiceId = &sid
	}
	if m.UpdatedPin != "" {
		p := m.UpdatedPin
		co.UpdatedPin = &p
	}
	if !m.UpdatedAt.IsZero() {
		s := timefmt.FormatJavaLocal(m.UpdatedAt)
		co.UpdatedAt = &s
	}
	return co
}

func FenceTagToCO(m model.FenceTag) dto.FenceTagCO {
	id := m.TagID
	name := m.TagName
	return dto.FenceTagCO{
		TagId:    &id,
		TagName:  &name,
		IzEnable: nullBoolPtr(m.IzEnable),
	}
}

func ResourceToCO(m model.ResourceManagement) dto.ResourceManagementCO {
	id, sid := m.ID, m.ServiceID
	co := dto.ResourceManagementCO{
		Id:            &id,
		ServiceId:     &sid,
		PageCode:      nullIntPtr(m.PageCode),
		Type:          nullIntPtr(m.Type),
		Status:        nullIntPtr(m.Status),
		Name:          nullStrPtr(m.Name),
		IzLimitTime:   nullBoolPtr(m.IzLimitTime),
		AdvId:         nullStrPtr(m.AdvID),
		Title:         nullStrPtr(m.Title),
		Appid:         nullStrPtr(m.AppID),
		SkipUrl:       nullStrPtr(m.SkipURL),
		Params:        nullStrPtr(m.Params),
		ImgUrl:        nullStrPtr(m.ImgURL),
		Sort:          nullIntPtr(m.Sort),
		ExposureCount: nullInt64Ptr(m.ExposureCount),
		ClickCount:    nullInt64Ptr(m.ClickCount),
		ClickPerson:   nullInt64Ptr(m.ClickPerson),
	}
	if m.StartTime.Valid {
		s := timefmt.FormatJavaLocal(m.StartTime.Time)
		co.StartTime = &s
	}
	if m.EndTime.Valid {
		s := timefmt.FormatJavaLocal(m.EndTime.Time)
		co.EndTime = &s
	}
	if m.UpdatedPin != "" {
		p := m.UpdatedPin
		co.UpdatedPin = &p
	}
	if !m.UpdatedAt.IsZero() {
		s := timefmt.FormatJavaLocalSpace(m.UpdatedAt)
		co.UpdatedAt = &s
	}
	return co
}

func MarshalRow(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

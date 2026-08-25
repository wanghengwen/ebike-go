package clientconfig

import (
	"encoding/json"

	"ebike-service-client-go/internal/pkg/javacompat"
)

// configBackCarCCO mirrors client ConfigBackCarCCO (Long id/serviceId as strings).
type configBackCarCCO struct {
	Id                          *string  `json:"id"`
	ServiceId                   *string  `json:"serviceId"`
	AllowOutofService           *bool    `json:"allowOutofService"`
	AllowInNostop               *bool    `json:"allowInNostop"`
	AllowOutofParking           *bool    `json:"allowOutofParking"`
	AllowInBanRiding            *bool    `json:"allowInBanRiding"`
	DispatchCost                *int     `json:"dispatchCost"`
	PenaltyInNostop             *int     `json:"penaltyInNostop"`
	PenaltyOutofService         *int     `json:"penaltyOutofService"`
	PenaltyInBanRiding          *int     `json:"penaltyInBanRiding"`
	IzCivilizationRemind        *bool    `json:"izCivilizationRemind"`
	CoefficientOfDifficult      *float64 `json:"coefficientOfDifficult"`
	BufferDistance              *float64 `json:"bufferDistance"`
	ParkingBackcar              *int     `json:"parkingBackcar"`
	NoParkingBackcar            *int     `json:"noParkingBackcar"`
	BanRidingBackcar            *int     `json:"banRidingBackcar"`
	IzHelmetLock                *bool    `json:"izHelmetLock"`
	IzHelmetPhoto               *bool    `json:"izHelmetPhoto"`
	IzBeacon                    *bool    `json:"izBeacon"`
	IzHelmetRemovalDetection    *bool    `json:"izHelmetRemovalDetection"`
	IzHelmetWearDetection       *bool    `json:"izHelmetWearDetection"`
	PowerOff                    *int     `json:"powerOff"`
	HelmetPenalty               *int     `json:"helmetPenalty"`
	DispatchFee                 *bool    `json:"dispatchFee"`
	RfidBeacon                  *bool    `json:"rfidBeacon"`
	Direction                   *bool    `json:"direction"`
	Helmet                      *bool    `json:"helmet"`
	Kickstand                   *bool    `json:"kickstand"`
	Camera                      *bool    `json:"camera"`
}

// configBaseItemCCO mirrors client ConfigBaseItemCCO.
type configBaseItemCCO struct {
	ServiceId                 *string `json:"serviceId"`
	OfflineTicketJudgeTime    *int    `json:"offlineTicketJudgeTime"`
	IzEnableLocalFence        *bool   `json:"izEnableLocalFence"`
	IzPeriodicalUpdates       *bool   `json:"izPeriodicalUpdates"`
	MsgSign                   *string `json:"msgSign"`
	MsgCode                   *string `json:"msgCode"`
	IzOnAbnormalMovement      *bool   `json:"izOnAbnormalMovement"`
	IzOnBatteryRemoval        *bool   `json:"izOnBatteryRemoval"`
	IzOfflineTicket           *bool   `json:"izOfflineTicket"`
	IzAutoRepairTicket        *bool   `json:"izAutoRepairTicket"`
	SwapBatteryThreshold      *int    `json:"swapBatteryThreshold"`
	IzOpenInvoice             *bool   `json:"izOpenInvoice"`
	IzTempUnlock              *bool   `json:"izTempUnlock"`
	TempUnlockTime            *int    `json:"tempUnlockTime"`
	IzCanInitiativeRepair     *bool   `json:"izCanInitiativeRepair"`
	IzCanAfterRidingRepair    *bool   `json:"izCanAfterRidingRepair"`
	CanRepairCon              *int    `json:"canRepairCon"`
	UserTicketPhotoWays       []int   `json:"userTicketPhotoWays"`
	IzWithdraw                *bool   `json:"izWithdraw"`
}

func convertConfigBackCarCCO(backCar, audit json.RawMessage) (json.RawMessage, bool) {
	var base map[string]json.RawMessage
	if json.Unmarshal(backCar, &base) != nil {
		return backCar, true
	}
	if len(audit) > 0 && !javacompat.IsNullJSON(audit) {
		var auditMap map[string]json.RawMessage
		if json.Unmarshal(audit, &auditMap) == nil {
			for k, v := range auditMap {
				base[k] = v
			}
		}
	}
	co := configBackCarCCO{
		Id:                       javacompat.LongStrPtr(base["id"]),
		ServiceId:                javacompat.LongStrPtr(base["serviceId"]),
		AllowOutofService:        rawBoolPtr(base["allowOutofService"]),
		AllowInNostop:            rawBoolPtr(base["allowInNostop"]),
		AllowOutofParking:        rawBoolPtr(base["allowOutofParking"]),
		AllowInBanRiding:         rawBoolPtr(base["allowInBanRiding"]),
		DispatchCost:             rawIntPtr(base["dispatchCost"]),
		PenaltyInNostop:          rawIntPtr(base["penaltyInNostop"]),
		PenaltyOutofService:      rawIntPtr(base["penaltyOutofService"]),
		PenaltyInBanRiding:       rawIntPtr(base["penaltyInBanRiding"]),
		IzCivilizationRemind:     rawBoolPtr(base["izCivilizationRemind"]),
		CoefficientOfDifficult:   rawFloatPtr(base["coefficientOfDifficult"]),
		BufferDistance:           rawFloatPtr(base["bufferDistance"]),
		ParkingBackcar:           rawIntPtr(base["parkingBackcar"]),
		NoParkingBackcar:         rawIntPtr(base["noParkingBackcar"]),
		BanRidingBackcar:         rawIntPtr(base["banRidingBackcar"]),
		IzHelmetLock:             rawBoolPtr(base["izHelmetLock"]),
		IzHelmetPhoto:            rawBoolPtr(base["izHelmetPhoto"]),
		IzBeacon:                 rawBoolPtr(base["izBeacon"]),
		IzHelmetRemovalDetection: rawBoolPtr(base["izHelmetRemovalDetection"]),
		IzHelmetWearDetection:    rawBoolPtr(base["izHelmetWearDetection"]),
		PowerOff:                 rawIntPtr(base["powerOff"]),
		HelmetPenalty:            rawIntPtr(base["helmetPenalty"]),
		DispatchFee:              rawBoolPtr(base["dispatchFee"]),
		RfidBeacon:               rawBoolPtr(base["rfidBeacon"]),
		Direction:                rawBoolPtr(base["direction"]),
		Helmet:                   rawBoolPtr(base["helmet"]),
		Kickstand:                rawBoolPtr(base["kickstand"]),
		Camera:                   rawBoolPtr(base["camera"]),
	}
	b, err := json.Marshal(co)
	if err != nil {
		return backCar, true
	}
	return b, true
}

func convertConfigBaseItemCCO(raw json.RawMessage) json.RawMessage {
	var base map[string]json.RawMessage
	if json.Unmarshal(raw, &base) != nil {
		return raw
	}
	var ways []int
	if w, ok := base["userTicketPhotoWays"]; ok && !javacompat.IsNullJSON(w) {
		_ = json.Unmarshal(w, &ways)
	}
	co := configBaseItemCCO{
		ServiceId:              javacompat.LongStrPtr(base["serviceId"]),
		OfflineTicketJudgeTime: rawIntPtr(base["offlineTicketJudgeTime"]),
		IzEnableLocalFence:     rawBoolPtr(base["izEnableLocalFence"]),
		IzPeriodicalUpdates:    rawBoolPtr(base["izPeriodicalUpdates"]),
		MsgSign:                javacompat.RawStringPtr(base["msgSign"]),
		MsgCode:                javacompat.RawStringPtr(base["msgCode"]),
		IzOnAbnormalMovement:   rawBoolPtr(base["izOnAbnormalMovement"]),
		IzOnBatteryRemoval:     rawBoolPtr(base["izOnBatteryRemoval"]),
		IzOfflineTicket:        rawBoolPtr(base["izOfflineTicket"]),
		IzAutoRepairTicket:     rawBoolPtr(base["izAutoRepairTicket"]),
		SwapBatteryThreshold:   rawIntPtr(base["swapBatteryThreshold"]),
		IzOpenInvoice:          rawBoolPtr(base["izOpenInvoice"]),
		IzTempUnlock:           rawBoolPtr(base["izTempUnlock"]),
		TempUnlockTime:         rawIntPtr(base["tempUnlockTime"]),
		IzCanInitiativeRepair:  rawBoolPtr(base["izCanInitiativeRepair"]),
		IzCanAfterRidingRepair: rawBoolPtr(base["izCanAfterRidingRepair"]),
		CanRepairCon:           rawIntPtr(base["canRepairCon"]),
		UserTicketPhotoWays:    ways,
		IzWithdraw:             rawBoolPtr(base["izWithdraw"]),
	}
	b, _ := json.Marshal(co)
	return b
}

type configPayCO struct {
	Id                      *javacompat.LongStr  `json:"id"`
	ServiceId               *javacompat.LongStr  `json:"serviceId"`
	IzBalanceEnoughReturnBike *bool   `json:"izBalanceEnoughReturnBike"`
	IzMakeupBalancePay      *bool   `json:"izMakeupBalancePay"`
	IzNotifyUnpaidOrder     *bool   `json:"izNotifyUnpaidOrder"`
	NotifyInterval          *int    `json:"notifyInterval"`
	TimesUpperBound         *int    `json:"timesUpperBound"`
	RemindWay               *string `json:"remindWay"`
}

func convertConfigPayCO(raw json.RawMessage) json.RawMessage {
	var co configPayCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

type configUseCarCO struct {
	Id                             *javacompat.LongStr  `json:"id"`
	ServiceId                      *javacompat.LongStr  `json:"serviceId"`
	RechargeBeforeUse              *bool                `json:"rechargeBeforeUse"`
	RechargeVisibleRange           *int                 `json:"rechargeVisibleRange"`
	RechargeByRegister             *bool                `json:"rechargeByRegister"`
	RechargeByTags                 *bool                `json:"rechargeByTags"`
	RechargeTagIds                 *string              `json:"rechargeTagIds"`
	RechargeCost                   *int                 `json:"rechargeCost"`
	IzStopService                  *bool                `json:"izStopService"`
	StopTimeStart                  *string              `json:"stopTimeStart"`
	StopTimeEnd                    *string              `json:"stopTimeEnd"`
	IzAutoRecovery                 *bool                `json:"izAutoRecovery"`
	RecoveryData                   *javacompat.DateTime `json:"recoveryData"`
	StopServiceNotice              *string              `json:"stopServiceNotice"`
	IzOnCertification              *bool                `json:"izOnCertification"`
	IzOnUseCar                     *bool                `json:"izOnUseCar"`
	RecognitionDegree              *int                 `json:"recognitionDegree"`
	EffectiveTime                  *int                 `json:"effectiveTime"`
	IzRidingStopTrigger            *bool                `json:"izRidingStopTrigger"`
	IzParkingTriggerReturnBike     *bool                `json:"izParkingTriggerReturnBike"`
	RidingStopTime                 *int                 `json:"ridingStopTime"`
	RidingStopEvent                *int                 `json:"ridingStopEvent"`
	ParkingTime                    *int                 `json:"parkingTime"`
	RemindWay                      *string              `json:"remindWay"`
	IzBeacon                       *bool                `json:"izBeacon"`
	OutServiceAreaAutoLock         *int                 `json:"outServiceAreaAutoLock"`
	IzRemoteUnlock                 *bool                `json:"izRemoteUnlock"`
	OutServiceAreaAutoLockRemindWay *string             `json:"outServiceAreaAutoLockRemindWay"`
	NearLine                       *int                 `json:"nearLine"`
	IzOpenSaddleOverloadMonitor    *bool                `json:"izOpenSaddleOverloadMonitor"`
	OverloadRemind                 *int                 `json:"overloadRemind"`
	HelmetConfig                   *string              `json:"helmetConfig"`
	IzOrderNotice                  *bool                `json:"izOrderNotice"`
	NoticeRidingTime               *int                 `json:"noticeRidingTime"`
	OrderRemindWay                 *string              `json:"orderRemindWay"`
	MinAge                         *int                 `json:"minAge"`
	MaxAge                         *int                 `json:"maxAge"`
	HideCarConfig                  *string              `json:"hideCarConfig"`
	IzAuth                         *bool                `json:"izAuth"`
	IzNeedAuth                     *bool                `json:"izNeedAuth"`
}

func convertConfigUseCarCO(raw json.RawMessage) json.RawMessage {
	var co configUseCarCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

type ridingPermissionCO struct {
	ServiceId                    *javacompat.LongStr  `json:"serviceId"`
	IzDefaultPermit              *bool                `json:"izDefaultPermit"`
	IzDeposit                    *bool                `json:"izDeposit"`
	Deposit                      *int                 `json:"deposit"`
	IzCareer                     *bool                `json:"izCareer"`
	Career                       *string              `json:"career"`
	IzDepositCard                *bool                `json:"izDepositCard"`
	IzWxScorePayDeposited        *bool                `json:"izWxScorePayDeposited"`
	IzWxScorePayNoPassword       *bool                `json:"izWxScorePayNoPassword"`
	IzFreeDeposit                *bool                `json:"izFreeDeposit"`
	UpdatedAt                    *javacompat.DateTime `json:"updatedAt"`
	UpdatedPin                   *string              `json:"updatedPin"`
	IzZhimaPayAfterUseDeposited  *bool                `json:"izZhimaPayAfterUseDeposited"`
	IzZhimaPayAfterUseNoPassword *bool                `json:"izZhimaPayAfterUseNoPassword"`
}

func convertRidingPermissionCO(raw json.RawMessage) json.RawMessage {
	var co ridingPermissionCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

func rawIntPtr(raw json.RawMessage) *int {
	if v, ok := javacompat.RawInt(raw); ok {
		n := int(v)
		return &n
	}
	return nil
}

func rawBoolPtr(raw json.RawMessage) *bool {
	if javacompat.IsNullJSON(raw) {
		return nil
	}
	var b bool
	if json.Unmarshal(raw, &b) == nil {
		return &b
	}
	return nil
}

func rawFloatPtr(raw json.RawMessage) *float64 {
	if javacompat.IsNullJSON(raw) {
		return nil
	}
	var f float64
	if json.Unmarshal(raw, &f) == nil {
		return &f
	}
	return nil
}

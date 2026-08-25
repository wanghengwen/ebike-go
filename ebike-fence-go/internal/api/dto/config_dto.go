package dto

import "encoding/json"

// --- Config Backcar ---

type ConfigBackcarCmd struct {
	Command
	Id                         *int64   `json:"id,omitempty"`
	ServiceId                  *int64   `json:"serviceId,omitempty"`
	AllowOutofService          *bool    `json:"allowOutofService,omitempty"`
	AllowInNostop              *bool    `json:"allowInNostop,omitempty"`
	AllowOutofParking          *bool    `json:"allowOutofParking,omitempty"`
	AllowInBanRiding           *bool    `json:"allowInBanRiding,omitempty"`
	DispatchCost               *int     `json:"dispatchCost,omitempty"`
	PenaltyInNostop            *int     `json:"penaltyInNostop,omitempty"`
	PenaltyOutofService        *int     `json:"penaltyOutofService,omitempty"`
	PenaltyInBanRiding         *int     `json:"penaltyInBanRiding,omitempty"`
	IzCivilizationRemind       *bool    `json:"izCivilizationRemind,omitempty"`
	CoefficientOfDifficult     *float64 `json:"coefficientOfDifficult,omitempty"`
	BufferDistance             *float64 `json:"bufferDistance,omitempty"`
	ParkingBackcar             *int     `json:"parkingBackcar,omitempty"`
	NoParkingBackcar           *int     `json:"noParkingBackcar,omitempty"`
	BanRidingBackcar           *int     `json:"banRidingBackcar,omitempty"`
	IzHelmetLock               *bool    `json:"izHelmetLock,omitempty"`
	IzHelmetPhoto              *bool    `json:"izHelmetPhoto,omitempty"`
	IzBeacon                   *bool    `json:"izBeacon,omitempty"`
	IzHelmetRemovalDetection   *bool    `json:"izHelmetRemovalDetection,omitempty"`
	IzHelmetWearDetection      *bool    `json:"izHelmetWearDetection,omitempty"`
	IzHelmetAutoRepair         *bool    `json:"izHelmetAutoRepair,omitempty"`
	PowerOff                   *int     `json:"powerOff,omitempty"`
	HelmetPenalty              *int     `json:"helmetPenalty,omitempty"`
	IzHelmetReign              *bool    `json:"izHelmetReign,omitempty"`
	IzCameraBackcar            *bool    `json:"izCameraBackcar,omitempty"`
	IzCameraDirectionalBackcar *bool    `json:"izCameraDirectionalBackcar,omitempty"`
	IzCameraPointBackcar       *bool    `json:"izCameraPointBackcar,omitempty"`
	CameraImpunityCount        *int     `json:"cameraImpunityCount,omitempty"`
	IzFullPileNoStop           *int     `json:"izFullPileNoStop,omitempty"`
	IzUseOtherParking          *bool    `json:"izUseOtherParking,omitempty"`
	BeaconEffect               *int     `json:"beaconEffect,omitempty"`
	RfidEffect                 *int     `json:"rfidEffect,omitempty"`
	DirectionEffect            *int     `json:"directionEffect,omitempty"`
	FootSupportEffect          *int     `json:"footSupportEffect,omitempty"`
	CameraEffect               *int     `json:"cameraEffect,omitempty"`
	CameraAngle                *int     `json:"cameraAngle,omitempty"`
	IzStandardReturn           *int     `json:"izStandardReturn,omitempty"`
}

type ConfigBackcarCO struct {
	Id                         *int64   `json:"id"`
	ServiceId                  *int64   `json:"serviceId"`
	AllowOutofService          *bool    `json:"allowOutofService"`
	AllowInNostop              *bool    `json:"allowInNostop"`
	AllowOutofParking          *bool    `json:"allowOutofParking"`
	AllowInBanRiding           *bool    `json:"allowInBanRiding"`
	DispatchCost               *int     `json:"dispatchCost"`
	PenaltyInNostop            *int     `json:"penaltyInNostop"`
	PenaltyOutofService        *int     `json:"penaltyOutofService"`
	PenaltyInBanRiding         *int     `json:"penaltyInBanRiding"`
	IzCivilizationRemind       *bool    `json:"izCivilizationRemind"`
	CoefficientOfDifficult     *float64 `json:"coefficientOfDifficult"`
	BufferDistance             *float64 `json:"bufferDistance"`
	ParkingBackcar             *int     `json:"parkingBackcar"`
	NoParkingBackcar           *int     `json:"noParkingBackcar"`
	BanRidingBackcar           *int     `json:"banRidingBackcar"`
	IzHelmetLock               *bool    `json:"izHelmetLock"`
	IzHelmetPhoto              *bool    `json:"izHelmetPhoto"`
	IzBeacon                   *bool    `json:"izBeacon"`
	IzHelmetRemovalDetection   *bool    `json:"izHelmetRemovalDetection"`
	IzHelmetWearDetection      *bool    `json:"izHelmetWearDetection"`
	IzHelmetAutoRepair         *bool    `json:"izHelmetAutoRepair"`
	PowerOff                   *int     `json:"powerOff"`
	HelmetPenalty              *int     `json:"helmetPenalty"`
	IzHelmetReign              *bool    `json:"izHelmetReign"`
	IzCameraBackcar            *bool    `json:"izCameraBackcar"`
	IzCameraDirectionalBackcar *bool    `json:"izCameraDirectionalBackcar"`
	IzCameraPointBackcar       *bool    `json:"izCameraPointBackcar"`
	CameraImpunityCount        *int     `json:"cameraImpunityCount"`
	IzFullPileNoStop           *int     `json:"izFullPileNoStop"`
	IzUseOtherParking          *bool    `json:"izUseOtherParking"`
	BeaconEffect               *int     `json:"beaconEffect"`
	RfidEffect                 *int     `json:"rfidEffect"`
	DirectionEffect            *int     `json:"directionEffect"`
	FootSupportEffect          *int     `json:"footSupportEffect"`
	CameraEffect               *int     `json:"cameraEffect"`
	CameraAngle                *int     `json:"cameraAngle"`
	IzStandardReturn           *int     `json:"izStandardReturn"`
}

// --- Config Pay / UseCar ---

type ConfigPayCmd struct {
	Command
	Id                      *int64  `json:"id,omitempty"`
	ServiceId               *int64  `json:"serviceId,omitempty"`
	IzBalanceEnoughReturnBike *bool `json:"izBalanceEnoughReturnBike,omitempty"`
	IzMakeupBalancePay      *bool   `json:"izMakeupBalancePay,omitempty"`
	IzNotifyUnpaidOrder     *bool   `json:"izNotifyUnpaidOrder,omitempty"`
	NotifyInterval          *int    `json:"notifyInterval,omitempty"`
	TimesUpperBound         *int    `json:"timesUpperBound,omitempty"`
	RemindWay               *string `json:"remindWay,omitempty"`
}

type ConfigPayCO struct {
	Id                      *int64  `json:"id"`
	ServiceId               *int64  `json:"serviceId"`
	IzBalanceEnoughReturnBike *bool `json:"izBalanceEnoughReturnBike"`
	IzMakeupBalancePay      *bool   `json:"izMakeupBalancePay"`
	IzNotifyUnpaidOrder     *bool   `json:"izNotifyUnpaidOrder"`
	NotifyInterval          *int    `json:"notifyInterval"`
	TimesUpperBound         *int    `json:"timesUpperBound"`
	RemindWay               *string `json:"remindWay"`
}

type ConfigUseCarCmd struct {
	Command
	Id                   *int64          `json:"id,omitempty"`
	ServiceId            *int64          `json:"serviceId,omitempty"`
	RechargeBeforeUse    *bool           `json:"rechargeBeforeUse,omitempty"`
	RechargeVisibleRange *int            `json:"rechargeVisibleRange,omitempty"`
	RechargeByRegister   *bool           `json:"rechargeByRegister,omitempty"`
	RechargeByTags       *bool           `json:"rechargeByTags,omitempty"`
	RechargeTagIds       *string         `json:"rechargeTagIds,omitempty"`
	RechargeCost         *int            `json:"rechargeCost,omitempty"`
	IzStopService        *bool           `json:"izStopService,omitempty"`
	IzAutoRecovery       *bool           `json:"izAutoRecovery,omitempty"`
	IzBeacon             *bool           `json:"izBeacon,omitempty"`
	IzNeedAuth           *bool           `json:"izNeedAuth,omitempty"`
	Extra                json.RawMessage `json:"-"`
}

type ConfigUseCarCO struct {
	Id                            *int64  `json:"id"`
	ServiceId                     *int64  `json:"serviceId"`
	RechargeBeforeUse             *bool   `json:"rechargeBeforeUse"`
	RechargeVisibleRange          *int    `json:"rechargeVisibleRange"`
	RechargeByRegister            *bool   `json:"rechargeByRegister"`
	RechargeByTags                *bool   `json:"rechargeByTags"`
	RechargeTagIds                *string `json:"rechargeTagIds"`
	RechargeCost                  *int    `json:"rechargeCost"`
	IzStopService                 *bool   `json:"izStopService"`
	StopTimeStart                 *string `json:"stopTimeStart"`
	StopTimeEnd                   *string `json:"stopTimeEnd"`
	IzAutoRecovery                *bool   `json:"izAutoRecovery"`
	RecoveryData                  *string `json:"recoveryData"`
	StopServiceNotice             *string `json:"stopServiceNotice"`
	IzOnCertification             *bool   `json:"izOnCertification"`
	IzOnUseCar                    *bool   `json:"izOnUseCar"`
	RecognitionDegree             *int    `json:"recognitionDegree"`
	EffectiveTime                 *int    `json:"effectiveTime"`
	IzRidingStopTrigger           *bool   `json:"izRidingStopTrigger"`
	IzParkingTriggerReturnBike    *bool   `json:"izParkingTriggerReturnBike"`
	RidingStopTime                *int    `json:"ridingStopTime"`
	RidingStopEvent               *int    `json:"ridingStopEvent"`
	ParkingTime                   *int    `json:"parkingTime"`
	RemindWay                     *string `json:"remindWay"`
	IzBeacon                      *bool   `json:"izBeacon"`
	OutServiceAreaAutoLock        *int    `json:"outServiceAreaAutoLock"`
	OutServiceAreaAutoLockRemindWay *string `json:"outServiceAreaAutoLockRemindWay"`
	NearLine                      *int    `json:"nearLine"`
	IzOpenSaddleOverloadMonitor   *bool   `json:"izOpenSaddleOverloadMonitor"`
	OverloadRemind                *int    `json:"overloadRemind"`
	IzRemoteUnlock                *bool   `json:"izRemoteUnlock"`
	HelmetConfig                  *string `json:"helmetConfig"`
	IzOrderNotice                 *bool   `json:"izOrderNotice"`
	NoticeRidingTime              *int    `json:"noticeRidingTime"`
	OrderRemindWay                *string `json:"orderRemindWay"`
	MinAge                        *int    `json:"minAge"`
	MaxAge                        *int    `json:"maxAge"`
	HideCarConfig                 *string `json:"hideCarConfig"`
	IzAuth                        *bool   `json:"izAuth"`
	IzNeedAuth                    *bool   `json:"izNeedAuth"`
}

// --- Config Base Item ---

type ConfigBaseItemCmd struct {
	Command
	Id                   *int64  `json:"id,omitempty"`
	ServiceId            *int64  `json:"serviceId,omitempty"`
	SwapBatteryThreshold *int    `json:"swapBatteryThreshold,omitempty"`
	IzAutoSwapBattery    *bool   `json:"izAutoSwapBattery,omitempty"`
	UserTicketPhotoWays  []int   `json:"userTicketPhotoWays,omitempty"`
	Extra                json.RawMessage `json:"-"`
}

type ConfigBaseItemCO struct {
	Id                      *int64  `json:"id"`
	ServiceId               *int64  `json:"serviceId"`
	OfflineTicketJudgeTime  *int    `json:"offlineTicketJudgeTime"`
	MsgSign                 *string `json:"msgSign"`
	MsgCode                 *string `json:"msgCode"`
	IzOnAbnormalMovement    *bool   `json:"izOnAbnormalMovement"`
	IzOnBatteryRemoval      *bool   `json:"izOnBatteryRemoval"`
	IzOfflineTicket         *bool   `json:"izOfflineTicket"`
	IzAutoRepairTicket      *bool   `json:"izAutoRepairTicket"`
	SwapBatteryThreshold    *int    `json:"swapBatteryThreshold"`
	IzAutoSwapBattery       *bool   `json:"izAutoSwapBattery"`
	IzOpenInvoice           *bool   `json:"izOpenInvoice"`
	IzTempUnlock            *bool   `json:"izTempUnlock"`
	TempUnlockTime          *int    `json:"tempUnlockTime"`
	DashboardSwitch         *bool   `json:"dashboardSwitch"`
	DashboardStartTime      *string `json:"dashboardStartTime"`
	DashboardEndTime        *string `json:"dashboardEndTime"`
	IzCanInitiativeRepair   *bool   `json:"izCanInitiativeRepair"`
	IzCanAfterRidingRepair  *bool   `json:"izCanAfterRidingRepair"`
	CanRepairCon            *int    `json:"canRepairCon"`
	IzHelmetAbnormalRiding  *bool   `json:"izHelmetAbnormalRiding"`
	IzEnableMoveAlarm       *bool   `json:"izEnableMoveAlarm"`
	MoveAlarmTime           *int    `json:"moveAlarmTime"`
	MoveAlarmCount          *int    `json:"moveAlarmCount"`
	MoveAlarmDistance       *int    `json:"moveAlarmDistance"`
	IzCreateShortOrder      *bool   `json:"izCreateShortOrder"`
	ShortOrderDuration      *int    `json:"shortOrderDuration"`
	ShortOrderCon           *int    `json:"shortOrderCon"`
	UserTicketPhotoWays     []int   `json:"userTicketPhotoWays"`
	IzWithdraw              *bool   `json:"izWithdraw"`
}

// --- Park Apply / Push Riding Card ---

type ConfigParkApplyCmd struct {
	Command
	Id          *int64 `json:"id,omitempty"`
	ServiceId   *int64 `json:"serviceId,omitempty"`
	IzApplyPark *int   `json:"izApplyPark,omitempty"`
}

type ConfigParkApplyCO struct {
	Id          *int64 `json:"id"`
	ServiceId   *int64 `json:"serviceId"`
	IzApplyPark *int   `json:"izApplyPark"`
}

type ConfigPushRidingCardCmd struct {
	Command
	Id        *int64 `json:"id,omitempty"`
	ServiceId *int64 `json:"serviceId,omitempty"`
	IzOpen    *bool  `json:"izOpen,omitempty"`
}

type ConfigPushRidingCardCO struct {
	Id         *int64  `json:"id"`
	ServiceId  *int64  `json:"serviceId"`
	IzOpen     *bool   `json:"izOpen"`
	UpdatedPin string  `json:"updatedPin"`
	UpdateName *string `json:"updateName"`
	UpdatedAt  string  `json:"updatedAt"`
}

// --- Credit Score / Big Screen ---

type CreditScoreConfigCmd struct {
	Command
	Id                 *int64   `json:"id,omitempty"`
	IzCreditScore      *bool    `json:"izCreditScore,omitempty"`
	Score              *float64 `json:"score,omitempty"`
	WarnScore          *float64 `json:"warnScore,omitempty"`
	NoRiddingScore     *float64 `json:"noRiddingScore,omitempty"`
	FirstNoRiddingDays *int     `json:"firstNoRiddingDays,omitempty"`
	SecondNoRiddingDays *int    `json:"secondNoRiddingDays,omitempty"`
	MoreNoRiddingDays  *int     `json:"moreNoRiddingDays,omitempty"`
	AddScore           *float64 `json:"addScore,omitempty"`
	AddScoreUpperLimit *float64 `json:"addScoreUpperLimit,omitempty"`
	RemindWay          *string  `json:"remindWay,omitempty"`
}

type CreditScoreConfigCO CreditScoreConfigCmd

type ConfigBigScreenCmd struct {
	Command
	Id                 *int64   `json:"id,omitempty"`
	DisplayCoefficient *float64 `json:"displayCoefficient,omitempty"`
}

type ConfigBigScreenCO struct {
	Id                 *int64   `json:"id"`
	ServiceId          *int64   `json:"serviceId"`
	DisplayCoefficient *float64 `json:"displayCoefficient"`
}

// --- Ad Config ---

type AdConfigCmd struct {
	Command
	Id        *int64      `json:"id,omitempty"`
	ServiceId *int64      `json:"serviceId,omitempty"`
	IzOn      interface{} `json:"izOn,omitempty"`
}

type AdConfigCO struct {
	Id        *int64      `json:"id"`
	ServiceId *int64      `json:"serviceId"`
	IzOn      interface{} `json:"izOn"`
}

type AdsConfigCmd struct {
	Command
	ServiceId *int64          `json:"serviceId,omitempty"`
	Raw       json.RawMessage `json:"-"`
}

type AdsConfigCO map[string]interface{}

// --- Alarm Contact ---

type AlarmContactCmd struct {
	Command
	Id          *int64  `json:"id,omitempty"`
	ServiceId   *int64  `json:"serviceId,omitempty"`
	Type        *int    `json:"type,omitempty"`
	Name        *string `json:"name,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	StartTime   *string `json:"startTime,omitempty"`
	EndTime     *string `json:"endTime,omitempty"`
	NotifyType  *string `json:"notifyType,omitempty"`
	TriggerArea *int    `json:"triggerArea,omitempty"`
}

type AlarmContactCO AlarmContactCmd

type ContactQuery struct {
	Command
	ServiceId *int64 `json:"serviceId,omitempty"`
	Type      *int   `json:"type,omitempty"`
	PageNum   int    `json:"pageNum,omitempty"`
	PageSize  int    `json:"pageSize,omitempty"`
}

type AlarmQuery struct {
	Command
	ServiceId *int64 `json:"serviceId,omitempty"`
	Type      *int   `json:"type,omitempty"`
}

// --- Riding Permission ---

type RidingPermissionCmd struct {
	Command
	Id                           *int64  `json:"id,omitempty"`
	ServiceId                    *int64  `json:"serviceId,omitempty"`
	IzDefaultPermit              *bool   `json:"izDefaultPermit,omitempty"`
	IzDeposit                    *bool   `json:"izDeposit,omitempty"`
	Deposit                      *int    `json:"deposit,omitempty"`
	IzCareer                     *bool   `json:"izCareer,omitempty"`
	Career                       *string `json:"career,omitempty"`
	IzDepositCard                *bool   `json:"izDepositCard,omitempty"`
	IzWxScorePayDeposited        *bool   `json:"izWxScorePayDeposited,omitempty"`
	IzWxScorePayNoPassword       *bool   `json:"izWxScorePayNoPassword,omitempty"`
	IzFreeDeposit                *bool   `json:"izFreeDeposit,omitempty"`
	IzZhimaPayAfterUseDeposited  *bool   `json:"izZhimaPayAfterUseDeposited,omitempty"`
	IzZhimaPayAfterUseNoPassword *bool   `json:"izZhimaPayAfterUseNoPassword,omitempty"`
}

type RidingPermissionCO struct {
	ServiceId                    *int64  `json:"serviceId"`
	IzDefaultPermit              *bool   `json:"izDefaultPermit"`
	IzDeposit                    *bool   `json:"izDeposit"`
	Deposit                      *int    `json:"deposit"`
	IzCareer                     *bool   `json:"izCareer"`
	Career                       *string `json:"career"`
	IzDepositCard                *bool   `json:"izDepositCard"`
	IzWxScorePayDeposited        *bool   `json:"izWxScorePayDeposited"`
	IzWxScorePayNoPassword       *bool   `json:"izWxScorePayNoPassword"`
	IzFreeDeposit                *bool   `json:"izFreeDeposit"`
	IzZhimaPayAfterUseDeposited  *bool   `json:"izZhimaPayAfterUseDeposited"`
	IzZhimaPayAfterUseNoPassword *bool   `json:"izZhimaPayAfterUseNoPassword"`
	UpdatedAt                    *string `json:"updatedAt"`
	UpdatedPin                   *string `json:"updatedPin"`
}

// --- Riding Car Config ---

type RidingCarConfigCO struct {
	UseCarCO *ConfigUseCarCO `json:"useCarCO"`
	PayCO    *ConfigPayCO    `json:"payCO"`
}

// --- Config Protocol ---

type ConfigProtocolCmd struct {
	Command
	ServiceId *int64 `json:"serviceId,omitempty"`
	Type      *int   `json:"type,omitempty"`
}

type ConfigProtocolUpdateCmd struct {
	Command
	Id        *int64  `json:"id,omitempty"`
	ServiceId *int64  `json:"serviceId,omitempty"`
	Type      *int    `json:"type,omitempty"`
	Title     *string `json:"title,omitempty"`
	Content   *string `json:"content,omitempty"`
}

type ConfigProtocolBatchUpdateCmd struct {
	Command
	List []ConfigProtocolUpdateCmd `json:"list,omitempty"`
}

type ConfigProtocolCO struct {
	Id         *int64  `json:"id"`
	ServiceId  *int64  `json:"serviceId"`
	Type       *int    `json:"type"`
	Title      *string `json:"title"`
	Content    *string `json:"content"`
	UpdatedPin *string `json:"updatedPin"`
	UpdatedAt  *string `json:"updatedAt"`
}

// --- Fence Tag ---

type FenceTagCO struct {
	TagId    *int64  `json:"tagId"`
	TagName  *string `json:"tagName"`
	IzEnable *bool   `json:"izEnable"`
}

// --- Resource Management ---

type ResourceManagementCmd struct {
	Command
	Id          *int64  `json:"id,omitempty"`
	ServiceId   *int64  `json:"serviceId,omitempty"`
	PageCode    *int    `json:"pageCode,omitempty"`
	Type        *int    `json:"type,omitempty"`
	Status      *int    `json:"status,omitempty"`
	Name        *string `json:"name,omitempty"`
	StartTime   *string `json:"startTime,omitempty"`
	EndTime     *string `json:"endTime,omitempty"`
	IzLimitTime *bool   `json:"izLimitTime,omitempty"`
	AdvId       *string `json:"advId,omitempty"`
	Title       *string `json:"title,omitempty"`
	Appid       *string `json:"appid,omitempty"`
	SkipUrl     *string `json:"skipUrl,omitempty"`
	Params      *string `json:"params,omitempty"`
	ImgUrl      *string `json:"imgUrl,omitempty"`
	Sort        *int    `json:"sort,omitempty"`
}

type ResourceManagementCO struct {
	Id             *int64  `json:"id"`
	ServiceId      *int64  `json:"serviceId"`
	PageCode       *int    `json:"pageCode"`
	Type           *int    `json:"type"`
	Status         *int    `json:"status"`
	Name           *string `json:"name"`
	StartTime      *string `json:"startTime"`
	EndTime        *string `json:"endTime"`
	IzLimitTime    *bool   `json:"izLimitTime"`
	AdvId          *string `json:"advId"`
	Title          *string `json:"title"`
	Appid          *string `json:"appid"`
	SkipUrl        *string `json:"skipUrl"`
	Params         *string `json:"params"`
	ImgUrl         *string `json:"imgUrl"`
	Sort           *int    `json:"sort"`
	ExposureCount  *int64  `json:"exposureCount"`
	ClickCount     *int64  `json:"clickCount"`
	ClickPerson    *int64  `json:"clickPerson"`
	UpdatedPin     *string `json:"updatedPin"`
	UpdatedAt      *string `json:"updatedAt"`
}

type ResourceManagementPageQuery struct {
	Command
	ServiceId *int64 `json:"serviceId,omitempty"`
	PageNum   int    `json:"pageNum,omitempty"`
	PageSize  int    `json:"pageSize,omitempty"`
}

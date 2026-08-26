package dto

// DeviceDetailQry is the request for /device/paas/device/detail.
type DeviceDetailQry struct {
	Imei           string          `json:"imei,omitempty"`
	CarId          string          `json:"carId,omitempty"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// DeviceDetailCo mirrors the Java DeviceDetailCo. The Java service serializes
// nulls, so no field uses omitempty; nil pointers emit JSON null and the three
// list fields emit [] (never null). Field order follows the Java class.
type DeviceDetailCo struct {
	Imei    *string `json:"imei"`
	Imsi    *string `json:"imsi"`
	Version *string `json:"version"`

	CarId     *string `json:"carId"`
	ServiceId *int64  `json:"serviceId"`
	TenantId  *string `json:"tenantId"`

	ReportTime *int64 `json:"reportTime"`

	GsmSignal       *int     `json:"gsmSignal"`
	Defend          *int     `json:"defend"`
	Acc             *int     `json:"acc"`
	Wgs84Lat        *float64 `json:"wgs84Lat"`
	Wgs84Lng        *float64 `json:"wgs84Lng"`
	Lat             *float64 `json:"lat"`
	Lng             *float64 `json:"lng"`
	Timestamp       *int64   `json:"timestamp"`
	Speed           *float64 `json:"speed"`
	Course          *float64 `json:"course"`
	Hdop            *float64 `json:"hdop"`
	BatteryLock     *int     `json:"batteryLock"`
	BatteryConnect  *int     `json:"batteryConnect"`
	BackWheelLock   *int     `json:"backWheelLock"`
	IsAutoLock      *int     `json:"isAutoLock"`
	Voltage         *int     `json:"voltage"`
	IsMoving        *int     `json:"isMoving"`
	TotalMiles      *float64 `json:"totalMiles"`
	MoveAlarmOn     *int     `json:"moveAlarmOn"`
	OverSpeedOn     *int     `json:"overSpeedOn"`
	RfidCarId       *string  `json:"rfidCarId"`
	HeadingAngle    *int     `json:"headingAngle"`
	HelmetType      *int     `json:"helmetType"`
	HelmetLock      *int     `json:"helmetLock"`
	HelmetReact     *int     `json:"helmetReact"`
	IsWheelSpan     *int     `json:"isWheelSpan"`
	IsFenceEnable   *int     `json:"isFenceEnable"`
	IsOutofServAera *int     `json:"isOutofServAera"`
	NoParkId        *int64   `json:"noParkId"`
	ForParkId       *int64   `json:"forParkId"`
	NoRideParkId    *int64   `json:"noRideParkId"`
	BatteryId       *int64   `json:"batteryId"`
	IsPowerExist    *int     `json:"isPowerExist"`

	IsOnline *int `json:"isOnline"`

	BmsSN        *string `json:"bmsSN"`
	Soc          *int    `json:"soc"`
	BmsTimeStamp *int64  `json:"bmsTimeStamp"`

	MoveAlarmTag *int `json:"moveAlarmTag"`
	RestBattery  *int `json:"restBattery"`
	RestMileage  *int `json:"restMileage"`

	LockTime   *int64 `json:"lockTime"`
	UnlockTime *int64 `json:"unlockTime"`

	RidingState    *int  `json:"ridingState"`
	OperationState []int `json:"operationState"`
	AlarmState     []int `json:"alarmState"`

	HelmetBind  *int `json:"helmetBind"`
	HelmetSOC   *int `json:"helmetSOC"`
	HelmetState *int `json:"helmetState"`

	HelmetAngleFault       *int `json:"helmetAngleFault"`
	HelmetCapacitanceFault *int `json:"helmetCapacitanceFault"`
	HelmetTinfraredFault   *int `json:"helmetTinfraredFault"`
	HelmetPressureFault    *int `json:"helmetPressureFault"`

	IsDisconnect *int `json:"isDisconnect"`

	RfidAck       *int   `json:"rfidAck"`
	RfidTimestamp *int64 `json:"rfidTimestamp"`

	MaintainAreaId *int64 `json:"maintainAreaId"`

	CarTagTypeIds []int `json:"carTagTypeIds"`

	Overload          *int `json:"overload"`
	OverloadThreshold *int `json:"overloadThreshold"`
	IsSupportOverload *int `json:"isSupportOverload"`

	CarHelmetState *int `json:"carHelmetState"`
}

// DeviceListCo mirrors the Java DeviceListCo (subset of device fields). Nulls are
// serialized; list fields emit [] (never null). Field order follows Java.
type DeviceListCo struct {
	Imei      *string `json:"imei"`
	CarId     *string `json:"carId"`
	ServiceId *int64  `json:"serviceId"`

	MaintainAreaId *int64 `json:"maintainAreaId"`

	Acc             *int     `json:"acc"`
	Defend          *int     `json:"defend"`
	Lat             *float64 `json:"lat"`
	Lng             *float64 `json:"lng"`
	Timestamp       *int64   `json:"timestamp"`
	IsOnline        *int     `json:"isOnline"`
	Voltage         *int     `json:"voltage"`
	ReportTime      *int64   `json:"reportTime"`
	IsOutofServAera *int     `json:"isOutofServAera"`
	NoParkId        *int64   `json:"noParkId"`
	ForParkId       *int64   `json:"forParkId"`
	NoRideParkId    *int64   `json:"noRideParkId"`
	BatteryId       *int64   `json:"batteryId"`
	BatteryLock     *int     `json:"batteryLock"`

	MoveAlarmTag *int `json:"moveAlarmTag"`
	RestBattery  *int `json:"restBattery"`

	LockTime   *int64 `json:"lockTime"`
	UnlockTime *int64 `json:"unlockTime"`

	RidingState    *int  `json:"ridingState"`
	OperationState []int `json:"operationState"`
	AlarmState     []int `json:"alarmState"`

	CarTagTypeIds []int `json:"carTagTypeIds"`
}

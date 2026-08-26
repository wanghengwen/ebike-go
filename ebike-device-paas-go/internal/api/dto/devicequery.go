package dto

// BlueToothTokenQry is the request for getBlueToothToken.
type BlueToothTokenQry struct {
	Imei           string          `json:"imei"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// BlueToothTokenCo mirrors the Java BlueToothTokenCo.
type BlueToothTokenCo struct {
	Token int `json:"token"`
}

// ImeiQry is the request for single-imei reads (querySaddleOverloadContact /
// queryCameraState).
type ImeiQry struct {
	Imei           string          `json:"imei"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// ImeiCmd is the request for removeDeviceTotalMiles.
type ImeiCmd struct {
	Imei           string          `json:"imei"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// SaddleOverloadContactCo mirrors the Java SaddleOverloadContactCo: the three
// saddle contact bits decoded from the stored payload.
type SaddleOverloadContactCo struct {
	FrontSaddleContact int `json:"frontSaddleContact"`
	CentSaddleContact  int `json:"centSaddleContact"`
	BackSaddleContact  int `json:"backSaddleContact"`
}

// CameraCacheCo mirrors the Java CameraCacheCo (parsed from the cached JSON).
type CameraCacheCo struct {
	Event     *int   `json:"event"`
	AngleDet  *int   `json:"angleDet"`
	Timestamp *int64 `json:"timestamp"`
}

// DeviceMapFakeQuery is the request for queryDeviceMapFake.
type DeviceMapFakeQuery struct {
	ServiceId      FlexInt64Value  `json:"serviceId"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// DeviceMapLocationCo mirrors the Java DeviceMapLocationCo.
type DeviceMapLocationCo struct {
	Lng   *float64 `json:"lng"`
	Lat   *float64 `json:"lat"`
	Imei  *string  `json:"imei"`
	CarId *string  `json:"carId"`
}

// DeviceMapFakeCo mirrors the Java DeviceMapFakeCo.
type DeviceMapFakeCo struct {
	Amount int                   `json:"amount"`
	Result []DeviceMapLocationCo `json:"result"`
}

// DeviceByBatteryQry is the request for queryDeviceByBattery.
type DeviceByBatteryQry struct {
	ServiceId      FlexInt64Value  `json:"serviceId"`
	MinRestBattery *int            `json:"minRestBattery"`
	MaxRestBattery *int            `json:"maxRestBattery"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// DeviceByBatteryCo mirrors the Java DeviceByBatteryCo.
type DeviceByBatteryCo struct {
	Imei           *string `json:"imei"`
	CarId          *string `json:"carId"`
	RestBattery    *int    `json:"restBattery"`
	ServiceId      *int64  `json:"serviceId"`
	MaintainAreaId *int64  `json:"maintainAreaId"`
	RidingState    *int    `json:"ridingState"`
	OperationState []int   `json:"operationState"`
	AlarmState     []int   `json:"alarmState"`
}

// DeviceByTotalMilesQry is the request for queryDeviceByTotalMiles.
type DeviceByTotalMilesQry struct {
	ServiceId      FlexInt64Value  `json:"serviceId"`
	MinTotalMiles  *float64        `json:"minTotalMiles"`
	MaxTotalMiles  *float64        `json:"maxTotalMiles"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// DeviceByTotalMilesCo mirrors the Java DeviceByTotalMilesCo.
type DeviceByTotalMilesCo struct {
	Imei           *string  `json:"imei"`
	CarId          *string  `json:"carId"`
	TotalMiles     *float64 `json:"totalMiles"`
	ServiceId      *int64   `json:"serviceId"`
	MaintainAreaId *int64   `json:"maintainAreaId"`
	RidingState    *int     `json:"ridingState"`
	OperationState []int    `json:"operationState"`
	AlarmState     []int    `json:"alarmState"`
}

// DeviceByNoOrderTimeQry is the request for queryDeviceByNoOrderTime.
type DeviceByNoOrderTimeQry struct {
	ServiceId      FlexInt64Value  `json:"serviceId"`
	MinNoOrderTime FlexInt64       `json:"minNoOrderTime"`
	MaxNoOrderTime FlexInt64       `json:"maxNoOrderTime"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// DeviceByNoOrderTimeCo mirrors the Java DeviceByNoOrderTimeCo.
type DeviceByNoOrderTimeCo struct {
	Imei           *string `json:"imei"`
	CarId          *string `json:"carId"`
	NoOrderTime    *int64  `json:"noOrderTime"`
	ServiceId      *int64  `json:"serviceId"`
	MaintainAreaId *int64  `json:"maintainAreaId"`
	RidingState    *int    `json:"ridingState"`
	OperationState []int   `json:"operationState"`
	AlarmState     []int   `json:"alarmState"`
}

// DeviceByStaticTimeQry is the request for queryDeviceByStaticTime.
type DeviceByStaticTimeQry struct {
	ServiceId      FlexInt64Value  `json:"serviceId"`
	MinStaticTime  FlexInt64       `json:"minStaticTime"`
	MaxStaticTime  FlexInt64       `json:"maxStaticTime"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// DeviceByStaticTimeCo mirrors the Java DeviceByStaticTimeCo.
type DeviceByStaticTimeCo struct {
	Imei           *string `json:"imei"`
	CarId          *string `json:"carId"`
	StaticTime     *int64  `json:"staticTime"`
	ServiceId      *int64  `json:"serviceId"`
	MaintainAreaId *int64  `json:"maintainAreaId"`
	RidingState    *int    `json:"ridingState"`
	OperationState []int   `json:"operationState"`
	AlarmState     []int   `json:"alarmState"`
}

// OneClickReturnNotifyCmd is the request for oneClickReturnNotify.
type OneClickReturnNotifyCmd struct {
	CarId          string          `json:"carId"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// GenDeviceMapFakeCmd is the request for genDeviceMapFake.
type GenDeviceMapFakeCmd struct {
	ServiceId      FlexInt64Value  `json:"serviceId"`
	Amount         int             `json:"amount"`
	FakeAmount     int             `json:"fakeAmount"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// DeviceOpeMapQry is the request for queryDeviceOpeMap.
type DeviceOpeMapQry struct {
	ServiceIdList  Int64Slice      `json:"serviceIdList"`
	RidingState    *int            `json:"ridingState"`
	OperationState *int            `json:"operationState"`
	AlarmState     *int            `json:"alarmState"`
	NoOrderTime    *float64        `json:"noOrderTime"`
	OrderTime      *float64        `json:"orderTime"`
	RestBattery    *int            `json:"restBattery"`
	BatteryId      FlexInt64       `json:"batteryId"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// DeviceGpsOpeMapCo mirrors the Java DeviceGpsOpeMapCo (map markers).
type DeviceGpsOpeMapCo struct {
	Imei           *string  `json:"imei"`
	CarId          *string  `json:"carId"`
	Lng            *float64 `json:"lng"`
	Lat            *float64 `json:"lat"`
	RidingState    *int     `json:"ridingState"`
	RestBattery    *int     `json:"restBattery"`
	OperationState []int    `json:"operationState"`
	AlarmState     []int    `json:"alarmState"`
	CarTagTypeIds  []int    `json:"carTagTypeIds"`
	CarTagNames    []string `json:"carTagNames"`
}

// DeviceOpeMapCo mirrors the Java DeviceOpeMapCo (operations dashboard aggregate).
type DeviceOpeMapCo struct {
	Operating  int `json:"operating"`
	UnSheleves int `json:"unSheleves"`
	CanRent    int `json:"canRent"`
	Booking    int `json:"booking"`
	Riding     int `json:"riding"`
	Parking    int `json:"parking"`
	Operation  int `json:"operation"`

	MoveCar       int `json:"moveCar"`
	ChangeBattery int `json:"changeBattery"`
	LowBattery    int `json:"lowBattery"`
	Fixing        int `json:"fixing"`
	DragBack      int `json:"dragBack"`

	MoveAlarm       int `json:"moveAlarm"`
	OutGfence       int `json:"outGfence"`
	NoParkingZone   int `json:"noParkingZone"`
	OutParkingZone  int `json:"outParkingZone"`
	PowerCut        int `json:"powerCut"`
	Offline         int `json:"offline"`
	OrderWithoutGps int `json:"orderWithoutGps"`
	Lost            int `json:"lost"`
	TooLongOrder    int `json:"tooLongOrder"`
	TooShortOrder   int `json:"tooShortOrder"`
	UnlockAbnormal  int `json:"unlockAbnormal"`
	HelmetLost      int `json:"helmetLost"`
	HelmetFault     int `json:"helmetFault"`

	RestBatteryLess10     int `json:"restBatteryLess10"`
	RestBatteryLess20     int `json:"restBatteryLess20"`
	RestBatteryLess30     int `json:"restBatteryLess30"`
	RestBatteryLess35     int `json:"restBatteryLess35"`
	RestBatteryLess40     int `json:"restBatteryLess40"`
	RestBatteryLessCustom int `json:"restBatteryLessCustom"`

	BatteryId map[string]int      `json:"batteryId"`
	Gps       []DeviceGpsOpeMapCo `json:"gps"`
}

// PartAnalysisResultCO mirrors the Java PartAnalysisResultCO. State serializes
// PartStateTypeA enum constants by name (Jackson default).
type PartAnalysisResultCO struct {
	Name    *string  `json:"name"`
	IzExist *bool    `json:"izExist"`
	CanUse  *bool    `json:"canUse"`
	UseType *int     `json:"useType"`
	Result  *bool    `json:"result"`
	State   []string `json:"state"`
}

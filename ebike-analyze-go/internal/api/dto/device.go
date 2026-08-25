package dto

type DeviceListQry struct {
	CommandContext *CommandContext `json:"commandContext"`
	ServiceIDList  []int64         `json:"serviceIdList"`
	ReportTime     *int64          `json:"reportTime"`
	ImeiList       []string        `json:"imeiList"`
	CarIDList      []string        `json:"carIdList"`
}

// DeviceListCo mirrors Java DeviceListCo. Most numeric fields are Java Integer/Long
// wrappers that serialize as JSON null when the corresponding protocol field was
// blank/absent, so they are pointer types here (not 0-defaulted primitives).
type DeviceListCo struct {
	Imei            string  `json:"imei"`
	CarID           string  `json:"carId"`
	ServiceID       int64   `json:"serviceId"`
	Acc             *int    `json:"acc"`
	Lat             float64 `json:"lat"`
	Lng             float64 `json:"lng"`
	Timestamp       int64   `json:"timestamp"`
	IsOnline        *int    `json:"isOnline"`
	Voltage         *int    `json:"voltage"`
	ReportTime      int64   `json:"reportTime"`
	IsOutofServAera *int    `json:"isOutofServAera"`
	NoParkID        *int64  `json:"noParkId"`
	ForParkID       *int64  `json:"forParkId"`
	NoRideParkID    *int64  `json:"noRideParkId"`
	BatteryID       *int64  `json:"batteryId"`
	BatteryLock     *int    `json:"batteryLock"`
	MoveAlarmTag    *int    `json:"moveAlarmTag"`
	RestBattery     *int    `json:"restBattery"`
	LockTime        *int64  `json:"lockTime"`
	UnlockTime      *int64  `json:"unlockTime"`
	RidingState     *int    `json:"ridingState"`
	OperationState  []int   `json:"operationState"`
	AlarmState      []int   `json:"alarmState"`
}

type DeviceCarStatisticsCo struct {
	TenantID                  string `json:"tenantId"`
	ServiceID                 int64  `json:"serviceId"`
	CanRent                   int    `json:"canRent"`
	Booking                   int    `json:"booking"`
	Riding                    int    `json:"riding"`
	Parking                   int    `json:"parking"`
	Operation                 int    `json:"operation"`
	FreeTimeOneToThree        int    `json:"freeTimeOneToThree"`
	FreeTimeThreeToSix        int    `json:"freeTimeThreeToSix"`
	FreeTimeSixToTwelve       int    `json:"freeTimeSixToTwelve"`
	FreeTimeHalfOrOneDay      int    `json:"freeTimeHalfOrOneDay"`
	FreeTimeOneOrTowDay       int    `json:"freeTimeOneOrTowDay"`
	FreeTimeTowDayMore        int    `json:"freeTimeTowDayMore"`
	VoltageZero               int    `json:"voltageZero"`
	VoltageZeroToTwenty       int    `json:"voltageZeroToTwenty"`
	VoltageTwentyToThirtyFive int    `json:"voltageTwentyToThirtyFive"`
	VoltageThirtyFiveMore     int    `json:"voltageThirtyFiveMore"`
}

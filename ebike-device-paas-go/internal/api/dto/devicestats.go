package dto

// IdListQry is the request for queryServiceStatisticsList.
type IdListQry struct {
	Ids            Int64Slice      `json:"ids"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// ServiceCarNumQry is the request for getCarNumByServiceId.
type ServiceCarNumQry struct {
	ServiceIdList  Int64Slice      `json:"serviceIdList"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// ServiceCarNumCo mirrors the Java ServiceCarNumCo (per-service on-shelf count).
type ServiceCarNumCo struct {
	ServiceId *int64 `json:"serviceId"`
	Count     int    `json:"count"`
}

// NoParamQuery is the request for parameterless endpoints (onlineNumByTenantId /
// carStatistics / getRackCarNumAll) — only the commandContext matters.
type NoParamQuery struct {
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// RackCarNumCo mirrors the Java RackCarNumCo (per-service on-shelf count).
type RackCarNumCo struct {
	TenantId  *string `json:"tenantId"`
	ServiceId *int64  `json:"serviceId"`
	Count     int     `json:"count"`
}

// CarStatisticsByServiceQry is the request for carStatisticsByService.
type CarStatisticsByServiceQry struct {
	ServiceIdList  Int64Slice      `json:"serviceIdList"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// CarServiceStatisticsCo mirrors the Java CarServiceStatisticsCo.
type CarServiceStatisticsCo struct {
	TenantId   *string `json:"tenantId"`
	ServiceId  *int64  `json:"serviceId"`
	CanRent    int     `json:"canRent"`
	Booking    int     `json:"booking"`
	Riding     int     `json:"riding"`
	Parking    int     `json:"parking"`
	Operation  int     `json:"operation"`
	LowBattery int     `json:"lowBattery"`

	FreeTimeOneToThree   int `json:"freeTimeOneToThree"`
	FreeTimeThreeToSix   int `json:"freeTimeThreeToSix"`
	FreeTimeSixToTwelve  int `json:"freeTimeSixToTwelve"`
	FreeTimeHalfOrOneDay int `json:"freeTimeHalfOrOneDay"`
	FreeTimeOneOrTowDay  int `json:"freeTimeOneOrTowDay"`
	FreeTimeTowDayMore   int `json:"freeTimeTowDayMore"`

	VoltageZero               int `json:"voltageZero"`
	VoltageZeroToTwenty       int `json:"voltageZeroToTwenty"`
	VoltageTwentyToThirtyFive int `json:"voltageTwentyToThirtyFive"`
	VoltageThirtyFiveMore     int `json:"voltageThirtyFiveMore"`
}

// CarParkingStatisticsCo mirrors the Java CarParkingStatisticsCo. Idle is the
// derived getIdle() value (freeTimeOneOrTowDay + freeTimeTowDayMore).
type CarParkingStatisticsCo struct {
	TenantId  *string `json:"tenantId"`
	ServiceId *int64  `json:"serviceId"`
	ParkingId *int64  `json:"parkingId"`
	CanRent   int     `json:"canRent"`
	Booking   int     `json:"booking"`
	Operation int     `json:"operation"`
	Alarm     int     `json:"alarm"`
	Fault     int     `json:"fault"`

	FreeTimeOneToThree   int `json:"freeTimeOneToThree"`
	FreeTimeThreeToSix   int `json:"freeTimeThreeToSix"`
	FreeTimeSixToTwelve  int `json:"freeTimeSixToTwelve"`
	FreeTimeHalfOrOneDay int `json:"freeTimeHalfOrOneDay"`
	FreeTimeOneOrTowDay  int `json:"freeTimeOneOrTowDay"`
	FreeTimeTowDayMore   int `json:"freeTimeTowDayMore"`

	Idle int `json:"idle"`
}

// CarStatisticsByServiceCo mirrors the Java CarStatisticsByServiceCo.
type CarStatisticsByServiceCo struct {
	ServiceStatistics []CarServiceStatisticsCo `json:"serviceStatistics"`
	ParkingStatistics []CarParkingStatisticsCo `json:"parkingStatistics"`
}

// CarStateServiceStatisticsCo mirrors the Java CarStateServiceStatisticsCo.
type CarStateServiceStatisticsCo struct {
	ServiceId      *int64 `json:"serviceId"`
	Operating      int    `json:"operating"`
	UnSheleves     int    `json:"unSheleves"`
	CanRent        int    `json:"canRent"`
	Booking        int    `json:"booking"`
	Riding         int    `json:"riding"`
	Parking        int    `json:"parking"`
	Operation      int    `json:"operation"`
	MoveCar        int    `json:"moveCar"`
	ChangeBattery  int    `json:"changeBattery"`
	LowBattery     int    `json:"lowBattery"`
	Fixing         int    `json:"fixing"`
	DragBack       int    `json:"dragBack"`
	OutParkingZone int    `json:"outParkingZone"`
}

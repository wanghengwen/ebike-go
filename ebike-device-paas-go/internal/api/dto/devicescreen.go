package dto

// DeviceScreenQry is the request for queryDeviceScreen / v2 / car_count.
type DeviceScreenQry struct {
	ServiceIdList  Int64Slice      `json:"serviceIdList"`
	RidingState    *int            `json:"ridingState,omitempty"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// DeviceScreenCo mirrors the Java DeviceScreenCo (single aggregate dashboard).
// All counters are always serialized (matching Java Integer fields set to counts).
type DeviceScreenCo struct {
	CanRent   int `json:"canRent"`
	Booking   int `json:"booking"`
	Riding    int `json:"riding"`
	Parking   int `json:"parking"`
	Operation int `json:"operation"`

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

	Gps []DeviceGpsCo `json:"gps"`
}

// DeviceScreenCoV2 mirrors the Java DeviceScreenCoV2. serviceId is nullable (set
// only by car_count); gps is nil for car_count.
type DeviceScreenCoV2 struct {
	ServiceId *int64 `json:"serviceId"`
	Total     int    `json:"total"`
	Offline   int    `json:"offline"`
	CanRent   int    `json:"canRent"`
	Booking   int    `json:"booking"`
	Riding    int    `json:"riding"`
	Parking   int    `json:"parking"`
	Operation int    `json:"operation"`

	FreeTimeZeroToThree      int `json:"freeTimeZeroToThree"`
	FreeTimeThreeToSix       int `json:"freeTimeThreeToSix"`
	FreeTimeSixToTwelve      int `json:"freeTimeSixToTwelve"`
	FreeTimeHalfOrOneDay     int `json:"freeTimeHalfOrOneDay"`
	FreeTimeOneOrTowDay      int `json:"freeTimeOneOrTowDay"`
	FreeTimeTowDayOrThreeDay int `json:"freeTimeTowDayOrThreeDay"`
	FreeTimeThreeDayMore     int `json:"freeTimeThreeDayMore"`

	VoltageZeroToTwenty  int `json:"voltageZeroToTwenty"`
	VoltageTwentyToForty int `json:"voltageTwentyToForty"`
	VoltageFortyToSixty  int `json:"voltageFortyToSixty"`
	VoltageSixtyToEighty int `json:"voltageSixtyToEighty"`
	VoltageEightyMore    int `json:"voltageEightyMore"`

	Gps []DeviceGpsCo `json:"gps"`
}

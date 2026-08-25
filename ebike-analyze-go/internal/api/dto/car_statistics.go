package dto

type CarStatisticsQuery struct {
	CommandContext *CommandContext `json:"commandContext"`
	CarID          string          `json:"carId"`
	ServiceID      *int64          `json:"serviceId"`
	PageNum        int             `json:"pageNum"`
	PageSize       int             `json:"pageSize"`
	SearchCount    *bool           `json:"searchCount"`
	Orders         []OrderItem     `json:"orders"`
	LastRecordID   string          `json:"lastRecordId"`
}

type CarStatisticsListQuery struct {
	CommandContext *CommandContext `json:"commandContext"`
	ServiceID      JSONInt64       `json:"serviceId"`
	CarIds         []string        `json:"carIds"`
	StartTime      DateTimeValue   `json:"startTime"`
	EndTime        DateTimeValue   `json:"endTime"`
}

type CarServiceStatisticsQuery struct {
	CommandContext *CommandContext `json:"commandContext"`
	ServiceIds     []int64         `json:"serviceIds"`
}

// CarStatisticsCO mirrors Java api.dto.CarStatisticsCO. /car-statistics/page only
// populates the CarStatisticsEntity field subset; /queryByCarList includes all
// aggregated counters below.
type CarStatisticsCO struct {
	CarID                   string  `json:"carId"`
	Imei                    *string `json:"imei"`
	OrderCount              *int    `json:"orderCount"`
	OrderCost               *int64 `json:"orderCost"`
	RidingDistance          *int64 `json:"ridingDistance"`
	RidingTime              *int64 `json:"ridingTime"`
	DdMissOrder             *int   `json:"ddMissOrder"`
	OperationMissOrderCount *int   `json:"operationMissOrderCount"`
	ChangeBatteryCount      *int   `json:"changeBatteryCount"`
	RepairCount             *int   `json:"repairCount"`
	MoveCarCount            *int   `json:"moveCarCount"`
}

// CarServiceStatisticsCo mirrors Java api.dto.CarServiceStatisticsCo, including
// the BaseDO audit fields copied by Java's ConvertorHelper.convert(...).
type CarServiceStatisticsCo struct {
	ID                        int64               `json:"id"`
	ServiceID                 int64               `json:"serviceId"`
	CanRent                   int                 `json:"canRent"`
	Booking                   int                 `json:"booking"`
	Riding                    int                 `json:"riding"`
	Parking                   int                 `json:"parking"`
	Operation                 int                 `json:"operation"`
	LowBattery                int                 `json:"lowBattery"`
	FreeTimeOneToThree        int                 `json:"freeTimeOneToThree"`
	FreeTimeThreeToSix        int                 `json:"freeTimeThreeToSix"`
	FreeTimeSixToTwelve       int                 `json:"freeTimeSixToTwelve"`
	FreeTimeHalfOrOneDay      int                 `json:"freeTimeHalfOrOneDay"`
	FreeTimeOneOrTowDay       int                 `json:"freeTimeOneOrTowDay"`
	FreeTimeTowDayMore        int                 `json:"freeTimeTowDayMore"`
	VoltageZero               int                 `json:"voltageZero"`
	VoltageZeroToTwenty       int                 `json:"voltageZeroToTwenty"`
	VoltageTwentyToThirtyFive int                 `json:"voltageTwentyToThirtyFive"`
	VoltageThirtyFiveMore     int                 `json:"voltageThirtyFiveMore"`
	TenantID                  string              `json:"tenantId"`
	CreatedPin                string              `json:"createdPin"`
	CreatedAt                 *LocalDateTimeValue `json:"createdAt"`
	UpdatedPin                string              `json:"updatedPin"`
	UpdatedAt                 *LocalDateTimeValue `json:"updatedAt"`
	Version                   int                 `json:"version"`
	IzDel                     bool                `json:"izDel"`
}

package dto

type SiteStatisticsCmd struct {
	CommandContext *CommandContext `json:"commandContext"`
	ParkingID      *int64          `json:"parkingId"`
}

type SiteStatisticsCO struct {
	ID         int64         `json:"id"`
	OrderCount int           `json:"orderCount"`
	ParkingID  int64         `json:"parkingId"`
	OrderCost  int64         `json:"orderCost"`
	StartTime  DateTimeValue `json:"startTime"`
	EndTime    DateTimeValue `json:"endTime"`
}

type ParkingStatisticalPageQuery struct {
	CommandContext *CommandContext `json:"commandContext"`
	PageNum        int             `json:"pageNum"`
	PageSize       int             `json:"pageSize"`
	SearchCount    *bool           `json:"searchCount"`
	Orders         []OrderItem     `json:"orders"`
	LastRecordID   string          `json:"lastRecordId"`
	ServiceID      int64           `json:"serviceId" binding:"required"`
	AreaIds        []int64         `json:"areaIds" binding:"required"`
	OrderStrategy  string          `json:"orderStrategy"`
}

type ParkingStationStatisticalCO struct {
	ServiceID   int64 `json:"serviceId"`
	ParkingID   int64 `json:"parkingId"`
	CanRent     int   `json:"canRent"`
	Idle        int   `json:"idle"`
	SiteOut     int   `json:"siteOut"`
	DdMissOrder int   `json:"ddMissOrder"`
	Booking     int   `json:"booking"`
	Operation   int   `json:"operation"`
	Alarm       int   `json:"alarm"`
	Fault       int   `json:"fault"`
	Idle13      int   `json:"idle13"`
	Idle36      int   `json:"idle36"`
	Idle612     int   `json:"idle612"`
	Idle1224    int   `json:"idle1224"`
	Idle2448    int   `json:"idle2448"`
	Idle48      int   `json:"idle48"`
}

type OneParkingAnalyzeCmd struct {
	CommandContext *CommandContext `json:"commandContext"`
	ServiceID      int64           `json:"serviceId" binding:"required"`
	ParkingID      int64           `json:"parkingId" binding:"required"`
}

type OneParkingAnalyzeCO struct {
	ServiceID int64            `json:"serviceId"`
	ParkingID int64            `json:"parkingId"`
	CanRent   int              `json:"canRent"`
	Booking   int              `json:"booking"`
	Operation int              `json:"operation"`
	CarMap    map[string][]int `json:"carMap"`
}

type ParkingOrderStatisticQry struct {
	CommandContext *CommandContext `json:"commandContext"`
	ServiceID      JSONInt64       `json:"serviceId"`
	MaintainAreaID *JSONInt64      `json:"maintainAreaId"`
	StartTime      DateTimeValue   `json:"startTime"`
	EndTime        DateTimeValue   `json:"endTime"`
	OrderStrategy  string          `json:"orderStrategy"`
	PageNum        int             `json:"pageNum"`
	PageSize       int             `json:"pageSize"`
	SearchCount    *bool           `json:"searchCount"`
	Orders         []OrderItem     `json:"orders"`
	LastRecordID   string          `json:"lastRecordId"`
}

type ParkingOrderStatisticListQry struct {
	CommandContext *CommandContext `json:"commandContext"`
	ServiceID      JSONInt64       `json:"serviceId" binding:"required"`
	ParkingIds     JSONInt64Slice  `json:"parkingIds" binding:"required"`
	StartTime      DateTimeValue   `json:"startTime" binding:"required"`
	EndTime        DateTimeValue   `json:"endTime" binding:"required"`
}

type ParkingOrderStatisticCO struct {
	ParkingID               int64    `json:"parkingId"`
	Name                    *string  `json:"name"`
	AreaSize                *float64 `json:"areaSize"`
	RideOrderCount          int      `json:"rideOrderCount"`
	RideOrderCost           int      `json:"rideOrderCost"`
	ReturnOrderCount        int      `json:"returnOrderCount"`
	LowPowerMissOrderCount  int      `json:"lowPowerMissOrderCount"`
	OperationMissOrderCount int      `json:"operationMissOrderCount"`
}

type ParkingInAndOutflowQry struct {
	CommandContext *CommandContext `json:"commandContext"`
	ParkingID      int64           `json:"parkingId"`
	TimeDimension  int             `json:"timeDimension"`
}

type ParkingInAndOutflowCo struct {
	Influx  []int `json:"influx"`
	OutFlow []int `json:"outFlow"`
	CanRent []int `json:"canRent"`
}

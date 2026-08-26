package dto

// DeviceListQry is the request for /device/paas/device/list.
type DeviceListQry struct {
	ServiceIdList  Int64Slice      `json:"serviceIdList"`
	ReportTime     FlexInt64       `json:"reportTime,omitempty"`
	ImeiList       []string        `json:"imeiList,omitempty"`
	CarIdList      []string        `json:"carIdList,omitempty"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// DeviceLocationQry is the request for /device/paas/device/eBikeLocation.
type DeviceLocationQry struct {
	ServiceId      FlexInt64Value  `json:"serviceId"`
	Lat            float64         `json:"lat"`
	Lng            float64         `json:"lng"`
	Radius         *float64        `json:"radius,omitempty"`
	Limit          *int            `json:"limit,omitempty"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// CarIdListQry is the request for /device/paas/device/listByCarIdList.
type CarIdListQry struct {
	CarIdList      []string        `json:"carIdList"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// DevicePageQry is the request for /device/paas/device/filterList and
// /device/paas/device/page (platform). PageNum/PageSize only matter for page.
type DevicePageQry struct {
	ServiceId      FlexInt64Value  `json:"serviceId"`
	PageNum        int             `json:"pageNum,omitempty"`
	PageSize       int             `json:"pageSize,omitempty"`
	RidingState    *int            `json:"ridingState,omitempty"`
	OperationState *int            `json:"operationState,omitempty"`
	AlarmState     *int            `json:"alarmState,omitempty"`
	NoOrderTime    *float64        `json:"noOrderTime,omitempty"`
	OrderTime      *float64        `json:"orderTime,omitempty"`
	RestBattery    *int            `json:"restBattery,omitempty"`
	BatteryId      FlexInt64       `json:"batteryId,omitempty"`
	HelmetSOC      *int            `json:"helmetSOC,omitempty"`
	CarHelmetState *int            `json:"carHelmetState,omitempty"`
	CarTagTypeIds  []int           `json:"carTagTypeIds,omitempty"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// DevicePageBusQry is the request for /device/paas/device/pageBus (merchant).
type DevicePageBusQry struct {
	ServiceId       FlexInt64Value  `json:"serviceId"`
	PageNum         int             `json:"pageNum,omitempty"`
	PageSize        int             `json:"pageSize,omitempty"`
	IzStateAnd      bool            `json:"izStateAnd"`
	RidingStates    []int           `json:"ridingStates,omitempty"`
	OperationStates []int           `json:"operationStates,omitempty"`
	AlarmStates     []int           `json:"alarmStates,omitempty"`
	NoOrderTime     *float64        `json:"noOrderTime,omitempty"`
	OrderTime       *float64        `json:"orderTime,omitempty"`
	RestBattery     *int            `json:"restBattery,omitempty"`
	MinCarId        string          `json:"minCarId,omitempty"`
	MaxCarId        string          `json:"maxCarId,omitempty"`
	CarTagTypeIds   []int           `json:"carTagTypeIds,omitempty"`
	IzFilterRiding  *bool           `json:"izFilterRiding,omitempty"`
	BatteryIds      Int64Slice      `json:"batteryIds,omitempty"`
	CommandContext  *CommandContext `json:"commandContext,omitempty"`
}

// DevicePageDTO mirrors the Java PageDTO<DevicePageCo> serialization.
type DevicePageDTO struct {
	Count       int64          `json:"count"`
	PageNum     int            `json:"pageNum"`
	PageSize    int            `json:"pageSize"`
	Orders      interface{}    `json:"orders"`
	SearchCount bool           `json:"searchCount"`
	List        []DevicePageCo `json:"list"`
}

// DevicePageCo mirrors the Java DevicePageCo. For filterList, address /
// scanAddress / carHelmetState are not populated (no amap, plain copy).
type DevicePageCo struct {
	Imei           *string  `json:"imei"`
	CarId          *string  `json:"carId"`
	Timestamp      *int64   `json:"timestamp"`
	IsOnline       *int     `json:"isOnline"`
	RestBattery    *int     `json:"restBattery"`
	RidingState    *int     `json:"ridingState"`
	OperationState []int    `json:"operationState"`
	AlarmState     []int    `json:"alarmState"`
	Lng            *float64 `json:"lng"`
	Lat            *float64 `json:"lat"`
	Address        *string  `json:"address"`
	ScanLng        *float64 `json:"scanLng"`
	ScanLat        *float64 `json:"scanLat"`
	ScanAddress    *string  `json:"scanAddress"`
	CarHelmetState *int     `json:"carHelmetState"`
	HelmetSOC      *int     `json:"helmetSOC"`
	CarTagTypeIds  []int    `json:"carTagTypeIds"`
}

// DeviceGpsCo mirrors the Java DeviceGpsCo (imei/carId + lng/lat with defaults).
type DeviceGpsCo struct {
	Imei  *string  `json:"imei"`
	CarId *string  `json:"carId"`
	Lng   *float64 `json:"lng"`
	Lat   *float64 `json:"lat"`
}

// CarImeiCo mirrors the Java CarImeiCo (imei <-> carId binding pair).
type CarImeiCo struct {
	Imei  *string `json:"imei"`
	CarId *string `json:"carId"`
}

// DeviceLocationCo mirrors the Java DeviceLocationCo (nearby usable bikes).
type DeviceLocationCo struct {
	CarId       *string  `json:"carId"`
	Imei        *string  `json:"imei"`
	Lat         *float64 `json:"lat"`
	Lng         *float64 `json:"lng"`
	Distance    *float64 `json:"distance"`
	RestBattery *int     `json:"restBattery"`
	RestMileage *int     `json:"restMileage"`
}

package dto

import "time"

// OrderQueryCmd mirrors Java OrderQueryCmd (extends PageQuery). Only a subset
// of fields are consumed by buildQueryBuilder (see order_query.go); the rest
// are kept here purely for request/API-contract completeness, matching Java.
type OrderQueryCmd struct {
	CommandContext   *CommandContext `json:"commandContext"`
	PageNum          int             `json:"pageNum"`
	PageSize         int             `json:"pageSize"`
	SearchCount      *bool           `json:"searchCount"`
	Orders           []OrderItem     `json:"orders"`
	LastRecordID     string          `json:"lastRecordId"`
	ID               *int64          `json:"id"`
	OrderID          *int64          `json:"orderId"`
	UserPin          string          `json:"userPin"`
	UserPinList      []string        `json:"userPinList"`
	ServiceID        *int64          `json:"serviceId"`
	EndMaintainArea  *int64          `json:"endMaintainArea"`
	CarType          *int            `json:"carType"`
	OriginCost       *int            `json:"originCost"`
	PayCost          *int            `json:"payCost"`
	ModifyCost       *int            `json:"modifyCost"`
	Mile             *int            `json:"mile"`
	RidingTime       *int64          `json:"ridingTime"`
	IzPaid           *int            `json:"izPaid"`
	IzComplained     *int            `json:"izComplained"`
	IzModified       *int            `json:"izModified"`
	InvoiceID        *int64          `json:"invoiceId"`
	Imei             string          `json:"imei"`
	CarID            string          `json:"carId"`
	PaidInfo         string          `json:"paidInfo"`
	StartLat         *float64        `json:"startLat"`
	StartLng         *float64        `json:"startLng"`
	EndLat           *float64        `json:"endLat"`
	EndLng           *float64        `json:"endLng"`
	PayTime          *time.Time      `json:"payTime"`
	StartParkingID   *int64          `json:"startParkingId"`
	EndParkingID     *int64          `json:"endParkingId"`
	StartTime        []int64         `json:"startTime"`
	EndTime          []int64         `json:"endTime"`
	Equal            *int            `json:"equal"`
	Phone            string          `json:"phone"`
	Name             string          `json:"name"`
	MinCost          *int            `json:"minCost"`
	MaxCost          *int            `json:"maxCost"`
	IzCanInvoiced    *bool           `json:"izCanInvoiced"`
	CostEqual        *int            `json:"costEqual"`
	TimeEqual        *int            `json:"timeEqual"`
	IzStandardReturn *int            `json:"izStandardReturn"`
}

// OrderItem mirrors Java com.baomidou...OrderItem (PageQuery.orders), used
// only when the caller supplies explicit sort instructions.
type OrderItem struct {
	Column string `json:"column"`
	Asc    bool   `json:"asc"`
}

type OrderCO struct {
	ID              int64         `json:"id"`
	OrderID         int64         `json:"orderId"`
	UserPin         string        `json:"userPin"`
	Phone           string        `json:"phone"`
	Name            string        `json:"name"`
	ServiceID       int64         `json:"serviceId"`
	CarType         int           `json:"carType"`
	OriginCost      int           `json:"originCost"`
	PayCost         int           `json:"payCost"`
	ModifyCost      int           `json:"modifyCost"`
	Mile            int           `json:"mile"`
	RidingDistance  int           `json:"ridingDistance"`
	RidingTime      int64         `json:"ridingTime"`
	IzPaid          int           `json:"izPaid"`
	Discount        *float32      `json:"discount"`
	IzComplained    int           `json:"izComplained"`
	IzModified      *int          `json:"izModified"`
	IzRepair        *int          `json:"izRepair"`
	InvoiceID       *int64        `json:"invoiceId"`
	Imei            string        `json:"imei"`
	CarID           string        `json:"carId"`
	PaidInfo        *string       `json:"paidInfo"`
	StartLat        float64       `json:"startLat"`
	StartLng        float64       `json:"startLng"`
	UserStartLat    *float64      `json:"userStartLat"`
	UserStartLng    *float64      `json:"userStartLng"`
	EndLat          float64       `json:"endLat"`
	EndLng          float64       `json:"endLng"`
	UserEndLat      *float64      `json:"userEndLat"`
	UserEndLng      *float64      `json:"userEndLng"`
	PayTime         DateTimeValue `json:"payTime"`
	StartTime       DateTimeValue `json:"startTime"`
	EndTime         DateTimeValue `json:"endTime"`
	StartParkingID  int64         `json:"startParkingId"`
	EndParkingID    int64         `json:"endParkingId"`
	IzDiscount      *int          `json:"izDiscount"`
	IzRidingCard    *int          `json:"izRidingCard"`
	IzActityFree    *int          `json:"izActityFree"`
	IzFavorable     *int          `json:"izFavorable"`
	RechargeCost    *int          `json:"rechargeCost"`
	PresentCost     *int          `json:"presentCost"`
	Penalty         *int          `json:"penalty"`
	PayType         *int          `json:"payType"`
	TotalRefundCost *int          `json:"totalRefundCost"`
	PenaltyType     int           `json:"penaltyType"`
	HelmetPenalty   *int          `json:"helmetPenalty"`
}

type OrderLatLng struct {
	L string `json:"l"`
	P *bool  `json:"p"`
}

type OrderLocationAnalyzeCO struct {
	Total  int           `json:"total"`
	PCount int           `json:"pcount"`
	NCount int           `json:"ncount"`
	OList  []OrderLatLng `json:"olist"`
}

func EmptyOrderLocationAnalyze() OrderLocationAnalyzeCO {
	return OrderLocationAnalyzeCO{OList: []OrderLatLng{}}
}

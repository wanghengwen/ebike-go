package dto

// LocationCmd matches Java LocationCmd.
type LocationCmd struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

// CarCmd matches Java CarCmd.
type CarCmd struct {
	CarLocation *LocationCmd `json:"carLocation"`
	Parts       []string     `json:"parts"`
	CarId       string       `json:"carId"`
	Imei        string       `json:"imei"`
	Version     string       `json:"version"`
	ReportTime  *int64       `json:"reportTime"`
}

// UserCmd matches Java UserCmd.
type UserCmd struct {
	UserLocation *LocationCmd `json:"userLocation"`
}

// BlueResult matches Java BlueResult.
type BlueResult struct {
	Name    string                 `json:"name"`
	IzExist *bool                  `json:"izExist"`
	CanUse  *bool                  `json:"canUse"`
	UseType *int                   `json:"useType"`
	Result  *bool                  `json:"result"`
	State   map[string]interface{} `json:"state"`
}

// ReturnCarCmd matches Java ReturnCarCmd extends Command.
type ReturnCarCmd struct {
	Command
	ServiceAreaId         *int64       `json:"serviceAreaId" binding:"required"`
	CarCmd                *CarCmd      `json:"carCmd" binding:"required"`
	UserCmd               *UserCmd     `json:"userCmd"`
	CheckType             *int         `json:"checkType"`
	OrderId               *int64       `json:"orderId"`
	IzReturn              *bool        `json:"izReturn"`
	Result                []BlueResult `json:"result"`
	IzFrontSuppotFullPile *bool        `json:"izFrontSuppotFullPile"`
}

// PartAnalysisResultCO matches Java PartAnalysisResultCO.
type PartAnalysisResultCO struct {
	Name    *string                  `json:"name"`
	IzExist *bool                    `json:"izExist"`
	CanUse  *bool                    `json:"canUse"`
	UseType *int                     `json:"useType"`
	Result  *bool                    `json:"result"`
	State   []map[string]interface{} `json:"state"`
	Ext     map[string]interface{}   `json:"ext"`
}

// PartNamePtr maps a part name to JSON null when empty, matching Java null name fields.
func PartNamePtr(name string) *string {
	if name == "" {
		return nil
	}
	s := name
	return &s
}

// ReturnCarCO matches Java ReturnCarCO.
type ReturnCarCO struct {
	ReturnType      interface{}            `json:"returnType"`
	IzCanReturn     *bool                  `json:"izCanReturn"`
	Parking         map[string]interface{} `json:"parking"`
	NoParkingCmd    map[string]interface{} `json:"noParkingCmd"`
	MaintainAreaCmd map[string]interface{} `json:"maintainAreaCmd"`
	IzNearService   *bool                  `json:"izNearService"`
	PartResult      []PartAnalysisResultCO `json:"partResult"`
	IzDeviceWeakNet *bool                  `json:"izDeviceWeakNet"`
	IzHelmetReign   *bool                  `json:"izHelmetReign"`
}

// RideCarCmd matches Java RideCarCmd extends Command.
type RideCarCmd struct {
	Command
	ServiceAreaId  *int64       `json:"serviceAreaId" binding:"required"`
	CarLocation    *LocationCmd `json:"carLocation" binding:"required"`
	UserLocation   *LocationCmd `json:"userLocation"`
	CarCmd         *CarCmd      `json:"carCmd" binding:"required"`
	Result         []BlueResult `json:"result"`
	IzRemoteUnlock *bool        `json:"izRemoteUnlock"`
}

// RideCarCO matches Java RideCarCO.
type RideCarCO struct {
	RideResult *bool                  `json:"rideResult"`
	Reason     interface{}            `json:"reason"`
	ParkingCO  map[string]interface{} `json:"parkingCO"`
	PartResult []PartAnalysisResultCO `json:"partResult"`
}

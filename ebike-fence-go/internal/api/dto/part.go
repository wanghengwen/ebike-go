package dto

// CheckingPartCmd matches Java com.xyy.ebike.fence.api.dto.packing.CheckingPartCmd.
type CheckingPartCmd struct {
	Command
	Parts     []string `json:"parts,omitempty"`
	FenceId   *int64   `json:"fenceId,omitempty"`
	ServiceId *int64   `json:"serviceId,omitempty"`
	Imei      string   `json:"imei,omitempty"`
	CarId     string   `json:"carId,omitempty"`
	CheckType *int     `json:"checkType,omitempty"`
	OrderId   *int64   `json:"orderId,omitempty"`
	Lat       *float64 `json:"lat,omitempty"`
	Lng       *float64 `json:"lng,omitempty"`
}

// HelmetCheckCmd matches Java com.xyy.ebike.fence.api.dto.packing.HelmetCheckCmd.
type HelmetCheckCmd struct {
	Command
	Imei      string `json:"imei" binding:"required"`
	ServiceId *int64 `json:"serviceId" binding:"required"`
}

// HelmetCmd matches Java com.xyy.ebike.fence.api.dto.part.HelmetCmd.
type HelmetCmd struct {
	Command
	Imei      string `json:"imei,omitempty"`
	ServiceId *int64 `json:"serviceId,omitempty"`
	CarId     string `json:"carId" binding:"required"`
	OrderId   *int64 `json:"orderId,omitempty"`
}

// ImeiCmd matches Java com.xyy.ebike.fence.api.dto.part.ImeiCmd.
type ImeiCmd struct {
	Command
	Imei string `json:"imei" binding:"required"`
}

// BlueTBeaconInfoCo matches Java com.xyy.ebike.fence.api.dto.part.BlueTBeaconInfoCo.
type BlueTBeaconInfoCo struct {
	Event       *int     `json:"event,omitempty"`
	TBeaconAddr string   `json:"tBeaconAddr,omitempty"`
	TBeaconId   string   `json:"tBeaconId,omitempty"`
	TBeaconSOC  *int     `json:"tBeaconSOC,omitempty"`
	RealRssi    *int     `json:"realRssi,omitempty"`
	TBeaconVsn  string   `json:"tBeaconVsn,omitempty"`
	Lng         *float64 `json:"lng,omitempty"`
	Lat         *float64 `json:"lat,omitempty"`
	Timestamp   *int64   `json:"timestamp,omitempty"`
}

// ConfigDto matches Java com.xyy.ebike.fence.api.dto.config.systemconfig.co.ConfigDto.
type ConfigDto struct {
	ConfigBaseItemCO *ConfigBaseItemCO `json:"configBaseItemCO"`
	ConfigBackcarCO  interface{}       `json:"configBackcarCO"`
	ConfigUseCarCO   interface{}       `json:"configUseCarCO"`
}

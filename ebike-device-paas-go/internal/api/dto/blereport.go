package dto

// BLE command-report DTOs (EcuBleCommandReportApi). Write/report endpoints; not
// shadow-compared. The shared base mirrors BaseEcuBleReportCommand.

// BaseBleReport is the common BLE report base (imei + carId + phone gps).
type BaseBleReport struct {
	CommandContext *CommandContext `json:"commandContext"`
	Imei           string          `json:"imei"`
	CarID          string          `json:"carId"`
	Lng            *float64        `json:"lng"`
	Lat            *float64        `json:"lat"`
}

// BleLockReportCmd mirrors BleLockReportCmd (izRiskControl defaults true).
type BleLockReportCmd struct {
	BaseBleReport
	Acc           *int  `json:"acc"`
	IzRiskControl *bool `json:"izRiskControl"`
}

// BleDefendReportCmd mirrors BleDefendReportCmd.
type BleDefendReportCmd struct {
	BaseBleReport
	Defend *int `json:"defend"`
}

// BleHelmetLockReportCmd mirrors BleHelmetLockReportCmd.
type BleHelmetLockReportCmd struct {
	BaseBleReport
	Sw *int `json:"sw"`
}

// BleRearWheelLockReportCmd mirrors BleRearWheelLockReportCmd.
type BleRearWheelLockReportCmd struct {
	BaseBleReport
	Sw *int `json:"sw"`
}

// BleBatteryCompartmentReportCmd mirrors BleBatteryCompartmentReportCmd.
type BleBatteryCompartmentReportCmd struct {
	BaseBleReport
	Sw *int `json:"sw"`
}

// BleDeviceInfoReportCmd mirrors BleDeviceInfoReportCmd (state snapshot upload).
type BleDeviceInfoReportCmd struct {
	BaseBleReport
	Gsm          *int   `json:"gsm"`
	Voltage      *int   `json:"voltage"`
	Timestamp    *int64 `json:"timestamp"`
	Speed        *int   `json:"speed"`
	Course       *int   `json:"course"`
	TotalMiles   *int   `json:"totalMiles"`
	Defend       *int   `json:"defend"`
	Acc          *int   `json:"acc"`
	Helmet6React *int   `json:"helmet6React"`
	Helmet6Lock  *int   `json:"helmet6Lock"`
}

// BleRfidInfoReportCmd mirrors BleRfidInfoReportCmd.
type BleRfidInfoReportCmd struct {
	BaseBleReport
	RfidAck       *int   `json:"rfidAck"`
	RfidCarId     string `json:"rfidCarId"`
	RfidTimestamp *int64 `json:"rfidTimestamp"`
}

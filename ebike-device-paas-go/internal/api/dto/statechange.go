package dto

// State-change command DTOs (StateChangeApi). These are write endpoints whose
// side effects (Redis device-info SETRANGE / C34 Kafka) are NOT shadow-compared.

// RidingStateChangeCmd mirrors RidingStateChangeCmd.
type RidingStateChangeCmd struct {
	CommandContext *CommandContext `json:"commandContext"`
	CarID          string          `json:"carId"`
	Imei           string          `json:"imei"`
	ServiceID      FlexInt64       `json:"serviceId"`
	RidingState    *int            `json:"ridingState"`
	OperationState []int           `json:"operationState"`
}

// AlarmStateChangeCmd mirrors AlarmStateChangeCmd.
type AlarmStateChangeCmd struct {
	CommandContext *CommandContext `json:"commandContext"`
	CarID          string          `json:"carId"`
	Imei           string          `json:"imei"`
	ServiceID      FlexInt64       `json:"serviceId"`
	AlarmType      []int           `json:"alarmType"`
}

// LocationChangeCmd mirrors LocationChangeCmd.
type LocationChangeCmd struct {
	CommandContext *CommandContext `json:"commandContext"`
	ImeiList       []string        `json:"imeiList"`
	Lng            *float64        `json:"lng"`
	Lat            *float64        `json:"lat"`
}

// ScanLocationChangeCmd mirrors ScanLocationChangeCmd.
type ScanLocationChangeCmd struct {
	CommandContext *CommandContext `json:"commandContext"`
	CarID          string          `json:"carId"`
	Lng            *float64        `json:"lng"`
	Lat            *float64        `json:"lat"`
}

// Package dto holds request/response DTOs ported from the Java api module
// (com.xyy.ebike.device.paas.api.dto). JSON tags match the Java field names so
// shadow diffing against the legacy service stays byte-comparable.
package dto

// CommandContext carries caller/tenant metadata, mirroring the Java
// CommandContext embedded in every request body (CommandContextAspect audit fields).
type CommandContext struct {
	TraceID  string `json:"traceId,omitempty"`
	TenantID string `json:"tenantId,omitempty"`
	Pin      string `json:"pin,omitempty"`
	UserId   string `json:"userId,omitempty"`
	// Platform mirrors CommandContext.platform (ios/android/wechat/pc/other);
	// the voice find-car rate limiter only triggers for platform=="wechat".
	Platform      string  `json:"platform,omitempty"`
	IP            *string `json:"ip,omitempty"`
	DeviceID      *string `json:"deviceId,omitempty"`
	Source        *string `json:"source,omitempty"`
	Name          *string `json:"name,omitempty"`
	StressTesting *bool   `json:"stressTesting,omitempty"`
}

// BaseEcuQuery is the common query base (imei + commandContext).
type BaseEcuQuery struct {
	Imei           string          `json:"imei"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// ImeiListQry is the request for /device/paas/getDeviceRealGpsList.
type ImeiListQry struct {
	ImeiList       []string        `json:"imeiList"`
	CarIdList      []string        `json:"carIdList,omitempty"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// DeviceRealGpsCo is the response element for batch realtime GPS, matching the
// Java DeviceRealGpsCo (imei, lng, lat).
type DeviceRealGpsCo struct {
	Imei string   `json:"imei"`
	Lng  *float64 `json:"lng"`
	Lat  *float64 `json:"lat"`
}

// EcuQuery is the shared request shape for the forwarded device-command queries
// (BaseEcuCommand fields). InnerParamQry / BlueTBeaconQry / BleHelmetInfoQry are
// structurally identical to this in the Java api module.
type EcuQuery struct {
	Async          *bool           `json:"async,omitempty"`
	CarId          string          `json:"carId,omitempty"`
	Imei           string          `json:"imei,omitempty"`
	Payload        string          `json:"payload,omitempty"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// DeviceInfoQry adds isCameraEnable (Java DeviceInfoQry).
type DeviceInfoQry struct {
	EcuQuery
	IsCameraEnable *int `json:"isCameraEnable,omitempty"`
}

// CommandResult mirrors the Java CommandResult<T> envelope returned by the
// command-query endpoints. Result is left as interface{} since each endpoint
// supplies its own typed (or map) payload. Java serializes with Jackson's
// default ALWAYS inclusion, so every field is emitted (null when absent); the
// optional jobId/payload are pointers so an absent value renders as null (not
// "" or a dropped key).
type CommandResult struct {
	JobId   *string     `json:"jobId"`
	Async   *bool       `json:"async"`
	EcuCode *string     `json:"ecuCode"`
	Payload *string     `json:"payload"`
	Result  interface{} `json:"result"`
}

// EcuCodeValue returns the ecuCode string, or "" when unset. Java's ecuCode is a
// nullable String (null when the gateway result has no inner result); both null
// and "0" are treated as success by DeviceResultHelper, so callers compare with
// "" or "0". Serialization keeps nil as JSON null to match Java's Jackson output.
func (c CommandResult) EcuCodeValue() string {
	if c.EcuCode == nil {
		return ""
	}
	return *c.EcuCode
}

// InnerParamQryCo mirrors the Java InnerParamQryCo. freqNorm/freqMove are pulled
// from the snake_case freq_norm/freq_move in the gateway result.
type InnerParamQryCo struct {
	Mode                       *int     `json:"mode"`
	FreqNorm                   *int     `json:"freqNorm"`
	FreqMove                   *int     `json:"freqMove"`
	IsOverSpeedOn              *int     `json:"isOverSpeedOn"`
	IsAutoLockOn               *int     `json:"isAutoLockOn"`
	IsMoveAlarmOn              *int     `json:"isMoveAlarmOn"`
	AutoLockPeriod             *int     `json:"autoLockPeriod"`
	AudioRatio                 *float64 `json:"audioRatio"`
	IsNightVoiceOn             *int     `json:"isNightVoiceOn"`
	IsFenceEnable              *int     `json:"isFenceEnable"`
	FenceVersion               *int64   `json:"fenceVersion"`
	IsTBeaconEnable            *int     `json:"isTBeaconEnable"`
	TBeaconthresholdCompensate *int     `json:"tBeaconthresholdCompensate"`
	IsTurnOverEnable           *int     `json:"isTurnOverEnable"`
	TurnOverAngle              *int     `json:"turnOverAngle"`
	TurnOverAudioPeriod        *int     `json:"turnOverAudioPeriod"`
	AudioRatioOverSpeed        *float64 `json:"audioRatioOverSpeed"`
	LowVoltageThreshold        *int     `json:"lowVoltageThreshold"`
	SlopeMinAngle              *int     `json:"slopeMinAngle"`
	SlopeMaxAngle              *int     `json:"slopeMaxAngle"`
	IsEtcSOC                   *int     `json:"isEtcSOC"`
	BmsType                    *int     `json:"bmsType"`
	CcuType                    *int     `json:"ccuType"`
	IsBleKeyEnable             *int     `json:"isBleKeyEnable"`
	Is433KeyEnable             *int     `json:"is433KeyEnable"`
	BeaconRssi                 *int     `json:"beaconRssi"`
	IsHelmetWearCheckEnable    *int     `json:"isHelmetWearCheckEnable"`
}

// BlueTBeaconInfoCo mirrors the Java BlueTBeaconInfoCo.
type BlueTBeaconInfoCo struct {
	Event       *int     `json:"event"`
	TBeaconAddr *string  `json:"tBeaconAddr"`
	TBeaconId   *string  `json:"tBeaconId"`
	TBeaconSOC  *int     `json:"tBeaconSOC"`
	RealRssi    *int     `json:"realRssi"`
	TBeaconVsn  *string  `json:"tBeaconVsn"`
	Lng         *float64 `json:"lng"`
	Lat         *float64 `json:"lat"`
	Timestamp   *int64   `json:"timestamp"`
}

// BleHelmetInfoCo mirrors the Java BleHelmetInfoCo.
type BleHelmetInfoCo struct {
	Event         *int    `json:"event"`
	Name          *string `json:"name"`
	Mac           *string `json:"mac"`
	Manufacturer  *int    `json:"manufacturer"`
	Version       *int    `json:"version"`
	Soc           *int    `json:"soc"`
	Voltage       *int    `json:"voltage"`
	Attitude      *int    `json:"attitude"`
	CapacityState *int    `json:"capacityState"`
	InfraredState *int    `json:"infraredState"`
	StressState   *int    `json:"stressState"`
	Fault         *int    `json:"fault"`
}

// RealRestBatteryCo mirrors the Java RealRestBatteryCo.
type RealRestBatteryCo struct {
	CarId       string `json:"carId"`
	Imei        string `json:"imei"`
	RestBattery *int   `json:"restBattery"`
}

// TenantOf extracts the tenantId from a CommandContext, empty when absent.
func TenantOf(cc *CommandContext) string {
	if cc == nil {
		return ""
	}
	return cc.TenantID
}

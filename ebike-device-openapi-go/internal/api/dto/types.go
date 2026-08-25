package dto

import "encoding/json"

// CommandContext simulates com.xyy.dto.CommandContext
type CommandContext struct {
	TraceId       string `json:"traceId,omitempty"`
	TenantId      string `json:"tenantId,omitempty"`
	Pin           string `json:"pin,omitempty"`
	Ip            string `json:"ip,omitempty"`
	Platform      string `json:"platform,omitempty"`
	DeviceId      string `json:"deviceId,omitempty"`
	Source        string `json:"source,omitempty"`
	Name          string `json:"name,omitempty"`
	StressTesting bool   `json:"stressTesting"`
}

// WildCmd corresponds to com.xyy.ebike.device.openapi.infrastructure.dto.WildCmd
type WildCmd struct {
	Imei           string                 `json:"imei" binding:"required"`
	Cmd            int16                  `json:"cmd" binding:"required"`
	Tm             *int                   `json:"tm,omitempty"`
	Dt             *int                   `json:"dt,omitempty"`
	Params         map[string]interface{} `json:"params,omitempty"`
	Async          bool                   `json:"async"`
	Timeout        *int                   `json:"timeout,omitempty"`
	Payload        string                 `json:"payload,omitempty"`
	JobId          string                 `json:"jobId,omitempty"`
	Ip             string                 `json:"ip,omitempty"`
	Port           int                    `json:"port,omitempty"`
	CommandContext *CommandContext        `json:"commandContext,omitempty"`
}

// WildResult corresponds to com.xyy.ebike.device.openapi.infrastructure.dto.WildResult
type WildResult struct {
	Payload string `json:"payload,omitempty"`
	JobId   string `json:"jobId,omitempty"`
	Imei    string `json:"imei,omitempty"`
	Cmd     int16  `json:"cmd,omitempty"`
	Json    string `json:"json,omitempty"`
}

type Login struct {
	TraceId    string `json:"traceId,omitempty"`
	Imei       string `json:"imei"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Supplier   *int   `json:"supplier,omitempty"`
	Imsi       string `json:"imsi,omitempty"`
	Version    string `json:"version,omitempty"`
	DeviceType *int   `json:"deviceType,omitempty"`
	Timestamp  *int64 `json:"timestamp,omitempty"`
}

type Logout struct {
	TraceId   string `json:"traceId,omitempty"`
	Imei      string `json:"imei"`
	Timestamp int64  `json:"timestamp"`
}

// EBikeRequest corresponds to com.xyy.ebike.device.openapi.api.dto.EBikeRequest
type EBikeRequest struct {
	TraceId        string          `json:"traceId,omitempty"`
	Imei           string          `json:"imei" binding:"required,len=15"`
	Async          bool            `json:"async"`
	Payload        interface{}     `json:"payload,omitempty"`
	Tm             *int            `json:"tm,omitempty"`
	Dt             *int            `json:"dt,omitempty"`
	CommandContext *CommandContext `json:"commandContext,omitempty"`
}

// stripNilValuesFromMap recursively removes keys with nil values from a map.
func stripNilValuesFromMap(m map[string]interface{}) {
	for k, v := range m {
		if v == nil {
			delete(m, k)
		} else if nestedMap, ok := v.(map[string]interface{}); ok {
			stripNilValuesFromMap(nestedMap)
		} else if nestedSlice, ok := v.([]interface{}); ok {
			for _, item := range nestedSlice {
				if nm, isMap := item.(map[string]interface{}); isMap {
					stripNilValuesFromMap(nm)
				}
			}
		}
	}
}

// ExtractParams returns all subclass fields as a map, excluding EBikeRequest base fields
func ExtractParams(req interface{}) map[string]interface{} {
	b, _ := json.Marshal(req)
	var m map[string]interface{}
	json.Unmarshal(b, &m)
	
	// Exclude common fields as defined in Java excludeField()
	delete(m, "imei")
	delete(m, "async")
	delete(m, "payload")
	delete(m, "traceId")
	delete(m, "tm")
	delete(m, "dt")
	delete(m, "commandContext")
	// Feign DTO fields not present on openapi Request types — Jackson ignores them.
	delete(m, "carId")
	delete(m, "izRiskControl")
	
	// Strip nil values from params to align with Java's NON_NULL serialization
	stripNilValuesFromMap(m)

	return m
}

type TransmissionBody struct {
	Async   bool   `json:"async"`
	JobId   string `json:"jobId"`
	Payload string `json:"payload,omitempty"`
	Result  string `json:"result"`
}

// Result corresponds to com.xyy.dto.Result
type Result struct {
	Success bool        `json:"success"`
	Code    string      `json:"code"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
}

// API Requests
type LockRequest struct {
	EBikeRequest
	Acc              *int   `json:"acc,omitempty" binding:"required,min=0,max=1"`
	Idx              *int   `json:"idx,omitempty"`
	Volume           *int   `json:"volume,omitempty" binding:"omitempty,min=1,max=100"`
	IsIgnoreFence    *int   `json:"isIgnoreFence,omitempty" binding:"omitempty,min=0,max=1"`
	IsTBeacon        *int   `json:"isTBeacon,omitempty" binding:"omitempty,min=0,max=1"`
	IsKickstand      *int   `json:"isKickstand,omitempty" binding:"omitempty,min=0,max=1"`
	SingleSpeedLimit *int   `json:"singleSpeedLimit,omitempty" binding:"omitempty,min=0,max=100"`
	Url              string `json:"url,omitempty" binding:"omitempty,url"`
	Crc              *int   `json:"crc,omitempty"`
	Type             string `json:"type,omitempty" binding:"omitempty,oneof=amr wav mp3"`
}

type DefendRequest struct {
	EBikeRequest
	Mode *int `json:"mode,omitempty" binding:"required"`
}

type HelmetLockRequest struct {
	EBikeRequest
	Opt *int `json:"opt,omitempty" binding:"required"`
}

// Uplink API DTOs

type MessageHeader struct {
	Magic    int   `json:"magic"`
	Cmd      int16 `json:"cmd"`
	Sequence int16 `json:"sequence"`
	Length   int   `json:"length"`
}

type LoginCmd struct {
	CommandContext *CommandContext `json:"commandContext,omitempty"`
	Imei           string          `json:"imei" binding:"required"`
	Host           string          `json:"host" binding:"required"`
	Port           int             `json:"port" binding:"required"`
	Imsi           string          `json:"imsi,omitempty"`
	Version        *int64          `json:"version,omitempty"`
	DeviceType     *int            `json:"deviceType,omitempty"`
	Timestamp      *int64          `json:"timestamp,omitempty"`
	RemoteAddress  string          `json:"remoteAddress,omitempty"`
	ChannelId      string          `json:"channelId" binding:"required"`
}

type LogoutCmd struct {
	CommandContext *CommandContext `json:"commandContext,omitempty"`
	Imei           string          `json:"imei" binding:"required"`
	ChannelId      string          `json:"channelId" binding:"required"`
}

type DecodeCmd struct {
	CommandContext *CommandContext `json:"commandContext,omitempty"`
	Imei           string          `json:"imei" binding:"required"`
	MessageHeader  *MessageHeader  `json:"messageHeader,omitempty" binding:"required"`
	HexBody        string          `json:"hexBody" binding:"required"`
	AutoReplay     bool            `json:"autoReplay,omitempty"`
	Timestamp      *int64          `json:"timestamp,omitempty"`
}

type ReplayCmd struct {
	CommandContext *CommandContext `json:"commandContext,omitempty"`
	Imei           string          `json:"imei" binding:"required"`
	JobId          string          `json:"jobId" binding:"required"`
	Data           string          `json:"data,omitempty"`
	Payload        string          `json:"payload,omitempty"`
}

type ReplayData struct {
	TraceId    string `json:"traceId,omitempty"`
	Imei       string `json:"imei"`
	NeedReplay bool   `json:"needReplay"`
	HexHeader  string `json:"hexHeader,omitempty"`
	HexBody    string `json:"hexBody,omitempty"`
	DecodeBody string `json:"decodeBody,omitempty"`
}

type DeviceReportMessage struct {
	Imei            string `json:"imei"`
	MsgType         string `json:"msgType"`
	BussinessType   string `json:"bussinessType"`
	Data            string `json:"data"`
	ReceiveDataTime int64  `json:"receiveDataTime"` // epoch millis at gateway ingress (worker PG timestamp source)
}

// DeviceReplyMessage corresponds to Java DeviceReplyMessage (used in replay)
type DeviceReplyMessage struct {
	Imei            string `json:"imei"`
	MsgType         string `json:"msgType"`
	BussinessType   string `json:"bussinessType"`
	Data            string `json:"data"`
	ReceiveDataTime int64  `json:"receiveDataTime"`
	MsgId           string `json:"msgId"`
	Identifier      string `json:"identifier,omitempty"`
	Payload         string `json:"payload,omitempty"`
}


// Result struct wrapper for downstream ECU response
// Gateway returns {"code":"200","message":"success","data":{...}} — no "success" field.
// Java's Result class defaults success=true on deserialization, so it works.
// In Go, bool defaults to false, so we check code instead.
type DownstreamResult struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    *WildResult `json:"data,omitempty"`
}

// Device protocol types stored in EcuLogin.Type / Redis ecu_login_*.
const (
	DeviceTypeXiaoan  = "xiaoan"
	DeviceTypeLuoping = "luoping"
)

type EcuLogin struct {
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Type          string `json:"type,omitempty"`
	Version       *int64 `json:"version,omitempty"`
	RemoteAddress string `json:"remoteAddress,omitempty"`
	ChannelId     string `json:"channelId,omitempty"`
	Timestamp     *int64 `json:"timestamp,omitempty"`
	// Luoping MQTT fields (deviceId != IMEI; topic uses deviceId).
	DeviceId  string `json:"deviceId,omitempty"`
	GroupName string `json:"groupName,omitempty"`
	ClientId  string `json:"clientId,omitempty"`
	LastSeenAt *int64 `json:"lastSeenAt,omitempty"`
}

// LuopingDeviceMapping binds MQTT deviceId <-> platform IMEI.
type LuopingDeviceMapping struct {
	Imei      string `json:"imei"`
	DeviceId  string `json:"deviceId"`
	GroupName string `json:"groupName,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty"`
}

// EmqxPresenceEvent is the EMQX $SYS connected/disconnected payload (MQTT shared sub).
type EmqxPresenceEvent struct {
	Event          string `json:"event,omitempty"`
	ClientID       string `json:"clientid"`
	Username       string `json:"username,omitempty"`
	Reason         string `json:"reason,omitempty"`
	ConnectedAt    int64  `json:"connected_at,omitempty"`
	DisconnectedAt int64  `json:"disconnected_at,omitempty"`
	Timestamp      int64  `json:"timestamp,omitempty"`
	Peername       string `json:"peername,omitempty"`
}

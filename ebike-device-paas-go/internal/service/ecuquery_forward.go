package service

import (
	"encoding/json"
	"errors"
	"time"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/client"
	"ebike-device-paas-go/internal/pkg/config"
	"ebike-device-paas-go/internal/repository"
)

// ErrDeviceNotFound is returned when the imei has no device-info entry in Redis.
var ErrDeviceNotFound = errors.New("device not found")

const c34Payload = `{"type":"c34"}`

// Gateway paths (DeviceCommandApiFeign).
const (
	pathInnerParam   = "/ebike/cmd/query_inner_param"
	pathDeviceInfo   = "/ebike/cmd/query_device_info"
	pathBlueTBeacon  = "/ebike/cmd/query_blueT_beacon"
	pathBleHelmetInf = "/ebike/cmd/query_ble_helmet_info"
)

// gatewayReq is the body forwarded to openapi. QueryDeviceInfoRequest extends
// Command, so commandContext is accepted even though Java BaseEcuCmdDto omits it;
// openapi resolves tenantId from commandContext before decrypting the appId header.
type gatewayReq struct {
	Async          *bool               `json:"async,omitempty"`
	CarId          string              `json:"carId,omitempty"`
	Imei           string              `json:"imei,omitempty"`
	TraceId        string              `json:"traceId,omitempty"`
	Payload        string              `json:"payload,omitempty"`
	IsCameraEnable *int                `json:"isCameraEnable,omitempty"`
	CommandContext *dto.CommandContext `json:"commandContext,omitempty"`
}

func reqFrom(q dto.EcuQuery, payload string, isCamera *int) gatewayReq {
	return gatewayReq{
		Async:          q.Async,
		CarId:          q.CarId,
		Imei:           q.Imei,
		TraceId:        traceOf(q.CommandContext),
		Payload:        payload,
		IsCameraEnable: isCamera,
		CommandContext: q.CommandContext,
	}
}

func traceOf(cc *dto.CommandContext) string {
	if cc == nil {
		return ""
	}
	return cc.TraceID
}

// GetInnerParam ports EcuQueryServiceImpl.getInnerParam.
func GetInnerParam(q dto.EcuQuery) (dto.CommandResult, error) {
	do, err := client.PostGateway(pathInnerParam, reqFrom(q, q.Payload, nil), dto.TenantOf(q.CommandContext))
	if err != nil {
		return dto.CommandResult{}, err
	}
	return buildCommandResult(do, mapInnerParam), nil
}

// GetDeviceInfo ports EcuQueryServiceImpl.getDeviceInfo (defaults payload to c34).
func GetDeviceInfo(q dto.DeviceInfoQry) (dto.CommandResult, error) {
	payload := q.Payload
	if payload == "" {
		payload = c34Payload
	}
	do, err := client.PostGateway(pathDeviceInfo, reqFrom(q.EcuQuery, payload, q.IsCameraEnable), dto.TenantOf(q.CommandContext))
	if err != nil {
		return dto.CommandResult{}, err
	}
	return buildCommandResult(do, mapDeviceInfo), nil
}

// GetBlueTBeacon ports EcuQueryServiceImpl.getBlueTBeacon.
func GetBlueTBeacon(q dto.EcuQuery) (dto.CommandResult, error) {
	do, err := client.PostGateway(pathBlueTBeacon, reqFrom(q, q.Payload, nil), dto.TenantOf(q.CommandContext))
	if err != nil {
		return dto.CommandResult{}, err
	}
	return buildCommandResult(do, mapBlueTBeacon), nil
}

// QueryBleHelmetInfo ports EcuQueryServiceImpl.queryBleHelmetInfo.
func QueryBleHelmetInfo(q dto.EcuQuery) (dto.CommandResult, error) {
	do, err := client.PostGateway(pathBleHelmetInf, reqFrom(q, q.Payload, nil), dto.TenantOf(q.CommandContext))
	if err != nil {
		return dto.CommandResult{}, err
	}
	return buildCommandResult(do, mapBleHelmet), nil
}

// QueryRealRestBattery ports EcuQueryServiceImpl.queryRealRestBattery:
//  1. fast path: fresh soc in Redis (bmsTimeStamp within 30min) -> return soc;
//  2. fallback: query device info -> voltageMv -> management computeRestBattery.
func QueryRealRestBattery(tenantID string, q dto.EcuQuery) (dto.RealRestBatteryCo, error) {
	dev := repository.GetDeviceByImei(tenantID, q.Imei)
	if dev == nil {
		return dto.RealRestBatteryCo{}, ErrDeviceNotFound
	}
	carID, _ := dev["carId"].(string)
	imei, _ := dev["imei"].(string)
	bms, _ := asInt64(dev["bmsTimeStamp"])

	if soc, ok := asInt(dev["soc"]); ok && (bms+30*60)*1000 > time.Now().UnixMilli() {
		s := soc
		return dto.RealRestBatteryCo{CarId: carID, Imei: imei, RestBattery: &s}, nil
	}

	// Carry the resolved tenant into the nested getDeviceInfo so its gateway call
	// can sign the AES auth headers even when the body context omitted tenantId.
	cc := q.CommandContext
	if cc == nil {
		cc = &dto.CommandContext{}
	}
	if cc.TenantID == "" {
		cc.TenantID = tenantID
	}
	di := dto.DeviceInfoQry{EcuQuery: dto.EcuQuery{Imei: q.Imei, CommandContext: cc}}
	cr, err := GetDeviceInfo(di)
	if err != nil {
		return dto.RealRestBatteryCo{}, err
	}
	voltageMv := extractVoltageMv(cr.Result)
	if voltageMv == nil {
		if v, ok := asInt64(dev["voltage"]); ok && v > 0 {
			iv := int(v)
			voltageMv = &iv
		}
	}
	// management CarRestBatteryCmd.voltage is @NotNull; calling with nil yields
	// "voltage must not be null" (seen in prod SHADOW DIFF). Skip the RPC and
	// degrade to Redis restBattery / 0 instead.
	if voltageMv == nil {
		rb := restBatteryFallback(dev)
		return dto.RealRestBatteryCo{CarId: carID, Imei: imei, RestBattery: &rb}, nil
	}
	rest, err := client.ComputeRestBattery(q.Imei, voltageMv, nil, q.CommandContext)
	if err != nil {
		return dto.RealRestBatteryCo{}, err
	}
	return dto.RealRestBatteryCo{CarId: carID, Imei: imei, RestBattery: rest}, nil
}

// restBatteryFallback returns Redis restBattery when present, otherwise 0.
func restBatteryFallback(dev map[string]interface{}) int {
	if rb, ok := asInt(dev["restBattery"]); ok {
		return rb
	}
	return 0
}

// buildCommandResult mirrors RpcResultToCo: unwrap CommandResultDo, parse the
// inner result JSON string for ecuCode + result object, and map the object.
func buildCommandResult(do *client.CommandResultDo, mapResult func(json.RawMessage) interface{}) dto.CommandResult {
	cr := dto.CommandResult{JobId: do.JobID, Async: do.Async, Payload: do.Payload}
	if do.Result == "" {
		return cr
	}
	var outer struct {
		Code   string          `json:"code"`
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal([]byte(do.Result), &outer); err != nil {
		return cr
	}
	cr.EcuCode = &outer.Code
	if len(outer.Result) > 0 && string(outer.Result) != "null" && mapResult != nil {
		cr.Result = mapResult(outer.Result)
	}
	return cr
}

func mapInnerParam(raw json.RawMessage) interface{} {
	var co dto.InnerParamQryCo
	_ = json.Unmarshal(raw, &co)
	var m map[string]json.RawMessage
	if json.Unmarshal(raw, &m) == nil {
		co.FreqNorm = readIntPtr(m, "freq_norm")
		co.FreqMove = readIntPtr(m, "freq_move")
	}
	return co
}

func mapBlueTBeacon(raw json.RawMessage) interface{} {
	var co dto.BlueTBeaconInfoCo
	_ = json.Unmarshal(raw, &co)
	return co
}

func mapBleHelmet(raw json.RawMessage) interface{} {
	var co dto.BleHelmetInfoCo
	_ = json.Unmarshal(raw, &co)
	return co
}

// mapDeviceInfo keeps all gateway result fields (map) and applies the two Java
// transforms: ecu.debug.isEnable -> helmet6React; coordinate.type==0 -> use WGS84.
func mapDeviceInfo(raw json.RawMessage) interface{} {
	var m map[string]interface{}
	if json.Unmarshal(raw, &m) != nil {
		return nil
	}
	if config.GlobalConfig.Ecu.Debug.IsEnable {
		if v, ok := m["helmet6Lock"]; ok {
			m["helmet6React"] = v
		}
	}
	if config.GlobalConfig.Coordinate.Type == 0 {
		if g, ok := m["gps"].(map[string]interface{}); ok {
			if v, ok := g["wgs84Lat"]; ok {
				g["lat"] = v
			}
			if v, ok := g["wgs84Lng"]; ok {
				g["lng"] = v
			}
		}
	}
	return m
}

func extractVoltageMv(result interface{}) *int {
	m, ok := result.(map[string]interface{})
	if !ok {
		return nil
	}
	if v, ok := asInt(m["voltageMv"]); ok {
		return &v
	}
	return nil
}

func readIntPtr(m map[string]json.RawMessage, key string) *int {
	v, ok := m[key]
	if !ok {
		return nil
	}
	var n int
	if json.Unmarshal(v, &n) != nil {
		return nil
	}
	return &n
}

func asInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}

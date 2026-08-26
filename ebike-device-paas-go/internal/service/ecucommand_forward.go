package service

import (
	"encoding/json"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/client"
	"ebike-device-paas-go/internal/repository"
)

// EcuCommand endpoints port EcuCommandServiceImpl: each maps the api command to
// the openapi gateway DTO, forwards it (DeviceCommandApiFeign), and reshapes the
// CommandResultDo via RpcResultToCo. They are live device commands with write
// side effects -> NOT shadow-compared.

// Gateway paths (DeviceCommandApiFeign).
const (
	pathCmdLock           = "/ebike/cmd/lock"
	pathCmdDefend         = "/ebike/cmd/defend"
	pathCmdHelmetLock     = "/ebike/cmd/helmet_lock"
	pathCmdSetInnerParam  = "/ebike/cmd/set_inner_param"
	pathCmdUpdateFence    = "/ebike/cmd/update_inner_fence"
	pathCmdUpgradeDevice  = "/ebike/cmd/upgrade_device"
	pathCmdUpgradeVoice   = "/ebike/cmd/upgrade_voice"
	pathCmdBroadcastVoice = "/ebike/cmd/broadcast_voice"
	pathCmdBatteryCompart = "/ebike/cmd/switch_battery_compartment"
	pathCmdSetBluetooth   = "/ebike/cmd/set_bluetooth"
	pathCmdScanBleHelmet  = "/ebike/cmd/scan_ble_helmet"
	pathCmdRestart        = "/ebike/cmd/restart"
	pathCmdTransmission   = "/ebike/cmd/transmission"
	pathCmdSetParkSite    = "/ebike/cmd/set_park_site"
	pathCmdTriggerTemp    = "/ebike/cmd/triggerTempState"
	pathCmdSetDashboard   = "/ebike/cmd/set_dashboard"
	pathCmdRearWheelLock  = "/ebike/cmd/switch_rear_wheel_lock"
)

// prepCommandBody mirrors ConvertorHelper.copyProperties + the Dto's traceId
// default: drop the paas-only commandContext and inject traceId from it.
func prepCommandBody(raw map[string]interface{}, cc *dto.CommandContext) map[string]interface{} {
	body := make(map[string]interface{}, len(raw)+1)
	for k, v := range raw {
		if k == "commandContext" {
			continue
		}
		body[k] = v
	}
	if cc != nil && cc.TraceID != "" {
		body["traceId"] = cc.TraceID
	}
	return body
}

func forwardCommand(path string, body map[string]interface{}, mapper func(json.RawMessage) interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	do, err := client.PostGateway(path, body, dto.TenantOf(cc))
	if err != nil {
		return dto.CommandResult{}, err
	}
	return buildCommandResult(do, mapper), nil
}

// forwardCommandLogged is forwardCommand plus the @OptLog audit record emitted by
// LogAspect on DeviceCommandApiRpcImpl: it always writes one operation-log line
// (content = the forwarded command body, result = the gateway CommandResultDo or
// the error), mirroring the aspect's finally block.
func forwardCommandLogged(path string, body map[string]interface{}, mapper func(json.RawMessage) interface{}, cc *dto.CommandContext, eventType string) (dto.CommandResult, error) {
	do, err := client.PostGateway(path, body, dto.TenantOf(cc))
	imei, carID, content := bodyStr(body, "imei"), bodyStr(body, "carId"), jsonStr(body)
	if err != nil {
		emitOpLog(cc, eventType, imei, carID, content, map[string]interface{}{"error": err.Error()})
		return dto.CommandResult{}, err
	}
	emitOpLog(cc, eventType, imei, carID, content, do)
	return buildCommandResult(do, mapper), nil
}

// LockCommand ports EcuCommandServiceImpl.lockCommand (+ no-risk-control flag).
func LockCommand(tenantID string, raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	accOpLog(raw, cc)
	cr, err := forwardCommand(pathCmdLock, prepCommandBody(raw, cc), nil, cc)
	if err != nil {
		return cr, err
	}
	imei := bodyStr(raw, "imei")
	// LockCmd.izRiskControl defaults to true in Java (field initializer); a missing
	// key must be treated as true, otherwise the no_risk_control_ex side effect
	// inverts (set vs delete).
	izRisk := true
	if v, ok := raw["izRiskControl"].(bool); ok {
		izRisk = v
	}
	if acc := bodyInt(raw, "acc"); acc == 1 {
		if !izRisk {
			repository.SetNoRiskControlEx(tenantID, imei)
		} else {
			repository.DeleteNoRiskControlEx(tenantID, imei)
		}
	}
	return cr, nil
}

// DefendCommand ports EcuCommandServiceImpl.defendCommand (+ clear no-risk flag).
func DefendCommand(tenantID string, raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	defendOpLog(raw, cc)
	cr, err := forwardCommand(pathCmdDefend, prepCommandBody(raw, cc), mapDefend, cc)
	if err != nil {
		return cr, err
	}
	repository.DeleteNoRiskControlEx(tenantID, bodyStr(raw, "imei"))
	return cr, nil
}

// HelmetCommand ports EcuCommandServiceImpl.helmetCommand (sw==0 -> idx=22).
func HelmetCommand(raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	body := prepCommandBody(raw, cc)
	// Java: Objects.equals(sw, 0) -> idx=22 only when sw is present AND 0; a
	// missing/null sw must NOT set idx (Objects.equals(null,0)==false). bodyInt
	// can't distinguish missing from 0, so check the key explicitly.
	if sw, ok := raw["sw"]; ok {
		if n, _ := asInt(sw); n == 0 {
			body["idx"] = 22
		}
	}
	return forwardCommandLogged(pathCmdHelmetLock, body, mapJobResultObject, cc, opHelmet)
}

// SetInnerParam ports EcuCommandServiceImpl.setInnerParam (freqMove/freqNorm ->
// freq_move/freq_norm snake_case for the gateway).
func SetInnerParam(raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	return forwardCommandLogged(pathCmdSetInnerParam, innerParamBody(raw, cc), nil, cc, opInnerParam)
}

func innerParamBody(raw map[string]interface{}, cc *dto.CommandContext) map[string]interface{} {
	body := prepCommandBody(raw, cc)
	if v, ok := raw["freqMove"]; ok {
		body["freq_move"] = v
	}
	if v, ok := raw["freqNorm"]; ok {
		body["freq_norm"] = v
	}
	return body
}

// UpdateInnerFence ports EcuCommandServiceImpl.updateInnerFence.
func UpdateInnerFence(raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	return forwardCommandLogged(pathCmdUpdateFence, prepCommandBody(raw, cc), nil, cc, opInnerFence)
}

// DeviceUpgrade ports EcuCommandServiceImpl.deviceUpgrade.
func DeviceUpgrade(raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	return forwardCommandLogged(pathCmdUpgradeDevice, prepCommandBody(raw, cc), nil, cc, opDeviceUpgrade)
}

// VoiceUpgrade ports EcuCommandServiceImpl.voiceUpgrade.
func VoiceUpgrade(raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	return forwardCommandLogged(pathCmdUpgradeVoice, prepCommandBody(raw, cc), nil, cc, opVoiceUpgrade)
}

// VoiceCommand ports EcuCommandServiceImpl.voiceCommand + DeviceCommandApiRpcImpl:
//   - PaasOperationLogUtil.findCarVoiceOperationLog (idx==6) is logged first.
//   - @RateLimiter (platform=="wechat" && idx==9) caps the voice find-car bell:
//     when limited it throws BizException(RATE_LIMIT_ERROR) before the forward.
//   - @OptLog(ECU_VOICE_COMMAND) records the call.
//   - tenant 250 + idx 32/34 + ecuCode 105 is treated as success (ecuCode 0).
func VoiceCommand(tenantID string, raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	findCarVoiceOpLog(raw, cc)
	idx := bodyInt(raw, "idx")
	platform := ""
	if cc != nil {
		platform = cc.Platform
	}
	if platform == "wechat" && idx == 9 {
		if ok, msg := repository.VoiceFindCarAllowed(tenantID, voicePin(cc)); !ok {
			err := &RateLimitError{Msg: msg}
			emitOpLog(cc, opVoice, bodyStr(raw, "imei"), bodyStr(raw, "carId"), jsonStr(prepCommandBody(raw, cc)), map[string]interface{}{"error": msg})
			return dto.CommandResult{}, err
		}
	}
	cr, err := forwardCommandLogged(pathCmdBroadcastVoice, prepCommandBody(raw, cc), nil, cc, opVoice)
	if err != nil {
		return cr, err
	}
	if (idx == 32 || idx == 34) && tenantID == "250" && cr.EcuCodeValue() == "105" {
		zero := "0"
		cr.EcuCode = &zero
	}
	return cr, nil
}

func voicePin(cc *dto.CommandContext) string {
	if cc != nil {
		return cc.Pin
	}
	return ""
}

// BatteryCompartmentCommand ports EcuCommandServiceImpl.batteryCompartmentCommand.
func BatteryCompartmentCommand(raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	batteryCompartmentOpLog(raw, cc)
	return forwardCommand(pathCmdBatteryCompart, prepCommandBody(raw, cc), nil, cc)
}

// Bluetooth ports EcuCommandServiceImpl.bluetooth.
func Bluetooth(raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	return forwardCommandLogged(pathCmdSetBluetooth, prepCommandBody(raw, cc), mapBluetooth, cc, opBluetooth)
}

// ScanBleHelmet ports EcuCommandServiceImpl.scanBleHelmet.
func ScanBleHelmet(raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	return forwardCommandLogged(pathCmdScanBleHelmet, prepCommandBody(raw, cc), mapJobResultObject, cc, opScanBleHelmet)
}

// RestartCommand ports EcuCommandServiceImpl.restartCommand.
func RestartCommand(raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	return forwardCommandLogged(pathCmdRestart, prepCommandBody(raw, cc), nil, cc, opRestart)
}

// RearWheelLock ports EcuCommandServiceImpl.rearWheelLock (@OptLog uses
// ECU_RESTART_COMMAND, matching the Java annotation).
func RearWheelLock(raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	rearWheelLockOpLog(raw, cc)
	return forwardCommandLogged(pathCmdRearWheelLock, prepCommandBody(raw, cc), nil, cc, opRestart)
}

// Transmission ports EcuCommandServiceImpl.transmission.
func Transmission(raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	return forwardCommandLogged(pathCmdTransmission, prepCommandBody(raw, cc), mapJobResultObject, cc, opTransmission)
}

// ReplyStopMove ports EcuCommandServiceImpl.replyStopMove.
func ReplyStopMove(raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	return forwardCommand(pathCmdSetParkSite, prepCommandBody(raw, cc), nil, cc)
}

// Dashboard ports EcuCommandServiceImpl.dashboard (field-by-field copy -> passthrough).
func Dashboard(raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	return forwardCommand(pathCmdSetDashboard, prepCommandBody(raw, cc), mapJobResultObject, cc)
}

// TriggerTempState ports EcuCommandServiceImpl.triggerTempState (clears the
// saddle-overload-contact cache first).
func TriggerTempState(tenantID string, raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	repository.ClearSaddleOverloadContact(tenantID, bodyStr(raw, "imei"))
	return forwardCommand(pathCmdTriggerTemp, prepCommandBody(raw, cc), nil, cc)
}

// TempUnLock ports EcuCommandServiceImpl.tempUnLock: set the temp-power-on tag,
// disable the fence (set_inner_param isFenceEnable=0), then optionally power on
// the acc (lock command). When izAccOn is false it short-circuits to ecuCode 0.
//
// Defaults mirror TempUnLockCommand field initializers: izAccOn=true, acc=1,
// isIgnoreFence=1. Upstream (ebike-rent) often omits acc/isIgnoreFence and relies
// on those Java defaults when copying into LockCmdDto.
func TempUnLock(tenantID string, raw map[string]interface{}, cc *dto.CommandContext) (dto.CommandResult, error) {
	imei := bodyStr(raw, "imei")
	carID := bodyStr(raw, "carId")
	repository.SetTempUnLockTag(tenantID, imei, bodyInt64(raw, "second"))

	fenceBody := prepCommandBody(map[string]interface{}{
		"imei":          imei,
		"carId":         carID,
		"isFenceEnable": 0,
	}, cc)
	// Java tempUnLock calls deviceCommandApiRpc.setInnerParam, which carries
	// @OptLog(ECU_INNER_PARAM_COMMAND); use the logged forward so the 17504 audit
	// line is emitted here too.
	if _, err := forwardCommandLogged(pathCmdSetInnerParam, fenceBody, nil, cc, opInnerParam); err != nil {
		return dto.CommandResult{}, err
	}

	if !tempUnLockIzAccOn(raw) {
		zero := "0"
		return dto.CommandResult{EcuCode: &zero}, nil
	}
	return forwardCommand(pathCmdLock, prepCommandBody(tempUnLockLockRaw(raw), cc), nil, cc)
}

// tempUnLockIzAccOn mirrors TempUnLockCommand.izAccOn = true: missing/null → true.
func tempUnLockIzAccOn(raw map[string]interface{}) bool {
	v, ok := raw["izAccOn"]
	if !ok || v == nil {
		return true
	}
	b, ok := v.(bool)
	if !ok {
		return true
	}
	return b
}

// tempUnLockLockRaw applies TempUnLockCommand defaults onto the lock payload
// (acc=1, isIgnoreFence=1) when the caller omitted them — same as Java
// ConvertorHelper.copyProperties from a command with field initializers.
func tempUnLockLockRaw(raw map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(raw)+2)
	for k, v := range raw {
		out[k] = v
	}
	if _, ok := out["acc"]; !ok || out["acc"] == nil {
		out["acc"] = 1
	}
	if _, ok := out["isIgnoreFence"]; !ok || out["isIgnoreFence"] == nil {
		out["isIgnoreFence"] = 1
	}
	return out
}

func bodyStr(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func bodyInt(m map[string]interface{}, key string) int {
	v, _ := asInt(m[key])
	return v
}

func bodyInt64(m map[string]interface{}, key string) int64 {
	switch n := m[key].(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	case json.Number:
		i, _ := n.Int64()
		return i
	default:
		return 0
	}
}

func bodyBool(m map[string]interface{}, key string) bool {
	b, _ := m[key].(bool)
	return b
}

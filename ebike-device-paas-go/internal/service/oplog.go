package service

import (
	"encoding/json"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/pkg/oplog"
)

// Operation-audit event (code, name) pairs, mirroring OperationTypeEnum. The
// name is the enum's Chinese msg (ResultHelper.getLocaleMessage falls back to it
// when no i18n bundle is configured, which is the prod behaviour).
const (
	opAccOn         = "17514" // 车辆启动
	opAccOff        = "17515" // 车辆断电
	opDefendOn      = "17516" // 车辆锁车
	opDefendOff     = "17517" // 关闭设防
	opBatteryOff    = "17519" // 电池仓上锁
	opBatteryOn     = "17518" // 电池仓解锁
	opRearWheelOn   = "17520" // 后轮锁开
	opRearWheelOff  = "17521" // 后轮锁关
	opFindCarVoice  = "17522" // 播放寻车音
	opHelmet        = "17503" // 设备头盔锁控制
	opInnerParam    = "17504" // 设置设备内部参数
	opInnerFence    = "17505" // 设备内置围栏更新
	opVoice         = "17508" // 设备语音播放
	opBluetooth     = "17510" // 设备蓝牙参数
	opTransmission  = "17513" // 透传命令
	opDeviceUpgrade = "17506" // 设备固件更新
	opVoiceUpgrade  = "17507" // 设备语音更新
	opScanBleHelmet = "17511" // 扫描并绑定蓝牙智能头盔
	opRestart       = "17512" // 设备重启
)

var opNames = map[string]string{
	opAccOn:         "车辆启动",
	opAccOff:        "车辆断电",
	opDefendOn:      "车辆锁车",
	opDefendOff:     "关闭设防",
	opBatteryOff:    "电池仓上锁",
	opBatteryOn:     "电池仓解锁",
	opRearWheelOn:   "后轮锁开",
	opRearWheelOff:  "后轮锁关",
	opFindCarVoice:  "播放寻车音",
	opHelmet:        "设备头盔锁控制",
	opInnerParam:    "设置设备内部参数",
	opInnerFence:    "设备内置围栏更新",
	opVoice:         "设备语音播放",
	opBluetooth:     "设备蓝牙参数",
	opTransmission:  "透传命令",
	opDeviceUpgrade: "设备固件更新",
	opVoiceUpgrade:  "设备语音更新",
	opScanBleHelmet: "扫描并绑定蓝牙智能头盔",
	opRestart:       "设备重启",
}

// RateLimitError mirrors BizException(MsgCodeEnum.RATE_LIMIT_ERROR, msg) thrown
// by the xyy-redis RateLimiterAspect: the carried Msg is the per-unit or window
// message and is rendered to the caller with code "00026".
type RateLimitError struct{ Msg string }

func (e *RateLimitError) Error() string { return e.Msg }

// jsonStr marshals v to a compact JSON string for the operation-log content
// field, mirroring JSONObject.toJSONString(cmd). Returns "" on error.
func jsonStr(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// emitOpLog builds and writes an operation-audit entry, mirroring
// OperationLog.builder()...build() which back-fills tenantId/traceId/pin/platform
// from the CommandContext.
func emitOpLog(cc *dto.CommandContext, eventType, imei, carID, content string, result interface{}) {
	e := oplog.Entry{
		EventType: eventType,
		EventName: opNames[eventType],
		Imei:      imei,
		CarId:     carID,
		Content:   content,
		Result:    result,
	}
	if cc != nil {
		e.TenantId = cc.TenantID
		e.TraceId = cc.TraceID
		e.Pin = cc.Pin
		e.Platform = cc.Platform
	}
	oplog.Emit(e)
}

// The following reproduce the PaasOperationLogUtil.*OperationLog calls made in
// EcuCommandServiceImpl BEFORE the gateway forward (no result, content = JSON of
// the inbound command body, mirroring JSONObject.toJSONString(cmd)).

// accAuditLog mirrors PaasOperationLogUtil.accOperationLog (acc 1 -> on, 0 -> off).
func accAuditLog(cc *dto.CommandContext, imei, carID string, acc int, content string) {
	switch acc {
	case 1:
		emitOpLog(cc, opAccOn, imei, carID, content, nil)
	case 0:
		emitOpLog(cc, opAccOff, imei, carID, content, nil)
	}
}

// defendAuditLog mirrors PaasOperationLogUtil.defendOperationLog (1 -> on, 0 -> off).
func defendAuditLog(cc *dto.CommandContext, imei, carID string, defend int, content string) {
	switch defend {
	case 1:
		emitOpLog(cc, opDefendOn, imei, carID, content, nil)
	case 0:
		emitOpLog(cc, opDefendOff, imei, carID, content, nil)
	}
}

// batteryCompartmentAuditLog mirrors PaasOperationLogUtil.batteryCompartmentOperationLog
// (sw 1 -> 上锁/off, 0 -> 解锁/on).
func batteryCompartmentAuditLog(cc *dto.CommandContext, imei, carID string, sw int, content string) {
	switch sw {
	case 1:
		emitOpLog(cc, opBatteryOff, imei, carID, content, nil)
	case 0:
		emitOpLog(cc, opBatteryOn, imei, carID, content, nil)
	}
}

// rearWheelLockAuditLog mirrors PaasOperationLogUtil.rearWheelLockOperationLog
// (sw 0 -> 锁开/on, 1 -> 锁关/off).
func rearWheelLockAuditLog(cc *dto.CommandContext, imei, carID string, sw int, content string) {
	switch sw {
	case 0:
		emitOpLog(cc, opRearWheelOn, imei, carID, content, nil)
	case 1:
		emitOpLog(cc, opRearWheelOff, imei, carID, content, nil)
	}
}

// The raw-map adapters below are used by the EcuCommand service (which forwards
// generic JSON bodies); the value-based cores above are reused by the typed BLE
// report commands.

func accOpLog(raw map[string]interface{}, cc *dto.CommandContext) {
	accAuditLog(cc, bodyStr(raw, "imei"), bodyStr(raw, "carId"), bodyInt(raw, "acc"), jsonStr(raw))
}

func defendOpLog(raw map[string]interface{}, cc *dto.CommandContext) {
	defendAuditLog(cc, bodyStr(raw, "imei"), bodyStr(raw, "carId"), bodyInt(raw, "defend"), jsonStr(raw))
}

func batteryCompartmentOpLog(raw map[string]interface{}, cc *dto.CommandContext) {
	batteryCompartmentAuditLog(cc, bodyStr(raw, "imei"), bodyStr(raw, "carId"), bodyInt(raw, "sw"), jsonStr(raw))
}

func rearWheelLockOpLog(raw map[string]interface{}, cc *dto.CommandContext) {
	rearWheelLockAuditLog(cc, bodyStr(raw, "imei"), bodyStr(raw, "carId"), bodyInt(raw, "sw"), jsonStr(raw))
}

// findCarVoiceOpLog mirrors PaasOperationLogUtil.findCarVoiceOperationLog
// (only idx == 6 is logged: 播放寻车音).
func findCarVoiceOpLog(raw map[string]interface{}, cc *dto.CommandContext) {
	if bodyInt(raw, "idx") == 6 {
		emitOpLog(cc, opFindCarVoice, bodyStr(raw, "imei"), bodyStr(raw, "carId"), jsonStr(raw), nil)
	}
}

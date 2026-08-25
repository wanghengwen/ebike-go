package validator

import (
	"encoding/json"
	"errors"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"ebike-device-openapi-go/internal/api/dto"
)

var voiceTypePattern = regexp.MustCompile(`^(amr|wav|mp3)$`)

func fail(msg string) *dto.Result {
	return &dto.Result{Success: false, Code: "10002", Msg: msg}
}

// ValidateCmdParams mirrors Java @Validated on per-command Request DTOs.
func ValidateCmdParams(cmdID int16, m map[string]interface{}) *dto.Result {
	if m == nil {
		m = map[string]interface{}{}
	}

	switch cmdID {
	case 33:
		return validateLock(m)
	case 4:
		return validateDefend(m)
	case 82:
		return validateHelmetLock(m)
	case 32:
		return validateSetInnerParam(m)
	case 50:
		return validateUpdateInnerFence(m)
	case 35:
		return validateUpgradeDevice(m)
	case 57:
		return validateUpgradeVoice(m)
	case 14:
		return validateBroadcastVoice(m)
	case 34:
		return validateQueryDeviceInfo(m)
	case 40:
		return validateSwitchBatteryCompartment(m)
	case 49:
		return validateSetBluetooth(m)
	case 107:
		return validateScanBleHelmet(m)
	case 28:
		return validateSwitchRearWheelLock(m)
	case 103:
		return validateSetDashboard(m)
	case 115:
		return validateSetParkSite(m)
	case 113:
		return validateTriggerTempState(m)
	case 121:
		return validateDynamicVoice(m)
	case 78:
		return requireIntRange(m, "forceLock", "forceLock不能为空", "forceLock应为0或1", 0, 1)
	case 45:
		return requireIntRange(m, "speed", "speed不能为空", "speed应为[1,100]之间的数字", 1, 100)
	case 93:
		return requireIntRange(m, "isLiftSpeed", "isLiftSpeed不能为空", "isLiftSpeed应为0或1", 0, 1)
	case 21, 25, 31, 85, 108, 124:
		return nil
	default:
		return nil
	}
}

func validateLock(m map[string]interface{}) *dto.Result {
	// Java LockRequest: @NotNull(message = "acc不能为空") @Range(0,1)
	if r := requireIntRange(m, "acc", "acc不能为空", "acc应为0或1", 0, 1); r != nil {
		return r
	}
	if r := optionalIntRange(m, "volume", "volume应在[1,100]之间的数字", 1, 100); r != nil {
		return r
	}
	if r := optionalIntRange(m, "isIgnoreFence", "isIgnoreFence应为0或1", 0, 1); r != nil {
		return r
	}
	if r := optionalIntRange(m, "isTBeacon", "isTBeacon应为0或1", 0, 1); r != nil {
		return r
	}
	if r := optionalIntRange(m, "isKickstand", "isKickstand应为0或1", 0, 1); r != nil {
		return r
	}
	if r := optionalIntRange(m, "singleSpeedLimit", "singleSpeedLimit不在取值范围内", 0, 100); r != nil {
		return r
	}
	if r := optionalURL(m, "url", "url格式不正确"); r != nil {
		return r
	}
	return optionalVoiceType(m, "type", "type不在取值范围内")
}

func validateDefend(m map[string]interface{}) *dto.Result {
	if r := requireIntRange(m, "defend", "defend不能为空", "defend应为0或1", 0, 1); r != nil {
		return r
	}
	if r := optionalIntRange(m, "volume", "volume应在[0,100]之间的数字", 0, 100); r != nil {
		return r
	}
	if r := optionalIntRange(m, "isTBeacon", "isTBeacon应为0或1", 0, 1); r != nil {
		return r
	}
	if r := optionalIntRange(m, "isSlopeStake", "isSlopeStake应为0或1", 0, 1); r != nil {
		return r
	}
	if r := optionalIntRange(m, "isRFID", "isRFID应为0或1", 0, 1); r != nil {
		return r
	}
	if r := optionalIntRange(m, "isKickstand", "isKickstand应为0或1", 0, 1); r != nil {
		return r
	}
	if r := optionalIntRange(m, "isCamera", "isCamera应为0或1", 0, 1); r != nil {
		return r
	}
	if r := optionalIntRange(m, "isHelmetLockOn", "isHelmetLockOn应为0或1", 0, 1); r != nil {
		return r
	}
	if r := optionalURL(m, "url", "url格式不正确"); r != nil {
		return r
	}
	return optionalVoiceType(m, "type", "type不在取值范围内")
}

func validateHelmetLock(m map[string]interface{}) *dto.Result {
	if r := requireIntRange(m, "sw", "sw不能为空", "sw应为0或1", 0, 1); r != nil {
		return r
	}
	if r := optionalIntRange(m, "volume", "volume应在[1,100]之间的数字", 1, 100); r != nil {
		return r
	}
	if r := optionalIntRange(m, "isAutoLock", "isAutoLock应为0或1", 0, 1); r != nil {
		return r
	}
	if r := optionalIntRange(m, "isAutoOpen", "isAutoOpen应为0或1", 0, 1); r != nil {
		return r
	}
	return optionalURL(m, "url", "url格式不正确")
}

func validateSetInnerParam(m map[string]interface{}) *dto.Result {
	rules := []struct {
		key, msg string
		min, max int
	}{
		{"mode", "mode应为[0,2]之间的数字", 0, 2},
		{"isOverSpeedOn", "isOverSpeedOn应为0或1", 0, 1},
		{"isEnablePhysicalKey", "物理钥匙失效或启用 应为0或1", 0, 1},
		{"isAutoLockOn", "isAutoLockOn应为0或1", 0, 1},
		{"isMoveAlarmOn", "isMoveAlarmOn应为0或1", 0, 1},
		{"isNightVoiceOn", "isMoveAlarmOn应为0或1", 0, 1},
		{"autoLockPeriod", "autoLockPeriod应为[300,3600]之间的数字", 300, 3600},
		{"isFenceEnable", "isFenceEnable应为0或1", 0, 1},
		{"isTBeaconEnable", "isTBeaconEnable应为0或1", 0, 1},
		{"tBeaconthresholdCompensate", "tBeaconthresholdCompensate应为[-128,127]之间的数字", -128, 127},
		{"isTurnOverEnable", "isTurnOverEnable应为0或1", 0, 1},
		{"turnOverAngle", "turnOverAngle应为[0,90]之间的数字", 0, 90},
		{"isEtcSOC", "isEtcSOC应为0或1", 0, 1},
		{"bmsType", "bmsType应为[0,4]之间的数字", 0, 4},
		{"ccuType", "ccuType应为[0,6]之间的数字", 0, 6},
		{"isBleKeyEnable", "isBleKeyEnable应为0或1", 0, 1},
		{"is433KeyEnable", "is433KeyEnable应为0或1", 0, 1},
		{"beaconRssi", "beaconRssi应为[20,128]之间的数字", 20, 128},
	}
	for _, rule := range rules {
		if r := optionalIntRange(m, rule.key, rule.msg, rule.min, rule.max); r != nil {
			return r
		}
	}
	if r := optionalFloatRange(m, "audioRatio", "audioRatio应为[0.1,1.0]之间的数字", 0.1, 1.0); r != nil {
		return r
	}
	return optionalFloatRange(m, "audioRatioOverSpeed", "audioRatio应为[0.1,1.0]之间的数字", 0.1, 1.0)
}

func validateUpdateInnerFence(m map[string]interface{}) *dto.Result {
	if r := requireInt(m, "crc", "crc不能为空"); r != nil {
		return r
	}
	if r := requireIntRange(m, "isFenceEnable", "isFenceEnable不能为空", "isFenceEnable应为0或1", 0, 1); r != nil {
		return r
	}
	return optionalURL(m, "url", "url格式不正确")
}

func validateUpgradeDevice(m map[string]interface{}) *dto.Result {
	if r := requireNonEmptyString(m, "url", "url不能为空"); r != nil {
		return r
	}
	if r := optionalURL(m, "url", "url格式不正确"); r != nil {
		return r
	}
	return requireInt(m, "crc", "crc不能为空")
}

func validateUpgradeVoice(m map[string]interface{}) *dto.Result {
	if _, ok := m["idx"]; !ok || m["idx"] == nil {
		return fail("idx不能为空")
	}
	switch v := m["idx"].(type) {
	case []interface{}:
		if len(v) == 0 {
			return fail("idx不能为空")
		}
	default:
		return fail("idx不能为空")
	}
	if r := requireNonEmptyString(m, "url", "url不能为空"); r != nil {
		return r
	}
	if r := optionalURL(m, "url", "url格式不正确"); r != nil {
		return r
	}
	if r := requireInt(m, "crc", "crc不能为空"); r != nil {
		return r
	}
	return requireInt(m, "type", "type不能为空")
}

func validateBroadcastVoice(m map[string]interface{}) *dto.Result {
	if r := requireInt(m, "idx", "idx不能为空"); r != nil {
		return r
	}
	if r := optionalIntRange(m, "volume", "volume不在取值范围内", 1, 100); r != nil {
		return r
	}
	if r := optionalURL(m, "url", "url格式不正确"); r != nil {
		return r
	}
	if v, ok := m["type"]; ok && v != nil {
		s, ok := v.(string)
		if !ok || !voiceTypePattern.MatchString(s) {
			return fail("type不在取值范围内")
		}
	}
	return nil
}

func validateQueryDeviceInfo(m map[string]interface{}) *dto.Result {
	return optionalIntRange(m, "isCameraEnable", "isCameraEnable应为0或1", 0, 1)
}

func validateSwitchBatteryCompartment(m map[string]interface{}) *dto.Result {
	if r := requireIntRange(m, "sw", "sw不能为空", "sw应为0或1", 0, 1); r != nil {
		return r
	}
	if r := optionalIntRange(m, "volume", "volume应在[1,100]之间的数字", 1, 100); r != nil {
		return r
	}
	if r := optionalIntRange(m, "isFenceEnable", "isFenceEnable应为0或1", 0, 1); r != nil {
		return r
	}
	return optionalURL(m, "url", "url格式不正确")
}

func validateSetBluetooth(m map[string]interface{}) *dto.Result {
	if v, ok := m["name"]; ok && v != nil {
		s, ok := v.(string)
		if ok && len(s) > 15 {
			return fail("name应在15个字符内")
		}
	}
	return nil
}

func validateScanBleHelmet(m map[string]interface{}) *dto.Result {
	if r := requireIntRange(m, "isUnbound", "isUnbound不能为空", "isUnbound应为0或1", 0, 1); r != nil {
		return r
	}
	if v, ok := m["mac"]; ok && v != nil {
		s, ok := v.(string)
		if ok && len(s) > 12 {
			return fail("mac应为12位16进制字符")
		}
	}
	return nil
}

func validateSwitchRearWheelLock(m map[string]interface{}) *dto.Result {
	if r := requireIntRange(m, "sw", "sw不能为空", "sw应为0或1", 0, 1); r != nil {
		return r
	}
	return optionalIntRange(m, "volume", "语音音量<1,100>, 跟随idx可选。如果无该字段，默认100", 1, 100)
}

func validateSetDashboard(m map[string]interface{}) *dto.Result {
	rules := []struct {
		key, msg string
		min, max int
	}{
		{"bikeState", "bikeState应为[0,5]区间的整数", 0, 5},
		{"parkState", "parkState应为[0,1]区间的整数", 0, 1},
		{"rideTime", "rideTime应为[0,99]区间的整数", 0, 99},
		{"dashboardLed", "dashboardLed应为[0,3]区间的整数", 0, 3},
		{"dashboardLight", "dashboardLight应为[0,3]区间的整数", 0, 3},
		{"socLedType", "socLedType应为[0,1]区间的整数", 0, 1},
	}
	for _, rule := range rules {
		if r := optionalIntRange(m, rule.key, rule.msg, rule.min, rule.max); r != nil {
			return r
		}
	}
	return optionalFloatRange(m, "rideCharge", "rideCharge应为[0.1,99.9]区间的数", 0.1, 99.9)
}

func validateSetParkSite(m map[string]interface{}) *dto.Result {
	if r := requireInt(m, "parkSiteType", "站点类型不能为空"); r != nil {
		return r
	}
	if v, ok := m["parkSiteGps"]; !ok || v == nil {
		return fail("站点描述定位点不能为空")
	}
	if r := optionalIntRange(m, "parkLedMode", "parkLedMode为[0,2]区间的整数", 0, 2); r != nil {
		return r
	}
	return optionalIntRange(m, "parkLedTime", "parkLedTime为[0,60]区间的整数", 0, 60)
}

func validateTriggerTempState(m map[string]interface{}) *dto.Result {
	return requireIntRange(m, "stateCmd", "acc不能为空", "stateCmd应为0或1", 0, 1)
}

func validateDynamicVoice(m map[string]interface{}) *dto.Result {
	if r := requireInt(m, "type", "type不能为空"); r != nil {
		return r
	}
	return optionalIntRange(m, "volume", "volume应为[1,100]", 1, 100)
}

func requireInt(m map[string]interface{}, key, emptyMsg string) *dto.Result {
	v, ok := m[key]
	if !ok || v == nil {
		return fail(emptyMsg)
	}
	if _, err := toInt(v); err != nil {
		return fail(emptyMsg)
	}
	return nil
}

func requireNonEmptyString(m map[string]interface{}, key, emptyMsg string) *dto.Result {
	v, ok := m[key]
	if !ok || v == nil {
		return fail(emptyMsg)
	}
	s, ok := v.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return fail(emptyMsg)
	}
	return nil
}

func requireIntRange(m map[string]interface{}, key, emptyMsg, rangeMsg string, min, max int) *dto.Result {
	v, ok := m[key]
	if !ok || v == nil {
		return fail(emptyMsg)
	}
	ival, err := toInt(v)
	if err != nil {
		return fail(emptyMsg)
	}
	if ival < min || ival > max {
		return fail(rangeMsg)
	}
	return nil
}

func optionalIntRange(m map[string]interface{}, key, rangeMsg string, min, max int) *dto.Result {
	v, ok := m[key]
	if !ok || v == nil || isBlankValue(v) {
		return nil
	}
	ival, err := toInt(v)
	if err != nil {
		return fail(rangeMsg)
	}
	if ival < min || ival > max {
		return fail(rangeMsg)
	}
	return nil
}

func optionalFloatRange(m map[string]interface{}, key, rangeMsg string, min, max float64) *dto.Result {
	v, ok := m[key]
	if !ok || v == nil || isBlankValue(v) {
		return nil
	}
	fval, err := toFloat(v)
	if err != nil {
		return fail(rangeMsg)
	}
	if fval < min || fval > max {
		return fail(rangeMsg)
	}
	return nil
}

func optionalURL(m map[string]interface{}, key, invalidMsg string) *dto.Result {
	v, ok := m[key]
	if !ok || v == nil || isBlankValue(v) {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return fail(invalidMsg)
	}
	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fail(invalidMsg)
	}
	return nil
}

func optionalVoiceType(m map[string]interface{}, key, invalidMsg string) *dto.Result {
	v, ok := m[key]
	if !ok || v == nil || isBlankValue(v) {
		return nil
	}
	s, ok := v.(string)
	if !ok || !voiceTypePattern.MatchString(s) {
		return fail(invalidMsg)
	}
	return nil
}

func isBlankValue(v interface{}) bool {
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val) == ""
	default:
		return false
	}
}

func toInt(v interface{}) (int, error) {
	switch n := v.(type) {
	case float64:
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return 0, errors.New("invalid number")
		}
		return int(n), nil
	case int:
		return n, nil
	case int32:
		return int(n), nil
	case int64:
		return int(n), nil
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return 0, err
		}
		return int(i), nil
	case string:
		s := strings.TrimSpace(n)
		if s == "" {
			return 0, errors.New("invalid number")
		}
		i, err := strconv.Atoi(s)
		if err != nil {
			return 0, err
		}
		return i, nil
	case bool:
		if n {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, errors.New("invalid number")
	}
}

func toFloat(v interface{}) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case int:
		return float64(n), nil
	case int32:
		return float64(n), nil
	case int64:
		return float64(n), nil
	case json.Number:
		return n.Float64()
	case string:
		s := strings.TrimSpace(n)
		if s == "" {
			return 0, errors.New("invalid number")
		}
		return strconv.ParseFloat(s, 64)
	default:
		return 0, errors.New("invalid number")
	}
}

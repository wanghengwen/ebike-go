// Package api defines the unified response envelope and error codes, kept
// compatible with the Java com.xyy.dto.Result so upstream callers see no change.
package api

import "net/http"

// Result is the standard response envelope.
//
//	{"success":true,"code":"0","msg":"成功","data":...}
type Result struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	// data is always emitted (even as null): ebike-device-paas serializes with
	// Spring's default Jackson and no default-property-inclusion=non_null is set
	// anywhere (app, bootstrap.yml, or xyy starters), so Jackson's ALWAYS policy
	// renders null data as "data":null. No omitempty here keeps that parity.
	Data interface{} `json:"data"`
}

// Known message codes observed in production logs / Java DevicePassMsgCode.
const (
	CodeSuccess = "0"
	// CodeParamError mirrors com.xyy.common.constant.MsgCodeEnum.PARAM_EXCEPTION
	// ("00002"), returned by GlobalExceptionHandler.handleHttpMessageNotReadable
	// when the request body cannot be parsed. (Java has no "-3" code.)
	CodeParamError          = "00002"
	CodeEcuExeFailed        = "17002"
	CodeEcuInfoFindFailed   = "17003"
	CodeImeiCarIdEmpty      = "17004"
	CodeImeiCarIdBind       = "17005"
	CodeDeviceNullHave      = "17006"
	CodeDeviceNullRack      = "170018"
	CodeDeviceOffline       = "17012"
	CodeDeviceNoResponse    = "17013"
	CodeTrajectory          = "17014"
	CodeTrajectoryDateRange = "17015"
	CodeDeviceGatewayError  = "17012"
	CodeGenFakeOutOfLimit   = "17016"
	CodeGenFakeNx           = "17017"
	// CodeRateLimit mirrors com.xyy.common.constant.MsgCodeEnum.RATE_LIMIT_ERROR.
	CodeRateLimit = "00026"
)

var msgByCode = map[string]string{
	CodeSuccess:           "成功",
	CodeParamError:        "http message not readable",
	CodeEcuExeFailed:      "ECU execution failed",
	CodeEcuInfoFindFailed: "ECU information query failed",
	CodeDeviceNullHave:    "对应设备不存在",
	CodeDeviceNullRack:    "车辆未上架",
	CodeDeviceOffline:     "设备离线",
	CodeDeviceNoResponse:  "设备未返回",
	CodeTrajectory:        "订单轨迹不存在",
}

// FailData builds a failure Result that still carries data (mirrors Java
// `new Result<>(data, msgCode, false)`).
func FailData(code, msg string, data interface{}) Result {
	if msg == "" {
		msg = msgByCode[code]
	}
	return Result{Success: false, Code: code, Msg: msg, Data: data}
}

// Ok wraps data in a success Result.
func Ok(data interface{}) Result {
	return Result{Success: true, Code: CodeSuccess, Msg: msgByCode[CodeSuccess], Data: data}
}

// Fail builds a failure Result for a known code (msg auto-filled if registered).
func Fail(code, msg string) Result {
	if msg == "" {
		msg = msgByCode[code]
	}
	return Result{Success: false, Code: code, Msg: msg}
}

// HTTPStatus maps a Result to an HTTP status. The Java service returns 200 for
// business errors too, so we mirror that.
func (r Result) HTTPStatus() int { return http.StatusOK }

// Package contract is the Xiaoan PaaS response envelope shared by HTTP handlers
// and middleware (auth / rate-limit). Kept separate from package api so that
// middleware does not import the handler package (which would create a cycle).
package contract

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the Xiaoan success/failure envelope.
type Response struct {
	Success bool         `json:"success"`
	Data    interface{}  `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

// ErrorDetail is Xiaoan's error object.
type ErrorDetail struct {
	ErrorMessage string `json:"errormessage"`
	Promot       string `json:"promot"`
}

// DeviceCodeData is the body of a device-communication response.
type DeviceCodeData struct {
	Code int `json:"code"`
}

const (
	ErrInputParamsMiss  = "INPUT_PARAMS_MISS"
	ErrSendCmdError     = "SEND_CMD_ERROR"
	ErrDeviceInfoIsNull = "DEVICE_INFO_IS_NULL"
	ErrUnauthorized     = "UNAUTHORIZED_ERROR"
	ErrImeiIllegal      = "IMEI_ILLEGAL"
	ErrBizError         = "BIZ_ERROR"
	// ErrRangeTooLarge carries business code 105 on the endpoints that return a
	// list rather than a data.code, where there is nowhere to put the number.
	ErrRangeTooLarge = "RANGE_TOO_LARGE"
)

const (
	CodeOK             = 0
	CodeServerInternal = 100
	CodeDeviceNoResp   = 108
	CodeDeviceOffline  = 109
	CodeDeviceMemory   = 110
	CodeUnsupported    = 111
	CodeBatteryUnset   = 113
	CodeParamInvalid   = 114
	CodeParamMissing   = 115
	CodeRidingNoLock   = 132
	CodeRangeTooLarge  = 105
	CodeNoTBeacon      = 136
	CodeNotBelongAgent = 1001
)

var promotByErrorType = map[string]string{
	ErrInputParamsMiss:  "request params missing!",
	ErrSendCmdError:     "send command error",
	ErrDeviceInfoIsNull: "device info is null",
	ErrUnauthorized:     "token unauthorized",
	ErrImeiIllegal:      "imei illegal",
	ErrBizError:         "biz error",
	ErrRangeTooLarge:    "request range too large",
}

// OK writes a success envelope.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Success: true, Data: data})
}

// OKDeviceCode writes a success envelope carrying only a device result code.
func OKDeviceCode(c *gin.Context, code int) {
	c.JSON(http.StatusOK, Response{Success: true, Data: DeviceCodeData{Code: code}})
}

// Fail writes a business-error envelope (always HTTP 200).
func Fail(c *gin.Context, errorType, promot string) {
	if promot == "" {
		promot = promotByErrorType[errorType]
	}
	c.JSON(http.StatusOK, Response{
		Success: false,
		Error:   &ErrorDetail{ErrorMessage: errorType, Promot: promot},
	})
}

// MapEcuCode translates an internal ecuCode / gateway code into Xiaoan data.code.
func MapEcuCode(ecuCode string) int {
	switch ecuCode {
	case "", "0", "00":
		return CodeOK
	case "105":
		return CodeRangeTooLarge
	case "110":
		return CodeDeviceMemory
	case "111":
		return CodeUnsupported
	case "113":
		return CodeBatteryUnset
	case "114":
		return CodeParamInvalid
	case "115":
		return CodeParamMissing
	case "132":
		return CodeRidingNoLock
	case "135", "137", "10007", "10013":
		return CodeDeviceNoResp
	case "136":
		return CodeServerInternal
	case "142":
		return CodeNoTBeacon
	case "10012":
		return CodeDeviceOffline
	case "10006":
		return CodeNotBelongAgent
	case "10001", "10002":
		return CodeServerInternal
	default:
		return CodeServerInternal
	}
}

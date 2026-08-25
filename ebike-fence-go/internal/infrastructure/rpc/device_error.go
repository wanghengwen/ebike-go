package rpc

import "fmt"

// DeviceAPIError mirrors Java DeviceApiRpcImpl throwing BizException with upstream code/msg.
type DeviceAPIError struct {
	Code string
	Msg  string
}

func (e *DeviceAPIError) Error() string {
	if e == nil {
		return ""
	}
	return e.Msg
}

func newDeviceAPIError(code int, msg string) *DeviceAPIError {
	return &DeviceAPIError{Code: fmt.Sprintf("%d", code), Msg: msg}
}

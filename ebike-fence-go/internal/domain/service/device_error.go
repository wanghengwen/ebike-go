package service

import (
	"errors"
	"strings"

	infrarpc "ebike-fence-go/internal/infrastructure/rpc"
)

// deviceInfoBizError maps device-paas / transport failures to Java BizException codes.
// Java DeviceApiRpcImpl.getDeviceInfo throws BizException(upstream code/msg) on failure.
func deviceInfoBizError(err error) error {
	if err == nil {
		return nil
	}
	var apiErr *infrarpc.DeviceAPIError
	if errors.As(err, &apiErr) {
		if apiErr.Code == "00004" {
			return newBizError(apiErr.Code, apiErr.Msg)
		}
		msg := apiErr.Msg
		if msg == "" {
			msg = "设备离线或操作失败"
		} else if !strings.Contains(msg, "设备离线或操作失败") {
			msg = "设备离线或操作失败: " + msg
		}
		code := apiErr.Code
		if code == "" {
			code = codeDeviceTimeout
		}
		return newBizError(code, msg)
	}
	msg := err.Error()
	if strings.Contains(msg, "context deadline exceeded") ||
		strings.Contains(msg, "Client.Timeout") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "Timeout") {
		return newBizError(codeDeviceTimeout, "设备离线或操作失败: "+msg)
	}
	return newBizError(codeDeviceTimeout, "设备离线或操作失败: "+msg)
}

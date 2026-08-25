package dto

import "encoding/json"

// NewSuccessResult mirrors Java's ResultHelper.success(data):
// success=true, code "0" (MsgCodeEnum.SUCCESS), msg from the i18n msgcode bundle
// (both zh_CN and en_US bundles resolve "success" to "成功").
// Only used for responses the gateway constructs locally; downstream responses
// are passed through verbatim.
func NewSuccessResult(data interface{}) Result {
	code := CodeSuccess
	msg := "成功"
	var raw json.RawMessage
	if data != nil {
		raw, _ = json.Marshal(data)
	}
	return Result{Success: true, Code: &code, Msg: &msg, Data: raw}
}

// ServiceDTO matches Java: com.xyy.ebike.service.client.client.dto.ServiceDTO
// (serviceId is @NotNull)
type ServiceDTO struct {
	ClientDTO
	ServiceId *int64 `json:"serviceId" binding:"required"`
}

// ServicePageDTO matches Java: com.xyy.ebike.service.client.client.dto.ServicePageDTO
// (serviceId is @NotNull)
type ServicePageDTO struct {
	PageClientDTO
	ServiceId *int64 `json:"serviceId" binding:"required"`
}

// TestCmd matches Java: com.xyy.ebike.service.client.client.dto.TestCmd
// (name is @NotEmpty)
type TestCmd struct {
	ClientDTO
	Name string `json:"name" binding:"required" form:"name"`
}

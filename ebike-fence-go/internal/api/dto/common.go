package dto

import "encoding/json"

// Error codes matching Java com.xyy.common.constant.MsgCodeEnum
const (
	CodeSuccess         = "0"
	CodeException       = "00001"
	CodeParamException  = "00002"
	CodeIllegalArgument = "00004"
	CodeNotNull         = "00018"
)

// Result matches Java com.xyy.dto.Result
type Result struct {
	Success bool            `json:"success"`
	Code    *string         `json:"code"`
	Msg     *string         `json:"msg"`
	Data    json.RawMessage `json:"data"`
}

func NewErrorResult(code, msg string) Result {
	return Result{Success: false, Code: &code, Msg: &msg}
}

// NewSuccessResult mirrors Java ResultHelper.success(data).
func NewSuccessResult(data interface{}) Result {
	code := CodeSuccess
	msg := "成功"
	var raw json.RawMessage
	if data != nil {
		raw, _ = json.Marshal(data)
	}
	return Result{Success: true, Code: &code, Msg: &msg, Data: raw}
}

// CommandContext matches Java com.xyy.dto.CommandContext
type CommandContext struct {
	TenantId      string `json:"tenantId,omitempty"`
	TraceId       string `json:"traceId,omitempty"`
	Pin           string `json:"pin,omitempty"`
	Ip            string `json:"ip,omitempty"`
	Platform      string `json:"platform,omitempty"`
	DeviceId      string `json:"deviceId,omitempty"`
	Source        string `json:"source,omitempty"`
	Name          string `json:"name,omitempty"`
	StressTesting bool   `json:"stressTesting"`
}

// Command matches Java com.xyy.dto.Command (all fence API request bodies embed this).
type Command struct {
	CommandContext *CommandContext `json:"commandContext" binding:"required"`
}

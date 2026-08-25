package dto

import "encoding/json"

// Error codes matching Java EbikeMsgCode / com.xyy.dto.Result
const (
	CodeSuccess                      = "0"
	CodeException                    = "10001"
	CodeParamException               = "10002"
	CodeIllegalArgument              = "10004"
	CodeOrderTrajectoryNotExist      = "10014"
	CodeTrajectoryQueryDateOutOfRange = "10015"
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

func NewSuccessResult(data interface{}) Result {
	code := CodeSuccess
	msg := "成功"
	var raw json.RawMessage
	if data != nil {
		raw, _ = json.Marshal(data)
	}
	return Result{Success: true, Code: &code, Msg: &msg, Data: raw}
}

package errcode

import "fmt"

// BizError represents a business-logic error with a machine-readable code
// and a human-readable message.
type BizError struct {
	Code string
	Msg  string
}

// Error implements the error interface.
func (e *BizError) Error() string {
	return fmt.Sprintf("BizError[%s]: %s", e.Code, e.Msg)
}

// NewBizError creates a new BizError.
func NewBizError(code string, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}

// Pre-defined business error codes.
var (
	ErrTenantConfigNotExist = &BizError{Code: "5000", Msg: "租户配置不存在"}
	ErrSupplierNotExist     = &BizError{Code: "5001", Msg: "供应商不存在"}
	ErrSignNameNotExist     = &BizError{Code: "5002", Msg: "短信签名不存在"}
	ErrTemplateNotExist     = &BizError{Code: "5003", Msg: "模板编码不存在"}
	ErrSecretMistake        = &BizError{Code: "5004", Msg: "秘钥错误"}
	ErrTaskRejected         = &BizError{Code: "5005", Msg: "任务被拒绝"}
	ErrChannelNotExist      = &BizError{Code: "5006", Msg: "供应商处理bean不存在"}
	ErrParamInvalid         = &BizError{Code: "4000", Msg: "参数错误"}
)

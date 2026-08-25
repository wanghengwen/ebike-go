package service

// BizError mirrors Java BizException with a business error code.
type BizError struct {
	Code string
	Msg  string
}

func (e *BizError) Error() string {
	return e.Msg
}

func newBizError(code, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}

// ExceptionError mirrors Java uncaught exceptions handled as code "00001".
type ExceptionError struct {
	Msg string
}

func (e *ExceptionError) Error() string { return e.Msg }

// ErrRidingCarConfigNullContext mirrors Java RidingCarConfigServiceImpl not passing
// commandContext into ConfigUseCarCmd, which NPEs when cancelAuthTenantIds is configured.
var ErrRidingCarConfigNullContext = &ExceptionError{Msg: "NullPointerException:null "}

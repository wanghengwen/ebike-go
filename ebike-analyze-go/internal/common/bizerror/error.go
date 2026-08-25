package bizerror

import (
	"fmt"
	"reflect"
	"strings"

	"ebike-analyze-go/internal/api/dto"
)

type BizError struct {
	ErrCode string
	Msg     string
	Args    []interface{}
}

func (e *BizError) Error() string { return e.FormattedMsg() }
func (e *BizError) Code() string  { return e.ErrCode }

func (e *BizError) FormattedMsg() string {
	if e == nil {
		return ""
	}
	if len(e.Args) == 0 {
		return e.Msg
	}
	return dto.FormatMsgCode(e.Msg, e.Args)
}

func New(code, msg string) *BizError {
	return &BizError{ErrCode: code, Msg: msg}
}

func NewWithArgs(code, msg string, args ...interface{}) *BizError {
	return &BizError{ErrCode: code, Msg: msg, Args: args}
}

// FormatException mirrors Java GlobalExceptionHandler:
// ClassName:message [optional first stack frame omitted in Go without debug.Stack].
func FormatException(err error) string {
	if err == nil {
		return ""
	}
	typeName := reflect.TypeOf(err).String()
	if idx := strings.LastIndex(typeName, "."); idx >= 0 {
		typeName = typeName[idx+1:]
	}
	typeName = strings.TrimPrefix(typeName, "*")
	return fmt.Sprintf("%s:%s", typeName, err.Error())
}

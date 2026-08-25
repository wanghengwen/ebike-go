package api

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func handleBindError(c *gin.Context, err error) {
	var verr validator.ValidationErrors
	if errors.As(err, &verr) && len(verr) > 0 {
		fe := verr[0]
		c.JSON(200, Error("-3", fe.Field()+" "+validationTagMessage(fe)))
		return
	}
	if isJSONReadableError(err) {
		c.JSON(200, Error("-3", "http message not readable"))
		return
	}
	c.JSON(200, Error("-3", "Invalid parameter"))
}

func validationTagMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "不能为空"
	default:
		if fe.Param() != "" {
			return fe.Tag() + "=" + fe.Param()
		}
		return fe.Tag()
	}
}

func isJSONReadableError(err error) bool {
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "json") ||
		strings.Contains(msg, "unexpected end of json input") ||
		strings.Contains(msg, "invalid character")
}

func handleError(c *gin.Context, err error) {
	errMsg := err.Error()
	if strings.HasPrefix(errMsg, "API_INVOKE_FAILED: ") {
		code := "00018"
		msg := strings.TrimPrefix(errMsg, "API_INVOKE_FAILED: ")
		c.JSON(200, Error(code, msg))
		return
	}

	errName := errorTypeName(err)
	frame := firstCallerFrame()
	code := fmt.Sprintf("%s:%s %s", errName, errMsg, frame)
	c.JSON(200, Error(code, errMsg))
}

func errorTypeName(err error) string {
	t := reflect.TypeOf(err)
	if t == nil {
		return "error"
	}
	if t.Kind() == reflect.Ptr {
		if t.Elem() == nil {
			return "error"
		}
		return t.Elem().Name()
	}
	return t.Name()
}

func firstCallerFrame() string {
	const depth = 32
	pcs := make([]uintptr, depth)
	n := runtime.Callers(3, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		if strings.Contains(frame.File, "runtime/") ||
			strings.Contains(frame.File, "gin-gonic") ||
			strings.Contains(frame.File, "middleware") {
			continue
		}
		return fmt.Sprintf("%s:%d", frame.File, frame.Line)
	}
	return ""
}

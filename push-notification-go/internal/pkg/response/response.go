package response

import (
	"net/http"
	"push-notification-go/internal/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// Result is the standard API response envelope.
type Result struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success creates a successful Result with code "0".
func Success(data interface{}) *Result {
	return &Result{
		Code:    "0",
		Message: "success",
		Data:    data,
	}
}

// Error creates a Result from a BizError.
func Error(err *errcode.BizError) *Result {
	return &Result{
		Code:    err.Code,
		Message: err.Msg,
	}
}

// ParamError creates a Result for parameter validation failures (code 4000).
func ParamError(msg string) *Result {
	return &Result{
		Code:    "4000",
		Message: msg,
	}
}

// Exception creates a Result for unexpected server errors (code 9999).
func Exception(msg string) *Result {
	return &Result{
		Code:    "9999",
		Message: msg,
	}
}

// SuccessJSON writes a successful JSON response to the Gin context.
func SuccessJSON(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Success(data))
}

// ErrorJSON writes a BizError JSON response to the Gin context.
func ErrorJSON(c *gin.Context, err *errcode.BizError) {
	c.JSON(http.StatusOK, Error(err))
}

package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Msg codes aligned with Java MsgCodeEnum.
const (
	CodeSuccess               = "0"
	CodeException             = "00001"
	CodeParamException        = "00002"
	CodeTokenInvalidOrExpired = "00005"
	CodeAccessUnauthorized    = "00006"
	CodeRequestOutOfDate      = "00007"
	CodeInvalidSign           = "00008"
	CodeInvalidToken          = "00013"
	CodeOtherDeviceLogin      = "00015"
	CodeAPIRefuseAccess       = "00024"
	CodeUserRoleChanged       = "00027"
)

// Result matches Java com.xyy.dto.Result JSON shape.
type Result struct {
	Success bool        `json:"success"`
	Code    string      `json:"code"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
}

func errorResult(code, msg string) Result {
	return Result{
		Success: false,
		Code:    code,
		Msg:     msg,
		Data:    nil,
	}
}

// Abort writes a unified error response and stops the handler chain.
func Abort(c *gin.Context, httpStatus int, code, msg string) {
	c.AbortWithStatusJSON(httpStatus, errorResult(code, msg))
}

// JSON writes a unified error response without aborting.
func JSON(c *gin.Context, httpStatus int, code, msg string) {
	c.JSON(httpStatus, errorResult(code, msg))
}

// AbortUnauthorized is a helper for missing/invalid bearer token.
func AbortUnauthorized(c *gin.Context) {
	Abort(c, http.StatusUnauthorized, CodeAccessUnauthorized, "请求未授权")
}

// AbortTokenInvalid is a helper for JWT validation failures.
func AbortTokenInvalid(c *gin.Context, msg string) {
	if msg == "" {
		msg = "token无效或过期"
	}
	Abort(c, http.StatusUnauthorized, CodeTokenInvalidOrExpired, msg)
}

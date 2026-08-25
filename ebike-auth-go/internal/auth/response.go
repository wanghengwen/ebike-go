package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommonResult struct {
	Success bool        `json:"success"`
	Code    string      `json:"code"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
}

func sendSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, CommonResult{
		Success: true,
		Code:    CodeSuccess,
		Msg:     defaultMsg(CodeSuccess),
		Data:    data,
	})
}

func sendError(c *gin.Context, code string, msg string) {
	if msg == "" {
		msg = defaultMsg(code)
	}
	c.JSON(http.StatusOK, CommonResult{
		Success: false,
		Code:    code,
		Msg:     msg,
		Data:    nil,
	})
}

func sendMustBindPhone(c *gin.Context, thirdLoginId int64) {
	c.JSON(http.StatusOK, CommonResult{
		Success: false,
		Code:    CodeSocialLoginMustBindPhone,
		Msg:     "需要绑定手机号",
		Data: map[string]interface{}{
			"thirdLoginId": thirdLoginId,
		},
	})
}

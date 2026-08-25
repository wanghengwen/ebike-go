package auth

// MsgCode constants aligned with Java com.xyy.common.constant.MsgCodeEnum
const (
	CodeSuccess                      = "0"
	CodeException                    = "00001"
	CodeParamException               = "00002"
	CodeIllegalArgument              = "00004"
	CodeTokenInvalidOrExpired        = "00005"
	CodeAccessUnauthorized           = "00006"
	CodeAuthenticationFailed         = "00009"
	CodeMessageCodeVerifyFailed      = "00010"
	CodeUserNotFound                 = "00011"
	CodeInvalidGrant                 = "00012"
	CodeInvalidToken                 = "00013"
	CodeInvalidUser                  = "00014"
	CodeOtherDeviceLogin             = "00015"
	CodeComponentVerifyTicketMissing = "00016"
	CodeSocialLoginFailed            = "00020"
	CodeSocialLoginMustBindPhone     = "00021"
	CodeUserDisabled                 = "00022"
	CodeUserNoPermission             = "00023"
	CodeUserRoleChanged              = "00027"
	CodeAKVerifyFailed               = "00009"
)

var defaultMsgByCode = map[string]string{
	CodeSuccess:                      "成功",
	CodeException:                    "异常",
	CodeParamException:               "参数异常",
	CodeIllegalArgument:              "非法参数",
	CodeTokenInvalidOrExpired:        "token无效或过期",
	CodeAccessUnauthorized:           "请求未授权",
	CodeAuthenticationFailed:         "http认证头信息失败",
	CodeMessageCodeVerifyFailed:      "短信码验证失败",
	CodeUserNotFound:                 "用户不存在",
	CodeInvalidGrant:                 "登录方式无效",
	CodeInvalidToken:                 "登录已过期，请重新登录",
	CodeInvalidUser:                  "无效用户",
	CodeOtherDeviceLogin:             "账号已在别处登录",
	CodeComponentVerifyTicketMissing: "component验证票据不存在",
	CodeSocialLoginFailed:            "三方登录失败",
	CodeSocialLoginMustBindPhone:     "三方登录必须绑定手机号",
	CodeUserDisabled:                 "用户已禁用",
	CodeUserNoPermission:             "用户没权限",
	CodeUserRoleChanged:              "账号权限已变更，请重新登录",
}

func defaultMsg(code string) string {
	if msg, ok := defaultMsgByCode[code]; ok {
		return msg
	}
	return "异常"
}

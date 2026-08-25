package validator

import (
	"errors"
	"regexp"
	"strings"

	"push-notification-go/internal/pkg/errcode"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Matches Java SendMessage/SendVoice @Pattern on phone field.
var chinaPhoneRegex = regexp.MustCompile(`^1(3\d|4[5-9]|5[0-35-9]|6[567]|7[0-8]|8\d|9[0-35-9])\d{8}$`)

type fieldRule struct {
	jsonName string
	required string
	invalid  string
}

var sendMessageRules = map[string]fieldRule{
	"TraceId":      {jsonName: "traceId", required: "不能为空"},
	"Source":       {jsonName: "source", required: "不能为空"},
	"TenantId":     {jsonName: "tenantId", required: "租户id不能为空"},
	"Phone":        {jsonName: "phone", required: "手机号不能为空", invalid: "手机号格式错误"},
	"TemplateCode": {jsonName: "templateCode", required: "模板编码不能为空"},
	"SignName":     {jsonName: "signName", required: "签名不能为空"},
	"Secret":       {jsonName: "secret", required: "不能为空"},
}

var sendVoiceRules = map[string]fieldRule{
	"TraceId":      {jsonName: "traceId", required: "不能为空"},
	"Source":       {jsonName: "source", required: "不能为空"},
	"TenantId":     {jsonName: "tenantId", required: "租户id不能为空"},
	"Phone":        {jsonName: "phone", required: "手机号不能为空", invalid: "手机号格式错误"},
	"TemplateCode": {jsonName: "templateCode", required: "模板编码不能为空"},
	"Secret":       {jsonName: "secret", required: "不能为空"},
}

// Register installs custom validators used by Gin binding.
func Register() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	_ = v.RegisterValidation("china_phone", validateChinaPhone)
}

func validateChinaPhone(fl validator.FieldLevel) bool {
	return chinaPhoneRegex.MatchString(fl.Field().String())
}

// FormatBindError converts Gin/validator binding errors into Java-style
// param errors: "{jsonField} {message}" with code 4000.
func FormatBindError(err error, forVoice bool) *errcode.BizError {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) || len(verrs) == 0 {
		return errcode.ErrParamInvalid
	}

	rules := sendMessageRules
	if forVoice {
		rules = sendVoiceRules
	}

	fe := verrs[0]
	rule, ok := rules[fe.Field()]
	if !ok {
		return errcode.NewBizError("4000", jsonFieldName(fe.Field())+" 参数错误")
	}

	msg := messageForTag(fe.Tag(), rule)
	return errcode.NewBizError("4000", rule.jsonName+" "+msg)
}

func jsonFieldName(structField string) string {
	if rule, ok := sendMessageRules[structField]; ok {
		return rule.jsonName
	}
	if rule, ok := sendVoiceRules[structField]; ok {
		return rule.jsonName
	}
	return strings.ToLower(structField)
}

func messageForTag(tag string, rule fieldRule) string {
	switch tag {
	case "china_phone":
		if rule.invalid != "" {
			return rule.invalid
		}
		return "格式错误"
	case "required":
		if rule.required != "" {
			return rule.required
		}
		return "不能为空"
	default:
		if rule.invalid != "" {
			return rule.invalid
		}
		return "参数错误"
	}
}

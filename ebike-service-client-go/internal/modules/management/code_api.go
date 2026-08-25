package management

// Ports Java CodeController (@RequestMapping("/client")):
//   - POST /client/code/send          -> ebike-management /code/send
//   - POST /client/code/sendEmailCode -> ebike-management /email/code/send
//   - POST /client/code/verify        -> ebike-management /code/verify

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/config"
	"ebike-service-client-go/internal/pkg/javacompat"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// smsSendCodeDTO matches Java SmsSendCodeDto extends ClientDTO
// (phone/scene are @NotBlank, controller has @Valid).
type smsSendCodeDTO struct {
	dto.ClientDTO
	Phone string `json:"phone" binding:"required"`
	Scene string `json:"scene" binding:"required"`
}

// codeVerifyDTO matches Java CodeVerifyDto extends ClientDTO
// (phone/code are @NotBlank, controller has @Valid).
type codeVerifyDTO struct {
	dto.ClientDTO
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// emailSendCodeDTO matches Java EmailSendCodeDto extends Command (NOT ClientDTO!):
// the request body itself carries the commandContext, like the Java
// CommandContextAspect expects for Command-typed controller arguments.
type emailSendCodeDTO struct {
	Email          string              `json:"email" binding:"required"`
	Scene          string              `json:"scene" binding:"required"`
	CommandContext *dto.CommandContext `json:"commandContext"`
}

var smsSendCodeMsgs = mergeMsgs(map[string]string{
	"phone": "must not be blank",
	"scene": "must not be blank",
})

var codeVerifyMsgs = mergeMsgs(map[string]string{
	"phone": "must not be blank",
	"code":  "must not be blank",
})

// emailSendCodeDTO does not extend ClientDTO, so no traceId/tenantId validation.
var emailSendCodeMsgs = map[string]string{
	"email": "must not be blank",
	"scene": "must not be blank",
}

// validPhone replicates com.xyy.common.utils.ValidUtils.validPhone:
// the phone must be "+<countryCode>-<number>" where both parts parse as longs.
func validPhone(phone string) bool {
	arr := javaSplit(phone, "-")
	if len(arr) != 2 || !strings.HasPrefix(arr[0], "+") {
		return false
	}
	cc := arr[0][1:]
	if cc == "" {
		return false
	}
	if _, err := strconv.ParseInt(cc, 10, 64); err != nil {
		return false
	}
	if _, err := strconv.ParseInt(arr[1], 10, 64); err != nil {
		return false
	}
	return true
}

// loginPhoneErrorMsg resolves ClientMsgCode.LOGIN_PHONE_ERROR ("10015") the way
// the Java ResourceBundleMessageSource does for the request locale:
// i18n/msg_en_US.properties has "login phone valid fail"; zh and any locale
// without the key fall back to the enum default "手机号格式错误".
func loginPhoneErrorMsg(acceptLanguage string) string {
	lang := strings.ToLower(strings.TrimSpace(strings.SplitN(strings.SplitN(acceptLanguage, ",", 2)[0], ";", 2)[0]))
	if lang == "en" || strings.HasPrefix(lang, "en-") || strings.HasPrefix(lang, "en_") {
		return "login phone valid fail"
	}
	return "手机号格式错误"
}

// sendCode ports CodeController.send.
// NOTE: the Java controller REPLACES the CommandContext with a fresh one
// holding only traceId/tenantId/pin(=phone) before calling the gateway, so the
// downstream commandContext intentionally lacks source/platform/deviceId.
func sendCode(c *gin.Context) {
	var req smsSendCodeDTO
	if !web.BindJSON(c, &req, smsSendCodeMsgs) {
		return
	}
	if !web.NotBlank(c, "phone", req.Phone) || !web.NotBlank(c, "scene", req.Scene) {
		return
	}
	// Java: Assert.isTrue(ValidUtils.validPhone(...), ClientMsgCode.LOGIN_PHONE_ERROR)
	if !validPhone(req.Phone) {
		c.JSON(http.StatusOK, dto.NewErrorResult("10015", loginPhoneErrorMsg(c.GetHeader("Accept-Language"))))
		return
	}
	cmdCtx := &dto.CommandContext{
		TraceId:  req.TraceId,
		TenantId: req.TenantId,
		Pin:      req.Phone,
	}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceManagement, "/code/send", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertBoolean(result.Data)
	}
	web.RespondResult(c, result, err)
}

// sendEmailCode ports CodeController.sendEmailCode.
// The DTO extends Command, so Java's CommandContextAspect takes the
// commandContext from the request body itself and only completes source.
func sendEmailCode(c *gin.Context) {
	var req emailSendCodeDTO
	if !web.BindJSON(c, &req, emailSendCodeMsgs) {
		return
	}
	if !web.NotBlank(c, "email", req.Email) || !web.NotBlank(c, "scene", req.Scene) {
		return
	}
	if req.CommandContext == nil {
		// Java: NPE inside CommandContextAspect (context.setSource on a null
		// commandContext) -> GlobalExceptionHandler -> code "00001"
		web.WriteException(c, `NullPointerException:Cannot invoke "com.xyy.dto.CommandContext.setSource(String)" because "context" is null`)
		return
	}
	cmdCtx := req.CommandContext
	cmdCtx.Source = config.GlobalConfig.Server.Name
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceManagement, "/email/code/send", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertBoolean(result.Data)
	}
	web.RespondResult(c, result, err)
}

// verifyCode ports CodeController.verify (pure passthrough; the downstream
// CodeVerifyCmd's scene stays null and cleanCache keeps its Java default true,
// exactly like BeanUtils.copyProperties produces).
func verifyCode(c *gin.Context) {
	var req codeVerifyDTO
	if !web.BindJSON(c, &req, codeVerifyMsgs) {
		return
	}
	if !web.NotBlank(c, "phone", req.Phone) || !web.NotBlank(c, "code", req.Code) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceManagement, "/code/verify", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertBoolean(result.Data)
	}
	web.RespondResult(c, result, err)
}

func convertBoolean(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var b *bool
	if json.Unmarshal(raw, &b) != nil {
		return raw
	}
	out, _ := json.Marshal(b)
	return out
}

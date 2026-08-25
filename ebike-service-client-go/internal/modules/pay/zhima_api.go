package pay

import (
	"encoding/json"
	"net/http"
	"strings"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// ZhimaPayAfterUseController (/client/zhima/payafteruse/sign*) — @Validated
// ---------------------------------------------------------------------------

// zhimaSign handles POST /client/zhima/payafteruse/sign.
func zhimaSign(c *gin.Context) {
	var req zhimaSignDTO
	if !web.BindJSON(c, &req, clientMsgs) {
		return
	}

	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/zhima/payafteruse/sign", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertZhimaPayAfterUseSignCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

// zhimaSignQuery handles POST /client/zhima/payafteruse/signQuery.
func zhimaSignQuery(c *gin.Context) {
	var req zhimaSignQueryDTO
	if !web.BindJSON(c, &req, clientMsgs) {
		return
	}

	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/zhima/payafteruse/signQuery", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertZhimaPayAfterUseSignRecordCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

// ---------------------------------------------------------------------------
// ZhimaPayAfterUseOrderController (/client/zhima/payafteruse/order*) — NOT validated
// ---------------------------------------------------------------------------

// zhimaOrderWithoutConfirm handles POST /client/zhima/payafteruse/orderWithoutConfirm.
func zhimaOrderWithoutConfirm(c *gin.Context) {
	// Java field default: orderAmount = 20000 (explicit JSON null overwrites it)
	defaultAmount := 20000
	req := zhimaOrderWithoutConfirmDTO{OrderAmount: &defaultAmount}
	if !web.BindJSON(c, &req, nil) {
		return
	}

	cmdCtx := middleware.CompleteCommandContext(c, req.toShared())
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/zhima/payafteruse/orderWithoutConfirm", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertZhimaPayAfterUseOrderWithoutConfirmCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

// zhimaOrderQuery handles POST /client/zhima/payafteruse/orderQuery.
func zhimaOrderQuery(c *gin.Context) {
	var req zhimaOrderQueryDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}

	cmdCtx := middleware.CompleteCommandContext(c, req.toShared())
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/zhima/payafteruse/orderQuery", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertZhimaPayAfterUseOrderQueryCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

// zhimaOrderFinish handles POST /client/zhima/payafteruse/orderFinish.
// Java: MayiPayGatewayImpl.orderFinish swallows ALL exceptions (including a
// failed downstream Result via getResultData) and always returns true, so the
// response is always ResultHelper.success(true).
func zhimaOrderFinish(c *gin.Context) {
	var req zhimaOrderFinishDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}

	cmdCtx := middleware.CompleteCommandContext(c, req.toShared())
	_, _ = rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/zhima/payafteruse/orderFinish", &req, cmdCtx)
	c.JSON(http.StatusOK, dto.NewSuccessResult(true))
}

// ---------------------------------------------------------------------------
// ZhimaPayAfterUseCallbackController (POST /callback/zhima/payafteruse/notify)
// ---------------------------------------------------------------------------

// zhimaCallbackCmd mirrors the downstream ebike-user ZhimaPayAfterUseCallbackCmd.
type zhimaCallbackCmd struct {
	CommandContext   *dto.CommandContext `json:"commandContext"`
	BizContentJson   *string             `json:"bizContentJson"`
	NotifyMethodName *string             `json:"notifyMethodName"`
}

const (
	zhimaNotifySuccess = "success"
	zhimaNotifyFail    = "fail"

	methodCreditAgreementChanged = "zhima.credit.payafteruse.creditagreement.changed"
	methodCreditBizOrderChanged  = "zhima.credit.payafteruse.creditbizorder.changed"
)

// zhimaNotify handles POST /callback/zhima/payafteruse/notify.
// The Java controller returns the plain strings "success"/"fail" (NOT a
// Result wrapper); any exception is swallowed and turned into "fail".
func zhimaNotify(c *gin.Context) {
	// Java binds a form ClientDTO (unused; a type-mismatch would raise a
	// BindException -> code "00004", exactly what BindForm reproduces).
	var bound clientDTO
	if !web.BindForm(c, &bound, nil) {
		return
	}

	// Java getParamMap: all request parameters (query string + form body).
	if c.Request.Form == nil {
		_ = c.Request.ParseForm()
	}
	params := map[string]string{}
	for k, vs := range c.Request.Form {
		if len(vs) > 0 {
			params[k] = vs[0]
		}
	}

	traceId := newUUID()
	cmdCtx := &dto.CommandContext{TraceId: traceId}

	notifyMethod, hasMethod := params["msg_method"]
	bizContent, hasBizContent := params["biz_content"]

	switch notifyMethod {
	case methodCreditAgreementChanged:
		cmdCtx.Pin = "zhima-pay-after-use-sign-notify"
		cmdCtx.TenantId = "0"
	case methodCreditBizOrderChanged:
		// Java: JSON.parseObject(biz_content).getString("out_order_no").split("@")
		// — any parse error / missing value / missing "@" part throws -> "fail".
		var biz map[string]interface{}
		if err := json.Unmarshal([]byte(bizContent), &biz); err != nil {
			c.String(http.StatusOK, zhimaNotifyFail)
			return
		}
		outOrderNo, _ := biz["out_order_no"].(string)
		if outOrderNo == "" {
			c.String(http.StatusOK, zhimaNotifyFail)
			return
		}
		// Java String.split drops trailing empty strings
		parts := strings.Split(outOrderNo, "@")
		for len(parts) > 0 && parts[len(parts)-1] == "" {
			parts = parts[:len(parts)-1]
		}
		if len(parts) < 2 {
			c.String(http.StatusOK, zhimaNotifyFail)
			return
		}
		cmdCtx.TenantId = parts[0]
		cmdCtx.Pin = parts[1]
	}

	// Java: AlipaySignature.rsaCheckV1(paramMap, publicKey, charset, "RSA2");
	// verification failure or any exception -> "fail".
	ok, err := verifyAlipaySign(params)
	if err != nil || !ok {
		c.String(http.StatusOK, zhimaNotifyFail)
		return
	}

	cmd := zhimaCallbackCmd{CommandContext: cmdCtx}
	if hasMethod {
		cmd.NotifyMethodName = &notifyMethod
	}
	if hasBizContent {
		cmd.BizContentJson = &bizContent
	}

	// Java ignores the downstream Result; only a thrown exception means "fail".
	if err := rpc.PostToService(c.Request.Context(), rpc.ServiceUser, "/callback/zhima/payafteruse/notify", &cmd, nil); err != nil {
		c.String(http.StatusOK, zhimaNotifyFail)
		return
	}
	c.String(http.StatusOK, zhimaNotifySuccess)
}

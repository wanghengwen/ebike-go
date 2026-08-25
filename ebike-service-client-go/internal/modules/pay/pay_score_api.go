package pay

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// PayScoreController (/client/pay/score/*) — downstream: ebike-user PayScoreApi
// ---------------------------------------------------------------------------

// payScorePermission handles POST /client/pay/score/permission.
// Java: no validation; PayScoreGatewayImpl sends genParam(PermissionRecordCmd.class)
// — context only, the request DTO is NOT copied downstream.
func payScorePermission(c *gin.Context) {
	var req clientDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}

	cmdCtx := middleware.CompleteCommandContext(c, req.toShared())
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/pay/score/permission", struct{}{}, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertString(result.Data)
	}
	web.RespondResult(c, result, err)
}

// getPermissionRecord handles POST /client/pay/score/getPermissionRecord.
// Java: @Valid PermissionRecordDTO (only inherited traceId/tenantId @NotEmpty).
func getPermissionRecord(c *gin.Context) {
	var req permissionRecordDTO
	if !web.BindJSON(c, &req, clientMsgs) {
		return
	}

	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/pay/score/getPermissionRecord", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertPermissionRecordCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

// terminatePermission handles POST /client/pay/score/terminatePermission.
// Java ignores the downstream Result (no getResultData) and always returns
// ResultHelper.success() — so even a failed downstream Result yields success.
// Only a transport-level error surfaces as an exception (00001).
func terminatePermission(c *gin.Context) {
	var req permissionRecordDTO
	if !web.BindJSON(c, &req, clientMsgs) {
		return
	}

	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	_, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/pay/score/terminatePermission", &req, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
}

// createPayScoreOrder handles POST /client/pay/score/createOrder.
// Java: no validation; plain passthrough via getResultData.
func createPayScoreOrder(c *gin.Context) {
	var req payScoreOrderDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}

	cmdCtx := middleware.CompleteCommandContext(c, req.toShared())
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/pay/score/createOrder", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertString(result.Data)
	}
	web.RespondResult(c, result, err)
}

// completePayScoreOrder handles POST /client/pay/score/completeOrder.
// Java ignores the downstream Result and returns ResultHelper.success().
func completePayScoreOrder(c *gin.Context) {
	var req payScoreOrderDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}

	cmdCtx := middleware.CompleteCommandContext(c, req.toShared())
	_, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/pay/score/completeOrder", &req, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
}

// cancelPayScoreOrder handles POST /client/pay/score/cancelOrder.
// Java ignores the downstream Result and returns ResultHelper.success().
func cancelPayScoreOrder(c *gin.Context) {
	var req payScoreOrderDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}

	cmdCtx := middleware.CompleteCommandContext(c, req.toShared())
	_, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/pay/score/cancelOrder", &req, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
}

// getPayScoreConfig handles POST /client/pay/score/getConfig.
// Java: no validation; genParam(PermissionRecordCmd.class) — context only.
func getPayScoreConfig(c *gin.Context) {
	var req clientDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}

	cmdCtx := middleware.CompleteCommandContext(c, req.toShared())
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/pay/score/getConfig", struct{}{}, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertWxScoreConfigCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

// ---------------------------------------------------------------------------
// PayScoreCallbackController (/callback/pay/score/{tenantId}/*)
// ---------------------------------------------------------------------------

// payScoreCallbackCmd mirrors the downstream ebike-user PayScoreCallbackCMD.
// Java's PayScoreGatewayImpl builds the CommandContext manually (tenantId from
// the path, a fresh UUID traceId and pin "支付分回调") — genParam is NOT used,
// so the auth/aspect context is irrelevant for the downstream body.
type payScoreCallbackCmd struct {
	CommandContext     *dto.CommandContext `json:"commandContext"`
	TenantId           string              `json:"tenantId"`
	WechatpaySerial    *string             `json:"wechatpaySerial"`
	WechatpayTimestamp *string             `json:"wechatpayTimestamp"`
	WechatpayNonce     *string             `json:"wechatpayNonce"`
	WechatpaySignature *string             `json:"wechatpaySignature"`
	Data               string              `json:"data"`
}

// headerOrNil mirrors Java request.getHeader: null when the header is absent.
func headerOrNil(c *gin.Context, name string) *string {
	v := c.GetHeader(name)
	if v == "" {
		return nil
	}
	return &v
}

// joinLines mirrors the Java BufferedReader readLine loop, which concatenates
// the body lines WITHOUT the line separators.
func joinLines(body []byte) string {
	s := strings.ReplaceAll(string(body), "\r", "")
	return strings.ReplaceAll(s, "\n", "")
}

// payScoreNotifyHandler builds the handler for the three WeChat pay-score
// callbacks (paySuccessNotify / permissionNotify / refundNotify). The Java
// controller returns the downstream PayScoreCallbackCO directly (NOT wrapped
// in Result); on a failed downstream Result, getResultData throws BizException
// and the global handler re-emits the downstream code/msg as a Result.
func payScoreNotifyHandler(downstreamPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantId := c.Param("tenantId")

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			web.WriteException(c, err.Error())
			return
		}

		// Java binds @RequestBody ClientDTO (re-readable thanks to
		// RequestBodyFilter); an unparseable body fails with
		// HttpMessageNotReadableException -> code "00002".
		var bound clientDTO
		if err := json.Unmarshal(body, &bound); err != nil {
			c.JSON(http.StatusOK, dto.NewErrorResult(dto.CodeParamException, "http message not readable"))
			return
		}

		cmd := payScoreCallbackCmd{
			CommandContext: &dto.CommandContext{
				TenantId: tenantId,
				TraceId:  newUUID(),
				Pin:      "支付分回调",
			},
			TenantId:           tenantId,
			WechatpaySerial:    headerOrNil(c, "Wechatpay-Serial"),
			WechatpayTimestamp: headerOrNil(c, "Wechatpay-Timestamp"),
			WechatpayNonce:     headerOrNil(c, "Wechatpay-Nonce"),
			WechatpaySignature: headerOrNil(c, "Wechatpay-Signature"),
			Data:               joinLines(body),
		}

		var result dto.Result
		if err := rpc.PostToService(c.Request.Context(), rpc.ServiceUser, downstreamPath, &cmd, &result); err != nil {
			web.WriteException(c, err.Error())
			return
		}

		if !result.Success {
			// Java: BizException(DefaultMsgCode(code, msg)) -> Result{code, msg, data:null}
			code, msg := "", ""
			if result.Code != nil {
				code = *result.Code
			}
			if result.Msg != nil {
				msg = *result.Msg
			}
			c.JSON(http.StatusOK, dto.NewErrorResult(code, msg))
			return
		}

		// Java returns the CO object directly; a null CO produces an empty body.
		data := bytes.TrimSpace(result.Data)
		if len(data) == 0 || string(data) == "null" {
			c.Status(http.StatusOK)
			return
		}
		c.Data(http.StatusOK, "application/json", data)
	}
}

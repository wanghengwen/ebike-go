package pay

import (
	"net/http"
	"strings"
	"time"

	"encoding/json"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/javacompat"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// payCmdForward mirrors the downstream ebike-pay PayCmd, which Java's
// PayGatewayImpl.create populates manually (genParam WITHOUT dto copy).
// Null Java fields are serialized as null (no NON_NULL), hence pointer types.
type payCmdForward struct {
	Pin             string  `json:"pin"`
	ServiceId       *int64  `json:"serviceId"`
	Channel         string  `json:"channel"`
	Source          int     `json:"source"`
	Description     string  `json:"description"` // Java field default "购买"
	Amount          int     `json:"amount"`
	SaleType        string  `json:"saleType"`
	Attach          *string `json:"attach"`
	PlatformAttach  *string `json:"platformAttach"`
	Openid          *string `json:"openid"`
	ActiveId        *int64  `json:"activeId"`
	DepositCardId   *int64  `json:"depositCardId"`
	FavorableCardId *int64  `json:"favorableCardId"`
	OrderId         *int64  `json:"orderId"`
	RidingCardId    *int64  `json:"ridingCardId"`
	OutTradeNo      *string `json:"outTradeNo"`
	WalletPwd       *string `json:"walletPwd"`
}

// platformCode mirrors Java PlatformEnum.of(platform).getCode():
// case-insensitive match, empty/unknown -> other(9).
func platformCode(platform string) int {
	switch strings.ToLower(platform) {
	case "wechat":
		return 0
	case "ios":
		return 1
	case "android":
		return 2
	case "pc":
		return 3
	default:
		return 9
	}
}

// createPay handles POST /client/ebike-pay/pay/create.
// Java: PayController.create -> PayServiceImpl.create -> PayGatewayImpl.create.
func createPay(c *gin.Context) {
	var req payDTO
	if !web.BindJSON(c, &req, payMsgs) {
		return
	}
	// @NotBlank also rejects whitespace-only values
	if !web.NotBlank(c, "channel_type", req.Channel_type) || !web.NotBlank(c, "sale_type", req.Sale_type) {
		return
	}

	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)

	cmd := payCmdForward{
		Pin:         cmdCtx.Pin,
		ServiceId:   req.Service_id,
		Channel:     req.Channel_type,
		Source:      platformCode(cmdCtx.Platform),
		Description: "购买",
		SaleType:    req.Sale_type,
		WalletPwd:   req.WalletPwd,
	}

	// Java: if (saleInfo != null && totalFee != null) { range check; amount=totalFee } else amount=0
	if req.Sale_info.Total_fee != nil {
		fee := *req.Sale_info.Total_fee
		if fee <= 0 || fee > 1000000 {
			// Java: throw BizException(ClientMsgCode.RECHARGE_PRICE_ERROR) -> code "50001"
			c.JSON(http.StatusOK, dto.NewErrorResult("50001", "充值金额错误"))
			return
		}
		cmd.Amount = fee
	}

	cmd.Openid = req.Channel_info.Open_id
	cmd.ActiveId = req.Sale_info.Active_id
	cmd.DepositCardId = req.Sale_info.Deposit_card_id
	cmd.FavorableCardId = req.Sale_info.Favorable_card_id
	cmd.RidingCardId = req.Sale_info.Riding_card_id
	cmd.OrderId = req.Sale_info.OrderId

	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServicePay, "/pay/create", &cmd, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertCreatePayCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

// accessTokenCmdForward mirrors ebike-pay AccessTokenCmd (code + type only).
// Java: genParam(AccessTokenCmd.class, dto) copies just those fields.
type accessTokenCmdForward struct {
	Code string `json:"code"`
	Type string `json:"type"`
}

// getOpenIdByJsCode handles POST /client/ebike-pay/pay/getOpenIdByJsCode.
// Java: PayGatewayImpl.getOpenIdByJsCode -> miniProgramApi.getAccessToken.
// NOTE: WeChat js_code is single-use; this endpoint must not be SHADOW-mirrored.
func getOpenIdByJsCode(c *gin.Context) {
	// Java field default: type = "WXLITE" (an explicit JSON null overwrites it)
	defaultType := "WXLITE"
	req := jsCodeDTO{Type: &defaultType}
	if !web.BindJSON(c, &req, jsCodeMsgs) {
		return
	}
	if !web.NotBlank(c, "code", req.Code) {
		return
	}

	typeVal := defaultType
	if req.Type != nil {
		typeVal = *req.Type
	}

	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	cmd := accessTokenCmdForward{Code: req.Code, Type: typeVal}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServicePay, "/miniProgram/getAccessToken", &cmd, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertString(result.Data)
	}
	web.RespondResult(c, result, err)
}

// cancelPay handles POST /client/ebike-pay/pay/cancel.
// Java: PayGatewayImpl.cancel sends genParam(BaseCmd.class) — context only,
// NO dto fields are copied into the downstream body.
func cancelPay(c *gin.Context) {
	var req baseDTO
	if !web.BindJSON(c, &req, clientMsgs) {
		return
	}

	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServicePay, "/pay/cancel", struct{}{}, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertBoolean(result.Data)
	}
	web.RespondResult(c, result, err)
}

// refundCmdForward mirrors the downstream ebike-pay RefundCmd: BeanUtils copies
// pin (same name), saleType/refundFee are set manually in Java (name mismatch
// with sale_type/refund_fee), and the Java field defaults paidAt=now()/
// izManualRefund=false/izWithdraw=false are serialized by the Feign client.
type refundCmdForward struct {
	Pin            string  `json:"pin"`
	SaleType       string  `json:"saleType"`
	TradeNo        *string `json:"tradeNo"`
	OutTradeNo     *string `json:"outTradeNo"`
	RefundFee      *int    `json:"refundFee"`
	PaidAt         string  `json:"paidAt"` // @JsonFormat "yyyy-MM-dd HH:mm:ss"
	IzManualRefund bool    `json:"izManualRefund"`
	IzWithdraw     bool    `json:"izWithdraw"`
}

// refundPay handles POST /client/ebike-pay/pay/refund.
func refundPay(c *gin.Context) {
	var req refundDTO
	if !web.BindJSON(c, &req, refundMsgs) {
		return
	}
	if !web.NotBlank(c, "pin", req.Pin) || !web.NotBlank(c, "sale_type", req.Sale_type) {
		return
	}

	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	cmd := refundCmdForward{
		Pin:       req.Pin,
		SaleType:  req.Sale_type,
		RefundFee: req.Refund_fee,
		PaidAt:    time.Now().Format("2006-01-02 15:04:05"),
	}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServicePay, "/pay/refund", &cmd, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertBoolean(result.Data)
	}
	web.RespondResult(c, result, err)
}

// withdraw handles POST /client/ebike-pay/pay/withdraw.
// Java: PayServiceImpl.withdraw has special logic for platform == "wechat".
func withdraw(c *gin.Context) {
	var req withdrawDTO
	if !web.BindJSON(c, &req, withdrawMsgs) {
		return
	}
	if !web.NotBlank(c, "channel", req.Channel) {
		return
	}

	// Java: if ("wechat".equals(dto.getPlatform())) { ... } (case-sensitive)
	if req.Platform == "wechat" {
		// Java: StringUtils.isBlank(openid) -> BizException(EXCEPTION, "openid不能为空")
		if req.Openid == nil || strings.TrimSpace(*req.Openid) == "" {
			web.WriteException(c, "openid不能为空")
			return
		}
		req.Channel = "WXLITE"
	}

	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServicePay, "/pay/withdrawal", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertBoolean(result.Data)
	}
	web.RespondResult(c, result, err)
}

// withdrawPage handles POST /client/ebike-pay/pay/withdraw/page.
// Java: PayGatewayImpl.withdrawPage = genParam(WithdrawQuery.class, dto) +
// setPin(context pin) -> payApi.withdrawalPage.
func withdrawPage(c *gin.Context) {
	// Java PageClientDTO field defaults pageNum=1, pageSize=10
	req := dto.PageClientDTO{PageNum: 1, PageSize: 10}
	if !web.BindJSON(c, &req, clientMsgs) {
		return
	}

	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	fwd := struct {
		dto.PageClientDTO
		Pin string `json:"pin"`
	}{req, cmdCtx.Pin}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServicePay, "/pay/withdrawal/page", &fwd, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertWithdrawalWaterPage(result.Data)
	}
	web.RespondResult(c, result, err)
}

type withdrawalWaterCO struct {
	Id           *javacompat.LongStr  `json:"id"`
	Amount       *int                 `json:"amount"`
	Present      *int                 `json:"present"`
	State        *int                 `json:"state"`
	Channel      *string              `json:"channel"`
	Openid       *string              `json:"openid"`
	OutTradeNo   *string              `json:"outTradeNo"`
	BatchId      *string              `json:"batchId"`
	Pin          *string              `json:"pin"`
	ServiceId    *javacompat.LongStr  `json:"serviceId"`
	Name         *string              `json:"name"`
	Phone        *string              `json:"phone"`
	InspectTime  *javacompat.DateTime `json:"inspectTime"`
	ErrorMessage *string              `json:"errorMessage"`
	WithdrawTime *javacompat.DateTime `json:"withdrawTime"`
	CreatedAt    *javacompat.DateTime `json:"createdAt"`
	UpdatedPin   *string              `json:"updatedPin"`
	UpdatedAt    *javacompat.DateTime `json:"updatedAt"`
}

type withdrawalPageCO struct {
	Count       *javacompat.LongStr `json:"count"`
	PageNum     *int                `json:"pageNum"`
	PageSize    *int                `json:"pageSize"`
	Orders      json.RawMessage     `json:"orders"`
	SearchCount *bool               `json:"searchCount"`
	List        []withdrawalWaterCO `json:"list"`
}

func convertWithdrawalWaterPage(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var page withdrawalPageCO
	if json.Unmarshal(raw, &page) != nil {
		return raw
	}
	b, _ := json.Marshal(page)
	return b
}

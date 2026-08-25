// Package pay ports the Java pay-related controllers:
// PayController, PayScoreController, PayScoreCallbackController,
// ZhimaPayAfterUseController, ZhimaPayAfterUseCallbackController,
// ZhimaPayAfterUseOrderController.
package pay

import (
	"encoding/json"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/pkg/javacompat"
)

// clientDTO is a NON-validated mirror of Java com.xyy.dto.ClientDTO, used for
// endpoints whose Java controller method has no @Validated/@Valid annotation
// (so the @NotEmpty annotations on traceId/tenantId are NOT enforced).
// The shared dto.ClientDTO carries binding:"required" tags and cannot be used
// for those endpoints.
type clientDTO struct {
	TraceId       string   `json:"traceId" form:"traceId"`
	TenantId      string   `json:"tenantId" form:"tenantId"`
	Platform      string   `json:"platform,omitempty" form:"platform"`
	DeviceId      string   `json:"deviceId,omitempty" form:"deviceId"`
	Version       string   `json:"version,omitempty" form:"version"`
	Ip            string   `json:"ip,omitempty" form:"ip"`
	Longitude     *float64 `json:"longitude,omitempty" form:"longitude"`
	Latitude      *float64 `json:"latitude,omitempty" form:"latitude"`
	Source        string   `json:"source,omitempty" form:"source"`
	StressTesting bool     `json:"stressTesting" form:"stressTesting"`
}

// toShared converts to the shared dto.ClientDTO so that
// middleware.CompleteCommandContext can be reused.
func (d *clientDTO) toShared() *dto.ClientDTO {
	return &dto.ClientDTO{
		TraceId:       d.TraceId,
		TenantId:      d.TenantId,
		Platform:      d.Platform,
		DeviceId:      d.DeviceId,
		Version:       d.Version,
		Ip:            d.Ip,
		Longitude:     d.Longitude,
		Latitude:      d.Latitude,
		Source:        d.Source,
		StressTesting: d.StressTesting,
	}
}

// clientMsgs covers the @NotEmpty annotations on Java ClientDTO.
var clientMsgs = map[string]string{
	"traceId":  "must not be empty",
	"tenantId": "must not be empty",
}

// ---------------------------------------------------------------------------
// PayController DTOs (Java: client.dto.pay.*, all endpoints @Validated)
// ---------------------------------------------------------------------------

// channelInfo mirrors Java ChannelInfo. Its @NotBlank/@NotNull annotations are
// NOT enforced because PayDTO.channel_info lacks @Valid (no cascade in Java).
type channelInfo struct {
	Open_id    *string `json:"open_id"`
	Service_id *int64  `json:"service_id"`
}

// saleInfo mirrors Java SaleInfo (constraints not cascaded, see channelInfo).
type saleInfo struct {
	Total_fee         *int   `json:"total_fee"`
	Active_id         *int64 `json:"active_id"`
	Deposit_card_id   *int64 `json:"deposit_card_id"`
	Favorable_card_id *int64 `json:"favorable_card_id"`
	Goods_id          *int64 `json:"goods_id"`
	Riding_card_id    *int64 `json:"riding_card_id"`
	OrderId           *int64 `json:"orderId"`
}

// payDTO mirrors Java PayDTO extends ClientDTO. Go field names intentionally
// keep the Java snake_case property names (Service_id...) so the validation
// error message field matches Java byte-for-byte (e.g. "service_id must not be null").
type payDTO struct {
	dto.ClientDTO
	Service_id   *int64       `json:"service_id" binding:"required"`   // @NotNull
	Channel_type string       `json:"channel_type" binding:"required"` // @NotBlank
	Channel_info *channelInfo `json:"channel_info" binding:"required"` // @NotNull
	Sale_type    string       `json:"sale_type" binding:"required"`    // @NotBlank
	Sale_info    *saleInfo    `json:"sale_info" binding:"required"`    // @NotNull
	WalletPwd    *string      `json:"walletPwd"`
}

var payMsgs = map[string]string{
	"traceId":      "must not be empty",
	"tenantId":     "must not be empty",
	"service_id":   "must not be null",
	"channel_type": "must not be blank",
	"channel_info": "must not be null",
	"sale_type":    "must not be blank",
	"sale_info":    "must not be null",
}

// jsCodeDTO mirrors Java JsCodeDTO (type has Java field default "WXLITE";
// pointer keeps Jackson's null-overwrite semantics when the client sends null).
type jsCodeDTO struct {
	dto.ClientDTO
	Code string  `json:"code" binding:"required"` // @NotBlank
	Type *string `json:"type"`
}

var jsCodeMsgs = map[string]string{
	"traceId":  "must not be empty",
	"tenantId": "must not be empty",
	"code":     "must not be blank",
}

// baseDTO mirrors Java fence.BaseDTO extends ClientDTO (no extra fields).
type baseDTO struct {
	dto.ClientDTO
}

// refundDTO mirrors Java RefundDTO.
type refundDTO struct {
	dto.ClientDTO
	Pin        string `json:"pin" binding:"required"`        // @NotBlank
	Sale_type  string `json:"sale_type" binding:"required"`  // @NotBlank
	Refund_fee *int   `json:"refund_fee" binding:"required"` // @NotNull
}

var refundMsgs = map[string]string{
	"traceId":    "must not be empty",
	"tenantId":   "must not be empty",
	"pin":        "must not be blank",
	"sale_type":  "must not be blank",
	"refund_fee": "must not be null",
}

// withdrawDTO mirrors Java WithdrawDTO.
type withdrawDTO struct {
	dto.ClientDTO
	Openid  *string `json:"openid"`
	Channel string  `json:"channel" binding:"required"` // @NotBlank
}

var withdrawMsgs = map[string]string{
	"traceId":  "must not be empty",
	"tenantId": "must not be empty",
	"channel":  "must not be blank",
}

// ---------------------------------------------------------------------------
// PayScoreController DTOs (Java: client.dto.user.*)
// ---------------------------------------------------------------------------

// permissionRecordDTO mirrors Java PermissionRecordDTO (used with @Valid:
// only the inherited ClientDTO @NotEmpty constraints apply).
type permissionRecordDTO struct {
	dto.ClientDTO
	Openid *string `json:"openid"`
	Reason *string `json:"reason"`
	Code   *string `json:"code"`
}

// payScoreOrderDTO mirrors Java PayScoreOrderDTO. All endpoints using it have
// NO validation annotation, hence the non-validated embedded clientDTO.
type payScoreOrderDTO struct {
	clientDTO
	Openid      *string `json:"openid"`
	TradeNo     *string `json:"tradeNo"`
	Amount      *int    `json:"amount"`
	Reason      *string `json:"reason"`
	Code        *string `json:"code"`
	Description *string `json:"description"`
}

// ---------------------------------------------------------------------------
// Zhima (ZhimaPayAfterUse*) DTOs
// ---------------------------------------------------------------------------

// zhimaSignDTO mirrors Java ZhimaPayAfterUseSignDto (@Validated, ClientDTO only).
type zhimaSignDTO struct {
	dto.ClientDTO
}

// zhimaSignQueryDTO mirrors Java ZhimaPayAfterUseSignQueryDto (@Validated).
// NOTE: the downstream Cmd field is outAgreementNo, which Java BeanUtils never
// populates from creditAgreementId (name mismatch) — same effective behavior here.
type zhimaSignQueryDTO struct {
	dto.ClientDTO
	CreditAgreementId *string `json:"creditAgreementId"`
}

// zhimaOrderWithoutConfirmDTO mirrors Java ZhimaPayAfterUseOrderWithoutConfirmDto.
// The controller method has NO @Validated, so @NotBlank/@Max are NOT enforced.
// orderAmount has the Java field default 20000 (preset before binding).
type zhimaOrderWithoutConfirmDTO struct {
	clientDTO
	OutOrderNo        *string `json:"outOrderNo"`
	CreditAgreementId *string `json:"creditAgreementId"`
	OrderAmount       *int    `json:"orderAmount"`
	Subject           *string `json:"subject"`
	OrderId           *int64  `json:"orderId"`
}

// zhimaOrderQueryDTO mirrors Java ZhimaPayAfterUseOrderQueryDto (NOT validated:
// no @Valid on the controller arg despite the @NotEmpty annotations).
type zhimaOrderQueryDTO struct {
	clientDTO
	Sign       *string `json:"sign"`
	OutOrderNo *string `json:"outOrderNo"`
}

// zhimaOrderFinishDTO mirrors Java ZhimaPayAfterUseOrderFinishDto (NOT validated).
type zhimaOrderFinishDTO struct {
	clientDTO
	OutOrderNo       *string `json:"outOrderNo"`
	CreditBizOrderId *string `json:"creditBizOrderId"`
	IzFulfilled      *bool   `json:"izFulfilled"`
}

// ---------------------------------------------------------------------------
// Response DTOs and converters
// ---------------------------------------------------------------------------

func convertBoolean(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var b bool
	if json.Unmarshal(raw, &b) != nil {
		return raw
	}
	out, _ := json.Marshal(b)
	return out
}

func convertString(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var s string
	if json.Unmarshal(raw, &s) != nil {
		return raw
	}
	out, _ := json.Marshal(s)
	return out
}

type createPayCO struct {
	TimeStamp  *string             `json:"timeStamp"`
	NonceStr   *string             `json:"nonceStr"`
	Packages   *string             `json:"packages"`
	SignType   *string             `json:"signType"`
	PaySign    *string             `json:"paySign"`
	WaterId    *javacompat.LongStr `json:"waterId"`
	OutTradeNo *string             `json:"outTradeNo"`
	TradeNo    *string             `json:"tradeNo"`
}

func convertCreatePayCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co createPayCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	out, _ := json.Marshal(co)
	return out
}

type permissionRecordCO struct {
	Openid                   *string `json:"openid"`
	Authorization_code       *string `json:"authorization_code"`
	Authorization_state      *string `json:"authorization_state"`
	Cancel_authorization_time  *string `json:"cancel_authorization_time"`
	Authorization_success_time *string `json:"authorization_success_time"`
}

func convertPermissionRecordCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co permissionRecordCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	out, _ := json.Marshal(co)
	return out
}

type wxScoreConfigCO struct {
	Id                  *javacompat.LongStr `json:"id"`
	Mchid               *string             `json:"mchid"`
	Appid               *string             `json:"appid"`
	AppV3Key            *string             `json:"appV3Key"`
	ServiceId           *string             `json:"serviceId"`
	PaymentsName        *string             `json:"paymentsName"`
	ServiceIntroduction *string             `json:"serviceIntroduction"`
	SerialNo            *string             `json:"serialNo"`
	PrivateKey          *string             `json:"privateKey"`
}

func convertWxScoreConfigCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co wxScoreConfigCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	out, _ := json.Marshal(co)
	return out
}

type zhimaPayAfterUseSignCO struct {
	Sign        *string `json:"sign"`
	ZmServiceId *string `json:"zmServiceId"`
}

func convertZhimaPayAfterUseSignCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co zhimaPayAfterUseSignCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	out, _ := json.Marshal(co)
	return out
}

type zhimaPayAfterUseSignRecordCO struct {
	Id                *javacompat.LongStr  `json:"id"`
	Name              *string              `json:"name"`
	Phone             *string              `json:"phone"`
	OutAgreementNo    *string              `json:"outAgreementNo"`
	CreditAgreementId *string              `json:"creditAgreementId"`
	AgreementStatus   *int                 `json:"agreementStatus"`
	BizTime           *javacompat.DateTime `json:"bizTime"`
	TenantIdList      []string             `json:"tenantIdList"`
}

func convertZhimaPayAfterUseSignRecordCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co zhimaPayAfterUseSignRecordCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	out, _ := json.Marshal(co)
	return out
}

type zhimaPayAfterUseOrderWithoutConfirmCO struct {
	CreditBizOrderId *string `json:"creditBizOrderId"`
	OutOrderNo       *string `json:"outOrderNo"`
}

func convertZhimaPayAfterUseOrderWithoutConfirmCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co zhimaPayAfterUseOrderWithoutConfirmCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	out, _ := json.Marshal(co)
	return out
}

type zhimaPayAfterUseOrderQueryCO struct {
	CreditBizOrderId  *string              `json:"creditBizOrderId"`
	CreditAgreementId *string              `json:"creditAgreementId"`
	TotalAmount       *string              `json:"totalAmount"`
	CreateTime        *javacompat.DateTime `json:"createTime"`
	ZmServiceId       *string              `json:"zmServiceId"`
	OrderStatus       *string              `json:"orderStatus"`
	TradeNo           *string              `json:"tradeNo"`
}

func convertZhimaPayAfterUseOrderQueryCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co zhimaPayAfterUseOrderQueryCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	out, _ := json.Marshal(co)
	return out
}

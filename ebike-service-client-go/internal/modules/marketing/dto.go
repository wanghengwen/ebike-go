package marketing

import (
	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/pkg/javacompat"
)

// ---------------------------------------------------------------------------
// Validation message maps (JSON field name -> hibernate-validator English
// default message of the Java annotation on that field).
// ---------------------------------------------------------------------------

// serviceMsgs covers ServiceDTO: traceId/tenantId @NotEmpty (ClientDTO),
// serviceId @NotNull.
var serviceMsgs = map[string]string{
	"traceId":   "must not be empty",
	"tenantId":  "must not be empty",
	"serviceId": "must not be null",
}

// baseMsgs covers BaseDTO (only the inherited ClientDTO annotations).
var baseMsgs = map[string]string{
	"traceId":  "must not be empty",
	"tenantId": "must not be empty",
}

// idMsgs covers IdDTO: id @NotNull.
var idMsgs = map[string]string{
	"traceId":  "must not be empty",
	"tenantId": "must not be empty",
	"id":       "must not be null",
}

// inviteAcceptMsgs covers InviteAcceptDTO: serviceId @NotNull,
// inviteePin/inviterPin @NotBlank.
var inviteAcceptMsgs = map[string]string{
	"traceId":    "must not be empty",
	"tenantId":   "must not be empty",
	"serviceId":  "must not be null",
	"inviteePin": "must not be blank",
	"inviterPin": "must not be blank",
}

// adRewardMsgs covers AdRewardDTO: serviceId/buyTime @NotNull.
var adRewardMsgs = map[string]string{
	"traceId":   "must not be empty",
	"tenantId":  "must not be empty",
	"serviceId": "must not be null",
	"buyTime":   "must not be null",
}

// voucherMsgs covers UserAddVoucherDTO: serviceId/voucherCode @NotNull.
var voucherMsgs = map[string]string{
	"traceId":     "must not be empty",
	"tenantId":    "must not be empty",
	"serviceId":   "must not be null",
	"voucherCode": "must not be null",
}

// ---------------------------------------------------------------------------
// Inbound request DTOs (matching the Java client DTO classes).
// dto.ServiceDTO (shared) is used directly for the @Validated ServiceDTO endpoints.
// ---------------------------------------------------------------------------

// inviteAcceptDTO matches Java InviteAcceptDTO extends ClientDTO.
type inviteAcceptDTO struct {
	dto.ClientDTO
	ServiceId  *int64 `json:"serviceId" binding:"required"`  // @NotNull
	InviteePin string `json:"inviteePin" binding:"required"` // @NotBlank
	InviterPin string `json:"inviterPin" binding:"required"` // @NotBlank
}

// adRewardDTO matches Java AdRewardDTO extends ClientDTO.
type adRewardDTO struct {
	dto.ClientDTO
	ServiceId *int64         `json:"serviceId" binding:"required"` // @NotNull
	BuyTime   *localDateTime `json:"buyTime" binding:"required"`   // @NotNull + @JsonFormat
}

// userAddVoucherDTO matches Java UserAddVoucherDTO extends ClientDTO.
type userAddVoucherDTO struct {
	dto.ClientDTO
	ServiceId   *int64  `json:"serviceId" binding:"required"`   // @NotNull
	VoucherCode *string `json:"voucherCode" binding:"required"` // @NotNull (String: null rejected, "" allowed in Java; pointer keeps "" passing)
}

// idDTO matches Java IdDTO extends ClientDTO (dto.fence.IdDTO).
type idDTO struct {
	dto.ClientDTO
	Id *int64 `json:"id" binding:"required"` // @NotNull
}

// userAccountDTO matches Java UserAccountDTO extends ClientDTO
// (serviceId has NO validation annotation).
type userAccountDTO struct {
	dto.ClientDTO
	ServiceId *int64 `json:"serviceId"`
}

// serviceDTONoValid matches Java ServiceDTO for InviteController.create, whose
// parameter has @RequestBody but NO @Validated -> nothing is validated.
// It cannot embed dto.ClientDTO because that would activate the required tags.
type serviceDTONoValid struct {
	TraceId       string   `json:"traceId"`
	TenantId      string   `json:"tenantId"`
	Platform      string   `json:"platform"`
	DeviceId      string   `json:"deviceId"`
	Version       string   `json:"version"`
	Ip            string   `json:"ip"`
	Longitude     *float64 `json:"longitude"`
	Latitude      *float64 `json:"latitude"`
	Source        string   `json:"source"`
	StressTesting bool     `json:"stressTesting"`
	ServiceId     *int64   `json:"serviceId"`
}

// clientDTO converts the unvalidated body into a dto.ClientDTO for
// middleware.CompleteCommandContext.
func (s *serviceDTONoValid) clientDTO() dto.ClientDTO {
	return dto.ClientDTO{
		TraceId:       s.TraceId,
		TenantId:      s.TenantId,
		Platform:      s.Platform,
		DeviceId:      s.DeviceId,
		Version:       s.Version,
		Ip:            s.Ip,
		Longitude:     s.Longitude,
		Latitude:      s.Latitude,
		Source:        s.Source,
		StressTesting: s.StressTesting,
	}
}

// ---------------------------------------------------------------------------
// Downstream command bodies. Each mirrors the exact field set of the Java Cmd
// class that the GatewayImpl builds (BeanUtils same-name copy + manual sets);
// rpc.ForwardCommand appends the commandContext field. Pointer fields without
// omitempty reproduce Jackson's null emission (NON_NULL is not enabled).
// ---------------------------------------------------------------------------

// idCmd matches com.xyy.ebike.marketing.api.dto.IdCmd.
type idCmd struct {
	Id *int64 `json:"id"`
}

// serviceCmd matches com.xyy.ebike.marketing.api.dto.ServiceCmd.
type serviceCmd struct {
	ServiceId *int64 `json:"serviceId"`
}

// userServiceCmd matches com.xyy.ebike.marketing.api.dto.invite.cmd.UserServiceCmd
// and com.xyy.ebike.marketing.api.dto.regularactivity.cmd.NotifyCmd
// (both declare exactly serviceId + pin).
type userServiceCmd struct {
	ServiceId *int64 `json:"serviceId"`
	Pin       string `json:"pin"`
}

// acceptInviteCmd matches com.xyy.ebike.marketing.api.dto.invite.cmd.AcceptInviteCmd.
type acceptInviteCmd struct {
	ServiceId *int64 `json:"serviceId"`
	InviteeId string `json:"inviteeId"`
	InviterId string `json:"inviterId"`
}

// rechargeCmd matches com.xyy.ebike.marketing.api.dto.recharge.cmd.RechargeCmd.
type rechargeCmd struct {
	ServiceId    *int64 `json:"serviceId"`
	State        *int   `json:"state"`
	ActivityType *int   `json:"activityType"` // never set by the gateway -> null
}

// userWalletJudgementCmd matches
// com.xyy.ebike.marketing.api.dto.recharge.cmd.UserWalletJudgementCmd.
// JudgementGatewayImpl never sets pin (AdRewardDTO has no pin field), so the
// Java body carries "pin": null.
type userWalletJudgementCmd struct {
	ServiceId *int64         `json:"serviceId"`
	Pin       *string        `json:"pin"`
	BuyTime   *localDateTime `json:"buyTime"`
}

// userAddVoucherCmd matches com.xyy.ebike.marketing.api.dto.voucher.cmd.UserAddVoucherCmd.
type userAddVoucherCmd struct {
	ServiceId   *int64  `json:"serviceId"`
	VoucherCode *string `json:"voucherCode"`
}

// accountUserServiceCmd matches com.xyy.ebike.account.api.cmd.UserServiceCmd.
type accountUserServiceCmd struct {
	Pin       string `json:"pin"`
	ServiceId *int64 `json:"serviceId"`
}

// userBaseCmd matches com.xyy.ebike.account.api.cmd.UserBaseCmd
// (Java declares `private Boolean izUserService = false;`, so the field is
// always emitted, never null).
type userBaseCmd struct {
	Pin           string `json:"pin"`
	IzUserService bool   `json:"izUserService"`
}

// ---------------------------------------------------------------------------
// Locally constructed response objects.
// ---------------------------------------------------------------------------

// rechargeCO matches com.xyy.ebike.marketing.api.dto.recharge.co.RechargeCO for
// the fallback template list built by RechargeConfigGatewayImpl.list when the
// downstream returns an empty list. Long fields (id/ridingCardId/serviceId)
// serialize as strings per the Java global ToStringSerializer; LocalDateTime
// fields would serialize as "yyyy-MM-dd HH:mm:ss" strings (always null here).
// Field order matches the Java class declaration order.
type rechargeCO struct {
	Id               *javacompat.LongStr  `json:"id"`
	State            *int                 `json:"state"`
	ActivityType     *int                 `json:"activityType"`
	CreatedAt        *javacompat.DateTime `json:"createdAt"`
	Amount           *int                 `json:"amount"`
	ActiveGiveAmount *int                 `json:"activeGiveAmount"`
	RidingCardId     *javacompat.LongStr  `json:"ridingCardId"`
	Details          *string              `json:"details"`
	UpdatedAt        *javacompat.DateTime `json:"updatedAt"`
	UpdatedName      *string              `json:"updatedName"`
	ServiceId        *javacompat.LongStr  `json:"serviceId"`
}

// userWalletDo matches the Java dataobject UserWalletDo (gateway/account/dataobject).
// All fields are Integer/String in Java (no Long), so they serialize as plain
// JSON numbers/strings; pointers without omitempty emit nulls like Jackson.
type userWalletDo struct {
	Pin            *string `json:"pin"`
	Balance        *int64  `json:"balance"`
	Recharge       *int64  `json:"recharge"`
	Present        *int64  `json:"present"`
	DepositedMount *int64  `json:"depositedMount"`
	DepositedStats *int64  `json:"depositedStats"`
}

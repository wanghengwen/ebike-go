package marketing

import (
	"encoding/json"
	"net/http"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/javacompat"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// inviteDetail ports InviteController.detail.
// Java gateway: UserServiceCmd = genParam(UserServiceCmd.class, dto)
// (copies serviceId) + pin = CommandContextHolder.getPin()
// -> inviteApi.clientInviteDetail.
func inviteDetail(c *gin.Context) {
	var req dto.ServiceDTO
	if !web.BindJSON(c, &req, serviceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := userServiceCmd{ServiceId: req.ServiceId, Pin: cmdCtx.Pin}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/invite/client_invite_detail", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertClientInviteDetailCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

// inviteRecord ports InviteController.record -> inviteApi.inviteRecord.
func inviteRecord(c *gin.Context) {
	var req dto.ServiceDTO
	if !web.BindJSON(c, &req, serviceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := userServiceCmd{ServiceId: req.ServiceId, Pin: cmdCtx.Pin}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/invite/invite_record", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertUserInviteCOList(result.Data)
	}
	web.RespondResult(c, result, err)
}

// inviteCreate ports InviteController.create.
// NOTE: the Java parameter is @RequestBody WITHOUT @Validated, so no bean
// validation runs (traceId/tenantId/serviceId may all be absent).
// Gateway: UserServiceCmd(serviceId copied, pin from context) -> inviteApi.createInvite.
func inviteCreate(c *gin.Context) {
	var req serviceDTONoValid
	if !web.BindJSON(c, &req, nil) {
		return
	}
	client := req.clientDTO()
	cmdCtx := middleware.CompleteCommandContext(c, &client)
	body := userServiceCmd{ServiceId: req.ServiceId, Pin: cmdCtx.Pin}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/invite/create_invite", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertLong(result.Data)
	}
	web.RespondResult(c, result, err)
}

// inviteAccept ports InviteController.accept.
// Java gateway: AcceptInviteCmd = genParam(AcceptInviteCmd.class, dto)
// (copies serviceId), inviteeId = dto.inviteePin, inviterId = dto.inviterPin;
// the downstream data is DISCARDED and the gateway returns `true` after
// getResultData (which throws on a failed Result). So on downstream success the
// response is always Result.success(true); on failure the downstream code/msg
// pass through.
func inviteAccept(c *gin.Context) {
	var req inviteAcceptDTO
	if !web.BindJSON(c, &req, inviteAcceptMsgs) {
		return
	}
	// Java @NotBlank also rejects whitespace-only values
	if !web.NotBlank(c, "inviteePin", req.InviteePin) || !web.NotBlank(c, "inviterPin", req.InviterPin) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := acceptInviteCmd{ServiceId: req.ServiceId, InviteeId: req.InviteePin, InviterId: req.InviterPin}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/invite/accept_invite", &body, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !result.Success {
		web.WriteBizResult(c, result)
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(true))
}

// inviteRule ports InviteController.rule.
// Java gateway: IdCmd = genParam(IdCmd.class) + id = dto.serviceId
// -> inviteApi.inviteRuleConfig.
func inviteRule(c *gin.Context) {
	var req dto.ServiceDTO
	if !web.BindJSON(c, &req, serviceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := idCmd{Id: req.ServiceId}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/invite/invite_rule_config", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertInviteRuleConfigCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

type clientInviteDetailCO struct {
	InviteNum     *int            `json:"inviteNum"`
	RewardNum     *int            `json:"rewardNum"`
	InviteeRecord json.RawMessage `json:"inviteeRecord"`
	RewardPool    *int            `json:"rewardPool"`
	ValidDay      *int            `json:"validDay"`
	Unit          *string         `json:"unit"`
	RewardType    *int            `json:"rewardType"`
	RewardInfo    *string         `json:"rewardInfo"`
}

type userInviteCO struct {
	Id            *javacompat.LongStr  `json:"id"`
	InviteId      *javacompat.LongStr  `json:"inviteId"`
	ParentId      *javacompat.LongStr  `json:"parentId"`
	ServiceId     *javacompat.LongStr  `json:"serviceId"`
	Pin           *string              `json:"pin"`
	PinName       *string              `json:"pinName"`
	PinPhone      *string              `json:"pinPhone"`
	Inviter       *bool                `json:"inviter"`
	Valid         *bool                `json:"valid"`
	RewardType    *int                 `json:"rewardType"`
	RewardInfo    *string              `json:"rewardInfo"`
	RewardCount   *int                 `json:"rewardCount"`
	RewardUnit    *int                 `json:"rewardUnit"`
	ShowType      *string              `json:"showType"`
	RewardTime    *javacompat.DateTime `json:"rewardTime"`
	RegisterTime  *javacompat.DateTime `json:"registerTime"`
	InviteeNum    *int                 `json:"inviteeNum"`
	CompleteCount *int                 `json:"completeCount"`
	Register      *bool                `json:"register"`
	RidingcardId  *javacompat.LongStr  `json:"ridingcardId"`
	State         *bool                `json:"state"`
	BeginTime     *javacompat.DateTime `json:"beginTime"`
	EndTime       *javacompat.DateTime `json:"endTime"`
}

type inviteRuleConfigCO struct {
	RewardType    *int                 `json:"rewardType"`
	BeginTime     *javacompat.DateTime `json:"beginTime"`
	EndTime       *javacompat.DateTime `json:"endTime"`
	ValidDay      *int                 `json:"validDay"`
	RidingCardNum *int                 `json:"ridingCardNum"`
	RewardMax     *int                 `json:"rewardMax"`
	InviteNum     *int                 `json:"inviteNum"`
	Amount        *int                 `json:"amount"`
}

func convertClientInviteDetailCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co clientInviteDetailCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

func convertUserInviteCOList(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var list []userInviteCO
	if json.Unmarshal(raw, &list) != nil {
		return raw
	}
	b, _ := json.Marshal(list)
	return b
}

func convertLong(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var v javacompat.LongStr
	if json.Unmarshal(raw, &v) != nil {
		return raw
	}
	b, _ := json.Marshal(v)
	return b
}

func convertInviteRuleConfigCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co inviteRuleConfigCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

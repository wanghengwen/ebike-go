package marketing

import (
	"encoding/json"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/javacompat"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// userRewardNotify ports UserRewardController.rewardNotify.
// Java gateway: UserServiceCmd = genParam(UserServiceCmd.class, dto)
// (copies serviceId) + pin = CommandContextHolder.getPin()
// -> userRewardsApi.notify (@PostMapping("notify") -> "/notify").
func userRewardNotify(c *gin.Context) {
	var req dto.ServiceDTO
	if !web.BindJSON(c, &req, serviceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := userServiceCmd{ServiceId: req.ServiceId, Pin: cmdCtx.Pin}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/notify", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertUserRewardNotifyCOList(result.Data)
	}
	web.RespondResult(c, result, err)
}

// regularGetRegisterReward ports UserRewardController.regularGetRegisterReward.
// Java gateway: IdCmd = genParam(IdCmd.class) + id = dto.serviceId
// -> regularActivityApi.getRegisterReward.
func regularGetRegisterReward(c *gin.Context) {
	var req dto.ServiceDTO
	if !web.BindJSON(c, &req, serviceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := idCmd{Id: req.ServiceId}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/regularActivity/getRegisterReward", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertBoolean(result.Data)
	}
	web.RespondResult(c, result, err)
}

// regularGetVerifyReward ports UserRewardController.regularGetVerifyReward
// -> regularActivityApi.getVerifyReward.
func regularGetVerifyReward(c *gin.Context) {
	var req dto.ServiceDTO
	if !web.BindJSON(c, &req, serviceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := idCmd{Id: req.ServiceId}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/regularActivity/getVerifyReward", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertGetVerifyRewardCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

// userWatchAdJudgement ports UserRewardController.adReward.
// Java gateway: UserWalletJudgementCmd = genParam(UserWalletJudgementCmd.class, dto)
// (copies serviceId + buyTime; pin is NEVER set -> "pin": null downstream)
// -> judgementApi.watchAdJudgement.
func userWatchAdJudgement(c *gin.Context) {
	var req adRewardDTO
	if !web.BindJSON(c, &req, adRewardMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := userWalletJudgementCmd{ServiceId: req.ServiceId, Pin: nil, BuyTime: req.BuyTime}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/judgement/watch_ad_judgement", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertBoolean(result.Data)
	}
	web.RespondResult(c, result, err)
}

// verifyNotify ports UserRewardController.verifyNotify.
// Java gateway (RegularActivityGatewayImpl): NotifyCmd = genParam(NotifyCmd.class, dto)
// (copies serviceId) + pin = CommandContextHolder.getPin()
// -> regularActivityApi.notify.
func verifyNotify(c *gin.Context) {
	var req dto.ServiceDTO
	if !web.BindJSON(c, &req, serviceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := userServiceCmd{ServiceId: req.ServiceId, Pin: cmdCtx.Pin}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/regularActivity/notify", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertBoolean(result.Data)
	}
	web.RespondResult(c, result, err)
}

type userRewardNotifyCO struct {
	RewardType *int     `json:"rewardType"`
	RewardShow []string `json:"rewardShow"`
}

type getVerifyRewardCO struct {
	NeedVerify *bool `json:"needVerify"`
}

func convertUserRewardNotifyCOList(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var list []userRewardNotifyCO
	if json.Unmarshal(raw, &list) != nil {
		return raw
	}
	b, _ := json.Marshal(list)
	return b
}

func convertGetVerifyRewardCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co getVerifyRewardCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

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

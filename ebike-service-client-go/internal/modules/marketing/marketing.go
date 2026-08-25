// Package marketing ports the Java marketing controllers
// (ActivityCenterController / CardCenterController / InviteController /
// RechargeConfigController / RedPaketCarController / UserRewardController /
// VoucherController) and AccountController to Go.
package marketing

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires all marketing + account module routes.
func RegisterRoutes(r *gin.RouterGroup) {
	// Java: @RequestMapping("/client/ebike-marketing/activity_center")
	activityCenter := r.Group("/client/ebike-marketing/activity_center")
	{
		activityCenter.POST("/default_activity_center_list", defaultActivityCenterList)
	}

	// Java: @RequestMapping("/client/ebike-marketing/card")
	card := r.Group("/client/ebike-marketing/card")
	{
		card.POST("/riding_config_list", ridingConfigList)
		card.POST("/riding_config_get_rule", ridingConfigGetRule)
	}

	// Java: @RequestMapping("/client/ebike-marketing/invite")
	invite := r.Group("/client/ebike-marketing/invite")
	{
		invite.POST("/detail", inviteDetail)
		invite.POST("/record", inviteRecord)
		invite.POST("/create", inviteCreate)
		invite.POST("/accept", inviteAccept)
		invite.POST("/rule", inviteRule)
	}

	// Java: @RequestMapping("/client/ebike-marketing/recharge_config")
	rechargeConfig := r.Group("/client/ebike-marketing/recharge_config")
	{
		rechargeConfig.POST("/list", rechargeList)
		rechargeConfig.POST("/recharge_scope", rechargeScope)
		rechargeConfig.POST("/get_recharge_config", getRechargeConfig)
	}

	// Java: @RequestMapping("/client/ebike-marketing/redPaketCar")
	redPaketCar := r.Group("/client/ebike-marketing/redPaketCar")
	{
		redPaketCar.POST("/getRule", redPaketCarGetRule)
	}

	// Java: @RequestMapping("/client/ebike-marketing/user_reward")
	userReward := r.Group("/client/ebike-marketing/user_reward")
	{
		userReward.POST("/user_reward_notify", userRewardNotify)
		userReward.POST("/regular_get_register_reward", regularGetRegisterReward)
		userReward.POST("/regular_get_verify_reward", regularGetVerifyReward)
		userReward.POST("/user_watch_ad_judgement", userWatchAdJudgement)
		userReward.POST("/verify_notify", verifyNotify)
	}

	// Java: @RequestMapping("/client/ebike-marketing/voucher")
	voucher := r.Group("/client/ebike-marketing/voucher")
	{
		voucher.POST("/user_add_voucher", userAddVoucher)
	}

	// Java: @RequestMapping("/client/ebike-account")
	account := r.Group("/client/ebike-account")
	{
		account.POST("/riding_card/get_service_riding_card", getServiceRidingCard)
		account.POST("/user_account", getUserAccount)
		account.POST("/riding_card/get_riding_card", getRidingCard)
		account.POST("/favorable_card/get_user_favorable_card", getUserFavorableCard)
		account.POST("/free_order/get_user_all_free_order", getUserFreeCard)
		account.POST("/discount/get_user_all_discount", getUserAllDiscount)
		account.POST("/deposit_card/get_user_deposit_card", getUserDepositCard)
		account.POST("/wallet/get_wallet_info", getWalletInfo)
	}
}

// localDateTimeLayout mirrors Java @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss").
const localDateTimeLayout = "2006-01-02 15:04:05"

// localDateTime mirrors a Java LocalDateTime field annotated with
// @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss"): a malformed value fails Jackson
// deserialization (HttpMessageNotReadableException -> code "00002"), which is
// reproduced here by returning a non-validator error from UnmarshalJSON so that
// web.BindJSON writes the "http message not readable" response.
// The value is kept as the original string so the downstream forward emits the
// exact same format the Java ObjectMapper would
// (global LocalDateTimeSerializer with pattern "yyyy-MM-dd HH:mm:ss").
type localDateTime string

func (t *localDateTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if _, err := time.Parse(localDateTimeLayout, s); err != nil {
		return fmt.Errorf("cannot parse %q as LocalDateTime(yyyy-MM-dd HH:mm:ss)", s)
	}
	*t = localDateTime(s)
	return nil
}

// isEmptyData reports whether a downstream Result.data is absent/null,
// mirroring Java's `data == null` checks after ResultHelper.getResultData.
func isEmptyData(data json.RawMessage) bool {
	return len(data) == 0 || string(data) == "null"
}

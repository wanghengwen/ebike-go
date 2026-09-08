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

// getServiceRidingCard ports AccountController.getServiceRidingCard.
// Java gateway (RidingCardGatewayImpl): UserBaseCmd = genParam(UserBaseCmd.class)
// (context only, the BaseDTO body is NOT copied) + pin = getPin() +
// izUserService = true -> ridingCardApi.getRidingCard.
func getServiceRidingCard(c *gin.Context) {
	var req dto.ClientDTO // Java BaseDTO adds no fields to ClientDTO
	if !web.BindJSON(c, &req, baseMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req)
	body := userBaseCmd{Pin: cmdCtx.Pin, IzUserService: true}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceAccount, "/riding_card/get_riding_card", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertAccountRidingCard(result.Data)
	}
	web.RespondResult(c, result, err)
}

// getRidingCard ports AccountController.getRidingCard: same downstream call as
// getServiceRidingCard but with izUserService = false.
func getRidingCard(c *gin.Context) {
	var req dto.ClientDTO
	if !web.BindJSON(c, &req, baseMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req)
	body := userBaseCmd{Pin: cmdCtx.Pin, IzUserService: false}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceAccount, "/riding_card/get_riding_card", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertAccountRidingCard(result.Data)
	}
	web.RespondResult(c, result, err)
}

// getUserAccount ports AccountController.getUserAccount.
// Java gateway (AccountGatewayImpl): account UserServiceCmd =
// genParam(UserServiceCmd.class, dto) (copies serviceId, which may be null —
// UserAccountDTO.serviceId has no validation) + pin = getPin()
// -> userAccountApi.info.
func getUserAccount(c *gin.Context) {
	var req userAccountDTO
	if !web.BindJSON(c, &req, baseMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := accountUserServiceCmd{Pin: cmdCtx.Pin, ServiceId: req.ServiceId}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceAccount, "/account/info", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertUserAccountCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

type walletCO struct {
	Id             *javacompat.LongStr `json:"id"`
	Pin            *string             `json:"pin"`
	Balance        *int                `json:"balance"`
	Recharge       *int                `json:"recharge"`
	Present        *int                `json:"present"`
	DepositedMount *int                `json:"depositedMount"`
	DepositedStats *int                `json:"depositedStats"`
	FreezeRecharge *int                `json:"freezeRecharge"`
	FreezePresent  *int                `json:"freezePresent"`
}

type ridingCardInfo struct {
	Id                    *javacompat.LongStr  `json:"id"`
	Pin                   *string              `json:"pin"`
	DeductionType         *int                 `json:"deductionType"`
	ConfigId              *javacompat.LongStr  `json:"configId"`
	Name                  *string              `json:"name"`
	ServiceId             *javacompat.LongStr  `json:"serviceId"`
	FreeTime              *int                 `json:"freeTime"`
	FreeDistance          *int                 `json:"freeDistance"`
	FreeMoney             *int                 `json:"freeMoney"`
	TotalTimes            *int                 `json:"totalTimes"`
	IzTotalTimes          *int                 `json:"izTotalTimes"`
	ReceTimes             *int                 `json:"receTimes"`
	EffectiveServiceIds   *string              `json:"effectiveServiceIds"`
	EffectiveServiceNames *string              `json:"effectiveServiceNames"`
	RemainTimes           *int                 `json:"remainTimes"`
	LastUseTime           *javacompat.DateTime `json:"lastUseTime"`
	StartTime             *javacompat.DateTime `json:"startTime"`
	CardExpiredDate       *javacompat.DateTime `json:"cardExpiredDate"`
	Content               *string              `json:"content"`
	State                 *int                 `json:"state"`
	PromotionTag          *string              `json:"promotionTag"`
	DeductionRules        *int                 `json:"deductionRules"`
	BackOfCardUrl         *string              `json:"backOfCardUrl"`
	BuyState              *int                 `json:"buyState"`
	SysTradeNo            *string              `json:"sysTradeNo"`
	UsedTimes             *int                 `json:"usedTimes"`
}

type depositCardCO struct {
	Id          *javacompat.LongStr  `json:"id"`
	Pin         *string              `json:"pin"`
	ServiceId   *javacompat.LongStr  `json:"serviceId"`
	ConfigId    *javacompat.LongStr  `json:"configId"`
	ExpiredDate *javacompat.DateTime `json:"expiredDate"`
	Content     *string              `json:"content"`
}

type favorableCardCO struct {
	Id        *javacompat.LongStr  `json:"id"`
	Pin       *string              `json:"pin"`
	ServiceId *javacompat.LongStr  `json:"serviceId"`
	ConfigId  *javacompat.LongStr  `json:"configId"`
	BeginTime *javacompat.DateTime `json:"beginTime"`
	EndTime   *javacompat.DateTime `json:"endTime"`
	Content   *string              `json:"content"`
}

type freeOrderCO struct {
	Id         *javacompat.LongStr `json:"id"`
	Pin        *string             `json:"pin"`
	FreeSecond *int                `json:"freeSecond"`
	FreeNum    *int                `json:"freeNum"`
}

type discountCO struct {
	Id           *javacompat.LongStr `json:"id"`
	Pin          *string             `json:"pin"`
	DiscountRate *int                `json:"discountRate"`
	DiscountType *int                `json:"discountType"`
}

type userAccountCO struct {
	UserWallet        *walletCO        `json:"userWallet"`
	UserRidingCard    *ridingCardInfo  `json:"userRidingCard"`
	UserDepositCard   *depositCardCO   `json:"userDepositCard"`
	UserFavorableCard *favorableCardCO `json:"userFavorableCard"`
	UserFreeOrder     *freeOrderCO     `json:"userFreeOrder"`
	UserDiscount      *discountCO      `json:"userDiscount"`
}

func convertUserAccountCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co userAccountCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

func convertRidingCardInfo(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co ridingCardInfo
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

type accountRidingCardCO struct {
	CostUse  json.RawMessage `json:"cost_use"`
	RuleInfo json.RawMessage `json:"rule_info"`
	Expired  json.RawMessage `json:"expired"`
	Used     json.RawMessage `json:"used"`
}

func convertAccountRidingCard(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co accountRidingCardCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

// getUserFavorableCard ports AccountController.getUserFavorableCard:
// Java validates the body then returns ResultHelper.success(null) without any
// downstream call.
func getUserFavorableCard(c *gin.Context) { successNull(c) }

// getUserFreeCard ports AccountController.getUserFreeCard (success(null)).
func getUserFreeCard(c *gin.Context) { successNull(c) }

// getUserAllDiscount ports AccountController.getUserAllDiscount (success(null)).
func getUserAllDiscount(c *gin.Context) { successNull(c) }

// getUserDepositCard ports AccountController.getUserDepositCard (success(null)).
func getUserDepositCard(c *gin.Context) { successNull(c) }

// successNull binds/validates a UserAccountDTO body exactly like the Java
// no-op endpoints, then responds {success:true, code:"0", msg:"成功", data:null}.
func successNull(c *gin.Context) {
	var req userAccountDTO
	if !web.BindJSON(c, &req, baseMsgs) {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
}

// getWalletInfo ports AccountController.getWalletInfo.
// Java (UserWalletGatewayImpl.getUserWallet): UserBaseCmd = genParam(UserBaseCmd.class)
// + pin = getPin() (izUserService keeps its Java field default false)
// -> walletApi.info. The whole call is wrapped in try/catch:
//   - RPC error, failed Result or null data  -> new UserWalletDo() (all nulls)
//   - success -> ConvertorHelper.copyProperties(WalletCO, UserWalletDo.class),
//     i.e. only the same-named fields pin/balance/recharge/present/
//     depositedMount/depositedStats are kept (WalletCO id/freezeRecharge/
//     freezePresent are dropped).
//
// The controller always wraps in ResultHelper.success, so this endpoint never
// surfaces a downstream failure.
func getWalletInfo(c *gin.Context) {
	var req dto.ClientDTO // Java BaseDTO adds no fields to ClientDTO
	if !web.BindJSON(c, &req, baseMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req)
	body := userBaseCmd{Pin: cmdCtx.Pin, IzUserService: false}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceAccount, "/wallet/info", &body, cmdCtx)

	do := userWalletDo{}
	if err == nil && result != nil && result.Success && !isEmptyData(result.Data) {
		// An undecodable body would have failed inside the Java try/catch too,
		// yielding the empty UserWalletDo — so the unmarshal error is ignored.
		_ = json.Unmarshal(result.Data, &do)
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(do))
}

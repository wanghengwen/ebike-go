package user

// Ports Java UserController + UserServiceImpl + UserGatewayImpl (and the
// fence/account/management gateways they aggregate).

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

const (
	fenceByLocationPath     = "/serviceArea/getServiceByLocation"
	fenceNearByLocationPath = "/serviceArea/getNearServiceByLocation"
)

// serviceAreaCO is the subset of the fence ServiceAreaCO/FenceCO consumed here.
type serviceAreaCO struct {
	Id        *LongStr `json:"id"`
	Name      *string  `json:"name"`
	ShapeType *string  `json:"shapeType"`
	CenterLat *float64 `json:"centerLat"`
	CenterLng *float64 `json:"centerLng"`
	PointList *string  `json:"pointList"`
	Pics      *string  `json:"pics"`
}

// fetchServiceFence ports ServiceFenceGatewayImpl.getServiceFence/getNearService:
// fence LocationCmd{lat,lng}, any failure (transport, failed Result, decode) -> nil.
func fetchServiceFence(ctx context.Context, path string, lng, lat *float64, cmdCtx *dto.CommandContext) *serviceAreaCO {
	body := map[string]interface{}{"lat": lat, "lng": lng}
	raw := callDataQuiet(ctx, rpc.ServiceFence, path, body, cmdCtx)
	if raw == nil || isNullJSON(raw) {
		return nil
	}
	var co serviceAreaCO
	if err := json.Unmarshal(raw, &co); err != nil {
		return nil
	}
	return &co
}

// userDetailCo is the subset of ebike-user UserDetailCo consumed by this module.
type userDetailCo struct {
	Pin                *string   `json:"pin"`
	AuthName           *string   `json:"authName"`
	AuthNo             *string   `json:"authNo"`
	Phone              *string   `json:"phone"`
	IzAuth             *bool     `json:"izAuth"`
	IzNeedAuth         *bool     `json:"izNeedAuth"`
	CreatedAt          *DateTime `json:"createdAt"`
	IzDeposited        *bool     `json:"izDeposited"`
	ServiceId          *LongStr  `json:"serviceId"`
	RidingState        *int      `json:"ridingState"`
	PayState           *int      `json:"payState"`
	IzRiding           *bool     `json:"izRiding"`
	IzRidingType       *int      `json:"izRidingType"`
	Description        *string   `json:"description"`
	Nickname           *string   `json:"nickname"`
	Avatar             *string   `json:"avatar"`
	IzCareer           *bool     `json:"izCareer"`
	CareerState        *int      `json:"careerState"`
	PlatformScore      *int      `json:"platformScore"`
	IzWxScorePay       *bool     `json:"izWxScorePay"`
	IzZhimaPayAfterUse *bool     `json:"izZhimaPayAfterUse"`
	DepositState       *int      `json:"depositState"`
	BlacklistExpire    *DateTime `json:"blacklistExpire"`
	DepositCardExpire  *DateTime `json:"depositCardExpire"`
	NeedFaceCheck      *int      `json:"needFaceCheck"`
}

// ---------------------------------------------------------------------------
// POST /client/user/user/location
// ---------------------------------------------------------------------------

// locationFenceCo mirrors LocationFenceCo (id is @JsonProperty("serviceId") + ToString).
type locationFenceCo struct {
	ServiceId *LongStr `json:"serviceId"`
	Name      *string  `json:"name"`
	ShapeType *string  `json:"shapeType"`
	CenterLat *float64 `json:"centerLat"`
	CenterLng *float64 `json:"centerLng"`
	PointList *string  `json:"pointList"`
	Pics      *string  `json:"pics"`
}

func location(c *gin.Context) {
	// Java UserLocationCmd only adds longitude/latitude, which ClientDTO already declares.
	var req dto.ClientDTO
	if !web.BindJSON(c, &req, baseMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req)
	ctx := c.Request.Context()

	fence := fetchServiceFence(ctx, fenceByLocationPath, req.Longitude, req.Latitude, cmdCtx)
	if fence == nil {
		// 没匹配上则获取附近服务区
		fence = fetchServiceFence(ctx, fenceNearByLocationPath, req.Longitude, req.Latitude, cmdCtx)
	}

	out := locationFenceCo{}
	var serviceId *LongStr
	if fence != nil {
		serviceId = fence.Id
		out = locationFenceCo{
			Name:      fence.Name,
			ShapeType: fence.ShapeType,
			CenterLat: fence.CenterLat,
			CenterLng: fence.CenterLng,
			PointList: fence.PointList,
			Pics:      fence.Pics,
		}
		// UserGatewayImpl.reportLocation: LocationCmd{pin, longitude/latitude as
		// String.valueOf(Double), serviceId}; the downstream Result is read with
		// getResultDataWithoutException (business failure ignored) but a transport
		// error propagates like any uncaught Feign exception (-> 00001).
		body := map[string]interface{}{
			"pin":       cmdCtx.Pin,
			"longitude": javaDoubleString(req.Longitude),
			"latitude":  javaDoubleString(req.Latitude),
			"serviceId": serviceId,
		}
		if _, err := rpc.ForwardCommand(ctx, rpc.ServiceUser, "/user/location", body, cmdCtx); err != nil {
			web.WriteException(c, err.Error())
			return
		}
	}
	out.ServiceId = serviceId
	c.JSON(http.StatusOK, dto.NewSuccessResult(out))
}

// ---------------------------------------------------------------------------
// POST /client/user/user/personInfo
// ---------------------------------------------------------------------------

type personInfoCmd struct {
	dto.ClientDTO
	IzDesensitization *bool `json:"izDesensitization"`
}

// personInfoCo mirrors PersonInfoCo (field declaration order preserved).
type personInfoCo struct {
	Pin               *string   `json:"pin"`
	Phone             *string   `json:"phone"`
	IzAuth            *bool     `json:"izAuth"`
	IzNeedAuth        *bool     `json:"izNeedAuth"`
	AuthName          *string   `json:"authName"`
	AuthNo            *string   `json:"authNo"`
	RidingState       *int      `json:"ridingState"`
	PayState          *int      `json:"payState"`
	ServiceId         *LongStr  `json:"serviceId"`
	IzRiding          *bool     `json:"izRiding"`
	IzRidingType      *int      `json:"izRidingType"`
	Description       *string   `json:"description"`
	Balance           *int      `json:"balance"`
	Recharge          *int      `json:"recharge"`
	Present           *int      `json:"present"`
	DepositedMount    *int      `json:"depositedMount"`
	DepositState      *int      `json:"depositState"`
	DepositCardExpire *DateTime `json:"depositCardExpire"`
	Nickname          *string   `json:"nickname"`
	Avatar            *string   `json:"avatar"`
	NeedFaceCheck     *int      `json:"needFaceCheck"`
	BlacklistExpire   *DateTime `json:"blacklistExpire"`
	InBlacklist       *bool     `json:"inBlacklist"`
	CreatedAt         *DateTime `json:"createdAt"`
	CareerState       *int      `json:"careerState"`
	TagNames          []string  `json:"tagNames"`
	Tags              *string   `json:"tags"`
}

// userWalletDo mirrors the infrastructure UserWalletDo copied from account WalletCO.
type userWalletDo struct {
	Pin            *string `json:"pin"`
	Balance        *int    `json:"balance"`
	Recharge       *int    `json:"recharge"`
	Present        *int    `json:"present"`
	DepositedMount *int    `json:"depositedMount"`
}

type userTagCo struct {
	Tags     *string  `json:"tags"`
	TagNames []string `json:"tagNames"`
}

func personInfo(c *gin.Context) {
	var req personInfoCmd
	if !web.BindJSON(c, &req, baseMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	ctx := c.Request.Context()
	pin := cmdCtx.Pin

	fence := fetchServiceFence(ctx, fenceByLocationPath, req.Longitude, req.Latitude, cmdCtx)
	if fence == nil {
		// Java quirk preserved: the fallback calls getServiceFence AGAIN
		// (not getNearService) — see UserServiceImpl.getPersonInfo.
		fence = fetchServiceFence(ctx, fenceByLocationPath, req.Longitude, req.Latitude, cmdCtx)
	}
	var serviceId *LongStr
	if fence != nil {
		serviceId = fence.Id
	}

	// UserGatewayImpl.getDetailAndUpdate: UserDetailQuery{pin, serviceId} -> /user/detailUpdate
	raw, ok := callData(c, rpc.ServiceUser, "/user/detailUpdate",
		map[string]interface{}{"pin": pin, "serviceId": serviceId}, cmdCtx)
	if !ok {
		return
	}

	var co *personInfoCo
	if !isNullJSON(raw) {
		var detail userDetailCo
		if err := json.Unmarshal(raw, &detail); err != nil {
			web.WriteException(c, err.Error())
			return
		}
		co = &personInfoCo{
			Pin: detail.Pin, Phone: detail.Phone, IzAuth: detail.IzAuth, IzNeedAuth: detail.IzNeedAuth,
			AuthName: detail.AuthName, AuthNo: detail.AuthNo, RidingState: detail.RidingState,
			PayState: detail.PayState, ServiceId: detail.ServiceId, IzRiding: detail.IzRiding,
			IzRidingType: detail.IzRidingType, Description: detail.Description,
			DepositState: detail.DepositState, DepositCardExpire: detail.DepositCardExpire,
			Nickname: detail.Nickname, Avatar: detail.Avatar, NeedFaceCheck: detail.NeedFaceCheck,
			BlacklistExpire: detail.BlacklistExpire, CreatedAt: detail.CreatedAt, CareerState: detail.CareerState,
		}
	}

	// UserWalletGatewayImpl.getUserWallet: every failure falls back to an empty
	// UserWalletDo whose null fields then overwrite the target (ConvertorHelper.addProperties).
	var wallet userWalletDo
	if rawW := callDataQuiet(ctx, rpc.ServiceAccount, "/wallet/info",
		map[string]interface{}{"pin": pin, "izUserService": false}, cmdCtx); rawW != nil && !isNullJSON(rawW) {
		if err := json.Unmarshal(rawW, &wallet); err != nil {
			wallet = userWalletDo{}
		}
	}
	if co != nil {
		co.Pin = wallet.Pin
		co.Balance, co.Recharge, co.Present, co.DepositedMount =
			wallet.Balance, wallet.Recharge, wallet.Present, wallet.DepositedMount
	}

	// UserCareerGatewayImpl.userTag uses getResultData -> downstream failures bubble up.
	rawT, ok := callData(c, rpc.ServiceUser, "/tag/user/me",
		map[string]interface{}{"pin": pin}, cmdCtx)
	if !ok {
		return
	}
	if co == nil {
		// Java: a null UserDetailCo yields a null PersonInfoCo and an NPE at
		// personInfoCo.setTags/setPin (after the wallet and tag calls were made).
		writeNPE(c)
		return
	}
	if !isNullJSON(rawT) {
		var tag userTagCo
		if err := json.Unmarshal(rawT, &tag); err != nil {
			web.WriteException(c, err.Error())
			return
		}
		co.Tags = tag.Tags
		co.TagNames = tag.TagNames
	}

	co.Pin = &pin
	inBl := co.BlacklistExpire != nil && co.BlacklistExpire.Time.After(time.Now())
	co.InBlacklist = &inBl
	if boolVal(req.IzDesensitization) {
		// PersonInfoCo.desenSitized
		co.AuthName = desensitizedName(co.AuthName)
		co.Nickname = desensitizedName(co.Nickname)
		co.Phone = desensitizedIdCard(co.Phone, 7, 2)
		co.AuthNo = desensitizedIdCard(co.AuthNo, 1, 1)
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(co))
}

// ---------------------------------------------------------------------------
// POST /client/user/user/thirdBindInfo
// ---------------------------------------------------------------------------

// clientThirdBindInfoCo mirrors ClientThirdBindInfoCo (drops the WeChat fields
// present on the downstream ThirdBindInfoCo).
type clientThirdBindInfoCo struct {
	Pin            *string `json:"pin"`
	IzBindApple    *bool   `json:"izBindApple"`
	IzBindGoogle   *bool   `json:"izBindGoogle"`
	IzBindFacebook *bool   `json:"izBindFacebook"`
}

func thirdBindInfo(c *gin.Context) {
	var req dto.ClientDTO
	if !web.BindJSON(c, &req, baseMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req)

	raw, ok := callData(c, rpc.ServiceUser, "/user/thirdBindInfo",
		map[string]interface{}{"pin": cmdCtx.Pin}, cmdCtx)
	if !ok {
		return
	}
	if isNullJSON(raw) {
		// Java: copyProperties(null, ...) -> null -> success(null)
		c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
		return
	}
	var out clientThirdBindInfoCo
	if err := json.Unmarshal(raw, &out); err != nil {
		web.WriteException(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(out))
}

// ---------------------------------------------------------------------------
// POST /client/user/user/qualificationList
// ---------------------------------------------------------------------------

type clientServiceQualificationListCmd struct {
	dto.ClientDTO
	ServiceId *int64 `json:"serviceId" binding:"required"` // @NotNull
}

type careerTagCO struct {
	Id      *LongStr `json:"id"`
	TagName *string  `json:"tagName"`
}

type careerTagServiceCo struct {
	EnableTags []careerTagCO `json:"enableTags"`
}

// ridingPermissionCO is the subset of the fence RidingPermissionCO consumed here.
type ridingPermissionCO struct {
	IzDefaultPermit             *bool   `json:"izDefaultPermit"`
	IzDeposit                   *bool   `json:"izDeposit"`
	Deposit                     *int    `json:"deposit"`
	IzCareer                    *bool   `json:"izCareer"`
	Career                      *string `json:"career"`
	IzDepositCard               *bool   `json:"izDepositCard"`
	IzWxScorePayDeposited       *bool   `json:"izWxScorePayDeposited"`
	IzZhimaPayAfterUseDeposited *bool   `json:"izZhimaPayAfterUseDeposited"`
}

// depositCardDo mirrors the infrastructure DepositCardDo with its snake_case
// @JsonProperty mappings used to decode the ebike-marketing payload.
type depositCardDo struct {
	Id               *LongStr  `json:"id"`
	CardDurationDays *string   `json:"card_duration_day"`
	DiscountMoney    *string   `json:"discount_money"`
	CurMoney         *string   `json:"money"`
	Name             *string   `json:"name"`
	BackOfCardUrl    *string   `json:"card_url"`
	CreatedAt        *DateTime `json:"created_at"`
	State            *int      `json:"state"`
}

// depositCardCo mirrors the client-facing DepositCardCo (camelCase output).
type depositCardCo struct {
	Id               *LongStr  `json:"id"`
	CardDurationDays *string   `json:"cardDurationDays"`
	DiscountMoney    *string   `json:"discountMoney"`
	CurMoney         *string   `json:"curMoney"`
	Name             *string   `json:"name"`
	BackOfCardUrl    *string   `json:"backOfCardUrl"`
	CreatedAt        *DateTime `json:"createdAt"`
	State            *int      `json:"state"`
}

// clientServiceQualificationListCo mirrors ClientServiceQualificationListCo
// (the Java field defaults are overwritten by copyProperties, so nulls pass through).
type clientServiceQualificationListCo struct {
	IzDeposit                   *bool           `json:"izDeposit"`
	Deposit                     *int            `json:"deposit"`
	IzCareer                    *bool           `json:"izCareer"`
	CareerTags                  []careerTagCO   `json:"careerTags"`
	IzWxScorePayDeposited       *bool           `json:"izWxScorePayDeposited"`
	IzZhimaPayAfterUseDeposited *bool           `json:"izZhimaPayAfterUseDeposited"`
	IzDepositCard               *bool           `json:"izDepositCard"`
	DepositCardCoList           []depositCardCo `json:"depositCardCoList"`
	CareerName                  *string         `json:"careerName"`
	Career                      *string         `json:"career"`
}

func qualificationList(c *gin.Context) {
	var req clientServiceQualificationListCmd
	if !web.BindJSON(c, &req, withBase(map[string]string{"serviceId": "must not be null"})) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	ctx := c.Request.Context()

	// UserGatewayImpl.getMembersConfig: fence /ridingPermission/get with
	// getResultDataWithoutException + try/catch -> empty RidingPermissionCO on any failure.
	rp := &ridingPermissionCO{}
	if raw := callDataQuiet(ctx, rpc.ServiceFence, "/ridingPermission/get",
		map[string]interface{}{"serviceId": req.ServiceId}, cmdCtx); raw != nil && !isNullJSON(raw) {
		var v ridingPermissionCO
		if err := json.Unmarshal(raw, &v); err == nil {
			rp = &v
		}
	}

	res := clientServiceQualificationListCo{
		IzDeposit:                   rp.IzDeposit,
		Deposit:                     rp.Deposit,
		IzCareer:                    rp.IzCareer,
		IzWxScorePayDeposited:       rp.IzWxScorePayDeposited,
		IzZhimaPayAfterUseDeposited: rp.IzZhimaPayAfterUseDeposited,
		IzDepositCard:               rp.IzDepositCard,
		Career:                      rp.Career,
	}

	// UserWalletGatewayImpl.getServiceDepositCardList: ebike-marketing feign,
	// body field is @JsonProperty("service_id"); any failure -> empty list.
	var cards []depositCardDo
	if raw := callDataQuiet(ctx, rpc.ServiceMarketing, "/ebike_marketing/deposit_config/platform/deposit_config_list",
		map[string]interface{}{"service_id": req.ServiceId}, cmdCtx); raw != nil && !isNullJSON(raw) {
		if err := json.Unmarshal(raw, &cards); err != nil {
			cards = nil
		}
	}
	active := make([]depositCardCo, 0)
	for _, d := range cards {
		if d.State == nil {
			// Java: `1 == d.getState()` unboxes a null Integer -> NPE -> 00001
			writeNPE(c)
			return
		}
		if *d.State == 1 {
			active = append(active, depositCardCo(d))
		}
	}
	res.DepositCardCoList = active

	// 兼容老版本职业认证
	if rp.IzDepositCard == nil {
		v := len(active) > 0
		res.IzDepositCard = &v
	}
	// 设置职业认证标签
	if boolVal(rp.IzCareer) {
		rawT, ok := callData(c, rpc.ServiceUser, "/tag/info",
			map[string]interface{}{"serviceIds": []string{strconv.FormatInt(*req.ServiceId, 10)}}, cmdCtx)
		if !ok {
			return
		}
		if isNullJSON(rawT) {
			// Java: serviceCo.getEnableTags() on a null CO -> NPE
			writeNPE(c)
			return
		}
		var tagSvc careerTagServiceCo
		if err := json.Unmarshal(rawT, &tagSvc); err != nil {
			web.WriteException(c, err.Error())
			return
		}
		res.CareerTags = tagSvc.EnableTags
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(res))
}

// ---------------------------------------------------------------------------
// POST /client/user/user/cancelAccount
// ---------------------------------------------------------------------------

func cancelAccount(c *gin.Context) {
	var req dto.ClientDTO
	if !web.BindJSON(c, &req, baseMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req)
	// UserGatewayImpl.cancelAccount: UserDetailQuery{pin} -> /user/cancelAccount (passthrough)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/user/cancelAccount",
		map[string]interface{}{"pin": cmdCtx.Pin}, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertUserCancelAccountCo(result.Data)
	}
	web.RespondResult(c, result, err)
}

// ---------------------------------------------------------------------------
// POST /client/user/user/submitCancel
// ---------------------------------------------------------------------------

type clientUserCancelAccountCmd struct {
	dto.ClientDTO
	Code *string `json:"code" binding:"required"` // @NotNull (empty string allowed)
}

func submitCancel(c *gin.Context) {
	var req clientUserCancelAccountCmd
	if !web.BindJSON(c, &req, withBase(map[string]string{"code": "must not be null"})) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	if _, ok := callData(c, rpc.ServiceUser, "/user/submitCancel",
		map[string]interface{}{"code": req.Code, "pin": cmdCtx.Pin}, cmdCtx); !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(1))
}

// ---------------------------------------------------------------------------
// POST /client/user/user/update
// ---------------------------------------------------------------------------

type clientUserUpdateCmd struct {
	dto.ClientDTO
	Nickname *string `json:"nickname"`
	Avatar   *string `json:"avatar"`
}

func updateUser(c *gin.Context) {
	var req clientUserUpdateCmd
	if !web.BindJSON(c, &req, baseMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	if _, ok := callData(c, rpc.ServiceUser, "/user/update",
		map[string]interface{}{"nickname": req.Nickname, "avatar": req.Avatar, "pin": cmdCtx.Pin}, cmdCtx); !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(1))
}

// ---------------------------------------------------------------------------
// POST /client/user/user/canRide
// ---------------------------------------------------------------------------

type canRidingCo struct {
	IzRidingQualification *bool   `json:"izRidingQualification"`
	IzCanRiding           *bool   `json:"izCanRiding"`
	ForbidType            *int    `json:"forbidType"`
	Reason                *string `json:"reason"`
	Phone                 *string `json:"phone"`
	Email                 *string `json:"email"`
	Pin                   *string `json:"pin"`
	UserTags              *string `json:"userTags"`
}

type userCancelAccountCo struct {
	Pin                       *string `json:"pin"`
	IzDeposit                 *bool   `json:"izDeposit"`
	IzToPay                   *bool   `json:"izToPay"`
	IzRecharge                *bool   `json:"izRecharge"`
	IzHaveUnauditedUserTicket *bool   `json:"izHaveUnauditedUserTicket"`
	Success                   *bool   `json:"success"`
}

type grayOrderCo struct {
	Phone     *string `json:"phone"`
	NodeState *int    `json:"nodeState"`
	SaasState *int    `json:"saasState"`
	Redirect  *int    `json:"redirect"`
}

func convertCanRidingCo(raw json.RawMessage) json.RawMessage {
	if isNullJSON(raw) {
		return raw
	}
	var co canRidingCo
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

func convertUserCancelAccountCo(raw json.RawMessage) json.RawMessage {
	if isNullJSON(raw) {
		return raw
	}
	var co userCancelAccountCo
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

func convertGrayOrderCo(raw json.RawMessage) json.RawMessage {
	if isNullJSON(raw) {
		return raw
	}
	var co grayOrderCo
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

type clientCanRidingQuery struct {
	dto.ClientDTO
	ServiceId *int64 `json:"serviceId" binding:"required"` // @NotNull
	Pin       string `json:"pin" binding:"required"`       // @NotBlank
}

func canRide(c *gin.Context) {
	var req clientCanRidingQuery
	if !web.BindJSON(c, &req, withBase(map[string]string{
		"serviceId": "must not be null",
		"pin":       "must not be blank",
	})) {
		return
	}
	if !web.NotBlank(c, "pin", req.Pin) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/user/izRidingQualification",
		map[string]interface{}{"serviceId": req.ServiceId, "pin": req.Pin}, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertCanRidingCo(result.Data)
	}
	web.RespondResult(c, result, err)
}

// ---------------------------------------------------------------------------
// POST /client/user/blacklist/info
// ---------------------------------------------------------------------------

type blacklistPageCo struct {
	Id        *LongStr  `json:"id"`
	Pin       *string   `json:"pin"`
	ServiceId *LongStr  `json:"serviceId"`
	CreatedAt *DateTime `json:"createdAt"`
	CancelAt  *DateTime `json:"cancelAt"`
	Reason    *string   `json:"reason"`
}

// clientBlacklistCo mirrors ClientBlacklistCo; days is a Java Long -> string output.
type clientBlacklistCo struct {
	Id           *LongStr  `json:"id"`
	Pin          *string   `json:"pin"`
	ServiceId    *LongStr  `json:"serviceId"`
	CreatedAt    *DateTime `json:"createdAt"`
	CancelAt     *DateTime `json:"cancelAt"`
	Reason       *string   `json:"reason"`
	State        *int      `json:"state"` // never populated: BlacklistPageCo has no state field
	Days         *LongStr  `json:"days"`
	SupportPhone *string   `json:"supportPhone"`
}

func blacklistInfo(c *gin.Context) {
	var req dto.ClientDTO
	if !web.BindJSON(c, &req, baseMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req)

	raw, ok := callData(c, rpc.ServiceUser, "/blacklist/info",
		map[string]interface{}{"pin": cmdCtx.Pin}, cmdCtx)
	if !ok {
		return
	}
	if isNullJSON(raw) {
		// Java dereferences blacklistPageCo.getServiceId() right away -> NPE
		writeNPE(c)
		return
	}
	var bl blacklistPageCo
	if err := json.Unmarshal(raw, &bl); err != nil {
		web.WriteException(c, err.Error())
		return
	}

	// HelpConfigGatewayImpl.getCustomerServiceByServiceId(Long): CustomerServiceCmd{serviceId}
	rawCS, ok := callData(c, rpc.ServiceFence, "/helpConfig/getCustomerServiceByServiceId",
		map[string]interface{}{"serviceId": bl.ServiceId}, cmdCtx)
	if !ok {
		return
	}
	if isNullJSON(rawCS) {
		// Java: customerServiceCO.getTel() on null -> NPE
		writeNPE(c)
		return
	}
	var cs struct {
		Tel *string `json:"tel"`
	}
	if err := json.Unmarshal(rawCS, &cs); err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if bl.CancelAt == nil {
		// Java: Duration.between(now, null) -> NPE
		writeNPE(c)
		return
	}
	// Duration.between(now, cancelAt).toDays() + 1, clamped to >= 0
	days := int64(bl.CancelAt.Time.Sub(time.Now())/(24*time.Hour)) + 1
	if days <= 0 {
		days = 0
	}
	daysVal := LongStr(days)
	out := clientBlacklistCo{
		Id:           bl.Id,
		Pin:          bl.Pin,
		ServiceId:    bl.ServiceId,
		CreatedAt:    bl.CreatedAt,
		CancelAt:     bl.CancelAt,
		Reason:       bl.Reason,
		Days:         &daysVal,
		SupportPhone: cs.Tel,
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(out))
}

// ---------------------------------------------------------------------------
// POST /client/user/cancel/smsCode
// ---------------------------------------------------------------------------

type clientUserCancelDTO struct {
	dto.ClientDTO
	Scene string `json:"scene" binding:"required"` // @NotBlank
}

func cancelSmsCode(c *gin.Context) {
	var req clientUserCancelDTO
	if !web.BindJSON(c, &req, withBase(map[string]string{"scene": "must not be blank"})) {
		return
	}
	if !web.NotBlank(c, "scene", req.Scene) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)

	raw, ok := callData(c, rpc.ServiceUser, "/user/detail",
		map[string]interface{}{"pin": cmdCtx.Pin}, cmdCtx)
	if !ok {
		return
	}
	if isNullJSON(raw) {
		// Java: user.getPhone() on null -> NPE
		writeNPE(c)
		return
	}
	var detail struct {
		Phone *string `json:"phone"`
	}
	if err := json.Unmarshal(raw, &detail); err != nil {
		web.WriteException(c, err.Error())
		return
	}
	// CodeGatewayImpl.sendMessageCode: ebike-management SmsSendCodeCmd{phone, scene}
	// (CodeApi path "code/send" has no leading slash in Java)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceManagement, "/code/send",
		map[string]interface{}{"phone": detail.Phone, "scene": req.Scene}, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertBoolean(result.Data)
	}
	web.RespondResult(c, result, err)
}

// ---------------------------------------------------------------------------
// POST /client/gray/queryOrder & POST /client/gray/userData
// ---------------------------------------------------------------------------

type cGrayOrderQuery struct {
	dto.ClientDTO
	Phone    string `json:"phone" binding:"required"`    // @NotEmpty + @Pattern
	GrayType string `json:"grayType" binding:"required"` // @NotEmpty
}

var grayPhonePattern = regexp.MustCompile(`^\+86\-1(3\d|4[5-9]|5[0-35-9]|6[567]|7[0-8]|8\d|9[0-35-9])\d{8}$`)

var grayMsgs = withBase(map[string]string{
	"phone":    "must not be empty",
	"grayType": "must not be empty",
})

func bindGrayQuery(c *gin.Context) (*cGrayOrderQuery, bool) {
	var req cGrayOrderQuery
	if !web.BindJSON(c, &req, grayMsgs) {
		return nil, false
	}
	if !grayPhonePattern.MatchString(req.Phone) {
		// @Pattern(message = "手机号格式错误")
		web.WriteParamError(c, "phone 手机号格式错误")
		return nil, false
	}
	return &req, true
}

func grayOrderQuery(c *gin.Context) {
	req, ok := bindGrayQuery(c)
	if !ok {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	// GrayOrderQuery copy: phone (getter guarantees the +86- prefix, which the
	// validated input already has), grayType, tenantId.
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/gray/queryOrder",
		map[string]interface{}{"phone": req.Phone, "grayType": req.GrayType, "tenantId": req.TenantId}, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertGrayOrderCo(result.Data)
	}
	web.RespondResult(c, result, err)
}

// grayUserDataCo mirrors GrayUserDataCo (field declaration order preserved).
type grayUserDataCo struct {
	Pin                *string   `json:"pin"`
	Phone              *string   `json:"phone"`
	IzAuth             *bool     `json:"izAuth"`
	AuthName           *string   `json:"authName"`
	AuthNo             *string   `json:"authNo"`
	IzFreeDeposited    *bool     `json:"izFreeDeposited"`
	IzDeposited        *bool     `json:"izDeposited"`
	DepositedInfo      *string   `json:"depositedInfo"`
	IzDepositedCard    *bool     `json:"izDepositedCard"`
	DepositCardExpire  *DateTime `json:"depositCardExpire"`
	IzWxPayScore       *bool     `json:"izWxPayScore"`
	IzZhimaPayAfterUse *bool     `json:"izZhimaPayAfterUse"`
	IzCareer           *bool     `json:"izCareer"`
	RidingState        *int      `json:"ridingState"`
	ServiceId          *LongStr  `json:"serviceId"`
	CreditScore        *float64  `json:"creditScore"` // never populated (no matching source field)
	Inblacklist        *bool     `json:"inblacklist"`
	IzRiding           *bool     `json:"izRiding"`
	Source             *int      `json:"source"`
}

// buildDepositedInfo replicates UserServiceImpl.getDepositedInfo: fastjson
// serialization of a HashMap. The key order matches Java HashMap iteration for
// these keys (table size 16), and fastjson omits null values, so a null phone
// drops transaction_id entirely.
func buildDepositedInfo(phone *string) string {
	var b strings.Builder
	b.WriteByte('{')
	first := true
	add := func(k, v string) {
		if !first {
			b.WriteByte(',')
		}
		first = false
		kb, _ := json.Marshal(k)
		vb, _ := json.Marshal(v)
		b.Write(kb)
		b.WriteByte(':')
		b.Write(vb)
	}
	if phone != nil {
		add("transaction_id", *phone)
	}
	add("time_expire", "")
	add("gmt_payment", "")
	add("out_trade_no", "")
	add("time_start", "")
	add("total_fee", "")
	add("channel", "")
	add("prepay_id", "")
	b.WriteByte('}')
	return b.String()
}

// computeSource ports UserServiceImpl.getSource; the second return value is
// false when Java would NPE on a null Boolean unboxing.
func computeSource(o *grayUserDataCo) (int, bool) {
	if o.IzRiding == nil {
		return 0, false
	}
	if !*o.IzRiding {
		return 0, true
	}
	if *o.IzDeposited {
		return 1, true
	}
	if *o.IzDepositedCard {
		return 2, true
	}
	if o.IzCareer == nil {
		return 0, false
	}
	if *o.IzCareer {
		return 3, true
	}
	if *o.IzFreeDeposited {
		return 4, true
	}
	if o.IzWxPayScore == nil {
		return 0, false
	}
	if *o.IzWxPayScore {
		return 5, true
	}
	return 0, true
}

func grayUserData(c *gin.Context) {
	req, ok := bindGrayQuery(c)
	if !ok {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	// UserGatewayImpl.getDetailByPhone mutates the CommandContext tenantId with
	// the request body value before the call.
	ctxCopy := *cmdCtx
	ctxCopy.TenantId = req.TenantId

	raw, ok := callData(c, rpc.ServiceUser, "/user/detailByPhone",
		map[string]interface{}{"phone": req.Phone}, &ctxCopy)
	if !ok {
		return
	}
	if isNullJSON(raw) {
		// Java: userDetailCo == null -> return null -> success(null)
		c.JSON(http.StatusOK, dto.NewSuccessResult(nil))
		return
	}
	var detail userDetailCo
	if err := json.Unmarshal(raw, &detail); err != nil {
		web.WriteException(c, err.Error())
		return
	}

	out := grayUserDataCo{
		Pin: detail.Pin, Phone: detail.Phone, IzAuth: detail.IzAuth,
		AuthName: detail.AuthName, AuthNo: detail.AuthNo,
		IzDeposited: detail.IzDeposited, DepositCardExpire: detail.DepositCardExpire,
		IzZhimaPayAfterUse: detail.IzZhimaPayAfterUse, IzCareer: detail.IzCareer,
		RidingState: detail.RidingState, ServiceId: detail.ServiceId, IzRiding: detail.IzRiding,
		IzWxPayScore: detail.IzWxScorePay,
	}
	// 押金信息 — Java unboxes getIzDeposited() -> NPE when null
	if out.IzDeposited == nil {
		writeNPE(c)
		return
	}
	if *out.IzDeposited {
		info := buildDepositedInfo(detail.Phone)
		out.DepositedInfo = &info
	}
	// 押金卡
	izDC := out.DepositCardExpire != nil && out.DepositCardExpire.Time.After(time.Now())
	out.IzDepositedCard = &izDC
	// 一键免押 (null-safe via Optional in Java)
	izFree := boolVal(out.IzRiding) &&
		!boolVal(out.IzDeposited) && !boolVal(out.IzDepositedCard) && !boolVal(out.IzCareer) &&
		!boolVal(out.IzWxPayScore) && !boolVal(out.IzZhimaPayAfterUse)
	out.IzFreeDeposited = &izFree
	// 押金来源
	src, okSrc := computeSource(&out)
	if !okSrc {
		writeNPE(c)
		return
	}
	out.Source = &src
	// 黑名单 — Java dereferences getBlacklistExpire() directly -> NPE when null
	if detail.BlacklistExpire == nil {
		writeNPE(c)
		return
	}
	inBl := detail.BlacklistExpire.Time.After(time.Now())
	out.Inblacklist = &inBl
	c.JSON(http.StatusOK, dto.NewSuccessResult(out))
}

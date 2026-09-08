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

// defaultActivityCenterList ports ActivityCenterController.defaultActivityCenterList.
// Java gateway: IdCmd = genParam(IdCmd.class) (context only, NO dto copy);
// idCmd.setId(serviceDTO.getServiceId()) -> activityCenterApi.getActivityCenterList.
func defaultActivityCenterList(c *gin.Context) {
	var req dto.ServiceDTO
	if !web.BindJSON(c, &req, serviceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := idCmd{Id: req.ServiceId}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/activityCenter/list", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertActivityCenterCOList(result.Data)
	}
	web.RespondResult(c, result, err)
}

type userRegularCO struct {
	ActivityId        *javacompat.LongStr  `json:"activityId"`
	EndTime           *javacompat.DateTime `json:"endTime"`
	Name              *string              `json:"name"`
	FixedActivityType *int                 `json:"fixedActivityType"`
	CareerTag         *string              `json:"careerTag"`
	CanComplete       *int                 `json:"canComplete"`
	UserCompleted     *int                 `json:"userCompleted"`
}

type ridingConfigListCO struct {
	CardId         *javacompat.LongStr  `json:"cardId"`
	RidingCardName *string              `json:"ridingCardName"`
	DeductionType  *int                 `json:"deductionType"`
	CurCost        *int                 `json:"curCost"`
	OriginCost     *int                 `json:"originCost"`
	ExpiryDate     *int                 `json:"expiryDate"`
	TotalTimes     *int                 `json:"totalTimes"`
	FreeMoney      *int                 `json:"freeMoney"`
	OpenStartTime  *javacompat.LongStr  `json:"openStartTime"`
	OpenEndTime    *javacompat.LongStr  `json:"openEndTime"`
	State          *int                 `json:"state"`
	CreatedAt      *javacompat.LongStr  `json:"createdAt"`
	DeductionRules *int                 `json:"deductionRules"`
	DescriptionTag *string              `json:"descriptionTag"`
	PromotionTag   *string              `json:"promotionTag"`
	OnSaleDays     *int                 `json:"onSaleDays"`
	BackOfCardUrl  *string              `json:"backOfCardUrl"`
	Sold           *int                 `json:"sold"`
	Used           *int                 `json:"used"`
	ExpireUnused   *int                 `json:"expireUnused"`
	IzMainPush     *bool                `json:"izMainPush"`
	UpdatedPin     *string              `json:"updatedPin"`
	UpdatedName    *string              `json:"updatedName"`
	DetailInfo     *string              `json:"detailInfo"`
	UpdatedAt      *javacompat.DateTime `json:"updatedAt"`
}

type inviteCO struct {
	Id              *javacompat.LongStr  `json:"id"`
	ServiceId       *javacompat.LongStr  `json:"serviceId"`
	Name            *string              `json:"name"`
	State           *bool                `json:"state"`
	RewardType      *int                 `json:"rewardType"`
	RewardInfo      *string              `json:"rewardInfo"`
	RewardCondition *string              `json:"rewardCondition"`
	RemindType      *string              `json:"remindType"`
	RidingcardId    *javacompat.LongStr  `json:"ridingcardId"`
	ShowType        *string              `json:"showType"`
	BeginTime       *javacompat.DateTime `json:"beginTime"`
	EndTime         *javacompat.DateTime `json:"endTime"`
}

type activityCenterCO struct {
	Name           *string              `json:"name"`
	Type           *int                 `json:"type"`
	Enabled        *int                 `json:"enabled"`
	ServiceId      *javacompat.LongStr  `json:"serviceId"`
	RegularList    []userRegularCO      `json:"regularList"`
	RidingCardList []ridingConfigListCO `json:"ridingCardList"`
	InviteDetail   *inviteCO            `json:"inviteDetail"`
}

func convertActivityCenterCOList(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var list []activityCenterCO
	if json.Unmarshal(raw, &list) != nil {
		return raw
	}
	b, _ := json.Marshal(list)
	return b
}

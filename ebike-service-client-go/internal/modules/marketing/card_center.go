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

// ridingConfigList ports CardCenterController.list.
// Java gateway: IdCmd = genParam(IdCmd.class, dto) (no same-named field, so
// nothing copies) + idCmd.setId(dto.getServiceId()) -> ridingConfigApi.usefulRidingCard.
func ridingConfigList(c *gin.Context) {
	var req dto.ServiceDTO
	if !web.BindJSON(c, &req, serviceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := idCmd{Id: req.ServiceId}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/ridingConfig/usefulRidingCard", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertRidingConfigListCOList(result.Data)
	}
	web.RespondResult(c, result, err)
}

// ridingConfigGetRule ports CardCenterController.getRidingConfigRule.
// Java gateway: IdCmd = genParam(IdCmd.class) + id = dto.serviceId -> ridingConfigApi.getRule.
func ridingConfigGetRule(c *gin.Context) {
	var req dto.ServiceDTO
	if !web.BindJSON(c, &req, serviceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := idCmd{Id: req.ServiceId}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/ridingConfig/getRule", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertRidingRuleCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

type ridingRuleCO struct {
	RuleInfo    *string `json:"ruleInfo"`
	MaxNum      *int    `json:"maxNum"`
	DefaultRule *bool   `json:"defaultRule"`
}

func convertRidingConfigListCOList(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var list []ridingConfigListCO
	if json.Unmarshal(raw, &list) != nil {
		return raw
	}
	b, _ := json.Marshal(list)
	return b
}

func convertRidingRuleCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co ridingRuleCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

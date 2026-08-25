package marketing

import (
	"encoding/json"

	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/javacompat"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// redPaketCarGetRule ports RedPaketCarController.getByActivityId.
// Java gateway: IdCmd = genParam(IdCmd.class, dto) (copies id from IdDTO)
// -> marketingActivityApi.getByActivityId.
func redPaketCarGetRule(c *gin.Context) {
	var req idDTO
	if !web.BindJSON(c, &req, idMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := idCmd{Id: req.Id}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/marketing_activity/getById", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertRedPaketActivityCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

type redPaketActivityCO struct {
	Id             *javacompat.LongStr  `json:"id"`
	ServiceId      *javacompat.LongStr  `json:"serviceId"`
	Name           *string              `json:"name"`
	Type           *int                 `json:"type"`
	PubType        *int                 `json:"pubType"`
	PubStartTime   *javacompat.DateTime `json:"pubStartTime"`
	PubEndTime     *javacompat.DateTime `json:"pubEndTime"`
	PubRate        *int                 `json:"pubRate"`
	ProdRuleId     *javacompat.LongStr  `json:"prodRuleId"`
	CarRestHour    *int                 `json:"carRestHour"`
	PointIds       []javacompat.LongStr `json:"pointIds"`
	Rate           *int                 `json:"rate"`
	Time           *string              `json:"time"`
	DurationHour   *int                 `json:"durationHour"`
	ReceiveRuleId  *javacompat.LongStr  `json:"receiveRuleId"`
	RidingTime     *javacompat.LongStr  `json:"ridingTime"`
	RidingDistance *int                 `json:"ridingDistance"`
	MinAmount      *int                 `json:"minAmount"`
	MaxAmount      *int                 `json:"maxAmount"`
}

func convertRedPaketActivityCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co redPaketActivityCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

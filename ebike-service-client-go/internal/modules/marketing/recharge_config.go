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

// rechargeList ports RechargeConfigController.list.
// Java gateway: RechargeCmd = genParam(RechargeCmd.class, dto) (copies
// serviceId) + state = 1 -> rechargeConfigApi.list. If the downstream succeeds
// but the list is null/empty, a hard-coded template list is returned instead
// (amounts 20/30/50/100/200 yuan, id=11, state=1, activityType=0).
func rechargeList(c *gin.Context) {
	var req dto.ServiceDTO
	if !web.BindJSON(c, &req, serviceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	state := 1
	body := rechargeCmd{ServiceId: req.ServiceId, State: &state}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/recharge/list", &body, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !result.Success {
		web.WriteBizResult(c, result)
		return
	}
	if rechargeDataIsEmpty(result.Data) {
		c.JSON(http.StatusOK, dto.NewSuccessResult(defaultRechargeList()))
		return
	}
	result.Data = convertRechargeCOList(result.Data)
	c.JSON(http.StatusOK, result)
}

// rechargeDataIsEmpty mirrors Java CollectionUtils.isEmpty(list): true for
// null or an empty list.
func rechargeDataIsEmpty(data json.RawMessage) bool {
	if isEmptyData(data) {
		return true
	}
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		// not a list; Java would have failed Feign decoding long before -> treat as non-empty passthrough
		return false
	}
	return len(items) == 0
}

// defaultRechargeList builds the Java fallback template:
// for amount in {20,30,50,100,200}: new RechargeCO().setId(11L).setState(1)
// .setActivityType(0).setAmount(amount*100) — every other field stays null.
func defaultRechargeList() []rechargeCO {
	amounts := []int{20, 30, 50, 100, 200}
	list := make([]rechargeCO, 0, len(amounts))
	for _, a := range amounts {
		id := javacompat.LongStr(11) // Java Long 11L -> global ToStringSerializer -> "11"
		state := 1
		activityType := 0
		amount := a * 100
		list = append(list, rechargeCO{
			Id:           &id,
			State:        &state,
			ActivityType: &activityType,
			Amount:       &amount,
		})
	}
	return list
}

// rechargeScope ports RechargeConfigController.rechargeScope.
// Java gateway: ServiceCmd = genParam(ServiceCmd.class, dto) (copies serviceId)
// -> rechargeConfigApi.getRechargeScope.
func rechargeScope(c *gin.Context) {
	var req dto.ServiceDTO
	if !web.BindJSON(c, &req, serviceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := serviceCmd{ServiceId: req.ServiceId}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/recharge/get_recharge_scope", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertRechargeScopeCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

// getRechargeConfig ports RechargeConfigController.getRechargeConfig
// -> rechargeConfigApi.getRechargeConfig.
func getRechargeConfig(c *gin.Context) {
	var req dto.ServiceDTO
	if !web.BindJSON(c, &req, serviceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := serviceCmd{ServiceId: req.ServiceId}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceMarketing, "/recharge/get_recharge_config", &body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertRechargeConfigCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

type rechargeScopeCO struct {
	Scope []int `json:"scope"`
}

type rechargeConfigCO struct {
	IzOpen  *bool   `json:"izOpen"`
	Title   *string `json:"title"`
	Content *string `json:"content"`
}

func convertRechargeCOList(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var list []rechargeCO
	if json.Unmarshal(raw, &list) != nil {
		return raw
	}
	b, _ := json.Marshal(list)
	return b
}

func convertRechargeScopeCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co rechargeScopeCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

func convertRechargeConfigCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co rechargeConfigCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, _ := json.Marshal(co)
	return b
}

package clientconfig

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

func getUseCarConfig(c *gin.Context) {
	var req getConfigDTO
	if !web.BindJSON(c, &req, getConfigMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/config/usecar/getConfigByServiceId", &req, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !result.Success {
		web.WriteBizResult(c, result)
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(convertConfigUseCarCO(result.Data))))
}

func getBackCarConfig(c *gin.Context) {
	var req configBackCarDTO
	if !web.BindJSON(c, &req, getConfigMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := map[string]*int64{"id": req.ServiceId}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/config/backcar/getConfigByServiceId", body, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !result.Success {
		web.WriteBizResult(c, result)
		return
	}
	merged, ok := mergeBackCarAudit(c, result.Data, req.ServiceId, cmdCtx)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(merged)))
}

func getBackCarConfigByCarId(c *gin.Context) {
	var req carIdDTO
	if !web.BindJSON(c, &req, map[string]string{"id": "must not be null", "traceId": "must not be empty", "tenantId": "must not be empty"}) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	carBody := map[string]interface{}{"carId": req.Id, "num": 0}
	carResult, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceManagement, "/car-info/detail", carBody, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !carResult.Success {
		web.WriteBizResult(c, carResult)
		return
	}
	var car struct {
		ServiceId *int64 `json:"serviceId"`
	}
	if json.Unmarshal(carResult.Data, &car) != nil || car.ServiceId == nil {
		writeNPE(c)
		return
	}
	body := map[string]*int64{"id": car.ServiceId}
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/config/backcar/getConfigByServiceId", body, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !result.Success {
		web.WriteBizResult(c, result)
		return
	}
	merged, _ := convertConfigBackCarCCO(result.Data, nil)
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(merged)))
}

func mergeBackCarAudit(c *gin.Context, backCar json.RawMessage, serviceId *int64, cmdCtx *dto.CommandContext) (json.RawMessage, bool) {
	auditBody := map[string]*int64{"serviceId": serviceId}
	auditResult, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceOrder, "/returnBikeAudit/getReturnBikeAuditIzOn", auditBody, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return nil, false
	}
	if !auditResult.Success {
		web.WriteBizResult(c, auditResult)
		return nil, false
	}
	var audit json.RawMessage
	if !javacompat.IsNullJSON(auditResult.Data) {
		audit = auditResult.Data
	}
	merged, _ := convertConfigBackCarCCO(backCar, audit)
	return merged, true
}

func getConfigPay(c *gin.Context) {
	var req configPayDTO
	if !bindConfigPayDTO(c, &req) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/config/pay/getConfigByServiceId", &req, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !result.Success {
		web.WriteBizResult(c, result)
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(convertConfigPayCO(result.Data))))
}

func getConfigBaseItem(c *gin.Context) {
	var req configBaseItemDTO
	if !bindConfigBaseItemDTO(c, &req) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/config/base/getConfigByServiceId", &req, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !result.Success {
		web.WriteBizResult(c, result)
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(convertConfigBaseItemCCO(result.Data))))
}

var writeNPE = javacompat.WriteNPE

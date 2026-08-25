package fence

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

// configProtocolDTO mirrors ConfigProtocolDTO; group validation is manual.
type configProtocolDTO struct {
	dto.ClientDTO
	ServiceId *int64 `json:"serviceId"`
	Type      *int   `json:"type"`
}

// resourceManagementDto mirrors ResourceManagementDto (no @Validated on controller).
// platform lives on embedded ClientDTO as string (ios/android/wechat); do not redeclare as int.
type resourceManagementDto struct {
	dto.ClientDTO
	Id          *int64  `json:"id,omitempty"`
	ServiceId   *int64  `json:"serviceId,omitempty"`
	PageCode    *int    `json:"pageCode,omitempty"`
	Type        *int    `json:"type,omitempty"`
	Status      *int    `json:"status,omitempty"`
	Name        *string `json:"name,omitempty"`
	StartTime   *string `json:"startTime,omitempty"`
	EndTime     *string `json:"endTime,omitempty"`
	IzLimitTime *bool   `json:"izLimitTime,omitempty"`
	AdvId       *string `json:"advId,omitempty"`
	Title       *string `json:"title,omitempty"`
	Appid       *string `json:"appid,omitempty"`
	SkipUrl     *string `json:"skipUrl,omitempty"`
	Params      *string `json:"params,omitempty"`
	ImgUrl      *string `json:"imgUrl,omitempty"`
	Sort        *int    `json:"sort,omitempty"`
}

// idDTO mirrors IdDTO (@NotNull id).
type idDTO struct {
	dto.ClientDTO
	Id *int64 `json:"id" binding:"required"`
}

// idsDTO mirrors IdsDTO (@NotEmpty ids).
type idsDTO struct {
	dto.ClientDTO
	Ids []int64 `json:"ids" binding:"required"`
}

var idMsgs = map[string]string{
	"traceId":  "must not be empty",
	"tenantId": "must not be empty",
	"id":       "must not be null",
	"ids":      "must not be empty",
}

func protocolByType(c *gin.Context) {
	var req configProtocolDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}
	if req.ServiceId == nil {
		web.WriteParamError(c, "serviceId must not be null")
		return
	}
	if req.Type == nil {
		web.WriteParamError(c, "type must not be null")
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/config/protocol/byType", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertConfigProtocolCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

func protocolDefault(c *gin.Context) {
	var req configProtocolDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}
	if req.Type == nil {
		web.WriteParamError(c, "type must not be null")
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/config/protocol/byType", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertConfigProtocolCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

func creditScoreGetConfig(c *gin.Context) {
	var req dto.ClientDTO
	if !web.BindJSON(c, &req, fenceMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceUser, "/creditScore/v2/getUse", map[string]interface{}{}, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !result.Success {
		web.WriteBizResult(c, result)
		return
	}
	raw := result.Data
	if isJSONNull(raw) {
		writeNPE(c)
		return
	}
	var use struct {
		LowScore     *float64 `json:"lowScore"`
		IzLowScoreOn *int     `json:"izLowScoreOn"`
	}
	if json.Unmarshal(raw, &use) != nil {
		writeNPE(c)
		return
	}
	co := cCreditScoreConfigCO{}
	if use.LowScore != nil {
		co.NoRiddingScore = use.LowScore
	}
	if use.IzLowScoreOn != nil {
		co.IzLowScoreOn = use.IzLowScoreOn
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(co))
}

func resourceAppList(c *gin.Context) {
	var req resourceManagementDto
	if !web.BindJSON(c, &req, nil) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/resource/management/appList", &req, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if result != nil && !result.Success {
		web.WriteBizResult(c, result)
		return
	}
	if result == nil || javacompat.IsNullJSON(result.Data) {
		c.JSON(http.StatusOK, dto.NewSuccessResult([]interface{}{}))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(convertResourceManagementList(result.Data))))
}

func resourceAddExposure(c *gin.Context) {
	var req idsDTO
	if !web.BindJSON(c, &req, idMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/resource/management/addExposure", &req, cmdCtx)
	web.RespondResult(c, result, err)
}

func resourceAddClick(c *gin.Context) {
	var req idDTO
	if !web.BindJSON(c, &req, idMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/resource/management/addClick", &req, cmdCtx)
	web.RespondResult(c, result, err)
}

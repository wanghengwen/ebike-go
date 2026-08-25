// Package operation ports Java RepairController (/client/operation/repair).
package operation

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

var repairMsgs = map[string]string{
	"traceId":   "must not be empty",
	"tenantId":  "must not be empty",
	"carId":     "must not be blank",
	"imei":      "must not be blank",
	"serviceId": "must not be null",
}

// RegisterRoutes wires repair endpoints.
func RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/client/operation/repair")
	{
		g.POST("/add", addRepair)
		g.POST("/page", pageRepair)
	}
}

func addRepair(c *gin.Context) {
	var req dto.UserRepairDTO
	if !web.BindJSON(c, &req, repairMsgs) {
		return
	}
	if !web.NotBlank(c, "carId", req.CarId) || !web.NotBlank(c, "imei", req.Imei) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	raw, ok := javacompat.CallData(c, rpc.ServiceOperation, "/repair/record/add", &req, cmdCtx)
	if !ok {
		return
	}
	var added bool
	_ = json.Unmarshal(raw, &added)
	c.JSON(http.StatusOK, dto.NewSuccessResult(added))
}

func pageRepair(c *gin.Context) {
	req := dto.PageClientDTO{PageNum: 1, PageSize: 10}
	if !web.BindJSON(c, &req, repairMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	body := dto.ConvertToBasePageCmd(&req, cmdCtx)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceOperation, "/repair/record/page", body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertRepairRecordPageDTO(result.Data)
	}
	web.RespondResult(c, result, err)
}

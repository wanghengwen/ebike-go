package management

// Ports Java RepairConfigController (@RequestMapping("/client/management/repairConfig")):
//   - POST /client/management/repairConfig/list      -> ebike-management /repair-config/list
//   - POST /client/management/repairConfig/listByCar -> ebike-management /car-info/detail (lookup)
//                                                       + ebike-management /repair-config/list
//
// IMPORTANT: the Java controller binds @RequestBody WITHOUT @Valid/@Validated,
// so NO bean validation runs (traceId/tenantId may be missing).

import (
	"encoding/json"
	"strings"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/javacompat"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// repairConfigQueryDTO matches Java RepairConfigQueryDTO extends ClientDTO with
// carModel/carId. ClientDTO fields are declared locally (instead of embedding
// dto.ClientDTO) because the shared struct carries binding:"required" tags that
// must NOT apply here (no @Valid in Java).
type repairConfigQueryDTO struct {
	TraceId       string   `json:"traceId"`
	TenantId      string   `json:"tenantId"`
	Platform      string   `json:"platform,omitempty"`
	DeviceId      string   `json:"deviceId,omitempty"`
	Version       string   `json:"version,omitempty"`
	Ip            string   `json:"ip,omitempty"`
	Longitude     *float64 `json:"longitude,omitempty"`
	Latitude      *float64 `json:"latitude,omitempty"`
	Source        string   `json:"source,omitempty"`
	StressTesting bool     `json:"stressTesting"`
	CarModel      *string  `json:"carModel"`
	CarId         *string  `json:"carId"`
}

func (r *repairConfigQueryDTO) toClientDTO() dto.ClientDTO {
	return dto.ClientDTO{
		TraceId:       r.TraceId,
		TenantId:      r.TenantId,
		Platform:      r.Platform,
		DeviceId:      r.DeviceId,
		Version:       r.Version,
		Ip:            r.Ip,
		Longitude:     r.Longitude,
		Latitude:      r.Latitude,
		Source:        r.Source,
		StressTesting: r.StressTesting,
	}
}

// repairConfigList ports RepairConfigController.getList (pure passthrough).
func repairConfigList(c *gin.Context) {
	var req repairConfigQueryDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}
	client := req.toClientDTO()
	cmdCtx := middleware.CompleteCommandContext(c, &client)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceManagement, "/repair-config/list", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertRepairConfigCOList(result.Data)
	}
	web.RespondResult(c, result, err)
}

// repairConfigListByCar ports RepairConfigServiceImpl.listByCar: when carId is
// non-blank, look the car up first (/car-info/detail) and overwrite carModel
// with the car's model — any lookup failure is swallowed (Java catch(Throwable)
// just logs a warning). Then forward to /repair-config/list.
func repairConfigListByCar(c *gin.Context) {
	var req repairConfigQueryDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}
	client := req.toClientDTO()
	cmdCtx := middleware.CompleteCommandContext(c, &client)

	if req.CarId != nil && strings.TrimSpace(*req.CarId) != "" {
		// Java: carInfoGateway.getCar(new CarInfoDto().setCarId(...)) ->
		// CarInfoCmd via BeanUtils (carId copied; primitive num defaults to 0)
		carCmd := map[string]interface{}{
			"carId": *req.CarId,
			"num":   0,
		}
		carResult, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceManagement, "/car-info/detail", carCmd, cmdCtx)
		if err == nil && carResult != nil && carResult.Success &&
			len(carResult.Data) > 0 && string(carResult.Data) != "null" {
			var car struct {
				Model *string `json:"model"`
			}
			if json.Unmarshal(carResult.Data, &car) == nil {
				// Java sets carModel even when the car's model is null
				req.CarModel = car.Model
			}
		}
		// failures fall through, keeping the client-provided carModel (Java
		// logs "get car info error,get repair config by car model")
	}

	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceManagement, "/repair-config/list", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertRepairConfigCOList(result.Data)
	}
	web.RespondResult(c, result, err)
}

type repairConfigCO struct {
	TenantId           *string              `json:"tenantId"`
	CreatedPin         *string              `json:"createdPin"`
	CreatedAt          *javacompat.DateTime `json:"createdAt"`
	UpdatedPin         *string              `json:"updatedPin"`
	UpdatedAt          *javacompat.DateTime `json:"updatedAt"`
	Version            *int                 `json:"version"`
	IzDel              *bool                `json:"izDel"`
	Id                 *javacompat.LongStr  `json:"id"`
	Content            *string              `json:"content"`
	Type               *int                 `json:"type"`
	IzStop             *bool                `json:"izStop"`
	CarModel           *string              `json:"carModel"`
	IzClientUserRepair *bool                `json:"izClientUserRepair"`
}

func convertRepairConfigCOList(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var list []repairConfigCO
	if json.Unmarshal(raw, &list) != nil {
		return raw
	}
	b, _ := json.Marshal(list)
	return b
}

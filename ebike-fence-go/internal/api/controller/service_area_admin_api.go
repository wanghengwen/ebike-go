package controller

import (
	"github.com/gin-gonic/gin"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/pkg/web"
)

func serviceAreaAdminHandlers() map[string]gin.HandlerFunc {
	initFenceAdmin()
	return map[string]gin.HandlerFunc{
		"/serviceArea/getById":                               handleServiceAreaGetByID,
		"/serviceArea/getListByCmd":                          handleServiceAreaGetListByCmd,
		"/serviceArea/getList":                               handleServiceAreaGetList,
		"/serviceArea/createServiceArea":                     handleServiceAreaCreate,
		"/serviceArea/updateServiceArea":                     handleServiceAreaUpdate,
		"/serviceArea/deleteServiceArea":                     handleServiceAreaDelete,
		"/serviceArea/getServiceByLocation":                  handleServiceAreaGetByLocation,
		"/serviceArea/getNearServiceByLocation":              handleServiceAreaGetNearByLocation,
		"/serviceArea/computeOutServiceDistance":             handleServiceAreaComputeDistance,
		"/serviceArea/getServiceAreaByIds":                   handleServiceAreaGetByIDs,
		"/serviceArea/getServiceAreaByRoleIds":               handleServiceAreaGetByRoleIDs,
		"/serviceArea/getServiceAreaByRoleIdsAndSubTenantId": handleServiceAreaGetByRoleIDsSubTenant,
		"/serviceArea/getAllService":                         handleServiceAreaGetAllService,
		"/serviceArea/getParkingPartStatistic":               handleServiceAreaParkingPartStat,
	}
}

func handleServiceAreaGetByID(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := svcAreaAdmin.GetByID(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, res, err)
}

func handleServiceAreaGetListByCmd(c *gin.Context) {
	var req dto.ServiceAreaCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req.Command); !ok {
		return
	}
	res, err := svcAreaAdmin.GetListByCmd(c.Request.Context(), req)
	respondAdmin(c, res, err)
}

func handleServiceAreaGetList(c *gin.Context) {
	var req dto.Command
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req)
	if !ok {
		return
	}
	res, err := svcAreaAdmin.GetList(c.Request.Context(), tenantID)
	respondAdmin(c, res, err)
}

func handleServiceAreaCreate(c *gin.Context) {
	var req dto.ServiceAreaCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	_, err := svcAreaAdmin.Create(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, nil, err)
}

func handleServiceAreaUpdate(c *gin.Context) {
	var req dto.ServiceAreaCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := svcAreaAdmin.Update(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, nil, err)
}

func handleServiceAreaDelete(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	err := svcAreaAdmin.Delete(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, nil, err)
}

func handleServiceAreaGetByLocation(c *gin.Context) {
	var req dto.ServiceAreaLocationCmd
	if !bindJSONCmd(c, &req, map[string]string{
		"commandContext": "must not be null",
	}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := svcAreaAdmin.GetByLocation(c.Request.Context(), tenantID, derefFloat64(req.Lat), derefFloat64(req.Lng))
	respondAdmin(c, res, err)
}

func handleServiceAreaGetNearByLocation(c *gin.Context) {
	var req dto.ServiceAreaLocationCmd
	if !bindJSONCmd(c, &req, map[string]string{
		"commandContext": "must not be null",
	}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := svcAreaAdmin.GetNearByLocation(c.Request.Context(), tenantID, derefFloat64(req.Lat), derefFloat64(req.Lng))
	respondAdmin(c, res, err)
}

func handleServiceAreaComputeDistance(c *gin.Context) {
	var req dto.ComputeDistanceCmd
	if !bindJSONCmd(c, &req, map[string]string{
		"commandContext": "must not be null",
		"imei":           "must not be null",
	}) {
		return
	}
	if req.Point == nil && req.PointJSON == "" {
		web.WriteParamError(c, "point must not be null")
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := svcAreaAdmin.ComputeOutServiceDistance(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleServiceAreaGetByIDs(c *gin.Context) {
	var req dto.IdsCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := svcAreaAdmin.GetByIDs(c.Request.Context(), tenantID, req.Ids)
	respondAdmin(c, res, err)
}

func handleServiceAreaGetByRoleIDs(c *gin.Context) {
	var req dto.IdsCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := svcAreaAdmin.GetByRoleIDs(c.Request.Context(), tenantID, req.CommandContext, req.Ids)
	respondAdmin(c, res, err)
}

func handleServiceAreaGetByRoleIDsSubTenant(c *gin.Context) {
	var req dto.TenantServiceCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := svcAreaAdmin.GetByRoleIDsAndSubTenant(c.Request.Context(), tenantID, cmdPin(c), req.CommandContext, req)
	respondAdmin(c, res, err)
}

func handleServiceAreaGetAllService(c *gin.Context) {
	var req dto.Command
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req); !ok {
		return
	}
	res, err := svcAreaAdmin.GetAllService(c.Request.Context())
	respondAdmin(c, res, err)
}

func handleServiceAreaParkingPartStat(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := svcAreaAdmin.GetParkingPartStatistic(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, res, err)
}

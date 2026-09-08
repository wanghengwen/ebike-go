package controller

import (
	"github.com/gin-gonic/gin"

	"ebike-fence-go/internal/api/dto"
)

func banRidingHandlers() map[string]gin.HandlerFunc {
	initFenceAdmin()
	return map[string]gin.HandlerFunc{
		"/banRiding/getById":               handleBanRidingGetByID,
		"/banRiding/getListByServiceId":    handleBanRidingGetListByServiceID,
		"/banRiding/getNearestBanRiding":   handleBanRidingGetNearest,
		"/banRiding/getNearBanRidingAreas": handleBanRidingGetNearAreas,
		"/banRiding/createBanRiding":       handleBanRidingCreate,
		"/banRiding/updateBanRiding":       handleBanRidingUpdate,
		"/banRiding/deleteBanRiding":       handleBanRidingDelete,
		"/banRiding/deleteBanRidingBatch":  handleBanRidingDeleteBatch,
	}
}

func handleBanRidingGetByID(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := banRidingAdmin.GetByID(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, res, err)
}

func handleBanRidingGetListByServiceID(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := banRidingAdmin.GetListByServiceID(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, res, err)
}

func handleBanRidingGetNearest(c *gin.Context) {
	var req dto.NearLocationCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := banRidingAdmin.GetNearest(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleBanRidingGetNearAreas(c *gin.Context) {
	var req dto.NearLocationCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := banRidingAdmin.GetNearAreas(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleBanRidingCreate(c *gin.Context) {
	var req dto.BanRidingCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := banRidingAdmin.Create(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, res, err)
}

func handleBanRidingUpdate(c *gin.Context) {
	var req dto.BanRidingCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := banRidingAdmin.Update(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, nil, err)
}

func handleBanRidingDelete(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	err := banRidingAdmin.Delete(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, nil, err)
}

func handleBanRidingDeleteBatch(c *gin.Context) {
	var req dto.IdsCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := banRidingAdmin.DeleteBatch(c.Request.Context(), tenantID, req.Ids)
	respondAdmin(c, nil, err)
}

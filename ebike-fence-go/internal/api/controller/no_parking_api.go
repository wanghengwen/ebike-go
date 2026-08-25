package controller

import (
	"github.com/gin-gonic/gin"

	"ebike-fence-go/internal/api/dto"
)

func noParkingHandlers() map[string]gin.HandlerFunc {
	initFenceAdmin()
	return map[string]gin.HandlerFunc{
		"/noParking/getById":              handleNoParkingGetByID,
		"/noParking/getList":              handleNoParkingGetList,
		"/noParking/getListByServiceId":   handleNoParkingGetListByServiceID,
		"/noParking/createNoParking":      handleNoParkingCreate,
		"/noParking/updateNoParking":      handleNoParkingUpdate,
		"/noParking/deleteNoParking":      handleNoParkingDelete,
		"/noParking/getNearNoParking":     handleNoParkingGetNear,
		"/noParking/deleteNoParkingBatch": handleNoParkingDeleteBatch,
	}
}

func handleNoParkingGetByID(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := noParkingAdmin.GetByID(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, res, err)
}

func handleNoParkingGetList(c *gin.Context) {
	var req dto.NoParkingCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req.Command); !ok {
		return
	}
	res, err := noParkingAdmin.GetList(c.Request.Context(), req)
	respondAdmin(c, res, err)
}

func handleNoParkingGetListByServiceID(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := noParkingAdmin.GetListByServiceID(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, res, err)
}

func handleNoParkingCreate(c *gin.Context) {
	var req dto.NoParkingCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := noParkingAdmin.Create(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, res, err)
}

func handleNoParkingUpdate(c *gin.Context) {
	var req dto.NoParkingCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := noParkingAdmin.Update(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, nil, err)
}

func handleNoParkingDelete(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	err := noParkingAdmin.Delete(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, nil, err)
}

func handleNoParkingGetNear(c *gin.Context) {
	var req dto.NearLocationCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := noParkingAdmin.GetNear(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleNoParkingDeleteBatch(c *gin.Context) {
	var req dto.IdsCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := noParkingAdmin.DeleteBatch(c.Request.Context(), tenantID, req.Ids)
	respondAdmin(c, nil, err)
}

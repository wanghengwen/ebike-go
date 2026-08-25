package controller

import (
	"github.com/gin-gonic/gin"

	"ebike-fence-go/internal/api/dto"
)

func maintainAreaHandlers() map[string]gin.HandlerFunc {
	initFenceAdmin()
	return map[string]gin.HandlerFunc{
		"/maintain/area/create":                 handleMaintainAreaCreate,
		"/maintain/area/update":                 handleMaintainAreaUpdate,
		"/maintain/area/delete":                 handleMaintainAreaDelete,
		"/maintain/area/deleteBatch":            handleMaintainAreaDeleteBatch,
		"/maintain/area/listByServiceId":        handleMaintainAreaListByServiceID,
		"/maintain/area/pageListByServiceId":    handleMaintainAreaPageList,
		"/maintain/area/personnelManagementList": handleMaintainAreaPersonnelList,
		"/maintain/area/personnelDetail":        handleMaintainAreaPersonnelDetail,
		"/maintain/area/personnelEdit":          handleMaintainAreaPersonnelEdit,
		"/maintain/area/personnelDelete":        handleMaintainAreaPersonnelDelete,
		"/maintain/area/getByUserPin":           handleMaintainAreaGetByUserPin,
	}
}

func handleMaintainAreaCreate(c *gin.Context) {
	var req dto.MaintainAreaCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	_, err := maintainAdmin.Create(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, nil, err)
}

func handleMaintainAreaUpdate(c *gin.Context) {
	var req dto.MaintainAreaCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := maintainAdmin.Update(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, nil, err)
}

func handleMaintainAreaDelete(c *gin.Context) {
	var req dto.MaintainAreaCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := maintainAdmin.Delete(c.Request.Context(), tenantID, req.Id)
	respondAdmin(c, nil, err)
}

func handleMaintainAreaDeleteBatch(c *gin.Context) {
	var req dto.IdsCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := maintainAdmin.DeleteBatch(c.Request.Context(), tenantID, req.Ids)
	respondAdmin(c, nil, err)
}

func handleMaintainAreaListByServiceID(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	res, err := maintainAdmin.ListByServiceID(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, res, err)
}

func handleMaintainAreaPageList(c *gin.Context) {
	var req dto.MaintainAreaPageQuery
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := maintainAdmin.PageListByServiceID(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleMaintainAreaPersonnelList(c *gin.Context) {
	var req dto.MaintainAreaPageQuery
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := maintainAdmin.PersonnelManagementList(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleMaintainAreaPersonnelDetail(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req.Command); !ok || !requireID(c, req.Id) {
		return
	}
	res, err := maintainAdmin.PersonnelDetail(c.Request.Context(), *req.Id)
	respondAdmin(c, res, err)
}

func handleMaintainAreaPersonnelEdit(c *gin.Context) {
	var req dto.PersonnelEditCMD
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := maintainAdmin.PersonnelEdit(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, nil, err)
}

func handleMaintainAreaPersonnelDelete(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	_, ok := applyCmd(c, &req.Command)
	if !ok || !requireID(c, req.Id) {
		return
	}
	err := maintainAdmin.PersonnelDelete(c.Request.Context(), cmdPin(c), *req.Id)
	respondAdmin(c, nil, err)
}

func handleMaintainAreaGetByUserPin(c *gin.Context) {
	var req dto.AreaEmployeeCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := maintainAdmin.GetByUserPin(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

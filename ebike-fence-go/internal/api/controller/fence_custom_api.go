package controller

import (
	"github.com/gin-gonic/gin"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/pkg/web"
)

func fenceCustomHandlers() map[string]gin.HandlerFunc {
	initFenceAdmin()
	return map[string]gin.HandlerFunc{
		"/fence/custom/type/create": handleFenceCustomTypeCreate,
		"/fence/custom/type/update": handleFenceCustomTypeUpdate,
		"/fence/custom/type/delete": handleFenceCustomTypeDelete,
		"/fence/custom/type/list":   handleFenceCustomTypeList,
		"/fence/custom/create":      handleFenceCustomCreate,
		"/fence/custom/update":      handleFenceCustomUpdate,
		"/fence/custom/delete":      handleFenceCustomDelete,
		"/fence/custom/list":        handleFenceCustomList,
		"/fence/custom/allTypeGeo":  handleFenceCustomAllTypeGeo,
	}
}

func handleFenceCustomTypeCreate(c *gin.Context) {
	var req dto.FenceCustomTypeCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null", "name": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := customAdmin.CreateType(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, 1, err)
}

func handleFenceCustomTypeUpdate(c *gin.Context) {
	var req dto.FenceCustomTypeCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null", "id": "must not be null", "name": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := customAdmin.UpdateType(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, 1, err)
}

func handleFenceCustomTypeDelete(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	if req.Id == nil {
		web.WriteParamError(c, "id must not be null")
		return
	}
	err := customAdmin.DeleteType(c.Request.Context(), tenantID, *req.Id)
	respondAdmin(c, 1, err)
}

func handleFenceCustomTypeList(c *gin.Context) {
	var req dto.Command
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req)
	if !ok {
		return
	}
	res, err := customAdmin.ListTypes(c.Request.Context(), tenantID)
	respondAdmin(c, res, err)
}

func handleFenceCustomCreate(c *gin.Context) {
	var req dto.FenceCustomCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := customAdmin.CreateFence(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, 1, err)
}

func handleFenceCustomUpdate(c *gin.Context) {
	var req dto.FenceCustomCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	if req.Id == 0 {
		web.WriteParamError(c, "id must not be null")
		return
	}
	err := customAdmin.UpdateFence(c.Request.Context(), tenantID, cmdPin(c), req)
	respondAdmin(c, 1, err)
}

func handleFenceCustomDelete(c *gin.Context) {
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	if req.Id == nil {
		web.WriteParamError(c, "id must not be null")
		return
	}
	err := customAdmin.DeleteFence(c.Request.Context(), tenantID, *req.Id, 0)
	respondAdmin(c, 1, err)
}

func handleFenceCustomList(c *gin.Context) {
	var req dto.CustomFenceListQry
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null", "serviceId": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := customAdmin.ListFences(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

func handleFenceCustomAllTypeGeo(c *gin.Context) {
	var req dto.CustomFenceListQry
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null", "serviceId": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := customAdmin.AllTypeGeo(c.Request.Context(), tenantID, req)
	respondAdmin(c, res, err)
}

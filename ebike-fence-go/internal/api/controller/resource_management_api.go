package controller

import (
	"ebike-fence-go/internal/api/dto"
	configsvc "ebike-fence-go/internal/domain/service/config"
	"ebike-fence-go/internal/infrastructure/persistence/repo"

	"github.com/gin-gonic/gin"
)

func resourceManagementHandlers() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"/resource/management/create":       handleResourceManagementCreate,
		"/resource/management/update":       handleResourceManagementUpdate,
		"/resource/management/delete":       handleResourceManagementDelete,
		"/resource/management/list":         handleResourceManagementList,
		"/resource/management/page":         handleResourceManagementPage,
		"/resource/management/appList":      handleResourceManagementAppList,
		"/resource/management/sort":         handleResourceManagementSort,
		"/resource/management/batchOffline": handleResourceManagementBatchOffline,
		"/resource/management/batchOnline":  handleResourceManagementBatchOnline,
		"/resource/management/addExposure":  handleResourceManagementAddExposure,
		"/resource/management/addClick":     handleResourceManagementAddClick,
	}
}

func handleResourceManagementCreate(c *gin.Context) {
	initConfigServices()
	var req dto.ResourceManagementCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := configsvc.Services.ResourceManagement.Create(c.Request.Context(), tenantID, repo.PinFromContext(c), req)
	respondConfig(c, nil, err)
}

func handleResourceManagementUpdate(c *gin.Context) {
	initConfigServices()
	var req dto.ResourceManagementCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantID, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	err := configsvc.Services.ResourceManagement.Update(c.Request.Context(), tenantID, repo.PinFromContext(c), req)
	respondConfig(c, nil, err)
}

func handleResourceManagementDelete(c *gin.Context) {
	initConfigServices()
	var req dto.IdsCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req.Command); !ok {
		return
	}
	err := configsvc.Services.ResourceManagement.Delete(c.Request.Context(), req.Ids)
	respondConfig(c, nil, err)
}

func handleResourceManagementList(c *gin.Context) {
	initConfigServices()
	var req dto.ResourceManagementCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req.Command); !ok {
		return
	}
	res, err := configsvc.Services.ResourceManagement.List(c.Request.Context(), req)
	respondConfig(c, res, err)
}

func handleResourceManagementPage(c *gin.Context) {
	initConfigServices()
	var req dto.ResourceManagementPageQuery
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req.Command); !ok {
		return
	}
	res, err := configsvc.Services.ResourceManagement.Page(c.Request.Context(), req)
	respondConfig(c, res, err)
}

func handleResourceManagementAppList(c *gin.Context) {
	initConfigServices()
	var req dto.ResourceManagementCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req.Command); !ok {
		return
	}
	res, err := configsvc.Services.ResourceManagement.AppList(c.Request.Context(), req)
	respondConfig(c, res, err)
}

func handleResourceManagementSort(c *gin.Context) {
	initConfigServices()
	var req dto.IdsCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req.Command); !ok {
		return
	}
	err := configsvc.Services.ResourceManagement.Sort(c.Request.Context(), repo.PinFromContext(c), req.Ids)
	respondConfig(c, nil, err)
}

func handleResourceManagementBatchOffline(c *gin.Context) {
	initConfigServices()
	var req dto.IdsCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req.Command); !ok {
		return
	}
	err := configsvc.Services.ResourceManagement.BatchOffline(c.Request.Context(), req.Ids)
	respondConfig(c, nil, err)
}

func handleResourceManagementBatchOnline(c *gin.Context) {
	initConfigServices()
	var req dto.IdsCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req.Command); !ok {
		return
	}
	err := configsvc.Services.ResourceManagement.BatchOnline(c.Request.Context(), req.Ids)
	respondConfig(c, nil, err)
}

func handleResourceManagementAddExposure(c *gin.Context) {
	initConfigServices()
	var req dto.IdsCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req.Command); !ok {
		return
	}
	okVal, err := configsvc.Services.ResourceManagement.AddExposure(c.Request.Context(), req.Ids)
	respondConfig(c, okVal, err)
}

func handleResourceManagementAddClick(c *gin.Context) {
	initConfigServices()
	var req dto.IdCmd
	if !bindJSONCmd(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	if _, ok := applyCmd(c, &req.Command); !ok || !requireID(c, req.Id) {
		return
	}
	pin := repo.PinFromContext(c)
	okVal, err := configsvc.Services.ResourceManagement.AddClick(c.Request.Context(), pin, *req.Id)
	respondConfig(c, okVal, err)
}

package controller

import (
	"context"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/service"
	configsvc "ebike-fence-go/internal/domain/service/config"
	"ebike-fence-go/internal/pkg/web"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type helmetConfigQueryAdapter struct{}

func (helmetConfigQueryAdapter) GetUseCarConfig(ctx context.Context, tenantID string, serviceID int64) (*dto.ConfigUseCarCO, error) {
	return configsvc.Services.UseCar.Get(ctx, tenantID, serviceID)
}

func (helmetConfigQueryAdapter) GetBackcarConfig(ctx context.Context, tenantID string, serviceID int64) (*dto.ConfigBackcarCO, error) {
	return configsvc.Services.Backcar.GetByServiceID(ctx, tenantID, serviceID)
}

func (helmetConfigQueryAdapter) GetBaseItemConfig(ctx context.Context, tenantID string, serviceID int64) (*dto.ConfigBaseItemCO, error) {
	return configsvc.Services.BaseItem.Get(ctx, tenantID, serviceID)
}

func ensureHelmetConfigQuery() {
	initConfigServices()
	if configsvc.Services != nil {
		globalHelmetService.SetConfigQuery(helmetConfigQueryAdapter{})
	}
}

func helmetHandlers() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"/helmet/unlock":            HelmetUnlock,
		"/helmet/lock":              HelmetLock,
		"/helmet/getHelmetLock":     HelmetGetHelmetLock,
		"/helmet/getHelmetReact":    HelmetGetHelmetReact,
		"/helmet/del":               HelmetDel,
		"/helmet/accUnlock":         HelmetAccUnlock,
		"/helmet/helmetStateChange": HelmetStateChange,
	}
}

func bindHelmetCmd(c *gin.Context) (dto.HelmetCmd, string, bool) {
	var req dto.HelmetCmd
	if !web.BindJSON(c, &req, map[string]string{
		"commandContext": "must not be null",
		"carId":          "must not be null",
	}) {
		return req, "", false
	}
	tenantId, ok := applyCmd(c, &req.Command)
	return req, tenantId, ok
}

func writeHelmetErr(c *gin.Context, err error) {
	if biz, ok := err.(*service.BizError); ok {
		web.WriteBizError(c, biz.Code, biz.Msg)
		return
	}
	zap.L().Error("helmet api failed", zap.Error(err))
	web.WriteException(c, "Internal Server Error")
}

func HelmetUnlock(c *gin.Context) {
	initGlobals()
	req, tenantId, ok := bindHelmetCmd(c)
	if !ok {
		return
	}
	res, err := globalHelmetService.Unlock(c.Request.Context(), tenantId, req)
	if err != nil {
		writeHelmetErr(c, err)
		return
	}
	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}

func HelmetLock(c *gin.Context) {
	initGlobals()
	req, tenantId, ok := bindHelmetCmd(c)
	if !ok {
		return
	}
	res, err := globalHelmetService.Lock(c.Request.Context(), tenantId, req)
	if err != nil {
		writeHelmetErr(c, err)
		return
	}
	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}

func HelmetGetHelmetLock(c *gin.Context) {
	initGlobals()
	req, tenantId, ok := bindHelmetCmd(c)
	if !ok {
		return
	}
	res, err := globalHelmetService.GetRecordHelmetLock(c.Request.Context(), tenantId, req.CarId)
	if err != nil {
		writeHelmetErr(c, err)
		return
	}
	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}

func HelmetGetHelmetReact(c *gin.Context) {
	initGlobals()
	req, tenantId, ok := bindHelmetCmd(c)
	if !ok {
		return
	}
	res, err := globalHelmetService.GetRecordHelmetReact(c.Request.Context(), tenantId, req.CarId)
	if err != nil {
		writeHelmetErr(c, err)
		return
	}
	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}

func HelmetDel(c *gin.Context) {
	initGlobals()
	req, tenantId, ok := bindHelmetCmd(c)
	if !ok {
		return
	}
	if err := globalHelmetService.DelHelmet(c.Request.Context(), tenantId, req.CarId); err != nil {
		writeHelmetErr(c, err)
		return
	}
	result := dto.NewSuccessResult(nil)
	web.RespondResult(c, &result, nil)
}

func HelmetAccUnlock(c *gin.Context) {
	initGlobals()
	req, tenantId, ok := bindHelmetCmd(c)
	if !ok {
		return
	}
	if req.ServiceId == nil {
		web.WriteParamError(c, "serviceId must not be null")
		return
	}
	res, err := globalHelmetService.AccUnlockByHelmetState(c.Request.Context(), tenantId, req.CommandContext, req.Imei, *req.ServiceId)
	if err != nil {
		writeHelmetErr(c, err)
		return
	}
	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}

func HelmetStateChange(c *gin.Context) {
	initGlobals()
	ensureHelmetConfigQuery()
	var req dto.IdCmd
	if !web.BindJSON(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return
	}
	tenantId, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	serviceId := req.Id
	if serviceId == nil {
		serviceId = req.ServiceId
	}
	if serviceId == nil {
		web.WriteParamError(c, "id must not be null")
		return
	}
	res, err := globalHelmetService.HelmetStateChange(c.Request.Context(), tenantId, *serviceId)
	if err != nil {
		writeHelmetErr(c, err)
		return
	}
	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}

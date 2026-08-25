package controller

import (
	"context"
	"strings"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/domain/service"
	"ebike-fence-go/internal/pkg/web"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func partHandlers() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"/checkPart":                        CheckPart,
		"/checkPartByApp":                   CheckPartByApp,
		"/part/checkHelmet":                 PartCheckHelmet,
		"/part/checkHelmetByTempParking":    PartCheckHelmetByTempParking,
		"/part/checkHelmetState":            PartCheckHelmetState,
		"/part/checkKickstand":              PartCheckKickstand,
		"/part/setHelmetAudit":              PartSetHelmetAudit,
		"/part/setPointAudit":               PartSetPointAudit,
		"/part/setTBeaconAudit":             PartSetTBeaconAudit,
		"/part/setDirectionAudit":           PartSetDirectionAudit,
		"/part/setKickstandAudit":           PartSetKickstandAudit,
		"/part/setCameraAudit":              PartSetCameraAudit,
		"/part/setLocationAudit":            PartSetLocationAudit,
		"/part/izCanCameraAudit":            PartIzCanCameraAudit,
		"/part/getBlueTBeaconInfoFromCache": PartGetBlueTBeaconInfoFromCache,
	}
}

func bindCheckingPart(c *gin.Context) (dto.CheckingPartCmd, string, bool) {
	var req dto.CheckingPartCmd
	if !web.BindJSON(c, &req, map[string]string{"commandContext": "must not be null"}) {
		return req, "", false
	}
	tenantId, ok := applyCmd(c, &req.Command)
	return req, tenantId, ok
}

func writePartErr(c *gin.Context, err error) {
	if biz, ok := err.(*service.BizError); ok {
		web.WriteBizError(c, biz.Code, biz.Msg)
		return
	}
	zap.L().Error("part api failed", zap.Error(err))
	web.WriteException(c, "Internal Server Error")
}

// requireCheckingPartImei mirrors Java PartApi @Validated endpoints that reject blank imei.
func requireCheckingPartImei(c *gin.Context, imei string) bool {
	if strings.TrimSpace(imei) == "" {
		web.WriteParamError(c, "imei 不能为空")
		return false
	}
	return true
}

func CheckPart(c *gin.Context) {
	initGlobals()
	req, tenantId, ok := bindCheckingPart(c)
	if !ok {
		return
	}
	res, err := globalPartService.PartAnalysisList(c.Request.Context(), tenantId, req)
	if err != nil {
		writePartErr(c, err)
		return
	}
	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}

func CheckPartByApp(c *gin.Context) {
	initGlobals()
	req, tenantId, ok := bindCheckingPart(c)
	if !ok {
		return
	}
	res, err := globalPartService.PartAnalysisByApp(c.Request.Context(), tenantId, req)
	if err != nil {
		writePartErr(c, err)
		return
	}
	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}

func PartCheckHelmet(c *gin.Context) {
	initGlobals()
	req, tenantId, ok := bindCheckingPart(c)
	if !ok {
		return
	}
	if !requireCheckingPartImei(c, req.Imei) {
		return
	}
	res, err := globalPartService.CheckHelmet(c.Request.Context(), tenantId, req)
	if err != nil {
		writePartErr(c, err)
		return
	}
	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}

func PartCheckHelmetByTempParking(c *gin.Context) {
	initGlobals()
	req, tenantId, ok := bindCheckingPart(c)
	if !ok {
		return
	}
	// Java checkHelmetByTempParking uses carId only; CheckingPartCmd.imei has no @NotBlank.
	res, err := globalPartService.CheckHelmetByTempParking(c.Request.Context(), tenantId, req)
	if err != nil {
		writePartErr(c, err)
		return
	}
	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}

func PartCheckKickstand(c *gin.Context) {
	initGlobals()
	req, _, ok := bindCheckingPart(c)
	if !ok {
		return
	}
	if !requireCheckingPartImei(c, req.Imei) {
		return
	}
	res, err := globalPartService.CheckKickstand(c.Request.Context(), req)
	if err != nil {
		writePartErr(c, err)
		return
	}
	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}

func PartCheckHelmetState(c *gin.Context) {
	initGlobals()
	var req dto.HelmetCheckCmd
	if !web.BindJSON(c, &req, map[string]string{
		"commandContext": "must not be null",
		"imei":           "must not be null",
		"serviceId":      "must not be null",
	}) {
		return
	}
	tenantId, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := globalPartService.CheckHelmetState(c.Request.Context(), tenantId, req)
	if err != nil {
		writePartErr(c, err)
		return
	}
	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}

func partSetAudit(c *gin.Context, setFn func(ctx context.Context, tenantId, carId string, orderId int64) error) {
	initGlobals()
	req, tenantId, ok := bindCheckingPart(c)
	if !ok {
		return
	}
	if req.OrderId == nil {
		web.WriteParamError(c, "orderId must not be null")
		return
	}
	if err := setFn(c.Request.Context(), tenantId, req.CarId, *req.OrderId); err != nil {
		writePartErr(c, err)
		return
	}
	result := dto.NewSuccessResult(nil)
	web.RespondResult(c, &result, nil)
}

func PartSetHelmetAudit(c *gin.Context) {
	partSetAudit(c, func(ctx context.Context, tenantId, carId string, orderId int64) error {
		return globalPartService.SetHelmetAudit(ctx, tenantId, carId, orderId)
	})
}

func PartSetPointAudit(c *gin.Context) {
	partSetAudit(c, func(ctx context.Context, tenantId, carId string, orderId int64) error {
		return globalPartService.SetPointAudit(ctx, tenantId, carId, orderId)
	})
}

func PartSetTBeaconAudit(c *gin.Context) {
	partSetAudit(c, func(ctx context.Context, tenantId, carId string, orderId int64) error {
		return globalPartService.SetTBeaconAudit(ctx, tenantId, carId, orderId)
	})
}

func PartSetDirectionAudit(c *gin.Context) {
	partSetAudit(c, func(ctx context.Context, tenantId, carId string, orderId int64) error {
		return globalPartService.SetDirectionAudit(ctx, tenantId, carId, orderId)
	})
}

func PartSetKickstandAudit(c *gin.Context) {
	partSetAudit(c, func(ctx context.Context, tenantId, carId string, orderId int64) error {
		return globalPartService.SetKickstandAudit(ctx, tenantId, carId, orderId)
	})
}

func PartSetCameraAudit(c *gin.Context) {
	partSetAudit(c, func(ctx context.Context, tenantId, carId string, orderId int64) error {
		return globalPartService.SetCameraAudit(ctx, tenantId, carId, orderId)
	})
}

func PartSetLocationAudit(c *gin.Context) {
	partSetAudit(c, func(ctx context.Context, tenantId, carId string, orderId int64) error {
		return globalPartService.SetLocationAudit(ctx, tenantId, carId, orderId)
	})
}

func PartIzCanCameraAudit(c *gin.Context) {
	initGlobals()
	req, tenantId, ok := bindCheckingPart(c)
	if !ok {
		return
	}
	res, err := globalPartService.IzCanCameraAudit(c.Request.Context(), tenantId, req.CarId, req.OrderId, req.ServiceId)
	if err != nil {
		writePartErr(c, err)
		return
	}
	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}

func PartGetBlueTBeaconInfoFromCache(c *gin.Context) {
	initGlobals()
	var req dto.ImeiCmd
	if !web.BindJSON(c, &req, map[string]string{
		"commandContext": "must not be null",
		"imei":           "must not be null",
	}) {
		return
	}
	tenantId, ok := applyCmd(c, &req.Command)
	if !ok {
		return
	}
	res, err := globalPartService.GetBlueTBeaconInfoFromCache(c.Request.Context(), tenantId, req.Imei)
	if err != nil {
		writePartErr(c, err)
		return
	}
	result := dto.NewSuccessResult(res)
	web.RespondResult(c, &result, nil)
}

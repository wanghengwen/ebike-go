package controller

import (
	"ebike-fence-go/internal/api"
	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/middleware"
	"ebike-fence-go/internal/pkg/web"

	"github.com/gin-gonic/gin"
)

func registerNativeHandlers() {
	handlers := map[string]gin.HandlerFunc{
		"/serviceArea/returnCar":       ReturnCar,
		"/serviceArea/ridingCar":       RidingCar,
		"/serviceArea/getFenceRelation": GetFenceRelation,
		"/actuator/deregisterService":  DeregisterService,
	}
	mergeHandlers(handlers, helpConfigHandlers())
	mergeHandlers(handlers, resourceManagementHandlers())
	mergeHandlers(handlers, configHandlers())
	mergeHandlers(handlers, serviceAreaAdminHandlers())
	mergeHandlers(handlers, parkingHandlers())
	mergeHandlers(handlers, noParkingHandlers())
	mergeHandlers(handlers, banRidingHandlers())
	mergeHandlers(handlers, maintainAreaHandlers())
	mergeHandlers(handlers, fenceCustomHandlers())
	mergeHandlers(handlers, fenceRfidHandlers())
	mergeHandlers(handlers, fenceBusinessHandlers())
	mergeHandlers(handlers, siteApplicationHandlers())
	mergeHandlers(handlers, partHandlers())
	mergeHandlers(handlers, helmetHandlers())
	api.RegisterHandlers(handlers)
}

func mergeHandlers(dst map[string]gin.HandlerFunc, src map[string]gin.HandlerFunc) {
	for k, v := range src {
		dst[k] = v
	}
}

func applyCmd(c *gin.Context, cmd *dto.Command) (tenantID string, ok bool) {
	cmdCtx, ok := middleware.ApplyCommand(c, cmd)
	if !ok {
		return "", false
	}
	return cmdCtx.TenantId, true
}

func bindJSONCmd[T any](c *gin.Context, req *T, msgs map[string]string) bool {
	return web.BindJSON(c, req, msgs)
}

// requireID mirrors Java @NotNull Long id on IdCmd. Probe raw JSON first so
// jsoniter fuzzy decoding cannot turn "" into 0 and bypass the null check.
func requireID(c *gin.Context, id *int64) bool {
	if !web.RequireJSONFieldNotNull(c, "id", "must not be null") {
		return false
	}
	if id != nil {
		return true
	}
	web.WriteParamError(c, "id must not be null")
	return false
}

// requireServiceID mirrors Java @NotNull Long serviceId.
func requireServiceID(c *gin.Context, serviceID *int64) bool {
	if !web.RequireJSONFieldNotNull(c, "serviceId", "must not be null") {
		return false
	}
	if serviceID != nil {
		return true
	}
	web.WriteParamError(c, "serviceId must not be null")
	return false
}

// derefFloat64 safely dereferences a *float64, returning 0 if nil.
// Used when the APP hasn't acquired GPS and omits lat/lng from the request.
func derefFloat64(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

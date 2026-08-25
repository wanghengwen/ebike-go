package clientconfig

import (
	"encoding/json"
	"net/http"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/middleware"
	"ebike-service-client-go/internal/pkg/rpc"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

func successInt1(c *gin.Context) {
	c.JSON(http.StatusOK, dto.NewSuccessResult(1))
}

// --- AdConfig ---

func getAdConfig(c *gin.Context) {
	var req getConfigDTO
	if !bindGetConfigDTO(c, &req) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/config/getAd", &req, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if err == nil && result != nil && result.Success {
		result.Data = convertAdConfigDTO(result.Data)
	}
	web.RespondResult(c, result, err)
}

func insAdConfig(c *gin.Context) {
	var req getConfigDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	if _, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/config/insAd", &req, cmdCtx); err != nil {
		web.WriteException(c, err.Error())
		return
	}
	successInt1(c)
}

func updAdConfig(c *gin.Context) {
	var req getConfigDTO
	if !web.BindJSON(c, &req, nil) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	if _, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/config/updAd", &req, cmdCtx); err != nil {
		web.WriteException(c, err.Error())
		return
	}
	successInt1(c)
}

func adSaveOrUpdate(c *gin.Context) {
	var req adsConfigDTO
	if !web.BindJSON(c, &req, map[string]string{"serviceId": "不能为null", "traceId": "must not be empty", "tenantId": "must not be empty"}) {
		return
	}
	body := map[string]json.RawMessage{}
	b, _ := json.Marshal(req)
	_ = json.Unmarshal(b, &body)
	if req.Ads != nil {
		body["ads"] = req.Ads
	}
	if req.HalfMiniProgramAds != nil {
		body["halfMiniProgramAds"] = req.HalfMiniProgramAds
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/ad_config/saveOrUpdate", body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertBoolean(result.Data)
	}
	web.RespondResult(c, result, err)
}

func adDetail(c *gin.Context) {
	if !web.RequireJSONFieldNotNull(c, "serviceId", "不能为null") {
		return
	}
	var req dto.ServiceDTO
	if !web.BindJSON(c, &req, serviceDTOBindMsgs) {
		return
	}
	if req.ServiceId == nil {
		web.WriteParamError(c, "serviceId 不能为null")
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/ad_config/detail", &req, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertAdsConfigCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

// --- Helmet ---

func helmetUnlock(c *gin.Context) {
	helmetAction(c, "/helmet/unlock")
}

func helmetLock(c *gin.Context) {
	helmetAction(c, "/helmet/lock")
}

func helmetAction(c *gin.Context, path string) {
	var req helmetCmd
	if !web.BindJSON(c, &req, nil) {
		return
	}
	body := map[string]string{"carId": req.CarId}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, path, body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertBoolean(result.Data)
	}
	web.RespondResult(c, result, err)
}

// --- RidingPermission ---

func ridingPermissionGet(c *gin.Context) {
	var req getConfigDTO
	if !web.BindJSON(c, &req, getConfigMsgs) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/ridingPermission/get", &req, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !result.Success {
		web.WriteBizResult(c, result)
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResult(json.RawMessage(convertRidingPermissionCO(result.Data))))
}

// --- SiteApplication ---

func siteApplication(c *gin.Context) {
	var req siteApplicationDTO
	if !web.BindJSON(c, &req, map[string]string{
		"traceId": "must not be empty", "tenantId": "must not be empty",
		"serviceId": "must not be null", "location": "must not be blank",
		"lat": "must not be null", "lng": "must not be null", "photoUrl": "must not be blank",
	}) {
		return
	}
	if !web.NotBlank(c, "location", req.Location) || !web.NotBlank(c, "photoUrl", req.PhotoUrl) {
		return
	}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/SiteApplication/SiteApplication", &req, cmdCtx)
	if err != nil {
		web.WriteException(c, err.Error())
		return
	}
	if !result.Success {
		// Java: SiteApplicationGatewayImpl.application -> ResultHelper.getResultData
		web.WriteBizResult(c, result)
		return
	}
	successInt1(c)
}

func applicationSiteGetConfig(c *gin.Context) {
	var req configParkApplyDTO
	if !web.BindJSON(c, &req, getConfigMsgs) {
		return
	}
	body := map[string]*int64{"id": req.ServiceId}
	cmdCtx := middleware.CompleteCommandContext(c, &req.ClientDTO)
	result, err := rpc.ForwardCommand(c.Request.Context(), rpc.ServiceFence, "/config/parkApply/getByServiceId", body, cmdCtx)
	if err == nil && result != nil && result.Success {
		result.Data = convertConfigParkApplyCO(result.Data)
	}
	web.RespondResult(c, result, err)
}

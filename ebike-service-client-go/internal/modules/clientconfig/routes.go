// Package clientconfig ports Java controllers bound under /client that are not
// covered by user/order/fence modules: AdConfig, Helmet, HelpConfig,
// SystemConfig, RidingPermission, SiteApplication.
package clientconfig

import "github.com/gin-gonic/gin"

// RegisterRoutes wires all clientconfig endpoints.
func RegisterRoutes(r *gin.RouterGroup) {
	client := r.Group("/client")
	{
		// AdConfigController
		client.POST("/helpConfig/getAdConfig", getAdConfig)
		client.POST("/config/getAd", getAdConfig)
		client.POST("/config/insAd", insAdConfig)
		client.POST("/config/updAd", updAdConfig)
		client.POST("/ad_config/saveOrUpdate", adSaveOrUpdate)
		client.POST("/ad_config/detail", adDetail)

		// HelmetController (no @Validated)
		client.POST("/helmet/unlock", helmetUnlock)
		client.POST("/helmet/lock", helmetLock)

		// HelpConfigController
		client.POST("/helpConfig/getFaqByServiceId", getFaqByServiceId)
		client.POST("/helpConfig/getFaqById", getFaqById)
		client.POST("/helpConfig/getHomeScrollerMsgByServiceId", getHomeScrollerMsgByServiceId)
		client.POST("/helpConfig/getHomeScrollerMsgByServiceId/v2", getHomeScrollerMsgByServiceIdV2)
		client.POST("/helpConfig/getHomeScrollerMsgById", getHomeScrollerMsgById)
		client.POST("/helpConfig/getGuidePageConfigByServiceId", getGuidePageConfigByServiceId)
		client.POST("/helpConfig/getSpecialTipsByServiceId", getSpecialTipsByServiceId)
		client.POST("/helpConfig/getCustomerServiceByServiceId", getCustomerServiceByServiceId)
		client.POST("/helpConfig/getHomeActivityEntranceByServiceId", getHomeActivityEntranceByServiceId)
		client.POST("/helpConfig/getHomeActivityById", getHomeActivityById)
		client.POST("/helpConfig/getHomeNavByServiceId", getHomeNavByServiceId)
		client.POST("/helpConfig/getHomeNavById", getHomeNavById)
		client.POST("/helpConfig/getIzMainPush", getIzMainPush)

		// SystemConfigController
		client.POST("/system/getUseCarConfig", getUseCarConfig)
		client.POST("/system/getbackCarConfig", getBackCarConfig)
		client.POST("/system/getbackCarConfigByCarId", getBackCarConfigByCarId)
		client.POST("/systemConfig/getConfigPay", getConfigPay)
		client.POST("/systemConfig/getConfigBaseItem", getConfigBaseItem)

		// RidingPermissionController
		client.POST("/ridingPermission/get", ridingPermissionGet)

		// SiteApplicationController
		client.POST("/SiteApplication/SiteApplication", siteApplication)
		client.POST("/applicationSite/getConfig", applicationSiteGetConfig)
	}
}

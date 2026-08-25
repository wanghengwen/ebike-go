// Package user ports the Java controllers UserController, UserNameAuthController,
// UserCareerController, ChangeBindController, CreditScoreController and
// TenantController (all class-level mapped under /client...).
package user

import "github.com/gin-gonic/gin"

// RegisterRoutes wires every endpoint of the user module.
func RegisterRoutes(r *gin.RouterGroup) {
	client := r.Group("/client")
	{
		// UserController  @RequestMapping("/client")
		client.POST("/user/user/location", location)
		client.POST("/user/user/personInfo", personInfo)
		client.POST("/user/user/thirdBindInfo", thirdBindInfo)
		client.POST("/user/user/qualificationList", qualificationList)
		client.POST("/user/user/cancelAccount", cancelAccount)
		client.POST("/user/user/submitCancel", submitCancel)
		client.POST("/user/user/update", updateUser)
		client.POST("/user/user/canRide", canRide)
		client.POST("/user/blacklist/info", blacklistInfo)
		client.POST("/user/cancel/smsCode", cancelSmsCode)
		client.POST("/gray/queryOrder", grayOrderQuery)
		client.POST("/gray/userData", grayUserData)

		// UserNameAuthController  @RequestMapping("/client")
		client.POST("/user/auth", nameAuth)
		client.POST("/user/rentCheck", rentCheck)
		client.POST("/user/auth/upload", uploadAuth)
		client.POST("/user/auth/cancel", cancelAuth)
		client.POST("/user/auth/izNeed", izNeedAuth)
		client.POST("/user/auth/state", authState)

		// UserCareerController  @RequestMapping("/client")
		client.POST("/user/career/upload", uploadCareer)
		client.POST("/user/career/cancel", cancelCareer)
		client.POST("/user/config/enable", configEnable)

		// ChangeBindController  @RequestMapping("/client")
		// NOTE: "/user/changBind/add" preserves the Java path typo (changBind).
		client.POST("/user/changBind/add", addChangeBind)
		client.POST("/user/changeBind/checkPreviousPhone", checkPreviousPhone)
		client.POST("/user/changeBind/withFace", changeBindWithFace)
		client.POST("/user/changeBind/withPhoneAviodAudit", changeBindPhoneAviodAudit)
		client.POST("/user/changeBind/cancel", cancelChangeBind)

		// CreditScoreController  @RequestMapping("/client/ebike-user")
		client.POST("/ebike-user/creditScore/noRidding/info", creditScoreInfo)
		client.POST("/ebike-user/creditScore/list", creditScoreList)
		client.POST("/ebike-user/creditScore/page", creditScorePage)

		// TenantController  @RequestMapping("/client/tenant")
		client.POST("/tenant/config", tenantConfig)
	}
}

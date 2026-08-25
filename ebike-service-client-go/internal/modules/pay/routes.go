package pay

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires all pay-module routes.
//
// Java sources:
//   - PayController                  @RequestMapping("/client/ebike-pay")
//   - PayScoreController             @RequestMapping("/client")
//   - PayScoreCallbackController     @RequestMapping("/callback")
//   - ZhimaPayAfterUseController     @RequestMapping("/client")
//   - ZhimaPayAfterUseOrderController @RequestMapping("/client")
//   - ZhimaPayAfterUseCallbackController (no class-level mapping)
func RegisterRoutes(r *gin.RouterGroup) {
	ebikePay := r.Group("/client/ebike-pay")
	{
		ebikePay.POST("/pay/create", createPay)
		ebikePay.POST("/pay/getOpenIdByJsCode", getOpenIdByJsCode)
		ebikePay.POST("/pay/cancel", cancelPay)
		ebikePay.POST("/pay/refund", refundPay)
		ebikePay.POST("/pay/withdraw", withdraw)
		ebikePay.POST("/pay/withdraw/page", withdrawPage)
	}

	client := r.Group("/client")
	{
		client.POST("/pay/score/permission", payScorePermission)
		client.POST("/pay/score/getPermissionRecord", getPermissionRecord)
		client.POST("/pay/score/terminatePermission", terminatePermission)
		client.POST("/pay/score/createOrder", createPayScoreOrder)
		client.POST("/pay/score/completeOrder", completePayScoreOrder)
		client.POST("/pay/score/cancelOrder", cancelPayScoreOrder)
		client.POST("/pay/score/getConfig", getPayScoreConfig)

		client.POST("/zhima/payafteruse/sign", zhimaSign)
		client.POST("/zhima/payafteruse/signQuery", zhimaSignQuery)
		client.POST("/zhima/payafteruse/orderWithoutConfirm", zhimaOrderWithoutConfirm)
		client.POST("/zhima/payafteruse/orderQuery", zhimaOrderQuery)
		client.POST("/zhima/payafteruse/orderFinish", zhimaOrderFinish)
	}

	callback := r.Group("/callback")
	{
		// Java: @PostMapping("/pay/score/{tenantId}/xxxNotify")
		callback.POST("/pay/score/:tenantId/paySuccessNotify", payScoreNotifyHandler("/pay/score/paySuccessNotify"))
		callback.POST("/pay/score/:tenantId/permissionNotify", payScoreNotifyHandler("/pay/score/permissionNotify"))
		callback.POST("/pay/score/:tenantId/refundNotify", payScoreNotifyHandler("/pay/score/refundNotify"))

		// Java: ZhimaPayAfterUseCallbackController.notify
		callback.POST("/zhima/payafteruse/notify", zhimaNotify)
	}
}

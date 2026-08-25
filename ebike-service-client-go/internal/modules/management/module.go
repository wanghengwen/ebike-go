// Package management ports the following Java controllers of ebike-service-client:
//
//   - CodeController          (@RequestMapping("/client") -> code/send, code/sendEmailCode, code/verify)
//   - FileController          (@RequestMapping("/client/file") -> upload, upload/async)
//   - MessageCenterController (@RequestMapping("/client/messageCenter"))
//   - management/GaodeMapController   (@RequestMapping("/client/management/gaode"))
//   - management/GoogleMapController  (@RequestMapping("/client/management/google"))
//   - management/RepairConfigController (@RequestMapping("/client/management/repairConfig"))
package management

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires all routes of this module. Paths replicate the Java
// class-level @RequestMapping + method-level @PostMapping verbatim.
func RegisterRoutes(r *gin.RouterGroup) {
	client := r.Group("/client")
	{
		// CodeController (method-level mappings have no leading slash in Java,
		// Spring resolves them relative to "/client")
		client.POST("/code/send", sendCode)
		client.POST("/code/sendEmailCode", sendEmailCode)
		client.POST("/code/verify", verifyCode)

		// FileController
		client.POST("/file/upload", fileUpload)
		client.POST("/file/upload/async", fileUploadAsync)

		// MessageCenterController
		client.POST("/messageCenter/page/list", msgRecordPageList)
		client.POST("/messageCenter/getById", msgRecordGetById)
		client.POST("/messageCenter/judgeIsHaveNoReadMsg", judgeIsHaveNoReadMsg)

		// GaodeMapController
		client.POST("/management/gaode/getAddress", gaodeGetAddress)
		client.POST("/management/gaode/getLocation", gaodeGetLocation)
		client.POST("/management/gaode/navigate", gaodeNavigate)
		client.POST("/management/gaode/v2/navigate", gaodeNavigateV2)
		client.POST("/management/gaode/v3/navigate", gaodeNavigateV3)

		// GoogleMapController
		client.POST("/management/google/getAddress", googleGetAddress)
		client.POST("/management/google/getLocation", googleGetLocation)

		// RepairConfigController
		client.POST("/management/repairConfig/list", repairConfigList)
		client.POST("/management/repairConfig/listByCar", repairConfigListByCar)
	}
}

package fence

import "github.com/gin-gonic/gin"

// RegisterRoutes wires all fence-module HTTP endpoints.
// Java controllers use @RequestMapping("/${ebike.fence.name}") -> "/client/fence"
// except CreditScoreConfigController -> "/client/ebike-fence".
func RegisterRoutes(r *gin.RouterGroup) {
	fence := r.Group("/client/fence")
	{
		// FenceController
		fence.POST("/serviceArea/getAll", getAllServiceAreas)
		fence.POST("/serviceArea/getByLocation", getServiceAreaByLocation)
		fence.POST("/serviceArea/getFenceByServiceId", getFenceByServiceId)
		fence.POST("/serviceArea/getServiceAreaById", getServiceAreaById)
		fence.POST("/serviceArea/getNearFence", getNearFence)

		fence.POST("/parking/getById", getParkingById)
		fence.POST("/parking/getByServiceId", getParkingByServiceId)
		fence.POST("/parking/createParking", createParking)
		fence.POST("/parking/updateParking", updateParking)
		fence.POST("/parking/deleteParking", deleteParking)
		fence.POST("/parking/enable", enableParking)
		fence.POST("/parking/disable", disableParking)
		fence.POST("/parking/nearParkingNum", getNearParkingNum)

		fence.POST("/noParking/getById", getNoParkingById)
		fence.POST("/noParking/getByServiceId", getNoParkingByServiceId)
		fence.POST("/noParking/createNoParking", createNoParking)
		fence.POST("/noParking/updateNoParking", updateNoParking)
		fence.POST("/noParking/deleteNoParking", deleteNoParking)

		// ProtocolConfigController @RequestMapping("/client/fence")
		fence.POST("/config/protocol/byType", protocolByType)
		fence.POST("/config/protocol/default", protocolDefault)

		// ResourceManagementController
		fence.POST("/resource/management/appList", resourceAppList)
		fence.POST("/resourceBit/addExposure", resourceAddExposure)
		fence.POST("/resourceBit/addClick", resourceAddClick)
	}

	// CreditScoreConfigController @RequestMapping("/client/ebike-fence")
	r.POST("/client/ebike-fence/creditScore/getConfig", creditScoreGetConfig)
}

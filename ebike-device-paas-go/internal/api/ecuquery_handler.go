package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/client"
	"ebike-device-paas-go/internal/service"

	"github.com/gin-gonic/gin"
)

// registerEcuQueryRoutes attaches the EcuQuery endpoints.
//
// Shadow / RECORD / live routing is owned by middleware.ProxyGateway (fence-go
// style). Handlers always implement Go logic; whether the client sees Go, Java,
// or a dual-run compare is decided solely by liveList / recordList / DryRun.
func registerEcuQueryRoutes(r *gin.Engine) {
	r.POST("/device/paas/getDeviceRealGpsList", getDeviceRealGpsList)
	r.POST("/device/paas/device/detail", getDeviceDetail)
	r.POST("/device/paas/device/list", getDeviceList)
	r.POST("/device/paas/device/listByImeiList", getDeviceListByImeiList)
	r.POST("/device/paas/device/listByCarIdList", getDeviceListByCarIdList)
	r.POST("/device/paas/device/filterList", getDeviceFilterList)
	r.POST("/device/paas/device/queryCarImeiBind", queryCarImeiBind)
	r.POST("/device/paas/device/queryDeviceCacheGps", queryDeviceCacheGps)
	r.POST("/device/paas/device/eBikeLocation", getUseableEbikeLocation)
	r.POST("/device/paas/device/queryDeviceScreen", queryDeviceScreen)
	r.POST("/device/paas/device/queryDeviceScreen/v2", queryDeviceScreenV2)
	r.POST("/device/paas/device/car_count", carCount)
	r.POST("/device/paas/device/queryServiceStatisticsList", queryServiceStatisticsList)
	r.POST("/device/paas/device/getCarNumByServiceId", getCarNumByServiceId)
	r.POST("/device/paas/device/carStatisticsByService", carStatisticsByService)
	r.POST("/device/paas/device/onlineNumByTenantId", onlineNumByTenantId)
	r.POST("/device/paas/device/carStatistics", carStatistics)
	r.POST("/device/paas/device/getRackCarNumAll", getRackCarNumAll)
	r.POST("/device/paas/device/getBlueToothToken", getBlueToothToken)
	r.POST("/device/paas/device/querySaddleOverloadContact", querySaddleOverloadContact)
	r.POST("/device/paas/device/queryCameraState", queryCameraState)
	r.POST("/device/paas/device/queryDeviceMapFake", queryDeviceMapFake)
	r.POST("/device/paas/device/oneClickReturnNotify", oneClickReturnNotify)
	r.POST("/device/paas/device/queryDeviceByBattery", queryDeviceByBattery)
	r.POST("/device/paas/device/queryDeviceByTotalMiles", queryDeviceByTotalMiles)
	r.POST("/device/paas/device/queryDeviceByNoOrderTime", queryDeviceByNoOrderTime)
	r.POST("/device/paas/device/queryDeviceByStaticTime", queryDeviceByStaticTime)
	r.POST("/device/paas/device/removeDeviceTotalMiles", removeDeviceTotalMiles)
	r.POST("/device/paas/device/queryDeviceOpeMap", queryDeviceOpeMap)
	r.POST("/device/paas/device/carParkingStatistics", carParkingStatistics)
	r.POST("/device/paas/device/page", getDevicePage)
	r.POST("/device/paas/device/pageBus", getDevicePageBus)
	r.POST("/device/paas/device/genDeviceMapFake", genDeviceMapFake)
	r.POST("/device/paas/innerParam", getInnerParam)
	r.POST("/device/paas/deviceInfo", getDeviceInfo)
	r.POST("/device/paas/blueTBeacon", getBlueTBeacon)
	r.POST("/device/paas/queryBleHelmetInfo", queryBleHelmetInfo)
	r.POST("/device/paas/queryRealRestBattery", queryRealRestBattery)
}

// getInnerParam ports EcuQueryController.getInnerParam.
func getInnerParam(c *gin.Context) {
	var q dto.EcuQuery
	if !bindBody(c, &q) {
		return
	}
	backfillTenant(c, &q.CommandContext)
	cr, err := service.GetInnerParam(q)
	writeCommandResult(c, cr, err)
}

// getDeviceInfo ports EcuQueryController.getDeviceInfo.
func getDeviceInfo(c *gin.Context) {
	var q dto.DeviceInfoQry
	if !bindBody(c, &q) {
		return
	}
	backfillTenant(c, &q.CommandContext)
	cr, err := service.GetDeviceInfo(q)
	writeCommandResult(c, cr, err)
}

// getBlueTBeacon ports EcuQueryController.getBlueTBeacon.
func getBlueTBeacon(c *gin.Context) {
	var q dto.EcuQuery
	if !bindBody(c, &q) {
		return
	}
	backfillTenant(c, &q.CommandContext)
	cr, err := service.GetBlueTBeacon(q)
	writeCommandResult(c, cr, err)
}

// queryBleHelmetInfo ports EcuQueryController.queryBleHelmetInfo.
func queryBleHelmetInfo(c *gin.Context) {
	var q dto.EcuQuery
	if !bindBody(c, &q) {
		return
	}
	backfillTenant(c, &q.CommandContext)
	cr, err := service.QueryBleHelmetInfo(q)
	writeCommandResult(c, cr, err)
}

// queryRealRestBattery ports EcuQueryController.queryRealRestBattery.
func queryRealRestBattery(c *gin.Context) {
	var q dto.EcuQuery
	if !bindBody(c, &q) {
		return
	}
	tenantID := dto.TenantOf(q.CommandContext)
	if tenantID == "" {
		if v, ok := c.Get("tenantId"); ok {
			tenantID, _ = v.(string)
		}
	}
	co, err := service.QueryRealRestBattery(tenantID, q)
	if err != nil {
		c.JSON(http.StatusOK, gatewayErrorResult(err, nil))
		return
	}
	c.JSON(http.StatusOK, Ok(co))
}

// bindBody reads + unmarshals the JSON body into v, writing a param error on
// failure. Returns false when the request should not proceed.
func bindBody(c *gin.Context, v interface{}) bool {
	body, _ := io.ReadAll(c.Request.Body)
	if len(body) == 0 {
		return true
	}
	if err := json.Unmarshal(body, v); err != nil {
		c.JSON(http.StatusOK, Fail(CodeParamError, ""))
		return false
	}
	return true
}

// resolveTenant prefers the request commandContext tenant, falling back to the
// gateway-injected tenantId header set by middleware.
func resolveTenant(c *gin.Context, cc *dto.CommandContext) string {
	tenantID := dto.TenantOf(cc)
	if tenantID == "" {
		if v, ok := c.Get("tenantId"); ok {
			tenantID, _ = v.(string)
		}
	}
	return tenantID
}

// writeShadowed marshals and writes the Go Result. Shadow compare (if any) is
// performed by ProxyGateway when the path is not in liveList/recordList.
func writeShadowed(c *gin.Context, _ []byte, result Result, _ ...string) {
	goJSON, err := json.Marshal(result)
	if err != nil {
		c.JSON(http.StatusOK, Fail(CodeParamError, "marshal error"))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", goJSON)
}

// writeCommandResult applies the Java toDeviceResult mapping: ecuCode 0/empty ->
// success; otherwise a 17003 failure that still carries the CommandResult.
func writeCommandResult(c *gin.Context, cr dto.CommandResult, err error) {
	if err != nil {
		c.JSON(http.StatusOK, gatewayErrorResult(err, cr))
		return
	}
	if ec := cr.EcuCodeValue(); ec == "" || ec == "0" {
		c.JSON(http.StatusOK, Ok(cr))
		return
	}
	c.JSON(http.StatusOK, FailData(CodeEcuInfoFindFailed, "", cr))
}

// gatewayErrorResult maps a *client.GatewayError code to the paas-facing code,
// mirroring DeviceResultHelper.getResultData.
func gatewayErrorResult(err error, data interface{}) Result {
	var rle *service.RateLimitError
	if errors.As(err, &rle) {
		// Java throws before producing a CommandResult, so data is null.
		return Fail(CodeRateLimit, rle.Msg)
	}
	var ge *client.GatewayError
	if errors.As(err, &ge) {
		switch ge.Code {
		case "10007":
			return FailData(CodeDeviceOffline, "请求设备超时", data)
		case "10012":
			return FailData(CodeDeviceOffline, "设备离线", data)
		case "10005", "10006", "10008":
			msg := "设备网关认证错误"
			if ge.Msg != "" {
				msg = ge.Msg
			}
			return FailData(CodeDeviceOffline, msg, data)
		case "10013":
			return FailData(CodeDeviceNoResponse, "", data)
		case "10014":
			return FailData(CodeTrajectory, "", data)
		case "10015":
			return FailData(CodeTrajectoryDateRange, "trajectory query only the data of the last six months", data)
		default:
			return FailData(CodeDeviceGatewayError, ge.Msg, data)
		}
	}
	if errors.Is(err, service.ErrDeviceNotFound) {
		return FailData(CodeEcuInfoFindFailed, "对应设备不存在", data)
	}
	if errors.Is(err, client.ErrAuthorization) {
		return FailData(CodeDeviceGatewayError, "设备网关认证错误", data)
	}
	return FailData(CodeDeviceGatewayError, "设备网关错误", data)
}

// getDeviceRealGpsList ports EcuQueryController.getDeviceRealGpsList: a pure
// Redis read, wrapped in the standard Result envelope, with shadow comparison
// against the legacy Java service.
func getDeviceRealGpsList(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)

	var qry dto.ImeiListQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}

	tenantID := dto.TenantOf(qry.CommandContext)
	if tenantID == "" {
		if v, ok := c.Get("tenantId"); ok {
			tenantID, _ = v.(string)
		}
	}

	list := service.GetDeviceRealGpsList(tenantID, qry.ImeiList)
	writeShadowed(c, body, Ok(list))
}

// getDeviceDetail ports DeviceInfoController.getDeviceDetail: a pure Redis read
// (carId/imei -> device). Shadow skipped when listed in xyy.proxy.liveList.
func getDeviceDetail(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)

	var qry dto.DeviceDetailQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	if qry.Imei == "" && qry.CarId == "" {
		c.JSON(http.StatusOK, Fail(CodeImeiCarIdEmpty, "imei和carId必须有一个不为空"))
		return
	}

	tenantID := dto.TenantOf(qry.CommandContext)
	if tenantID == "" {
		if v, ok := c.Get("tenantId"); ok {
			tenantID, _ = v.(string)
		}
	}

	co, err := service.GetDeviceDetail(tenantID, qry)
	if err != nil {
		c.JSON(http.StatusOK, deviceDetailError(err))
		return
	}

	writeShadowed(c, body, Ok(co))
}

// getDeviceListByImeiList ports DeviceInfoController.getDeviceListByImeiList:
// a pure Redis read, shadow-compared against the legacy Java service.
func getDeviceListByImeiList(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)

	var qry dto.ImeiListQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}

	tenantID := dto.TenantOf(qry.CommandContext)
	if tenantID == "" {
		if v, ok := c.Get("tenantId"); ok {
			tenantID, _ = v.(string)
		}
	}

	list := service.GetDeviceListByImeiList(tenantID, qry.ImeiList)
	writeShadowed(c, body, Ok(list))
}

// getDeviceList ports DeviceInfoController.getDeviceList: a pure Redis read
// (service-area sets + devices), shadow-compared against the legacy Java service.
func getDeviceList(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)

	var qry dto.DeviceListQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}

	tenantID := dto.TenantOf(qry.CommandContext)
	if tenantID == "" {
		if v, ok := c.Get("tenantId"); ok {
			tenantID, _ = v.(string)
		}
	}

	list := service.GetDeviceList(tenantID, qry)
	writeShadowed(c, body, Ok(list))
}

// getDeviceListByCarIdList ports DeviceInfoController.getDeviceListByCarIdList:
// carId -> imei -> device-info, shadow-compared.
func getDeviceListByCarIdList(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)

	var qry dto.CarIdListQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}

	tenantID := resolveTenant(c, qry.CommandContext)
	list := service.GetDeviceListByCarIdList(tenantID, qry.CarIdList)
	writeShadowed(c, body, Ok(list))
}

// getDeviceFilterList ports DeviceInfoController.getDeviceFilterList: read +
// filter service-area devices (no amap/cache), shadow-compared.
func getDeviceFilterList(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)

	var qry dto.DevicePageQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}

	tenantID := resolveTenant(c, qry.CommandContext)
	list := service.GetDeviceFilterList(tenantID, qry)
	writeShadowed(c, body, Ok(list))
}

// queryCarImeiBind ports DeviceInfoController.queryCarImeiBind: positional
// imei<->carId pairs, shadow-compared.
func queryCarImeiBind(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)

	var qry dto.ImeiListQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}

	tenantID := resolveTenant(c, qry.CommandContext)
	list, err := service.QueryCarImeiBind(tenantID, qry)
	if err != nil {
		c.JSON(http.StatusOK, Fail(CodeParamError, ""))
		return
	}
	writeShadowed(c, body, Ok(list))
}

// queryDeviceCacheGps ports DeviceInfoController.queryDeviceCacheGps:
// imei -> device-info cached GPS, shadow-compared.
func queryDeviceCacheGps(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)

	var qry dto.ImeiListQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}

	tenantID := resolveTenant(c, qry.CommandContext)
	list := service.QueryDeviceCacheGps(tenantID, qry.ImeiList)
	writeShadowed(c, body, Ok(list))
}

// queryDeviceScreen ports DeviceInfoController.queryDeviceScreen (V1 dashboard
// aggregate over the service list), shadow-compared.
func queryDeviceScreen(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.DeviceScreenQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	writeShadowed(c, body, Ok(service.QueryDeviceScreen(tenantID, qry)))
}

// queryDeviceScreenV2 ports DeviceInfoController.queryDeviceScreenV2, shadow-compared.
func queryDeviceScreenV2(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.DeviceScreenQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	writeShadowed(c, body, Ok(service.QueryDeviceScreenV2(tenantID, qry)))
}

// carCount ports DeviceInfoController.carCount (per-service totals), shadow-compared.
func carCount(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.DeviceScreenQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	writeShadowed(c, body, Ok(service.CarCount(tenantID, qry)))
}

// queryServiceStatisticsList ports DeviceInfoController.queryServiceStatisticsList
// (per-service state tally, deterministic input order), shadow-compared.
func queryServiceStatisticsList(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.IdListQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	writeShadowed(c, body, Ok(service.QueryServiceStatisticsList(tenantID, qry.Ids)))
}

// getCarNumByServiceId ports DeviceInfoController.getCarNumByServiceId (per-service
// on-shelf counts), shadow-compared.
func getCarNumByServiceId(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.ServiceCarNumQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	// Java returns the per-service groups in HashMap order; compare the top-level
	// list order-insensitively.
	writeShadowed(c, body, Ok(service.GetCarNumByServiceId(tenantID, []int64(qry.ServiceIdList))), "data")
}

// carStatisticsByService ports DeviceInfoController.carStatisticsByService.
// Result lists follow Java HashMap order; the serviceStatistics / parkingStatistics
// arrays are compared order-insensitively (configured unordered keys).
func carStatisticsByService(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.CarStatisticsByServiceQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	writeShadowed(c, body, Ok(service.CarStatisticsByService(tenantID, []int64(qry.ServiceIdList))))
}

// onlineNumByTenantId ports DeviceInfoController.onlineNumByTenantId (current
// tenant on-shelf device count), shadow-compared.
func onlineNumByTenantId(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.NoParamQuery
	if len(body) > 0 {
		_ = json.Unmarshal(body, &qry)
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	writeShadowed(c, body, Ok(service.OnlineNumByTenantId(tenantID, qry.CommandContext)))
}

// carStatistics ports DeviceInfoController.carStatistics (cross-tenant per-service
// aggregation), shadow-compared with order-insensitive top-level list.
func carStatistics(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.NoParamQuery
	if len(body) > 0 {
		_ = json.Unmarshal(body, &qry)
	}
	writeShadowed(c, body, Ok(service.CarStatistics(qry.CommandContext)), "data")
}

// getRackCarNumAll ports DeviceInfoController.getRackCarNumAll (cross-tenant
// per-service on-shelf counts), shadow-compared with order-insensitive list.
func getRackCarNumAll(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.NoParamQuery
	if len(body) > 0 {
		_ = json.Unmarshal(body, &qry)
	}
	writeShadowed(c, body, Ok(service.GetRackCarNumAll(qry.CommandContext)), "data")
}

// getBlueToothToken ports DeviceInfoController.getBlueToothToken: validate the
// device + tenant, read the cached token. Read-only -> shadow-compared.
func getBlueToothToken(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.BlueToothTokenQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	co, err := service.GetBlueToothToken(tenantID, qry.Imei)
	if err != nil {
		c.JSON(http.StatusOK, Fail(CodeDeviceNullHave, "对应设备不存在"))
		return
	}
	writeShadowed(c, body, Ok(co))
}

// querySaddleOverloadContact ports DeviceInfoController.querySaddleOverloadContact:
// decode the saddle contact bits. Read-only -> shadow-compared.
func querySaddleOverloadContact(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.ImeiQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	writeShadowed(c, body, Ok(service.QuerySaddleOverloadContact(tenantID, qry.Imei)))
}

// queryCameraState ports DeviceInfoController.queryCameraState: parse cached
// camera state (null when absent). Read-only -> shadow-compared.
func queryCameraState(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.ImeiQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	writeShadowed(c, body, Ok(service.QueryCameraState(tenantID, qry.Imei)))
}

// queryDeviceMapFake ports DeviceInfoController.queryDeviceMapFake: cached amount
// + fake locations. Read-only -> shadow-compared.
func queryDeviceMapFake(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.DeviceMapFakeQuery
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	writeShadowed(c, body, Ok(service.QueryDeviceMapFake(tenantID, qry.ServiceId.Int64())))
}

// oneClickReturnNotify ports DeviceInfoController.oneClickReturnNotify: a pure
// Redis read of the cached fail-notify code (despite being a *Cmd). Shadow-compared.
func oneClickReturnNotify(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var cmd dto.OneClickReturnNotifyCmd
	if len(body) > 0 {
		if err := json.Unmarshal(body, &cmd); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, cmd.CommandContext)
	writeShadowed(c, body, Ok(service.OneClickReturnNotify(tenantID, cmd.CarId)))
}

// queryDeviceByBattery ports DeviceInfoController.queryDeviceByBattery (zset range
// + device filter). Read-only; result order follows Java -> compare unordered.
func queryDeviceByBattery(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.DeviceByBatteryQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	writeShadowed(c, body, Ok(service.QueryDeviceByBattery(tenantID, qry)), "data")
}

// queryDeviceByTotalMiles ports DeviceInfoController.queryDeviceByTotalMiles.
func queryDeviceByTotalMiles(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.DeviceByTotalMilesQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	writeShadowed(c, body, Ok(service.QueryDeviceByTotalMiles(tenantID, qry)), "data")
}

// queryDeviceByNoOrderTime ports DeviceInfoController.queryDeviceByNoOrderTime.
// noOrderTime is now-derived (volatile) and ignored by shadow diff.
func queryDeviceByNoOrderTime(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.DeviceByNoOrderTimeQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	writeShadowed(c, body, Ok(service.QueryDeviceByNoOrderTime(tenantID, qry)), "data")
}

// queryDeviceByStaticTime ports DeviceInfoController.queryDeviceByStaticTime.
// staticTime is now-derived (volatile) and ignored by shadow diff.
func queryDeviceByStaticTime(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.DeviceByStaticTimeQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	writeShadowed(c, body, Ok(service.QueryDeviceByStaticTime(tenantID, qry)), "data")
}

// removeDeviceTotalMiles ports DeviceInfoController.removeDeviceTotalMiles: a ZREM
// write. NOT shadow-compared (mirroring would issue a second real delete to Java).
func removeDeviceTotalMiles(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var cmd dto.ImeiCmd
	if len(body) > 0 {
		if err := json.Unmarshal(body, &cmd); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, cmd.CommandContext)
	service.RemoveDeviceTotalMiles(tenantID, cmd.Imei)
	c.JSON(http.StatusOK, Ok(nil))
}

// queryDeviceOpeMap ports DeviceInfoController.queryDeviceOpeMap: an operations
// dashboard aggregate (+ filtered GPS markers). Read-only -> shadow-compared.
func queryDeviceOpeMap(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.DeviceOpeMapQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	writeShadowed(c, body, Ok(service.QueryDeviceOpeMap(tenantID, qry)))
}

// carParkingStatistics ports DeviceInfoController.carParkingStatistics: cross-tenant
// per-parking aggregation. Result follows HashMap order -> compare unordered ("data").
func carParkingStatistics(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.NoParamQuery
	if len(body) > 0 {
		_ = json.Unmarshal(body, &qry)
	}
	writeShadowed(c, body, Ok(service.CarParkingStatistics(qry.CommandContext)), "data")
}

// genDeviceMapFake ports DeviceInfoController.genDeviceMapFake: generate + cache
// a fake device map. WRITE endpoint (lock + SETs) — NOT shadow-compared (random
// output + mirroring would double-write to Java).
func genDeviceMapFake(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var cmd dto.GenDeviceMapFakeCmd
	if len(body) > 0 {
		if err := json.Unmarshal(body, &cmd); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, cmd.CommandContext)
	switch err := service.GenDeviceMapFake(tenantID, cmd); {
	case errors.Is(err, service.ErrGenFakeOutOfLimit):
		c.JSON(http.StatusOK, Fail(CodeGenFakeOutOfLimit, "生成车辆数超出限制"))
	case errors.Is(err, service.ErrGenFakeNx):
		c.JSON(http.StatusOK, Fail(CodeGenFakeNx, "正在生成数据中,请耐心等待!"))
	case err != nil:
		c.JSON(http.StatusOK, Fail(CodeParamError, ""))
	default:
		c.JSON(http.StatusOK, Ok(nil))
	}
}

// getDevicePage ports DeviceInfoController.getDevicePage (platform pagination).
// Read-only -> shadow-compared (address/scanAddress are amap-derived and ignored).
func getDevicePage(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.DevicePageQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	writeShadowed(c, body, Ok(service.GetDevicePage(tenantID, qry)))
}

// getDevicePageBus ports DeviceInfoController.getDevicePageBus (merchant pagination).
func getDevicePageBus(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var qry dto.DevicePageBusQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}
	tenantID := resolveTenant(c, qry.CommandContext)
	writeShadowed(c, body, Ok(service.GetDevicePageBus(tenantID, qry)))
}

// getUseableEbikeLocation ports DeviceInfoController.getUseableEbikeLocation: a
// Redis GEO read (+ best-effort fence config), shadow-compared.
func getUseableEbikeLocation(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)

	var qry dto.DeviceLocationQry
	if len(body) > 0 {
		if err := json.Unmarshal(body, &qry); err != nil {
			c.JSON(http.StatusOK, Fail(CodeParamError, ""))
			return
		}
	}

	tenantID := dto.TenantOf(qry.CommandContext)
	if tenantID == "" {
		if v, ok := c.Get("tenantId"); ok {
			tenantID, _ = v.(string)
		}
	}

	list := service.GetUseableEbikeLocation(tenantID, qry)
	writeShadowed(c, body, Ok(list))
}

func deviceDetailError(err error) Result {
	switch {
	case errors.Is(err, service.ErrCarImeiBind):
		return Fail(CodeImeiCarIdBind, "请输入正确的车辆号")
	case errors.Is(err, service.ErrDeviceNotFound):
		return Fail(CodeDeviceNullHave, "对应设备不存在")
	default:
		return Fail(CodeDeviceGatewayError, "设备网关错误")
	}
}

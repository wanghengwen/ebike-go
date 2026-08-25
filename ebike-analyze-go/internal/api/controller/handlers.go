package controller

import (
	"ebike-analyze-go/internal/api"
	"ebike-analyze-go/internal/api/dto"
	"ebike-analyze-go/internal/app"
	"ebike-analyze-go/internal/common/bizerror"
	"ebike-analyze-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

var (
	orderSvc      app.OrderQueryService
	userQrySvc    app.UserQueryService
	userStatSvc   app.UserStatisticsService
	carSvc        app.CarStatisticsService
	siteSvc       app.SiteStatisticsService
	visualSvc     app.VisualQueryService
	deviceSvc     app.DeviceInfoService
	hotSvc        app.HotPointService
	operateLogSvc app.OperateLogService
)

func registerNativeHandlers() {
	handlers := map[string]gin.HandlerFunc{
		"/orderQuery/selectOrderCount":      bindOrder(selectOrderCount),
		"/orderQuery/selectOrderAnalyze":    bindOrder(selectOrderAnalyze),
		"/orderQuery/selectOrderList":       bindOrder(selectOrderList),
		"/orderDelete/deleteByQuery":        bindOrder(deleteByQuery),
		"/userQuery/selectUserCount":        bindUserQry(selectUserCount),
		"/analyze/member":                   bindMember(memberStatistics),
		"/user/ageStatistic":                bindAge(ageStatistic),
		"/user/ageStatistic/v2":             bindAge(ageStatisticV2),
		"/user/allUserStatistic":            bindAllUser(allUserStatistic),
		"/user/userStatisticHour":           bindHour(userStatisticHour),
		"/car-statistics/page":              bindCarPage(carStatisticsPage),
		"/car-statistics/queryByCarList":    bindCarList(carStatisticsByCarList),
		"/carServiceStatistics/getList":     bindCarService(carServiceStatistics),
		"/site-statistics/list":             bindSiteList(siteStatisticsList),
		"/parking/statistics":               bindParkingStat(parkingStatistics),
		"/parking/analyze/one":              bindParkingOne(parkingAnalyzeOne),
		"/parking/orderStatistics/page":     bindParkingOrderPage(parkingOrderPage),
		"/parking/orderStatistics/list":     bindParkingOrderList(parkingOrderList),
		"/parking/getParkingInAndOutflow":   bindInOutflow(parkingInOutflow),
		"/riding_card_order/page":           bindVisualPage(ridingCardPage),
		"/riding_card_order/list":           bindVisualPage(ridingCardList),
		"/wallet_order/page":                bindVisualPage(walletPage),
		"/wallet_order/list":                bindVisualPage(walletList),
		"/deposit_order/page":               bindVisualPage(depositPage),
		"/device/paas/device/list":          bindDeviceList(deviceList),
		"/device/paas/device/carStatistics": bindDeviceCarStats(deviceCarStatistics),
		"/hot/point/start/list":             bindHotPoint(hotPointStart),
		"/hot/point/end/list":               bindHotPoint(hotPointEnd),
		"/log/page":                         bindOperateLog(operateLogPage),
	}
	api.RegisterHandlers(handlers)
}

func bindOrder(fn func(*gin.Context, *dto.OrderQueryCmd, string) (interface{}, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cmd dto.OrderQueryCmd
		if err := c.ShouldBindJSON(&cmd); err != nil {
			c.JSON(200, dto.NewErrorResult(dto.CodeParamException, "http message not readable"))
			return
		}
		if _, ok := middleware.ApplyCommand(c, &dto.Command{CommandContext: cmd.CommandContext}); !ok {
			return
		}
		res, err := fn(c, &cmd, cmd.CommandContext.TenantId)
		if err != nil {
			writeErr(c, err)
			return
		}
		writeOK(c, res)
	}
}

func bindUserQry(fn func(*gin.Context, *dto.UserQueryCmd, string) (interface{}, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cmd dto.UserQueryCmd
		if err := c.ShouldBindJSON(&cmd); err != nil {
			c.JSON(200, dto.NewErrorResult(dto.CodeParamException, "http message not readable"))
			return
		}
		if _, ok := middleware.ApplyCommand(c, &dto.Command{CommandContext: cmd.CommandContext}); !ok {
			return
		}
		res, err := fn(c, &cmd, cmd.CommandContext.TenantId)
		if err != nil {
			writeErr(c, err)
			return
		}
		writeOK(c, res)
	}
}

func bindMember(fn func(*gin.Context, *dto.MemberStatisticQuery, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}
func bindAge(fn func(*gin.Context, *dto.AgeStatisticCmd, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}
func bindAllUser(fn func(*gin.Context, *dto.AllUserStatisticCmd, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}
func bindHour(fn func(*gin.Context, *dto.UserStatisticHourQry, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}
func bindCarPage(fn func(*gin.Context, *dto.CarStatisticsQuery, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}
func bindCarList(fn func(*gin.Context, *dto.CarStatisticsListQuery, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}
func bindCarService(fn func(*gin.Context, *dto.CarServiceStatisticsQuery, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}
func bindSiteList(fn func(*gin.Context, *dto.SiteStatisticsCmd, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}
func bindParkingStat(fn func(*gin.Context, *dto.ParkingStatisticalPageQuery, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}
func bindParkingOne(fn func(*gin.Context, *dto.OneParkingAnalyzeCmd, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}
func bindParkingOrderPage(fn func(*gin.Context, *dto.ParkingOrderStatisticQry, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}
func bindParkingOrderList(fn func(*gin.Context, *dto.ParkingOrderStatisticListQry, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}
func bindInOutflow(fn func(*gin.Context, *dto.ParkingInAndOutflowQry, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}
func bindVisualPage(fn func(*gin.Context, *dto.OrderQueryPageCmd, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}
func bindDeviceList(fn func(*gin.Context, *dto.DeviceListQry, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}
func bindDeviceCarStats(fn func(*gin.Context, *dto.Command, string) (interface{}, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cmd dto.Command
		if err := c.ShouldBindJSON(&cmd); err != nil {
			c.JSON(200, dto.NewErrorResult(dto.CodeParamException, "http message not readable"))
			return
		}
		if _, ok := middleware.ApplyCommand(c, &cmd); !ok {
			return
		}
		res, err := fn(c, &cmd, cmd.CommandContext.TenantId)
		if err != nil {
			writeErr(c, err)
			return
		}
		writeOK(c, res)
	}
}
func bindHotPoint(fn func(*gin.Context, *dto.HotPointCmd, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}

func bindOperateLog(fn func(*gin.Context, *dto.OperateLogCmd, string) (interface{}, error)) gin.HandlerFunc {
	return bindGeneric(fn)
}

func bindGeneric[T any](fn func(*gin.Context, *T, string) (interface{}, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cmd T
		if err := c.ShouldBindJSON(&cmd); err != nil {
			c.JSON(200, dto.NewErrorResult(dto.CodeParamException, "http message not readable"))
			return
		}
		cc := extractCommandContext(&cmd)
		if _, ok := middleware.ApplyCommand(c, &dto.Command{CommandContext: cc}); !ok {
			return
		}
		res, err := fn(c, &cmd, cc.TenantId)
		if err != nil {
			writeErr(c, err)
			return
		}
		writeOK(c, res)
	}
}

func extractCommandContext(cmd interface{}) *dto.CommandContext {
	switch v := cmd.(type) {
	case *dto.MemberStatisticQuery:
		return v.CommandContext
	case *dto.AgeStatisticCmd:
		return v.CommandContext
	case *dto.AllUserStatisticCmd:
		return v.CommandContext
	case *dto.UserStatisticHourQry:
		return v.CommandContext
	case *dto.CarStatisticsQuery:
		return v.CommandContext
	case *dto.CarStatisticsListQuery:
		return v.CommandContext
	case *dto.CarServiceStatisticsQuery:
		return v.CommandContext
	case *dto.SiteStatisticsCmd:
		return v.CommandContext
	case *dto.ParkingStatisticalPageQuery:
		return v.CommandContext
	case *dto.OneParkingAnalyzeCmd:
		return v.CommandContext
	case *dto.ParkingOrderStatisticQry:
		return v.CommandContext
	case *dto.ParkingOrderStatisticListQry:
		return v.CommandContext
	case *dto.ParkingInAndOutflowQry:
		return v.CommandContext
	case *dto.OrderQueryPageCmd:
		return v.CommandContext
	case *dto.DeviceListQry:
		return v.CommandContext
	case *dto.HotPointCmd:
		return v.CommandContext
	case *dto.OperateLogCmd:
		return v.CommandContext
	}
	return nil
}

func writeOK(c *gin.Context, data interface{}) {
	c.JSON(200, dto.NewSuccessResult(data))
}

func writeErr(c *gin.Context, err error) {
	if biz, ok := err.(*bizerror.BizError); ok {
		code := biz.Code()
		msg := biz.FormattedMsg()
		c.JSON(200, dto.Result{Success: false, Code: &code, Msg: &msg})
		return
	}
	c.JSON(200, dto.NewErrorResult(dto.CodeException, bizerror.FormatException(err)))
}

func selectOrderCount(c *gin.Context, cmd *dto.OrderQueryCmd, tenantID string) (interface{}, error) {
	return orderSvc.SelectOrderCount(c.Request.Context(), cmd, tenantID), nil
}
func selectOrderAnalyze(c *gin.Context, cmd *dto.OrderQueryCmd, tenantID string) (interface{}, error) {
	return orderSvc.SelectOrderAnalyze(c.Request.Context(), cmd, tenantID), nil
}
func selectOrderList(c *gin.Context, cmd *dto.OrderQueryCmd, tenantID string) (interface{}, error) {
	return orderSvc.SelectOrderList(c.Request.Context(), cmd, tenantID), nil
}
func deleteByQuery(c *gin.Context, cmd *dto.OrderQueryCmd, tenantID string) (interface{}, error) {
	return orderSvc.DeleteByQuery(c.Request.Context(), cmd, tenantID), nil
}
func selectUserCount(c *gin.Context, cmd *dto.UserQueryCmd, tenantID string) (interface{}, error) {
	return userQrySvc.SelectUserCount(c.Request.Context(), cmd, tenantID), nil
}
func memberStatistics(c *gin.Context, cmd *dto.MemberStatisticQuery, tenantID string) (interface{}, error) {
	return userStatSvc.GetMemberStatistics(cmd, tenantID), nil
}
func ageStatistic(c *gin.Context, cmd *dto.AgeStatisticCmd, tenantID string) (interface{}, error) {
	return userStatSvc.GetAgeStatistic(cmd, tenantID), nil
}
func ageStatisticV2(c *gin.Context, cmd *dto.AgeStatisticCmd, tenantID string) (interface{}, error) {
	return userStatSvc.GetAgeStatisticV2(cmd, tenantID), nil
}
func allUserStatistic(c *gin.Context, cmd *dto.AllUserStatisticCmd, tenantID string) (interface{}, error) {
	return userStatSvc.AllUserStatistic(cmd, tenantID), nil
}
func userStatisticHour(c *gin.Context, cmd *dto.UserStatisticHourQry, tenantID string) (interface{}, error) {
	return userStatSvc.UserStatisticHour(cmd, tenantID), nil
}
func carStatisticsPage(c *gin.Context, cmd *dto.CarStatisticsQuery, tenantID string) (interface{}, error) {
	return carSvc.Page(cmd, tenantID), nil
}
func carStatisticsByCarList(c *gin.Context, cmd *dto.CarStatisticsListQuery, tenantID string) (interface{}, error) {
	return carSvc.QueryByCarList(cmd, tenantID), nil
}
func carServiceStatistics(c *gin.Context, cmd *dto.CarServiceStatisticsQuery, tenantID string) (interface{}, error) {
	return carSvc.GetServiceList(cmd, tenantID), nil
}
func siteStatisticsList(c *gin.Context, cmd *dto.SiteStatisticsCmd, tenantID string) (interface{}, error) {
	return siteSvc.List(cmd, tenantID), nil
}
func parkingStatistics(c *gin.Context, cmd *dto.ParkingStatisticalPageQuery, tenantID string) (interface{}, error) {
	return siteSvc.QueryStationStatistical(cmd, tenantID), nil
}
func parkingAnalyzeOne(c *gin.Context, cmd *dto.OneParkingAnalyzeCmd, tenantID string) (interface{}, error) {
	return siteSvc.AnalyzeOneParking(cmd, tenantID), nil
}
func parkingOrderPage(c *gin.Context, cmd *dto.ParkingOrderStatisticQry, tenantID string) (interface{}, error) {
	return siteSvc.ParkingOrderStatisticPage(cmd, tenantID), nil
}
func parkingOrderList(c *gin.Context, cmd *dto.ParkingOrderStatisticListQry, tenantID string) (interface{}, error) {
	return siteSvc.ParkingOrderStatisticList(cmd, tenantID), nil
}
func parkingInOutflow(c *gin.Context, cmd *dto.ParkingInAndOutflowQry, tenantID string) (interface{}, error) {
	return siteSvc.GetParkingInAndOutflow(cmd, tenantID), nil
}
func ridingCardPage(c *gin.Context, cmd *dto.OrderQueryPageCmd, tenantID string) (interface{}, error) {
	return visualSvc.RidingCardPage(cmd, tenantID), nil
}
func ridingCardList(c *gin.Context, cmd *dto.OrderQueryPageCmd, tenantID string) (interface{}, error) {
	return visualSvc.RidingCardList(cmd, tenantID), nil
}
func walletPage(c *gin.Context, cmd *dto.OrderQueryPageCmd, tenantID string) (interface{}, error) {
	return visualSvc.WalletPage(cmd, tenantID), nil
}
func walletList(c *gin.Context, cmd *dto.OrderQueryPageCmd, tenantID string) (interface{}, error) {
	return visualSvc.WalletList(cmd, tenantID), nil
}
func depositPage(c *gin.Context, cmd *dto.OrderQueryPageCmd, tenantID string) (interface{}, error) {
	return visualSvc.DepositPage(cmd, tenantID), nil
}
func deviceList(c *gin.Context, cmd *dto.DeviceListQry, tenantID string) (interface{}, error) {
	return deviceSvc.GetDeviceList(cmd, tenantID), nil
}
func deviceCarStatistics(c *gin.Context, _ *dto.Command, _ string) (interface{}, error) {
	return deviceSvc.CarStatistics(c.Request.Context())
}
func hotPointStart(c *gin.Context, cmd *dto.HotPointCmd, tenantID string) (interface{}, error) {
	return hotSvc.StartList(cmd, tenantID)
}
func hotPointEnd(c *gin.Context, cmd *dto.HotPointCmd, tenantID string) (interface{}, error) {
	return hotSvc.EndList(cmd, tenantID)
}
func operateLogPage(c *gin.Context, cmd *dto.OperateLogCmd, tenantID string) (interface{}, error) {
	return operateLogSvc.Page(c.Request.Context(), cmd, tenantID), nil
}

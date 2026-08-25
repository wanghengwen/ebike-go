package api

// GeneratedRoute describes a single API endpoint in the ebike-analyze service.
type GeneratedRoute struct {
	Path        string
	Method      string
	HandlerName string
	ApiModule   string
}

// GeneratedRoutes returns the full manifest of ebike-analyze POST endpoints.
func GeneratedRoutes() []GeneratedRoute {
	return []GeneratedRoute{
		// CarStatistics (3)
		{Path: "/car-statistics/page", Method: "POST", HandlerName: "CarStatisticsPage", ApiModule: "car_statistics"},
		{Path: "/car-statistics/queryByCarList", Method: "POST", HandlerName: "CarStatisticsQueryByCarList", ApiModule: "car_statistics"},
		{Path: "/carServiceStatistics/getList", Method: "POST", HandlerName: "CarServiceStatisticsGetList", ApiModule: "car_statistics"},
		// DeviceInfo (2)
		{Path: "/device/paas/device/list", Method: "POST", HandlerName: "DeviceList", ApiModule: "device_info"},
		{Path: "/device/paas/device/carStatistics", Method: "POST", HandlerName: "DeviceCarStatistics", ApiModule: "device_info"},
		// HotPoint (2)
		{Path: "/hot/point/start/list", Method: "POST", HandlerName: "HotPointStartList", ApiModule: "hot_point"},
		{Path: "/hot/point/end/list", Method: "POST", HandlerName: "HotPointEndList", ApiModule: "hot_point"},
		// OrderQuery (4)
		{Path: "/orderQuery/selectOrderCount", Method: "POST", HandlerName: "SelectOrderCount", ApiModule: "order_query"},
		{Path: "/orderQuery/selectOrderAnalyze", Method: "POST", HandlerName: "SelectOrderAnalyze", ApiModule: "order_query"},
		{Path: "/orderQuery/selectOrderList", Method: "POST", HandlerName: "SelectOrderList", ApiModule: "order_query"},
		{Path: "/orderDelete/deleteByQuery", Method: "POST", HandlerName: "DeleteByQuery", ApiModule: "order_query"},
		// SiteStatistics (6)
		{Path: "/site-statistics/list", Method: "POST", HandlerName: "SiteStatisticsList", ApiModule: "site_statistics"},
		{Path: "/parking/statistics", Method: "POST", HandlerName: "ParkingStatistics", ApiModule: "site_statistics"},
		{Path: "/parking/analyze/one", Method: "POST", HandlerName: "ParkingAnalyzeOne", ApiModule: "site_statistics"},
		{Path: "/parking/orderStatistics/page", Method: "POST", HandlerName: "ParkingOrderStatisticsPage", ApiModule: "site_statistics"},
		{Path: "/parking/orderStatistics/list", Method: "POST", HandlerName: "ParkingOrderStatisticsList", ApiModule: "site_statistics"},
		{Path: "/parking/getParkingInAndOutflow", Method: "POST", HandlerName: "ParkingInAndOutflow", ApiModule: "site_statistics"},
		// UserQuery (1)
		{Path: "/userQuery/selectUserCount", Method: "POST", HandlerName: "SelectUserCount", ApiModule: "user_query"},
		// UserStatistic (5)
		{Path: "/analyze/member", Method: "POST", HandlerName: "MemberStatistics", ApiModule: "user_statistic"},
		{Path: "/user/ageStatistic", Method: "POST", HandlerName: "AgeStatistic", ApiModule: "user_statistic"},
		{Path: "/user/ageStatistic/v2", Method: "POST", HandlerName: "AgeStatisticV2", ApiModule: "user_statistic"},
		{Path: "/user/allUserStatistic", Method: "POST", HandlerName: "AllUserStatistic", ApiModule: "user_statistic"},
		{Path: "/user/userStatisticHour", Method: "POST", HandlerName: "UserStatisticHour", ApiModule: "user_statistic"},
		// VisualQuery (5)
		{Path: "/riding_card_order/page", Method: "POST", HandlerName: "RidingCardOrderPage", ApiModule: "visual_query"},
		{Path: "/riding_card_order/list", Method: "POST", HandlerName: "RidingCardOrderList", ApiModule: "visual_query"},
		{Path: "/wallet_order/list", Method: "POST", HandlerName: "WalletOrderList", ApiModule: "visual_query"},
		{Path: "/wallet_order/page", Method: "POST", HandlerName: "WalletOrderPage", ApiModule: "visual_query"},
		{Path: "/deposit_order/page", Method: "POST", HandlerName: "DepositOrderPage", ApiModule: "visual_query"},
		// OperateLog (1)
		{Path: "/log/page", Method: "POST", HandlerName: "OperateLogPage", ApiModule: "operate_log"},
	}
}

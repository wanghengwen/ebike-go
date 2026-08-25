package app

import (
	"time"

	"ebike-analyze-go/internal/api/dto"
	"ebike-analyze-go/internal/infrastructure/persistence/model"
	"ebike-analyze-go/internal/infrastructure/persistence/repo"
)

type SiteStatisticsService struct{}

func (s *SiteStatisticsService) List(cmd *dto.SiteStatisticsCmd, tenantID string) []dto.SiteStatisticsCO {
	rows := repo.SiteStatisticsList(cmd.ParkingID, tenantID)
	out := make([]dto.SiteStatisticsCO, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.SiteStatisticsCO{
			ID: r.ID, ParkingID: r.ParkingID, OrderCount: r.OrderCount,
			OrderCost: r.OrderCost, StartTime: dto.NewDateTime(r.StartTime), EndTime: dto.NewDateTime(r.EndTime),
		})
	}
	return out
}

func (s *SiteStatisticsService) QueryStationStatistical(cmd *dto.ParkingStatisticalPageQuery, tenantID string) []dto.ParkingStationStatisticalCO {
	orderStrategy := cmd.OrderStrategy
	if orderStrategy == "" {
		orderStrategy = "canRent_asc"
	}
	orderBy := "can_rent desc"
	switch orderStrategy {
	case "canRent_asc":
		orderBy = "can_rent asc"
	case "canRent_desc":
		orderBy = "can_rent desc"
	case "idle_asc":
		orderBy = "idle asc"
	case "idle_desc":
		orderBy = "idle desc"
	case "siteOut_asc":
		orderBy = "site_out asc"
	case "siteOut_desc":
		orderBy = "site_out desc"
	case "ddMissOrder_asc":
		orderBy = "dd_miss_order asc"
	case "ddMissOrder_desc":
		orderBy = "dd_miss_order desc"
	}
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := formatDataTime()
	list1, list2 := repo.QueryStationStatistical(tenantID, cmd.ServiceID, cmd.AreaIds, start, end, orderBy)
	for i := range list1 {
		for _, d2 := range list2 {
			if list1[i].TenantID == d2.TenantID && list1[i].ServiceID == d2.ServiceID && list1[i].ParkingID == d2.ParkingID {
				list1[i].DdMissOrder = d2.DdMissOrder
			}
		}
	}
	out := make([]dto.ParkingStationStatisticalCO, 0, len(list1))
	for _, r := range list1 {
		out = append(out, dto.ParkingStationStatisticalCO{
			ServiceID: r.ServiceID, ParkingID: r.ParkingID,
			CanRent: r.CanRent, Idle: r.Idle, SiteOut: r.SiteOut, DdMissOrder: r.DdMissOrder,
			Booking: r.Booking, Operation: r.Operation, Alarm: r.Alarm, Fault: r.Fault,
			Idle13: r.Idle13, Idle36: r.Idle36, Idle612: r.Idle612,
			Idle1224: r.Idle1224, Idle2448: r.Idle2448, Idle48: r.Idle48,
		})
	}
	return out
}

func (s *SiteStatisticsService) AnalyzeOneParking(cmd *dto.OneParkingAnalyzeCmd, tenantID string) dto.OneParkingAnalyzeCO {
	result := dto.OneParkingAnalyzeCO{ServiceID: cmd.ServiceID, ParkingID: cmd.ParkingID, CarMap: map[string][]int{}}
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	latest := repo.AnalyzeLatestCar(tenantID, cmd.ServiceID, cmd.ParkingID, start)
	var idleList []int
	if latest != nil {
		idleList = []int{latest.Idle13, latest.Idle36, latest.Idle612, latest.Idle1224, latest.Idle2448, latest.Idle48}
		result.CanRent = latest.CanRent
		result.Booking = latest.Booking
		result.Operation = latest.Operation
	}
	end := formatDataTime()
	series := repo.AnalyzeCarSeries(tenantID, cmd.ServiceID, cmd.ParkingID, start, end)
	if len(series) == 0 {
		return result
	}
	fill := fillList(series[0].DataTime)
	canRentList := pluckInt(series, func(r model.ParkingStationStatistic) int { return r.CanRent })
	operationList := pluckInt(series, func(r model.ParkingStationStatistic) int { return r.Operation })
	bookingList := pluckInt(series, func(r model.ParkingStationStatistic) int { return r.Booking })
	if latest != nil && len(canRentList) > 0 {
		canRentList[len(canRentList)-1] = latest.CanRent
		operationList[len(operationList)-1] = latest.Operation
		bookingList[len(bookingList)-1] = latest.Booking
	}
	result.CarMap = map[string][]int{
		"canRent":     concatInt(fill, canRentList),
		"operation":   concatInt(fill, operationList),
		"booking":     concatInt(fill, bookingList),
		"alarm":       concatInt(fill, pluckInt(series, func(r model.ParkingStationStatistic) int { return r.Alarm })),
		"fault":       concatInt(fill, pluckInt(series, func(r model.ParkingStationStatistic) int { return r.Fault })),
		"siteOut":     concatInt(fill, pluckInt(series, func(r model.ParkingStationStatistic) int { return r.SiteOut })),
		"ddMissOrder": concatInt(fill, pluckInt(series, func(r model.ParkingStationStatistic) int { return r.DdMissOrder })),
		"idle":        idleList,
	}
	return result
}

func (s *SiteStatisticsService) ParkingOrderStatisticPage(qry *dto.ParkingOrderStatisticQry, tenantID string) dto.PageDTO[dto.ParkingOrderStatisticCO] {
	orderStrategy := qry.OrderStrategy
	if orderStrategy == "" {
		orderStrategy = "id_asc"
	}
	if orderStrategy == "id_asc" || orderStrategy == "id_desc" || orderStrategy == "areaSize_asc" || orderStrategy == "areaSize_desc" {
		fences, count := repo.GetParkingPage(tenantID, qry.ServiceID.Int64(), qry.MaintainAreaID.PtrInt64(), orderStrategy, qry.PageNum, qry.PageSize)
		if count == 0 {
			return dto.NewPageDTO(qry.PageNum, qry.PageSize, 0, []dto.ParkingOrderStatisticCO{})
		}
		ids := make([]int64, len(fences))
		for i, f := range fences {
			ids[i] = f.ID
		}
		// Path A mirrors Java's raw stationStatisticsMapper.analyzeOrderCarList call
		// directly (SiteStatisticsServiceImpl), which has NO +1h offset.
		stats := repo.AnalyzeOrderCarList(tenantID, qry.ServiceID.Int64(), ids, qry.StartTime.AsTime(), qry.EndTime.AsTime(), false)
		statMap := map[int64]model.ParkingStationStatistic{}
		for _, st := range stats {
			statMap[st.ParkingID] = st
		}
		list := make([]dto.ParkingOrderStatisticCO, 0, len(fences))
		for _, f := range fences {
			st := statMap[f.ID]
			name := f.Name
			areaSize := f.AreaSize
			// Java path A copies FenceDO → CO without mapping id → parkingId.
			list = append(list, dto.ParkingOrderStatisticCO{
				Name: &name, AreaSize: &areaSize,
				RideOrderCount: st.RideOrderCount, RideOrderCost: st.RideOrderCost,
				ReturnOrderCount: st.ReturnOrderCount, LowPowerMissOrderCount: st.DdMissOrder,
				OperationMissOrderCount: st.OperationMissOrderCount,
			})
		}
		return dto.NewPageDTO(qry.PageNum, qry.PageSize, count, list)
	}
	var parkingIDs []int64
	if qry.MaintainAreaID != nil {
		parkingIDs = repo.GetParkingIDsByMaintainArea(tenantID, qry.MaintainAreaID.Int64())
		if len(parkingIDs) == 0 {
			return dto.NewPageDTO(qry.PageNum, qry.PageSize, 0, []dto.ParkingOrderStatisticCO{})
		}
	}
	stats, count := repo.AnalyzeOrderCarPage(tenantID, qry.ServiceID.Int64(), parkingIDs, qry.StartTime.AsTime(), qry.EndTime.AsTime(), orderStrategy, qry.PageNum, qry.PageSize)
	if count == 0 {
		return dto.NewPageDTO(qry.PageNum, qry.PageSize, 0, []dto.ParkingOrderStatisticCO{})
	}
	ids := make([]int64, len(stats))
	for i, st := range stats {
		ids[i] = st.ParkingID
	}
	fences := repo.GetParkingByIDs(tenantID, ids)
	fenceMap := map[int64]model.Fence{}
	for _, f := range fences {
		fenceMap[f.ID] = f
	}
	list := make([]dto.ParkingOrderStatisticCO, 0, len(stats))
	for _, st := range stats {
		f := fenceMap[st.ParkingID]
		name := f.Name
		areaSize := f.AreaSize
		list = append(list, dto.ParkingOrderStatisticCO{
			ParkingID: st.ParkingID, Name: &name, AreaSize: &areaSize,
			RideOrderCount: st.RideOrderCount, RideOrderCost: st.RideOrderCost,
			ReturnOrderCount: st.ReturnOrderCount, LowPowerMissOrderCount: st.DdMissOrder,
			OperationMissOrderCount: st.OperationMissOrderCount,
		})
	}
	return dto.NewPageDTO(qry.PageNum, qry.PageSize, count, list)
}

func (s *SiteStatisticsService) ParkingOrderStatisticList(qry *dto.ParkingOrderStatisticListQry, tenantID string) []dto.ParkingOrderStatisticCO {
	// This endpoint calls through ParkingStationStatisticsGatewayImpl.analyzeOrderCarList
	// in Java, which DOES apply the +1h offset before delegating to the mapper.
	stats := repo.AnalyzeOrderCarList(tenantID, qry.ServiceID.Int64(), []int64(qry.ParkingIds), qry.StartTime.AsTime(), qry.EndTime.AsTime(), true)
	list := make([]dto.ParkingOrderStatisticCO, 0, len(stats))
	for _, st := range stats {
		list = append(list, dto.ParkingOrderStatisticCO{
			ParkingID: st.ParkingID, RideOrderCount: st.RideOrderCount, RideOrderCost: st.RideOrderCost,
			ReturnOrderCount: st.ReturnOrderCount, LowPowerMissOrderCount: st.DdMissOrder,
			OperationMissOrderCount: st.OperationMissOrderCount,
		})
	}
	return list
}

func (s *SiteStatisticsService) GetParkingInAndOutflow(qry *dto.ParkingInAndOutflowQry, tenantID string) dto.ParkingInAndOutflowCo {
	rows := repo.GetParkingInAndOutflow(qry.ParkingID, qry.TimeDimension, tenantID)
	influx := make([]int, len(rows))
	outFlow := make([]int, len(rows))
	canRent := make([]int, len(rows))
	for i, r := range rows {
		influx[i] = r.ReturnOrderCount
		outFlow[i] = r.RideOrderCount
		canRent[i] = r.CanRent
	}
	return dto.ParkingInAndOutflowCo{Influx: influx, OutFlow: outFlow, CanRent: canRent}
}

func formatDataTime() time.Time {
	now := time.Now().Add(time.Hour)
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
}

func fillList(dataTime time.Time) []int {
	h, m := dataTime.Hour(), dataTime.Minute()
	if m == 0 && h >= 1 && h <= 23 {
		return make([]int, h-1)
	}
	return make([]int, 23)
}

func pluckInt(rows []model.ParkingStationStatistic, fn func(model.ParkingStationStatistic) int) []int {
	out := make([]int, len(rows))
	for i, r := range rows {
		out[i] = fn(r)
	}
	return out
}

func concatInt(a, b []int) []int {
	out := make([]int, 0, len(a)+len(b))
	out = append(out, a...)
	out = append(out, b...)
	return out
}

type VisualQueryService struct{}

func (s *VisualQueryService) defaultRange(start, end *time.Time) (time.Time, time.Time) {
	if start != nil && end != nil {
		return *start, *end
	}
	now := time.Now()
	return now.AddDate(0, -1, 0), now
}

func (s *VisualQueryService) RidingCardPage(cmd *dto.OrderQueryPageCmd, tenantID string) dto.PageDTO[dto.RidingCardCO] {
	start, end := s.defaultRange(cmd.Start, cmd.End)
	raw, count := repo.QueryVisualPage("t_ebike_visual_riding_card_detail", tenantID, cmd.ServiceID, cmd.Name, cmd.Phone, cmd.Type, start, end, cmd.PageNum, cmd.PageSize, nil)
	rows, _ := raw.([]model.RidingCardDetail)
	return dto.NewPageDTO(cmd.PageNum, cmd.PageSize, count, mapRidingCards(rows))
}

func (s *VisualQueryService) RidingCardList(cmd *dto.OrderQueryPageCmd, tenantID string) []dto.RidingCardCO {
	start, end := s.defaultRange(cmd.Start, cmd.End)
	raw := repo.QueryVisualList("t_ebike_visual_riding_card_detail", tenantID, cmd.ServiceID, cmd.Name, cmd.Phone, cmd.Type, start, end, nil)
	rows, _ := raw.([]model.RidingCardDetail)
	return mapRidingCards(rows)
}

func (s *VisualQueryService) WalletPage(cmd *dto.OrderQueryPageCmd, tenantID string) dto.PageDTO[dto.WalletCO] {
	start, end := s.defaultRange(cmd.Start, cmd.End)
	raw, count := repo.QueryVisualPage("t_ebike_visual_wallet_detail", tenantID, cmd.ServiceID, cmd.Name, cmd.Phone, cmd.Type, start, end, cmd.PageNum, cmd.PageSize, nil)
	rows, _ := raw.([]model.WalletDetail)
	return dto.NewPageDTO(cmd.PageNum, cmd.PageSize, count, mapWallets(rows))
}

func (s *VisualQueryService) WalletList(cmd *dto.OrderQueryPageCmd, tenantID string) []dto.WalletCO {
	start, end := s.defaultRange(cmd.Start, cmd.End)
	raw := repo.QueryVisualList("t_ebike_visual_wallet_detail", tenantID, cmd.ServiceID, cmd.Name, cmd.Phone, cmd.Type, start, end, nil)
	rows, _ := raw.([]model.WalletDetail)
	return mapWallets(rows)
}

func (s *VisualQueryService) DepositPage(cmd *dto.OrderQueryPageCmd, tenantID string) dto.PageDTO[dto.DepositCO] {
	start, end := s.defaultRange(cmd.Start, cmd.End)
	izCard := 1
	raw, count := repo.QueryVisualPage("t_ebike_visual_deposit_card_detail", tenantID, cmd.ServiceID, cmd.Name, cmd.Phone, cmd.Type, start, end, cmd.PageNum, cmd.PageSize, &izCard)
	rows, _ := raw.([]model.DepositDetail)
	return dto.NewPageDTO(cmd.PageNum, cmd.PageSize, count, mapDeposits(rows))
}

func mapRidingCards(rows []model.RidingCardDetail) []dto.RidingCardCO {
	out := make([]dto.RidingCardCO, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.RidingCardCO{
			ID: r.ID, PinID: r.PinID, PinPhone: r.PinPhone, PinName: r.PinName,
			ConfigID: r.ConfigID, ServiceID: r.ServiceID, Type: r.Type, Channel: r.Channel,
			SysTradeNo: r.SysTradeNo, MerchantTradeNo: r.MerchantTradeNo, Amount: r.Amount,
			Name: r.Name, Duration: r.Duration, PaidAt: dto.NewDateTime(r.PaidAt), IzRefund: r.IzRefund,
		})
	}
	return out
}

func mapWallets(rows []model.WalletDetail) []dto.WalletCO {
	out := make([]dto.WalletCO, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.WalletCO{
			ID: r.ID, PinID: r.PinID, PinPhone: r.PinPhone, PinName: r.PinName,
			ServiceID: r.ServiceID, Type: r.Type, Channel: r.Channel,
			SysTradeNo: r.SysTradeNo, MerchantTradeNo: r.MerchantTradeNo, Amount: r.Amount,
			RechargeAmount: r.RechargeAmount, PresentAmount: r.PresentAmount,
			PaidAt: dto.NewDateTime(r.PaidAt), IzRefund: r.IzRefund,
		})
	}
	return out
}

func mapDeposits(rows []model.DepositDetail) []dto.DepositCO {
	out := make([]dto.DepositCO, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.DepositCO{
			ID: r.ID, PinID: r.PinID, PinPhone: r.PinPhone, PinName: r.PinName,
			ConfigID: r.ConfigID, ServiceID: r.ServiceID, Type: r.Type, Channel: r.Channel,
			SysTradeNo: r.SysTradeNo, MerchantTradeNo: r.MerchantTradeNo, Amount: r.Amount,
			Name: r.Name, Duration: r.Duration, IzCard: r.IzCard, PaidAt: dto.NewDateTime(r.PaidAt), IzRefund: r.IzRefund,
		})
	}
	return out
}

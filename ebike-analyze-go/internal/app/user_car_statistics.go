package app

import (
	"time"

	"ebike-analyze-go/internal/api/dto"
	"ebike-analyze-go/internal/infrastructure/persistence/model"
	"ebike-analyze-go/internal/infrastructure/persistence/repo"
)

type UserStatisticsService struct{}

func (s *UserStatisticsService) AllUserStatistic(cmd *dto.AllUserStatisticCmd, tenantID string) dto.AllUserStatisticCo {
	result := dto.AllUserStatisticCo{List: []dto.TodayUserStatisticCo{}}
	begin, end := resolveAllUserStatisticRange(cmd.Begin, cmd.End)
	rows := repo.SelectDayUserStatistic(cmd.ServiceID, begin, end, tenantID)
	for _, r := range rows {
		result.List = append(result.List, dto.TodayUserStatisticCo{
			QueryDate: dto.NewDateOnly(r.StatisticDate),
			Total:     r.TotalNum,
			ActiveCou: r.ActiveNum,
			NewCou:    r.CreateNum,
		})
	}
	if end.Format("2006-01-02") == todayDate().Format("2006-01-02") {
		if recent := repo.SelectRecentlyUserStatistic(cmd.ServiceID, tenantID); recent != nil {
			result.List = append(result.List, dto.TodayUserStatisticCo{
				QueryDate: dto.NewDateOnly(recent.StatisticDate),
				Total:     recent.TotalNum,
				ActiveCou: recent.ActiveNum,
				NewCou:    recent.CreateNum,
			})
		}
	}
	if len(result.List) > 0 {
		last := result.List[len(result.List)-1]
		result.TotalUser = int64(last.Total)
		result.ActiveUser = last.ActiveCou
		result.NewUser = last.NewCou
	}
	return result
}

func (s *UserStatisticsService) GetAgeStatistic(cmd *dto.AgeStatisticCmd, tenantID string) dto.AgeStatisticCO {
	co := dto.AgeStatisticCO{
		HaveRidingQualification: dto.RidingQualification{},
		NoRidingQualification:   dto.RidingQualification{},
	}
	rows := repo.GetAgeStatistics(cmd.ServiceIds, tenantID)
	for _, r := range rows {
		num := r.TotalNum
		var target *dto.AgeMap
		if r.IzRiding == 1 {
			target = pickGender(&co.HaveRidingQualification, r.Gender)
		} else {
			target = pickGender(&co.NoRidingQualification, r.Gender)
		}
		setScopeNumV1(r.Scope, num, target)
	}
	co.HaveRidingQualification.Total = co.HaveRidingQualification.Male.Total + co.HaveRidingQualification.Female.Total + co.HaveRidingQualification.Unknown.Total
	co.NoRidingQualification.Total = co.NoRidingQualification.Male.Total + co.NoRidingQualification.Female.Total + co.NoRidingQualification.Unknown.Total
	return co
}

func pickGender(rq *dto.RidingQualification, gender int) *dto.AgeMap {
	switch gender {
	case 0:
		return &rq.Male
	case 1:
		return &rq.Female
	default:
		return &rq.Unknown
	}
}

func setScopeNumV1(scope, personNum int, m *dto.AgeMap) {
	switch scope {
	case 0, 1:
		m.OneAgeScope += personNum
		m.Total += personNum
	case 2:
		m.TwoAgeScope = personNum
		m.Total += personNum
	case 3, 4:
		m.ThreeAgeScope += personNum
		m.Total += personNum
	case 5, 6, 7:
		m.FourAgeScope += personNum
		m.Total += personNum
	}
}

func (s *UserStatisticsService) GetAgeStatisticV2(cmd *dto.AgeStatisticCmd, tenantID string) dto.AgeStatisticCOV2 {
	result := dto.AgeStatisticCOV2{}
	rows := repo.GetAgeStatistics(cmd.ServiceIds, tenantID)
	group := map[int]map[int]int{}
	for _, r := range rows {
		if group[r.Scope] == nil {
			group[r.Scope] = map[int]int{}
		}
		group[r.Scope][r.Gender] += r.TotalNum
	}
	for scope, genderMaps := range group {
		switch scope {
		case 0:
			result.OneAgeScope = mergeGender(genderMaps, result.OneAgeScope)
		case 1, 2, 3:
			result.TwoAgeScope = mergeGender(genderMaps, result.TwoAgeScope)
		case 4:
			result.ThreeAgeScope = mergeGender(genderMaps, result.ThreeAgeScope)
		case 5:
			result.FourAgeScope = mergeGender(genderMaps, result.FourAgeScope)
		case 6:
			result.FiveAgeScope = mergeGender(genderMaps, result.FiveAgeScope)
		default:
			result.SixAgeScope = mergeGender(genderMaps, result.SixAgeScope)
		}
	}
	return result
}

func mergeGender(genderMaps map[int]int, gm dto.GenderMapV2) dto.GenderMapV2 {
	if v, ok := genderMaps[0]; ok {
		gm.Male += v
	}
	if v, ok := genderMaps[1]; ok {
		gm.Female += v
	}
	if v, ok := genderMaps[2]; ok {
		gm.Unknown += v
	}
	return gm
}

func (s *UserStatisticsService) GetMemberStatistics(cmd *dto.MemberStatisticQuery, tenantID string) map[string]interface{} {
	rows := repo.GetMemberStatistics(cmd.ServiceID, tenantID)
	var auth, authMember, authNoMember, authValidMember, authInvalidMember int
	var authHistoricalMember, authHistoricalNoMember, authCareer, authDeposit, authDepositCard, authFreeDeposit int
	var noAuth, noAuthMember, noAuthNoMember, noAuthValidMember, noAuthInvalidMember int
	var noAuthHistoricalMember, noAuthHistoricalNoMember, noAuthCareer, noAuthDeposit, noAuthDepositCard, noAuthFreeDeposit int
	for _, m := range rows {
		auth += m.Auth
		authMember += m.AuthMember
		authNoMember += m.AuthNoMember
		authValidMember += m.AuthValidMember
		authInvalidMember += m.AuthInvalidMember
		authHistoricalMember += m.AuthHistoricalMember
		authHistoricalNoMember += m.AuthHistoricalNoMember
		authCareer += m.AuthCareer
		authDeposit += m.AuthDeposit
		authDepositCard += m.AuthDepositCard
		authFreeDeposit += m.AuthFreeDeposit
		noAuth += m.NoAuth
		noAuthMember += m.NoAuthMember
		noAuthNoMember += m.NoAuthNoMember
		noAuthValidMember += m.NoAuthValidMember
		noAuthInvalidMember += m.NoAuthInvalidMember
		noAuthHistoricalMember += m.NoAuthHistoricalMember
		noAuthHistoricalNoMember += m.NoAuthHistoricalNoMember
		noAuthCareer += m.NoAuthCareer
		noAuthDeposit += m.NoAuthDeposit
		noAuthDepositCard += m.NoAuthDepositCard
		noAuthFreeDeposit += m.NoAuthFreeDeposit
	}
	sunburst := []map[string]interface{}{
		buildAuthTree("实名用户", auth, authMember, authNoMember, authValidMember, authInvalidMember,
			authHistoricalMember, noAuthHistoricalMember, authCareer, authDeposit, authDepositCard, authFreeDeposit),
		buildAuthTree("非实名用户", noAuth, noAuthMember, noAuthNoMember, noAuthValidMember, noAuthInvalidMember,
			noAuthHistoricalMember, noAuthHistoricalNoMember, noAuthCareer, noAuthDeposit, noAuthDepositCard, noAuthFreeDeposit),
	}
	info := map[string]interface{}{
		"noAuthentication": noAuth, "authentication": auth,
		"invalidMember": authInvalidMember, "validMember": authValidMember,
		"member": authMember, "nonMember": authNoMember, "historicalMember": authHistoricalMember,
		"career": authCareer, "deposit": authDeposit, "memberCard": authDepositCard,
		"noAuthInvalidMember": noAuthInvalidMember, "noAuthValidMember": noAuthValidMember,
		"noAuthMember": noAuthMember, "noAuthNonMember": noAuthNoMember,
		"noAuthHistoricalMember": noAuthHistoricalMember, "noAuthCareer": noAuthCareer,
		"noAuthDeposit": noAuthDeposit, "noAuthMemberCard": noAuthDepositCard,
		"freeUser": authFreeDeposit,
	}
	return map[string]interface{}{"sunburst": sunburst, "info": info}
}

func buildAuthTree(rootName string, auth, authMember, authNoMember, authValidMember, authInvalidMember,
	authHistoricalMember, historyNoMember, authCareer, authDeposit, authDepositCard, authFreeDeposit int) map[string]interface{} {
	validChildren := []map[string]interface{}{
		sunburstNode("职业认证", authCareer),
		sunburstNode("诚信金", authDeposit),
		sunburstNode("会员卡", authDepositCard),
		sunburstNode("一键免押会员", authFreeDeposit),
	}
	validMember := sunburstNode("有效会员", authValidMember)
	validMember["children"] = validChildren
	invalidMember := sunburstNode("待失效会员", authInvalidMember)
	noMemberChildren := []map[string]interface{}{
		sunburstNode("历史会员", authHistoricalMember),
		sunburstNode("历史非会员", historyNoMember),
	}
	noMember := sunburstNode("非会员用户", authNoMember)
	noMember["children"] = noMemberChildren
	memberChildren := []map[string]interface{}{validMember, invalidMember}
	member := sunburstNode("会员用户", authMember)
	member["children"] = memberChildren
	rootChildren := []map[string]interface{}{member, noMember}
	root := sunburstNode(rootName, auth)
	root["children"] = rootChildren
	return root
}

func sunburstNode(name string, value int) map[string]interface{} {
	return map[string]interface{}{"name": name, "value": value}
}

func (s *UserStatisticsService) UserStatisticHour(cmd *dto.UserStatisticHourQry, tenantID string) []dto.UserStatisticHourCo {
	queryDate := time.Now()
	if cmd.QueryDate != nil {
		queryDate = cmd.QueryDate.AsTime()
	}
	rows := repo.SelectHourUserStatistic(cmd.ServiceID, queryDate, tenantID)
	byHour := map[int]model.UserStatistic{}
	maxHour := 24
	for _, r := range rows {
		byHour[r.StatisticHour] = r
		if r.StatisticHour > maxHour {
			maxHour = r.StatisticHour
		}
	}
	var out []dto.UserStatisticHourCo
	for i := 1; i <= 24; i++ {
		co := dto.UserStatisticHourCo{StatisticHour: i}
		if i == 1 {
			if r, ok := byHour[i]; ok {
				co.ActiveNum = r.HourActiveNum
				co.CreateNum = r.CreateNum
			}
		} else if i <= maxHour {
			end, ok := byHour[i]
			if ok {
				co.ActiveNum = end.HourActiveNum
				prevCreate := 0
				if start, ok2 := byHour[i-1]; ok2 {
					prevCreate = start.CreateNum
				}
				co.CreateNum = end.CreateNum - prevCreate
			}
		}
		out = append(out, co)
	}
	return out
}

type CarStatisticsService struct{}

func (s *CarStatisticsService) Page(cmd *dto.CarStatisticsQuery, tenantID string) dto.PageDTO[dto.CarStatisticsCO] {
	pageNum, pageSize := cmd.PageNum, cmd.PageSize
	rows, count := repo.CarStatisticsPage(cmd.CarID, cmd.ServiceID, pageNum, pageSize, tenantID)
	list := make([]dto.CarStatisticsCO, 0, len(rows))
	for _, r := range rows {
		list = append(list, mapCarStatPage(r))
	}
	return dto.NewPageDTO(pageNum, pageSize, count, list)
}

func (s *CarStatisticsService) QueryByCarList(cmd *dto.CarStatisticsListQuery, tenantID string) []dto.CarStatisticsCO {
	rows := repo.CarStatisticsByCarList(cmd.ServiceID.Int64(), cmd.CarIds, cmd.StartTime.AsTime(), cmd.EndTime.AsTime(), tenantID)
	list := make([]dto.CarStatisticsCO, 0, len(rows))
	for _, r := range rows {
		list = append(list, mapCarStat(r))
	}
	return list
}

func (s *CarStatisticsService) GetServiceList(cmd *dto.CarServiceStatisticsQuery, tenantID string) []dto.CarServiceStatisticsCo {
	rows := repo.ServiceCarStatisticsList(cmd.ServiceIds, tenantID)
	list := make([]dto.CarServiceStatisticsCo, 0, len(rows))
	for _, r := range rows {
		createdAt := dto.NewLocalDateTime(r.CreatedAt)
		updatedAt := dto.NewLocalDateTime(r.UpdatedAt)
		list = append(list, dto.CarServiceStatisticsCo{
			ID: r.ID, ServiceID: r.ServiceID, CanRent: r.CanRent, Booking: r.Booking, Riding: r.Riding,
			Parking: r.Parking, Operation: r.Operation, LowBattery: r.LowBattery,
			FreeTimeOneToThree: r.FreeTimeOneToThree, FreeTimeThreeToSix: r.FreeTimeThreeToSix,
			FreeTimeSixToTwelve: r.FreeTimeSixToTwelve, FreeTimeHalfOrOneDay: r.FreeTimeHalfOrOneDay,
			FreeTimeOneOrTowDay: r.FreeTimeOneOrTowDay, FreeTimeTowDayMore: r.FreeTimeTowDayMore,
			VoltageZero: r.VoltageZero, VoltageZeroToTwenty: r.VoltageZeroToTwenty,
			VoltageTwentyToThirtyFive: r.VoltageTwentyToThirtyFive, VoltageThirtyFiveMore: r.VoltageThirtyFiveMore,
			TenantID: r.TenantID, CreatedPin: r.CreatedPin, CreatedAt: &createdAt,
			UpdatedPin: r.UpdatedPin, UpdatedAt: &updatedAt, Version: r.Version, IzDel: r.IzDel,
		})
	}
	return list
}

// mapCarStatPage mirrors Java page: CarStatisticsDO → CarStatisticsEntity → CO,
// which omits ddMissOrder/operationMissOrderCount/changeBatteryCount/repairCount/moveCarCount.
func mapCarStatPage(r model.CarStatistic) dto.CarStatisticsCO {
	return dto.CarStatisticsCO{
		CarID: r.CarID, Imei: stringPtrOrNil(r.Imei),
		OrderCount: intPtr(r.OrderCount), OrderCost: int64Ptr(r.OrderCost),
		RidingDistance: int64Ptr(r.RidingDistance), RidingTime: int64Ptr(r.RidingTime),
	}
}

func mapCarStat(r model.CarStatistic) dto.CarStatisticsCO {
	return dto.CarStatisticsCO{
		CarID: r.CarID, Imei: stringPtrOrNil(r.Imei),
		OrderCount: intPtr(r.OrderCount), OrderCost: int64Ptr(r.OrderCost),
		RidingDistance: int64Ptr(r.RidingDistance), RidingTime: int64Ptr(r.RidingTime),
		DdMissOrder: intPtr(r.DdMissOrder), OperationMissOrderCount: intPtr(r.OperationMissOrderCount),
		ChangeBatteryCount: intPtr(r.ChangeBatteryCount), RepairCount: intPtr(r.RepairCount),
		MoveCarCount: intPtr(r.MoveCarCount),
	}
}

func intPtr(v int) *int       { return &v }
func int64Ptr(v int64) *int64 { return &v }

func stringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func todayDate() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// resolveAllUserStatisticRange mirrors Java UserStatisticsServiceImpl.allUserStatistic:
// when begin is null, both begin and end default to today; when only end is null, end defaults to today.
func resolveAllUserStatisticRange(begin, end *dto.DateOnlyValue) (time.Time, time.Time) {
	today := todayDate()
	if begin == nil {
		return today, today
	}
	b := begin.LocalDate()
	if end == nil {
		return b, today
	}
	return b, end.LocalDate()
}

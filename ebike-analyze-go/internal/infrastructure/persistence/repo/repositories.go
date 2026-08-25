package repo

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"ebike-analyze-go/internal/api/dto"
	"ebike-analyze-go/internal/infrastructure/persistence/model"
	"ebike-analyze-go/internal/infrastructure/persistence/tenant"
	"ebike-analyze-go/internal/pkg/mysql"
	"gorm.io/gorm"
)

type UserStatisticDto struct {
	StatisticDate time.Time
	TotalNum      int
	ActiveNum     int
	CreateNum     int
}

func db() *gorm.DB { return mysql.DB }

func SelectDayUserStatistic(serviceIDs []int64, begin, end time.Time, tenantID string) []UserStatisticDto {
	if len(serviceIDs) == 0 || db() == nil {
		return nil
	}
	begin = dto.NormalizeLocalDate(begin)
	end = dto.NormalizeLocalDate(end)
	var rows []model.UserStatistic
	q := model.ScopedDB(db(), tenantID).Model(&model.UserStatistic{}).
		Where("service_id IN ?", serviceIDs).
		Where("statistic_date BETWEEN ? AND ?", begin, end).
		Where("statistic_hour = ?", 24)
	q.Find(&rows)
	if len(rows) == 0 {
		return nil
	}
	seen := map[string]bool{}
	byDate := map[string][]model.UserStatistic{}
	for _, r := range rows {
		key := fmt.Sprintf("%d;%s;%d", r.ServiceID, r.StatisticDate.Format("2006-01-02"), r.StatisticHour)
		if seen[key] {
			continue
		}
		seen[key] = true
		dk := r.StatisticDate.Format("2006-01-02")
		byDate[dk] = append(byDate[dk], r)
	}
	var result []UserStatisticDto
	for _, list := range byDate {
		dto := aggregateUserStats(list)
		if dto != nil {
			result = append(result, *dto)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].StatisticDate.Before(result[j].StatisticDate)
	})
	return result
}

func SelectRecentlyUserStatistic(serviceIDs []int64, tenantID string) *UserStatisticDto {
	if len(serviceIDs) == 0 || db() == nil {
		return nil
	}
	var latest model.UserStatistic
	err := model.ScopedDB(db(), tenantID).Model(&model.UserStatistic{}).
		Where("service_id IN ?", serviceIDs).
		Order("statistic_date DESC, statistic_hour DESC").
		First(&latest).Error
	if err != nil {
		return nil
	}
	var rows []model.UserStatistic
	model.ScopedDB(db(), tenantID).Model(&model.UserStatistic{}).
		Where("service_id IN ?", serviceIDs).
		Where("statistic_date = ? AND statistic_hour = ?", latest.StatisticDate, latest.StatisticHour).
		Find(&rows)
	seen := map[string]bool{}
	var filtered []model.UserStatistic
	for _, r := range rows {
		key := fmt.Sprintf("%d;%s;%d", r.ServiceID, r.StatisticDate.Format("2006-01-02"), r.StatisticHour)
		if !seen[key] {
			seen[key] = true
			filtered = append(filtered, r)
		}
	}
	return aggregateUserStats(filtered)
}

func aggregateUserStats(list []model.UserStatistic) *UserStatisticDto {
	if len(list) == 0 {
		return nil
	}
	dto := &UserStatisticDto{StatisticDate: list[0].StatisticDate}
	for _, r := range list {
		dto.TotalNum += r.TotalNum
		dto.ActiveNum += r.ActiveNum
		dto.CreateNum += r.CreateNum
	}
	return dto
}

func SelectHourUserStatistic(serviceID int64, date time.Time, tenantID string) []model.UserStatistic {
	if db() == nil {
		return nil
	}
	var rows []model.UserStatistic
	model.ScopedDB(db(), tenantID).Model(&model.UserStatistic{}).
		Where("service_id = ? AND statistic_date = ?", serviceID, date).
		Order("statistic_hour ASC").
		Find(&rows)
	return rows
}

func GetAgeStatistics(serviceIDs []int64, tenantID string) []model.AgeStatistic {
	if len(serviceIDs) == 0 || db() == nil {
		return nil
	}
	var latest model.AgeStatistic
	err := model.ScopedDB(db(), tenantID).Model(&model.AgeStatistic{}).
		Where("service_id IN ?", serviceIDs).
		Order("statistic_at DESC").
		First(&latest).Error
	if err != nil {
		return nil
	}
	var rows []model.AgeStatistic
	model.ScopedDB(db(), tenantID).Model(&model.AgeStatistic{}).
		Where("service_id IN ? AND statistic_at = ?", serviceIDs, latest.StatisticAt).
		Find(&rows)
	seen := map[string]bool{}
	var out []model.AgeStatistic
	for _, r := range rows {
		key := fmt.Sprintf("%d;%s;%d;%d;%d", r.ServiceID, r.StatisticAt.Format(time.RFC3339), r.Gender, r.IzRiding, r.Scope)
		if !seen[key] {
			seen[key] = true
			out = append(out, r)
		}
	}
	return out
}

func GetMemberStatistics(serviceIDs []int64, tenantID string) []model.MemberStatistic {
	if len(serviceIDs) == 0 || db() == nil {
		return nil
	}
	var latest model.MemberStatistic
	err := model.ScopedDB(db(), tenantID).Model(&model.MemberStatistic{}).
		Where("service_id IN ?", serviceIDs).
		Order("statistic_at DESC").
		First(&latest).Error
	if err != nil {
		return nil
	}
	var rows []model.MemberStatistic
	model.ScopedDB(db(), tenantID).Model(&model.MemberStatistic{}).
		Where("service_id IN ? AND statistic_at = ?", serviceIDs, latest.StatisticAt).
		Order("id DESC").
		Find(&rows)
	seen := map[string]bool{}
	var out []model.MemberStatistic
	for _, r := range rows {
		key := fmt.Sprintf("%d;%s", r.ServiceID, r.StatisticAt.Format(time.RFC3339))
		if !seen[key] {
			seen[key] = true
			out = append(out, r)
		}
	}
	return out
}

func CarStatisticsPage(carID string, serviceID *int64, pageNum, pageSize int, tenantID string) ([]model.CarStatistic, int64) {
	if db() == nil {
		return nil, 0
	}
	table := model.CarStatistic{}.ResolvedTable(tenantID)
	q := tenant.WithTenant(db(), tenantID).Table(table)
	if carID != "" {
		q = q.Where("car_id = ?", carID)
	}
	if serviceID != nil {
		q = q.Where("service_id = ?", *serviceID)
	}
	var count int64
	q.Count(&count)
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var rows []model.CarStatistic
	// Java's page() only orders when the caller passes explicit `orders`
	// (rarely used); absent that, MySQL/InnoDB scans in clustered PK order.
	// Ordering by id gives deterministic, closely-equivalent behavior.
	q.Order("id ASC").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&rows)
	return rows, count
}

func CarStatisticsByCarList(serviceID int64, carIDs []string, start, end time.Time, tenantID string) []model.CarStatistic {
	if db() == nil || len(carIDs) == 0 {
		return nil
	}
	table := model.CarStatistic{}.ResolvedTable(tenantID)
	var rows []model.CarStatistic
	tenant.WithTenant(db(), tenantID).Table(table).
		Select(`service_id, car_id, tenant_id,
			sum(order_count) as order_count, sum(order_cost) as order_cost,
			sum(riding_distance) as riding_distance, sum(riding_time) as riding_time,
			sum(dd_miss_order) as dd_miss_order, sum(operation_miss_order_count) as operation_miss_order_count,
			sum(change_battery_count) as change_battery_count, sum(repair_count) as repair_count,
			sum(move_car_count) as move_car_count`).
		Where("service_id = ? AND car_id IN ?", serviceID, carIDs).
		Where("data_time >= ? AND data_time <= ?", start, end).
		Group("car_id").
		Find(&rows)
	return rows
}

func ServiceCarStatisticsList(serviceIDs []int64, tenantID string) []model.ServiceCarStatistic {
	if db() == nil || len(serviceIDs) == 0 {
		return nil
	}
	now := time.Now()
	hour := now.Hour()
	// Java: endTime = today.atTime(hour, 59, 0);
	//       startTime = yesterday.atTime(now.plusHours(1).getHour(), 0, 0);
	// The wrapped hour (now+1h, mod 24) is combined with YESTERDAY's date, not
	// with whatever date the +1h would roll onto. At hour==23 this makes
	// start = yesterday 00:00 (not today 00:00 as naive date-arithmetic would give).
	yesterday := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	startHour := now.Add(time.Hour).Hour()
	start := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), startHour, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), hour, 59, 0, 0, now.Location())
	var rows []model.ServiceCarStatistic
	model.ScopedDB(db(), tenantID).Model(&model.ServiceCarStatistic{}).
		Where("service_id IN ?", serviceIDs).
		Where("created_at BETWEEN ? AND ?", start, end).
		Order("created_at ASC").
		Find(&rows)
	return rows
}

func SiteStatisticsList(parkingID *int64, tenantID string) []model.SiteStatistic {
	if db() == nil {
		return nil
	}
	q := model.ScopedDB(db(), tenantID).Model(&model.SiteStatistic{})
	if parkingID != nil {
		q = q.Where("parking_id = ?", *parkingID)
	}
	var rows []model.SiteStatistic
	q.Find(&rows)
	return rows
}

func parkingTable(tenantID string) string {
	return model.ParkingStationStatistic{}.ResolvedTable(tenantID)
}

func QueryStationStatistical(tenantID string, serviceID int64, parkingIDs []int64, start, end time.Time, orderBy string) ([]model.ParkingStationStatistic, []model.ParkingStationStatistic) {
	if db() == nil || len(parkingIDs) == 0 {
		return nil, nil
	}
	table := parkingTable(tenantID)
	inClause := buildInClause(parkingIDs)
	var list1 []model.ParkingStationStatistic
	sql1 := fmt.Sprintf(`SELECT id,tenant_id,service_id,parking_id,data_time,can_rent,idle,site_out
		FROM %s t1 WHERE t1.tenant_id = ? AND t1.service_id = ? AND t1.parking_id IN (%s)
		AND t1.data_time = (SELECT max(t2.data_time) FROM %s t2 WHERE t2.tenant_id = ? AND t2.service_id = ?
		AND t2.parking_id IN (%s) AND t2.data_time BETWEEN ? AND ?)
		ORDER BY %s`, table, inClause, table, inClause, orderBy)
	args := []interface{}{tenantID, serviceID}
	for _, id := range parkingIDs {
		args = append(args, id)
	}
	args = append(args, tenantID, serviceID)
	for _, id := range parkingIDs {
		args = append(args, id)
	}
	args = append(args, start, end)
	db().Raw(sql1, args...).Scan(&list1)

	sql2 := fmt.Sprintf(`SELECT tenant_id,service_id,parking_id,sum(dd_miss_order) as dd_miss_order
		FROM %s WHERE tenant_id = ? AND service_id = ? AND parking_id IN (%s)
		AND data_time BETWEEN ? AND ? GROUP BY tenant_id,service_id,parking_id`, table, inClause)
	args2 := []interface{}{tenantID, serviceID}
	for _, id := range parkingIDs {
		args2 = append(args2, id)
	}
	args2 = append(args2, start, end)
	var list2 []model.ParkingStationStatistic
	db().Raw(sql2, args2...).Scan(&list2)
	return list1, list2
}

func AnalyzeLatestCar(tenantID string, serviceID, parkingID int64, start time.Time) *model.ParkingStationStatistic {
	if db() == nil {
		return nil
	}
	table := parkingTable(tenantID)
	var row model.ParkingStationStatistic
	db().Raw(fmt.Sprintf(`SELECT service_id,parking_id,data_time,idle1_3,idle3_6,idle6_12,idle12_24,idle24_48,idle48,can_rent,booking,operation
		FROM %s WHERE tenant_id = ? AND service_id = ? AND parking_id = ? AND data_time >= ?
		ORDER BY data_time DESC LIMIT 1`, table), tenantID, serviceID, parkingID, start).Scan(&row)
	if row.ParkingID == 0 {
		return nil
	}
	return &row
}

func AnalyzeCarSeries(tenantID string, serviceID, parkingID int64, start, end time.Time) []model.ParkingStationStatistic {
	if db() == nil {
		return nil
	}
	table := parkingTable(tenantID)
	var rows []model.ParkingStationStatistic
	db().Raw(fmt.Sprintf(`SELECT service_id,parking_id,data_time,can_rent,operation,booking,alarm,fault,site_out,dd_miss_order
		FROM %s WHERE tenant_id = ? AND service_id = ? AND parking_id = ?
		AND data_time > ? AND data_time <= ? ORDER BY data_time ASC`, table),
		tenantID, serviceID, parkingID, start, end).Scan(&rows)
	return rows
}

// AnalyzeOrderCarList mirrors Java's stationStatisticsMapper.analyzeOrderCarList
// (raw Mapper SQL, no time offset) when applyHourOffset is false. Java's
// ParkingStationStatisticsGatewayImpl.analyzeOrderCarList wraps that same
// mapper call but shifts start/end by +1h before querying; pass
// applyHourOffset=true to reproduce that Gateway path instead.
func AnalyzeOrderCarList(tenantID string, serviceID int64, parkingIDs []int64, start, end time.Time, applyHourOffset bool) []model.ParkingStationStatistic {
	if db() == nil {
		return nil
	}
	if applyHourOffset {
		start = start.Add(time.Hour)
		end = end.Add(time.Hour)
	}
	table := parkingTable(tenantID)
	q := tenant.WithTenant(db(), tenantID).Table(table).
		Select(`parking_id, sum(dd_miss_order) as dd_miss_order,
			sum(operation_miss_order_count) as operation_miss_order_count,
			sum(ride_order_count) as ride_order_count, sum(ride_order_cost) as ride_order_cost,
			sum(return_order_count) as return_order_count`).
		Where("service_id = ?", serviceID).
		Where("data_time >= ? AND data_time <= ?", start, end)
	if len(parkingIDs) > 0 {
		q = q.Where("parking_id IN ?", parkingIDs)
	}
	var rows []model.ParkingStationStatistic
	q.Group("parking_id").Find(&rows)
	return rows
}

func AnalyzeOrderCarPage(tenantID string, serviceID int64, parkingIDs []int64, start, end time.Time, orderStrategy string, pageNum, pageSize int) ([]model.ParkingStationStatistic, int64) {
	if db() == nil {
		return nil, 0
	}
	table := parkingTable(tenantID)
	q := tenant.WithTenant(db(), tenantID).Table(table).
		Select(`parking_id, sum(dd_miss_order) as dd_miss_order,
			sum(operation_miss_order_count) as operation_miss_order_count,
			sum(ride_order_count) as ride_order_count, sum(ride_order_cost) as ride_order_cost,
			sum(return_order_count) as return_order_count`).
		Where("service_id = ?", serviceID).
		Where("data_time > ? AND data_time < ?", start, end)
	if len(parkingIDs) > 0 {
		q = q.Where("parking_id IN ?", parkingIDs)
	}
	q = q.Group("parking_id")
	// Count before ORDER BY: with GROUP BY, GORM keeps ORDER BY in count SQL,
	// which can fail under ONLY_FULL_GROUP_BY (Java MP pagination strips it).
	var count int64
	q.Count(&count)
	orderCol := orderStrategyToColumn(orderStrategy)
	if orderCol != "" {
		if strings.HasSuffix(orderStrategy, "_desc") {
			q = q.Order(orderCol + " DESC")
		} else {
			q = q.Order(orderCol + " ASC")
		}
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var rows []model.ParkingStationStatistic
	q.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&rows)
	return rows, count
}

func orderStrategyToColumn(s string) string {
	switch s {
	case "rideOrderCount_asc", "rideOrderCount_desc":
		return "ride_order_count"
	case "rideOrderCost_asc", "rideOrderCost_desc":
		return "ride_order_cost"
	case "returnOrderCount_asc", "returnOrderCount_desc":
		return "return_order_count"
	case "lowPowerMissOrderCount_asc", "lowPowerMissOrderCount_desc":
		return "dd_miss_order"
	case "operationMissOrderCount_asc", "operationMissOrderCount_desc":
		return "operation_miss_order_count"
	}
	return ""
}

func GetParkingPage(tenantID string, serviceID int64, maintainAreaID *int64, orderStrategy string, pageNum, pageSize int) ([]model.Fence, int64) {
	if db() == nil {
		return nil, 0
	}
	q := model.ScopedDB(db(), tenantID).Model(&model.Fence{}).Where("type = 2 AND service_id = ?", serviceID)
	if maintainAreaID != nil {
		q = q.Where("maintain_area_id = ?", *maintainAreaID)
	}
	switch orderStrategy {
	case "id_desc":
		q = q.Order("id DESC")
	case "areaSize_asc":
		q = q.Order("area_size ASC")
	case "areaSize_desc":
		q = q.Order("area_size DESC")
	default:
		q = q.Order("id ASC")
	}
	var count int64
	q.Count(&count)
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var rows []model.Fence
	q.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&rows)
	return rows, count
}

func GetParkingIDsByMaintainArea(tenantID string, maintainAreaID int64) []int64 {
	if db() == nil {
		return nil
	}
	var ids []int64
	model.ScopedDB(db(), tenantID).Model(&model.Fence{}).
		Where("type = 2 AND maintain_area_id = ?", maintainAreaID).
		Pluck("id", &ids)
	return ids
}

func GetParkingByIDs(tenantID string, ids []int64) []model.Fence {
	if db() == nil || len(ids) == 0 {
		return nil
	}
	var rows []model.Fence
	model.ScopedDB(db(), tenantID).Model(&model.Fence{}).Where("id IN ?", ids).Find(&rows)
	return rows
}

func GetParkingInAndOutflow(parkingID int64, timeDimension int, tenantID string) []model.ParkingStationStatistic {
	if db() == nil {
		return nil
	}
	end := formatDataTime()
	table := parkingTable(tenantID)
	q := tenant.WithTenant(db(), tenantID).Table(table)
	var rows []model.ParkingStationStatistic
	switch timeDimension {
	case 0:
		start := end.Add(-24 * time.Hour)
		q.Select("data_time, ride_order_count, return_order_count, can_rent").
			Where("parking_id = ?", parkingID).
			Where("data_time > ? AND data_time <= ?", start, end).
			Order("data_time DESC").
			Find(&rows)
	case 1:
		start := end.AddDate(0, 0, -13)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
		q.Select("DATE_FORMAT(data_time,'%Y-%m-%d') as data_time, sum(ride_order_count) as ride_order_count, sum(return_order_count) as return_order_count").
			Where("parking_id = ?", parkingID).
			Where("data_time > ? AND data_time <= ?", start, end).
			Group("DATE_FORMAT(data_time,'%Y-%m-%d')").
			Order("data_time DESC").
			Find(&rows)
	case 2:
		start := end.AddDate(0, -5, 0)
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
		q.Select("DATE_FORMAT(data_time,'%Y-%m-01') as data_time, sum(ride_order_count) as ride_order_count, sum(return_order_count) as return_order_count").
			Where("parking_id = ?", parkingID).
			Where("data_time > ? AND data_time <= ?", start, end).
			Group("DATE_FORMAT(data_time,'%Y-%m-01')").
			Order("data_time DESC").
			Find(&rows)
	}
	return parkingInAndOutflowPerfect(rows, timeDimension, end)
}

func formatDataTime() time.Time {
	now := time.Now().Add(time.Hour)
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
}

func parkingInAndOutflowPerfect(rows []model.ParkingStationStatistic, timeDimension int, end time.Time) []model.ParkingStationStatistic {
	flowSize := 0
	switch timeDimension {
	case 0:
		flowSize = 24
	case 1:
		flowSize = 14
	case 2:
		flowSize = 6
	}
	if len(rows) >= flowSize {
		return rows
	}
	collect := map[string]model.ParkingStationStatistic{}
	for _, r := range rows {
		collect[r.DataTime.Format("2006-01-02 15:04:05")] = r
	}
	var out []model.ParkingStationStatistic
	for i := 0; i < flowSize; i++ {
		var slot time.Time
		switch timeDimension {
		case 0:
			slot = end.Add(-time.Duration(i) * time.Hour)
		case 1:
			t := end.AddDate(0, 0, -i)
			slot = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		case 2:
			t := end.AddDate(0, -i, 0)
			slot = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		}
		key := slot.Format("2006-01-02 15:04:05")
		if r, ok := collect[key]; ok {
			out = append(out, r)
		} else {
			out = append(out, model.ParkingStationStatistic{RideOrderCount: 0, ReturnOrderCount: 0, CanRent: 0})
		}
	}
	return out
}

func buildInClause(ids []int64) string {
	parts := make([]string, len(ids))
	for i := range ids {
		parts[i] = "?"
	}
	return strings.Join(parts, ",")
}

// Visual query repos use VisualDB
func visualDB() *gorm.DB { return mysql.VisualDB }

func QueryVisualPage(table, tenantID string, serviceID int64, name, phone string, typ *int, start, end time.Time, pageNum, pageSize int, izCard *int) (interface{}, int64) {
	return queryVisualOrders(table, tenantID, serviceID, name, phone, typ, start, end, pageNum, pageSize, izCard)
}

func QueryVisualList(table, tenantID string, serviceID int64, name, phone string, typ *int, start, end time.Time, izCard *int) interface{} {
	return queryVisualList(table, tenantID, serviceID, name, phone, typ, start, end, izCard)
}

func queryVisualOrders(table, tenantID string, serviceID int64, name, phone string, typ *int, start, end time.Time, pageNum, pageSize int, izCard *int) (interface{}, int64) {
	if visualDB() == nil {
		return nil, 0
	}
	tbl := tenant.ResolveTableName(table, tenantID)
	q := tenant.WithTenant(visualDB(), tenantID).Table(tbl).Where("service_id = ?", serviceID)
	if name != "" {
		q = q.Where("pin_name = ?", name)
	}
	if phone != "" {
		q = q.Where("pin_phone = ?", phone)
	}
	if typ != nil {
		q = q.Where("type = ?", *typ)
	}
	if izCard != nil {
		q = q.Where("iz_card = ?", *izCard)
	}
	q = q.Where("paid_at BETWEEN ? AND ?", start, end)
	if table != "t_ebike_visual_deposit_card_detail" {
		q = q.Order("paid_at DESC")
	}
	var count int64
	q.Count(&count)
	if pageNum > 0 && pageSize > 0 {
		var page interface{}
		switch table {
		case "t_ebike_visual_riding_card_detail":
			var rows []model.RidingCardDetail
			q.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&rows)
			page = rows
		case "t_ebike_visual_wallet_detail":
			var rows []model.WalletDetail
			q.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&rows)
			page = rows
		case "t_ebike_visual_deposit_card_detail":
			var rows []model.DepositDetail
			q.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&rows)
			page = rows
		}
		return page, count
	}
	return nil, count
}

func queryVisualList(table, tenantID string, serviceID int64, name, phone string, typ *int, start, end time.Time, izCard *int) interface{} {
	if visualDB() == nil {
		return nil
	}
	tbl := tenant.ResolveTableName(table, tenantID)
	q := tenant.WithTenant(visualDB(), tenantID).Table(tbl).Where("service_id = ?", serviceID)
	if name != "" {
		q = q.Where("pin_name = ?", name)
	}
	if phone != "" {
		q = q.Where("pin_phone = ?", phone)
	}
	if typ != nil {
		q = q.Where("type = ?", *typ)
	}
	if izCard != nil {
		q = q.Where("iz_card = ?", *izCard)
	}
	q = q.Where("paid_at BETWEEN ? AND ?", start, end).Order("paid_at DESC")
	switch table {
	case "t_ebike_visual_riding_card_detail":
		var rows []model.RidingCardDetail
		q.Find(&rows)
		return rows
	case "t_ebike_visual_wallet_detail":
		var rows []model.WalletDetail
		q.Find(&rows)
		return rows
	}
	return nil
}

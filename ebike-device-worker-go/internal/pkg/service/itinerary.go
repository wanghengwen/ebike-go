package service

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"ebike-device-worker-go/internal/api/dto"
	"ebike-device-worker-go/internal/pkg/db"
	"ebike-device-worker-go/internal/pkg/model"
	"ebike-device-worker-go/internal/pkg/utils"
	"ebike-device-worker-go/internal/pkg/web"

	"gorm.io/gorm"
)

// GetTrajectory mirrors Java EBikeItineraryServiceImpl.getTrajectoryByImeiAndTime / buildTrajectoryListByPg.
func GetTrajectory(cmd dto.TrajectoryCmd) ([]dto.TrajectoryCo, error) {
	if err := validateTrajectoryTimeRange(cmd.StartTime, cmd.EndTime); err != nil {
		return nil, err
	}

	gpsList, err := queryGpsByImeiAndTime(cmd.Imei, cmd.StartTime, cmd.EndTime)
	if err != nil {
		return nil, err
	}

	trajectories := make([]dto.TrajectoryCo, 0, len(gpsList))
	tempMap := make(map[int64]bool)

	// Java iterates from end to start for dedup preference.
	for i := len(gpsList) - 1; i >= 0; i-- {
		domain := gpsList[i]
		point, err := utils.ParsePointWKT(domain.Geometry)
		if err != nil {
			continue
		}

		var trajectory dto.TrajectoryCo
		trajectory.Timestamp = int64(point.M)
		if trajectory.Timestamp == 0 {
			trajectory.Timestamp = domain.Timestamp.UnixMilli()
		}

		var metricData map[string]interface{}
		metricStr := domain.Metric
		if err := json.Unmarshal([]byte(metricStr), &metricData); err == nil {
			if speed, ok := metricData["speed"].(float64); ok {
				trajectory.Speed = float32(speed)
			}
			if course, ok := metricData["course"].(float64); ok {
				trajectory.Course = float32(course)
			}
		}
		// Java buildTrajectoryListByPg: itinerary points never set totalMiles (null in JSON).
		applyGpsTotalMiles(&trajectory, metricStr)

		lng, lat := point.Lng, point.Lat
		if cmd.Type != nil && *cmd.Type == 1 {
			lng, lat = utils.TransformWGS84ToGCJ02(lng, lat)
		}
		trajectory.Lng = lng
		trajectory.Lat = lat

		if trajectory.Lng == 0 && trajectory.Lat == 0 {
			continue
		}
		if tempMap[trajectory.Timestamp] {
			continue
		}
		tempMap[trajectory.Timestamp] = true
		trajectories = append(trajectories, trajectory)
	}

	if len(trajectories) > 1 {
		sort.Slice(trajectories, func(i, j int) bool {
			return trajectories[i].Timestamp < trajectories[j].Timestamp
		})
	}
	return trajectories, nil
}

// GetTrajectoryDistance mirrors Java EBikeItineraryServiceImpl.getDistance(TrajectoryDTO).
func GetTrajectoryDistance(cmd dto.TrajectoryCmd) (*dto.TrajectoryDistanceCo, error) {
	trajectories, err := GetTrajectory(cmd)
	if err != nil {
		return nil, err
	}
	return &dto.TrajectoryDistanceCo{Distance: orderDistance(trajectories, *cmd.Type)}, nil
}

// GetMetric mirrors Java EBikeItineraryServiceImpl.getMetric / buildMetricListByPg.
func GetMetric(cmd dto.MetricCmd) ([]dto.MetricCo, error) {
	if err := validateMetricTimeRange(cmd.StartTime, cmd.EndTime); err != nil {
		return nil, err
	}

	gpsList, err := queryGpsByImeiAndTime(cmd.Imei, cmd.StartTime, cmd.EndTime)
	if err != nil {
		return nil, err
	}

	metrics := make([]dto.MetricCo, 0, len(gpsList))

	for i := len(gpsList) - 1; i >= 0; i-- {
		domain := gpsList[i]
		var metricData map[string]interface{}
		if err := json.Unmarshal([]byte(domain.Metric), &metricData); err != nil {
			continue
		}

		var metric dto.MetricCo
		if ts := number(metricData["timestamp"]); ts != 0 {
			metric.Timestamp = int64(ts)
		}
		if speed := number(metricData["speed"]); speed != 0 {
			metric.Speed = int(speed)
		}
		if course := number(metricData["course"]); course != 0 {
			metric.Course = int(course)
		}
		if voltage := number(metricData["voltage"]); voltage != 0 {
			metric.Voltage = int(voltage)
		}
		if gsm := number(metricData["gsm"]); gsm != 0 {
			metric.Gsm = int(gsm)
		}

		wgs84Lng, _ := metricData["wgs84Lng"].(float64)
		wgs84Lat, _ := metricData["wgs84Lat"].(float64)
		if wgs84Lng == 0 && wgs84Lat == 0 {
			continue
		}

		lng, lat := wgs84Lng, wgs84Lat
		if cmd.Type != nil && *cmd.Type == 1 {
			lng, lat = utils.TransformWGS84ToGCJ02(wgs84Lng, wgs84Lat)
		}
		metric.Lng = lng
		metric.Lat = lat

		// Java buildMetricListByPg dedup is broken: tempMap.get(Timestamp) never hits Long keys.
		metrics = append(metrics, metric)
	}

	if len(metrics) > 1 {
		sort.Slice(metrics, func(i, j int) bool {
			return metrics[i].Timestamp < metrics[j].Timestamp
		})
	}
	return metrics, nil
}

// GetOrderTrajectory mirrors Java getTrajectoryByOrderAndPg.
func GetOrderTrajectory(cmd dto.OrderTrajectoryCmd) ([]dto.TrajectoryCo, error) {
	return fetchOrderTrajectory(cmd.OrderId, *cmd.Type, cmd.Imei)
}

// GetOrderTrajectoryDistance mirrors Java getDistance(OrderTrajectoryDTO).
func GetOrderTrajectoryDistance(cmd dto.OrderTrajectoryCmd) (*dto.TrajectoryDistanceCo, error) {
	trajectories, err := GetOrderTrajectory(cmd)
	if err != nil {
		return nil, err
	}
	return &dto.TrajectoryDistanceCo{Distance: orderDistance(trajectories, *cmd.Type)}, nil
}

// GetBatchOrderTrajectory mirrors Java batchOrderTrajectoryByPg.
func GetBatchOrderTrajectory(cmd dto.BatchOrderTrajectoryCmd) (map[string][]dto.TrajectoryCo, error) {
	if len(cmd.OrderTrajectoryRequests) == 0 {
		return map[string][]dto.TrajectoryCo{}, nil
	}

	coordType := *cmd.Type
	out := make(map[string][]dto.TrajectoryCo, len(cmd.OrderTrajectoryRequests))
	seenMiss := make(map[string]struct{}, len(cmd.OrderTrajectoryRequests))

	for _, req := range cmd.OrderTrajectoryRequests {
		if _, dup := seenMiss[req.OrderId]; dup {
			continue
		}
		seenMiss[req.OrderId] = struct{}{}

		traj, err := fetchOrderTrajectory(req.OrderId, coordType, req.Imei)
		if err != nil {
			if biz, ok := err.(*web.BizError); ok && biz.Code == dto.CodeOrderTrajectoryNotExist {
				continue
			}
			return nil, err
		}
		out[req.OrderId] = traj
	}
	return out, nil
}

// GetBatchOrderTrajectoryDistance mirrors Java batchOrderTrajectoryDistanceByPg.
func GetBatchOrderTrajectoryDistance(cmd dto.BatchOrderTrajectoryCmd) (map[string]float64, error) {
	trajMap, err := GetBatchOrderTrajectory(cmd)
	if err != nil {
		return nil, err
	}
	out := make(map[string]float64, len(trajMap))
	for orderID, trajectories := range trajMap {
		out[orderID] = orderDistance(trajectories, *cmd.Type)
	}
	return out, nil
}

func queryItinerary(orderID, imei string) (model.EBikeItineraryDO, error) {
	query := db.DB.Select("id, order_id, imei, start_time, end_time, metric, create_time, ST_AsText(geometry) as geometry").
		Where("order_id = ?", orderID)
	if imei != "" {
		query = query.Where("imei = ?", imei)
	}
	var itinerary model.EBikeItineraryDO
	if err := query.First(&itinerary).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return itinerary, &web.BizError{Code: dto.CodeOrderTrajectoryNotExist, Msg: "轨迹不存在"}
		}
		return itinerary, err
	}
	return itinerary, nil
}

func parseItineraryTrajectory(itinerary model.EBikeItineraryDO, coordType int) ([]dto.TrajectoryCo, error) {
	points, err := utils.ParseLineStringWKT(itinerary.Geometry)
	if err != nil {
		return nil, err
	}

	type attribute struct {
		Speed  []float32 `json:"speed"`
		Course []float32 `json:"course"`
	}
	var metric attribute
	_ = json.Unmarshal([]byte(itinerary.Metric), &metric)

	trajectories := make([]dto.TrajectoryCo, 0, len(points))
	for i, point := range points {
		trajectory := dto.TrajectoryCo{
			Timestamp: int64(point.M),
		}
		if trajectory.Timestamp == 0 {
			continue
		}

		lng, lat := point.Lng, point.Lat
		if coordType == 1 {
			lng, lat = utils.TransformWGS84ToGCJ02(lng, lat)
		}
		if lng == 0 && lat == 0 {
			continue
		}
		trajectory.Lng = lng
		trajectory.Lat = lat

		if len(metric.Course) > i {
			trajectory.Course = metric.Course[i]
		}
		if len(metric.Speed) > i {
			trajectory.Speed = metric.Speed[i]
		}
		trajectories = append(trajectories, trajectory)
	}
	return trajectories, nil
}

func orderDistance(trajectories []dto.TrajectoryCo, coordType int) float64 {
	if len(trajectories) <= 1 {
		return 0
	}
	var total float64
	for i := 1; i < len(trajectories); i++ {
		prev := trajectories[i-1]
		curr := trajectories[i]
		total += utils.GetDistance(prev.Lng, prev.Lat, curr.Lng, curr.Lat, coordType)
	}
	return total
}

func validateTrajectoryTimeRange(startTime, endTime int64) error {
	if startTime > endTime {
		return &web.BizError{Code: dto.CodeTrajectoryQueryDateOutOfRange, Msg: "起始时间不能大于截止时间"}
	}
	now := time.Now()
	sixMonthsAgo := utils.ClampedAddMonths(now, -6)
	if startTime < sixMonthsAgo.UnixMilli() || endTime < sixMonthsAgo.UnixMilli() {
		return &web.BizError{Code: dto.CodeTrajectoryQueryDateOutOfRange, Msg: "只能获取最近6个月的数据"}
	}
	end := time.UnixMilli(endTime)
	if startTime < utils.ClampedAddMonths(end, -6).UnixMilli() {
		return &web.BizError{Code: dto.CodeTrajectoryQueryDateOutOfRange, Msg: "只能获取最近6个月的数据"}
	}
	return nil
}

func validateMetricTimeRange(startTime, endTime int64) error {
	if startTime > endTime {
		return &web.BizError{Code: dto.CodeTrajectoryQueryDateOutOfRange, Msg: "起始时间不能大于截止时间"}
	}
	if endTime-startTime > 7*24*3600*1000 {
		return &web.BizError{Code: dto.CodeTrajectoryQueryDateOutOfRange, Msg: "只能查询1周之内的数据"}
	}
	return nil
}

// applyGpsTotalMiles mirrors Java buildTrajectoryListByPg totalMiles handling.
func applyGpsTotalMiles(trajectory *dto.TrajectoryCo, metricStr string) {
	if v, ok := parseTotalMilesValue(metricStr); ok {
		trajectory.TotalMiles = &v
	}
}

// parseTotalMiles mirrors Java buildTrajectoryListByPg string split (including parse quirks).
func parseTotalMiles(metric string) int {
	if v, ok := parseTotalMilesValue(metric); ok {
		return v
	}
	return 0
}

func parseTotalMilesValue(metric string) (int, bool) {
	pairs := strings.Split(metric, ",")
	for _, pair := range pairs {
		if !strings.Contains(pair, "totalMiles") {
			continue
		}
		parts := strings.Split(pair, ":")
		if len(parts) < 2 {
			continue
		}
		totalMiles := strings.TrimSpace(parts[1])
		if totalMiles == "null" {
			return 0, false
		}
		miles, err := strconv.Atoi(totalMiles)
		if err != nil {
			return 0, false
		}
		return miles, true
	}
	return 0, false
}

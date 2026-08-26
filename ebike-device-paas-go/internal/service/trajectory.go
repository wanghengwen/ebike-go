package service

import (
	"encoding/json"
	"math"
	"strings"
	"time"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/client"
	"ebike-device-paas-go/internal/pkg/config"
	"ebike-device-paas-go/internal/repository"
)

// DeviceTrajectory endpoints port DeviceTrajectoryServiceImpl: each forwards to
// the ebike-device-worker gateway (setting type=coordinate.type) and reshapes
// the response. Forwarded RPCs -> NOT shadow-compared.

// Worker gateway paths (DeviceTrajectoryApiFeign).
const (
	pathWorkerTrajectory      = "/ebike/gps/getTrajectory"
	pathWorkerOrderTrajectory = "/ebike/gps/getOrderTrajectory"
	pathWorkerBatchTrajectory = "/ebike/gps/getBatchOrderTrajectory"
	pathWorkerSaveTrajectory  = "/ebike/gps/saveOrderTrajectory"
	pathWorkerDistance        = "/ebike/gps/getTrajectoryDistance"
	pathWorkerMetric          = "/ebike/gps/getMetric"
)

func coordinateType() int { return config.GlobalConfig.Coordinate.Type }

// TrajectoryRealTime ports DeviceTrajectoryServiceImpl.trajectoryRealTime.
func TrajectoryRealTime(traceID string, q dto.TrajectoryRealTimeQry) ([]dto.DeviceTrajectoryCo, error) {
	t := coordinateType()
	body := map[string]interface{}{
		"imei":      q.Imei,
		"type":      t,
		"startTime": q.StartTime.Ptr(),
		"endTime":   q.EndTime.Ptr(),
		"traceId":   traceID,
	}
	out := []dto.DeviceTrajectoryCo{}
	if err := client.PostWorker(pathWorkerTrajectory, body, &out, dto.TenantOf(q.CommandContext)); err != nil {
		return nil, err
	}
	return out, nil
}

// TrajectoryHistory ports DeviceTrajectoryServiceImpl.trajectoryHistory.
func TrajectoryHistory(traceID string, q dto.TrajectoryHistoryQry) ([]dto.DeviceTrajectoryCo, error) {
	body := historyReq(traceID, q, coordinateType())
	out := []dto.DeviceTrajectoryCo{}
	if err := client.PostWorker(pathWorkerOrderTrajectory, body, &out, dto.TenantOf(q.CommandContext)); err != nil {
		return nil, err
	}
	return out, nil
}

// TrajectoryHistoryBatch ports DeviceTrajectoryServiceImpl.trajectoryHistoryBatch.
func TrajectoryHistoryBatch(traceID string, q dto.TrajectoryHistoryBatchQry) (map[string][]dto.DeviceTrajectoryCo, error) {
	t := coordinateType()
	// Java only sets type on the wrapper; inner requests carry orderId/time only.
	reqs := make([]map[string]interface{}, 0, len(q.OrderBatch))
	for _, h := range q.OrderBatch {
		reqs = append(reqs, map[string]interface{}{
			"orderId":   h.OrderId.Ptr(),
			"startTime": h.StartTime.Ptr(),
			"endTime":   h.EndTime.Ptr(),
			"traceId":   traceID,
		})
	}
	body := map[string]interface{}{
		"orderTrajectoryRequests": reqs,
		"type":                    t,
		"traceId":                 traceID,
	}
	out := map[string][]dto.DeviceTrajectoryCo{}
	if err := client.PostWorker(pathWorkerBatchTrajectory, body, &out, dto.TenantOf(q.CommandContext)); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTrajectoryDistance ports DeviceTrajectoryApiRpcImpl.getTrajectoryDistance:
// any error degrades to distance 0 (mirrors the Java try/catch).
func GetTrajectoryDistance(traceID string, q dto.TrajectoryRealTimeQry) dto.TrajectoryDistanceCo {
	body := map[string]interface{}{
		"imei":      q.Imei,
		"type":      coordinateType(),
		"startTime": q.StartTime.Ptr(),
		"endTime":   q.EndTime.Ptr(),
		"traceId":   traceID,
	}
	var out struct {
		Distance *float64 `json:"distance"`
	}
	if err := client.PostWorker(pathWorkerDistance, body, &out, dto.TenantOf(q.CommandContext)); err != nil || out.Distance == nil {
		return dto.TrajectoryDistanceCo{Distance: 0}
	}
	return dto.TrajectoryDistanceCo{Distance: *out.Distance}
}

// TrajectorySaveDb ports DeviceTrajectoryServiceImpl.trajectorySaveDb (write).
func TrajectorySaveDb(traceID string, cmd dto.TrajectoryDbCmd) error {
	body := map[string]interface{}{
		"orderId":   cmd.OrderId.Ptr(),
		"imei":      cmd.Imei,
		"startTime": cmd.StartTime.Ptr(),
		"endTime":   cmd.EndTime.Ptr(),
		"traceId":   traceID,
	}
	return client.PostWorker(pathWorkerSaveTrajectory, body, nil, dto.TenantOf(cmd.CommandContext))
}

// GetMetric ports DeviceTrajectoryApiRpcImpl.getMetric + the controller reshape:
// resolve imei (carId -> imei when not 86-prefixed), validate, forward, reshape
// the worker JSON list into course/speed/gps/gsm/voltage metrics.
func GetMetric(tenantID, traceID string, q dto.MetricQry) (map[string]interface{}, error) {
	deviceID := q.Imei
	if !strings.HasPrefix(deviceID, "86") {
		deviceID = repository.GetImeiByCarId(tenantID, deviceID)
	}
	if deviceID == "" || !strings.HasPrefix(deviceID, "86") || len(deviceID) != 15 {
		return nil, ErrCarImeiBind
	}

	body := map[string]interface{}{
		"imei":      deviceID,
		"type":      coordinateType(),
		"startTime": q.StartTime.Ptr(),
		"endTime":   q.EndTime.Ptr(),
		"traceId":   traceID,
	}
	var list []map[string]json.Number
	if err := client.PostWorker(pathWorkerMetric, body, &list, tenantID); err != nil {
		return nil, err
	}
	return reshapeMetric(list), nil
}

func historyReq(traceID string, q dto.TrajectoryHistoryQry, t int) map[string]interface{} {
	return map[string]interface{}{
		"type":      t,
		"orderId":   q.OrderId.Ptr(),
		"startTime": q.StartTime.Ptr(),
		"endTime":   q.EndTime.Ptr(),
		"traceId":   traceID,
	}
}

// reshapeMetric mirrors getMetric's loop: timestamp*1000, formatted time, and
// course/speed/gps/gsm/voltage(=mV/1000 @2dp HALF_EVEN) series.
func reshapeMetric(list []map[string]json.Number) map[string]interface{} {
	n := len(list)
	timeList := make([]string, 0, n)
	courseList := make([]int, 0, n)
	speedList := make([]int, 0, n)
	gpsList := make([]dto.MetricGps, 0, n)
	gsmList := make([]int, 0, n)
	voltageList := make([]float64, 0, n)

	for _, j := range list {
		ts := numInt64(j["timestamp"]) * 1000
		timeList = append(timeList, time.UnixMilli(ts).Format("2006-01-02 15:04:05"))

		course := numInt(j["course"])
		courseList = append(courseList, course)
		speed := numInt(j["speed"])
		speedList = append(speedList, speed)

		lng := numFloat(j["lng"])
		lat := numFloat(j["lat"])
		gpsList = append(gpsList, dto.MetricGps{Timestamp: ts, Lng: lng, Lat: lat, Course: course, Speed: speed})

		gsmList = append(gsmList, numInt(j["gsm"]))
		voltageList = append(voltageList, roundHalfEven(numFloat(j["voltage"])/1000, 2))
	}

	return map[string]interface{}{
		"course":  dto.MetricCourse{Course: courseList, Time: timeList},
		"speed":   dto.MetricSpeed{Speed: speedList, Time: timeList},
		"gps":     gpsList,
		"gsm":     dto.MetricGsm{Gsm: gsmList, Time: timeList},
		"voltage": dto.MetricVoltage{Voltage: voltageList, Time: timeList},
	}
}

func numInt64(n json.Number) int64 {
	if n == "" {
		return 0
	}
	v, err := n.Int64()
	if err != nil {
		f, _ := n.Float64()
		return int64(f)
	}
	return v
}

func numInt(n json.Number) int { return int(numInt64(n)) }

func numFloat(n json.Number) float64 {
	if n == "" {
		return 0
	}
	f, _ := n.Float64()
	return f
}

// roundHalfEven rounds x to places decimals using banker's rounding (matches
// Java BigDecimal RoundingMode.HALF_EVEN).
func roundHalfEven(x float64, places int) float64 {
	pow := math.Pow(10, float64(places))
	scaled := x * pow
	floor := math.Floor(scaled)
	diff := scaled - floor
	switch {
	case diff > 0.5:
		floor++
	case diff == 0.5:
		if math.Mod(floor, 2) != 0 {
			floor++
		}
	}
	return floor / pow
}

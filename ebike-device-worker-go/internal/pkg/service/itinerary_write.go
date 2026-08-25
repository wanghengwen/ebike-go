package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"ebike-device-worker-go/internal/api/dto"
	"ebike-device-worker-go/internal/pkg/async"
	"ebike-device-worker-go/internal/pkg/db"
	"ebike-device-worker-go/internal/pkg/fastid"
	"ebike-device-worker-go/internal/pkg/model"
	"ebike-device-worker-go/internal/pkg/utils"
	"ebike-device-worker-go/internal/pkg/web"

	"gorm.io/gorm"
)

// SaveOrderTrajectory mirrors Java EBikeItineraryServiceImpl.saveTrajectory (itineraryType=0).
func SaveOrderTrajectory(cmd dto.SaveOrderTrajectoryCmd) error {
	return saveTrajectory(cmd.OrderId, cmd.Imei, cmd.StartTime, cmd.EndTime, 0, cmd.Gps)
}

// SaveMovingEBikeTrajectory mirrors Java saveMovingEBikeTrajectory (itineraryType=1).
func SaveMovingEBikeTrajectory(cmd dto.SaveMovingTrajectoryCmd) error {
	return saveTrajectory(cmd.MovingEBikeTaskId, cmd.Imei, cmd.StartTime, cmd.EndTime, 1, nil)
}

func saveTrajectory(orderID, imei string, startMs, endMs int64, itineraryType int, endGps *dto.TrajectoryPoint) error {
	if startMs > endMs {
		return &web.BizError{Code: "10015", Msg: "时间不在范围内"}
	}
	async.Go(func() {
		if err := doSaveTrajectory(orderID, imei, startMs, endMs, itineraryType, endGps); err != nil {
			log.Printf("[itinerary] save failed order=%s imei=%s: %v", orderID, imei, err)
		}
	})
	return nil
}

func doSaveTrajectory(orderID, imei string, startMs, endMs int64, itineraryType int, endGps *dto.TrajectoryPoint) error {
	if db.DB == nil {
		return fmt.Errorf("database not configured")
	}
	startTime := time.UnixMilli(startMs)
	endTime := time.UnixMilli(endMs)

	gpsList, err := queryGpsByImeiAndTime(imei, startMs, endMs)
	if err != nil {
		return err
	}
	if len(gpsList) < 2 {
		log.Printf("[itinerary] skip save, gps count < 2 order=%s", orderID)
		return nil
	}

	vertices, speeds, courses := buildItineraryPoints(gpsList)
	if endGps != nil && utils.ValidWGS84(endGps.Lng, endGps.Lat) {
		endSecs := endGpsGeometryM(endMs)
		vertices = append(vertices, utils.Point{Lng: endGps.Lng, Lat: endGps.Lat, M: endSecs})
		speeds = append(speeds, 0)
		courses = append(courses, 0)
	}
	_ = itineraryType // mirrors Java ItineraryDTO.itineraryType; not stored in PG ebike_itinerary.

	lineWKT, err := utils.FormatLineStringMUnfiltered(vertices)
	if err != nil {
		return err
	}
	// mirrors Java Attribute{speed, course} — no timestamp array.
	metricObj := map[string]interface{}{
		"speed":  speeds,
		"course": courses,
	}
	metricJSON, _ := json.Marshal(metricObj)

	var existing model.EBikeItineraryDO
	err = db.DB.Where("order_id = ?", orderID).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	createTime := time.Now()
	table := db.PgTable("ebike_itinerary")
	if err := db.DB.Exec(fmt.Sprintf(`
		INSERT INTO %s(id, order_id, imei, start_time, end_time, geometry, metric, create_time)
		VALUES (?, ?, ?, ?, ?, ST_GeomFromText(?,4326), ?::jsonb, ?)`, table),
		fastid.Next(), orderID, imei, startTime, endTime, lineWKT, string(metricJSON), createTime,
	).Error; err != nil {
		return err
	}
	invalidateOrderTrajectoryCache(orderID)
	return nil
}

// buildItineraryPoints mirrors Java EBikeItineraryServiceImpl.create() + Attribute.addMetric().
// Returns vertices (WKT geometry points), speeds and courses (metric arrays).
// Java Attribute only tracks speed and course — no timestamp array.
func buildItineraryPoints(gpsList []model.EBikeGpsDO) (vertices []utils.Point, speeds, courses []float32) {
	for _, g := range gpsList {
		point, err := parsePointForItinerary(g)
		if err != nil {
			continue
		}
		vertices = append(vertices, point.vertex)
		var metric map[string]interface{}
		_ = json.Unmarshal([]byte(g.Metric), &metric)
		speeds = append(speeds, float32(number(metric["speed"])))
		courses = append(courses, float32(number(metric["course"])))
	}
	return vertices, speeds, courses
}

type itineraryPoint struct {
	vertex utils.Point
	ts     int64
}

func parsePointForItinerary(g model.EBikeGpsDO) (itineraryPoint, error) {
	fromWKT, err := utils.ParsePointWKT(g.Geometry)
	if err != nil {
		return itineraryPoint{}, err
	}
	ts := int64(fromWKT.M)
	if ts == 0 {
		ts = g.Timestamp.Unix()
	}
	return itineraryPoint{
		vertex: utils.Point{Lng: fromWKT.Lng, Lat: fromWKT.Lat, M: float64(ts)},
		ts:     ts,
	}, nil
}

// endGpsGeometryM mirrors Java EBikeItineraryServiceImpl.savePg: point.setM(endTime.getTime()/1000).
func endGpsGeometryM(endMs int64) float64 {
	return float64(endMs / 1000)
}

func number(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case json.Number:
		f, _ := t.Float64()
		return f
	default:
		return 0
	}
}

package service

import (
	"ebike-device-worker-go/internal/pkg/db"
	"ebike-device-worker-go/internal/pkg/model"
	"ebike-device-worker-go/internal/pkg/utils"
)

// javaGpsTimestampQuery mirrors EBikeGpsMapper.queryByImeiAndTimestamp (second-precision bounds).
const javaGpsTimestampQuery = "imei = ? AND timestamp >= to_timestamp(?, 'YYYY-MM-DD HH24:MI:SS') AND timestamp <= to_timestamp(?, 'YYYY-MM-DD HH24:MI:SS')"

// queryGpsByImeiAndTime mirrors Java EBikeGpsServiceImpl.getPointList.
func queryGpsByImeiAndTime(imei string, startMs, endMs int64) ([]model.EBikeGpsDO, error) {
	var gpsList []model.EBikeGpsDO
	result := db.DB.Select("id, imei, metric, timestamp, create_time, ST_AsText(geometry) as geometry").
		Where(javaGpsTimestampQuery, imei, utils.FormatJavaGpsQueryTime(startMs), utils.FormatJavaGpsQueryTime(endMs)).
		Order("timestamp ASC").
		Find(&gpsList)
	return gpsList, result.Error
}

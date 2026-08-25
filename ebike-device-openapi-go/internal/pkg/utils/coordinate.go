package utils

import (
	"encoding/json"
	"math"

	"ebike-device-openapi-go/internal/pkg/logger"
	"go.uber.org/zap"
)

const (
	x_PI = 3.14159265358979324 * 3000.0 / 180.0
	PI   = 3.1415926535897932384626
	a    = 6378245.0
	ee   = 0.00669342162296594323
)

// TransformWGS84ToGCJ02 converts WGS84 coordinates to GCJ02.
// Returns (lng, lat)
func TransformWGS84ToGCJ02(lng, lat float64) (float64, float64) {
	if OutOfChina(lng, lat) {
		return lng, lat
	}

	dLat := transformLat(lng-105.0, lat-35.0)
	dLng := transformLng(lng-105.0, lat-35.0)
	radLat := lat / 180.0 * PI
	magic := math.Sin(radLat)
	magic = 1 - ee*magic*magic
	sqrtMagic := math.Sqrt(magic)

	dLat = (dLat * 180.0) / ((a * (1 - ee)) / (magic * sqrtMagic) * PI)
	dLng = (dLng * 180.0) / (a / sqrtMagic * math.Cos(radLat) * PI)

	mgLat := lat + dLat
	mgLng := lng + dLng

	return mgLng, mgLat
}

func transformLat(lng, lat float64) float64 {
	ret := -100.0 + 2.0*lng + 3.0*lat + 0.2*lat*lat + 0.1*lng*lat + 0.2*math.Sqrt(math.Abs(lng))
	ret += (20.0*math.Sin(6.0*lng*PI) + 20.0*math.Sin(2.0*lng*PI)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lat*PI) + 40.0*math.Sin(lat/3.0*PI)) * 2.0 / 3.0
	ret += (160.0*math.Sin(lat/12.0*PI) + 320.0*math.Sin(lat*PI/30.0)) * 2.0 / 3.0
	return ret
}

func transformLng(lng, lat float64) float64 {
	ret := 300.0 + lng + 2.0*lat + 0.1*lng*lng + 0.1*lng*lat + 0.1*math.Sqrt(math.Abs(lng))
	ret += (20.0*math.Sin(6.0*lng*PI) + 20.0*math.Sin(2.0*lng*PI)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lng*PI) + 40.0*math.Sin(lng/3.0*PI)) * 2.0 / 3.0
	ret += (150.0*math.Sin(lng/12.0*PI) + 300.0*math.Sin(lng/30.0*PI)) * 2.0 / 3.0
	return ret
}

// OutOfChina checks if the coordinates are outside China
func OutOfChina(lng, lat float64) bool {
	return (lng < 72.004 || lng > 137.8347) || (lat < 0.8293 || lat > 55.8271)
}

// AddWGS84 parses the json string, looks for result.gps, and adds wgs84Lng and wgs84Lat while converting lng/lat to GCJ02.
func AddWGS84(jsonStr string) string {
	if jsonStr == "" {
		return jsonStr
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		logger.Log.Warn("AddWGS84 json unmarshal error", zap.Error(err), zap.String("payload", jsonStr))
		return jsonStr
	}

	resultVal, ok := data["result"]
	if !ok || resultVal == nil {
		return jsonStr
	}
	resultMap, ok := resultVal.(map[string]interface{})
	if !ok {
		return jsonStr
	}

	gpsVal, ok := resultMap["gps"]
	if !ok || gpsVal == nil {
		return jsonStr
	}
	gpsMap, ok := gpsVal.(map[string]interface{})
	if !ok {
		return jsonStr
	}

	var lngOpt, latOpt float64
	var hasLng, hasLat bool

	if lngVal, ok := gpsMap["lng"]; ok && lngVal != nil {
		if lf, ok := lngVal.(float64); ok {
			lngOpt = lf
			hasLng = true
		}
	}
	if latVal, ok := gpsMap["lat"]; ok && latVal != nil {
		if lf, ok := latVal.(float64); ok {
			latOpt = lf
			hasLat = true
		}
	}

	if hasLng && hasLat {
		gcjLng, gcjLat := TransformWGS84ToGCJ02(lngOpt, latOpt)
		gpsMap["wgs84Lng"] = lngOpt
		gpsMap["wgs84Lat"] = latOpt
		gpsMap["lng"] = gcjLng
		gpsMap["lat"] = gcjLat
	}

	newBytes, err := json.Marshal(data)
	if err != nil {
		logger.Log.Warn("AddWGS84 json marshal error", zap.Error(err))
		return jsonStr
	}
	return string(newBytes)
}

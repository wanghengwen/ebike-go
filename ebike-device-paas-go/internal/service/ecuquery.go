// Package service implements the EcuQuery business logic, ported from the Java
// EcuQueryServiceImpl. P1 covers the read-only Redis-backed queries.
package service

import (
	"time"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/repository"
)

// realGpsFreshnessSec matches the Java filter: keep devices whose timestamp is
// within 15s of now (timestamp + 15 > now/1000).
const realGpsFreshnessSec int64 = 15

// GetDeviceRealGpsList mirrors EcuQueryServiceImpl.getDeviceRealGpsList: read
// device info from Redis, drop stale GPS, and convert to DeviceRealGpsCo.
func GetDeviceRealGpsList(tenantID string, imeiList []string) []dto.DeviceRealGpsCo {
	devices := repository.GetDeviceInfoList(tenantID, imeiList)
	return FilterRealGps(devices, time.Now().Unix())
}

// FilterRealGps is the pure transform (testable without Redis): keep devices
// with a fresh timestamp and project to DeviceRealGpsCo.
func FilterRealGps(devices []map[string]interface{}, nowSec int64) []dto.DeviceRealGpsCo {
	out := make([]dto.DeviceRealGpsCo, 0, len(devices))
	for _, dev := range devices {
		ts, ok := asInt64(dev["timestamp"])
		if !ok || ts+realGpsFreshnessSec <= nowSec {
			continue
		}
		co := dto.DeviceRealGpsCo{}
		if imei, ok := dev["imei"].(string); ok {
			co.Imei = imei
		}
		if lng, ok := asFloat64(dev["lng"]); ok {
			co.Lng = &lng
		}
		if lat, ok := asFloat64(dev["lat"]); ok {
			co.Lat = &lat
		}
		out = append(out, co)
	}
	return out
}

func asInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case int64:
		return n, true
	case int:
		return int64(n), true
	case float64:
		return int64(n), true
	default:
		return 0, false
	}
}

func asFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int64:
		return float64(n), true
	case int:
		return float64(n), true
	default:
		return 0, false
	}
}

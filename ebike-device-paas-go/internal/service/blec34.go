package service

import (
	"encoding/json"
	"log"
	"time"

	"ebike-device-paas-go/internal/client"
	"ebike-device-paas-go/internal/pkg/config"
	"ebike-device-paas-go/internal/pkg/geo"
	"ebike-device-paas-go/internal/pkg/kafka"
)

// bleC34Payload is the constant CommandResultDo.payload for BLE c34 messages
// (BleInfoToResult.BLE_C34).
const bleC34Payload = `{"type":"c34","online":false}`

// sendC34 mirrors DeviceStateDateImpl.bleDeviceInfoReport -> KafkaProducer.sendC34:
// wrap the CommandResultDo (payload + result) JSON, enrich it with
// imei/deviceDataType/code/appId, and publish to ecu.kafka.parent-topic keyed by
// imei. A nil result is dropped (matches the Java null guard).
func sendC34(tenantID, imei string, result map[string]interface{}) {
	if result == nil {
		return
	}
	msg := map[string]interface{}{
		"payload":        bleC34Payload,
		"result":         result,
		"imei":           imei,
		"deviceDataType": "cmd",
		"code":           0,
		"appId":          client.IotPlatformAppID(tenantID),
	}
	value, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[c34] marshal failed (imei=%s): %v", imei, err)
		return
	}
	kafka.Send(config.GlobalConfig.Ecu.Kafka.ParentTopic, value)
}

// nowSeconds returns the current unix time in seconds (Java System.currentTimeMillis()/1000).
func nowSeconds() int64 { return time.Now().UnixMilli() / 1000 }

// c34Location mirrors BleInfoToResult.locationReportToDeviceInfoDto: only gps,
// converting the phone GCJ-02 fix to wgs84.
func c34Location(lng, lat float64) map[string]interface{} {
	wgsLat, wgsLng := geo.Gcj02ToWgs84(lat, lng)
	return map[string]interface{}{
		"gps": map[string]interface{}{
			"wgs84Lat":  wgsLat,
			"wgs84Lng":  wgsLng,
			"lng":       lng,
			"lat":       lat,
			"timestamp": nowSeconds(),
		},
	}
}

// c34Lock mirrors BleInfoToResult.bleLockReportToDeviceInfoDto.
func c34Lock(acc *int, lng, lat float64) map[string]interface{} {
	result := map[string]interface{}{}
	if acc != nil {
		result["acc"] = *acc
		if *acc == 1 {
			result["defend"] = 0
			result["wheelLock"] = 0
		}
	}
	wgsLat, wgsLng := geo.Gcj02ToWgs84(lat, lng)
	result["gps"] = map[string]interface{}{
		"wgs84Lat":  wgsLat,
		"wgs84Lng":  wgsLng,
		"lng":       lng,
		"lat":       lat,
		"timestamp": nowSeconds(),
	}
	return result
}

// c34Defend mirrors BleInfoToResult.bleDefendReportToDeviceInfoDto.
func c34Defend(defend *int, lng, lat float64) map[string]interface{} {
	result := map[string]interface{}{}
	if defend != nil {
		result["defend"] = *defend
		if *defend == 1 {
			result["acc"] = 0
			result["wheelLock"] = 1
		}
	}
	wgsLat, wgsLng := geo.Gcj02ToWgs84(lat, lng)
	result["gps"] = map[string]interface{}{
		"wgs84Lat":  wgsLat,
		"wgs84Lng":  wgsLng,
		"lng":       lng,
		"lat":       lat,
		"timestamp": nowSeconds(),
	}
	return result
}

// c34DeviceInfo mirrors BleInfoToResult.toBleDeviceInfoDto: a full state snapshot
// with the wgs84 fix converted to GCJ-02 for lng/lat.
func c34DeviceInfo(cmd *dtoBleDeviceInfo) map[string]interface{} {
	result := map[string]interface{}{}
	putIntPtr(result, "acc", cmd.Acc)
	putIntPtr(result, "defend", cmd.Defend)
	putIntPtr(result, "gsm", cmd.Gsm)
	if cmd.Voltage != nil {
		result["voltageMv"] = *cmd.Voltage * 10
	}
	putIntPtr(result, "helmet6Lock", cmd.Helmet6Lock)
	putIntPtr(result, "helmet6React", cmd.Helmet6React)
	putIntPtr(result, "totalMiles", cmd.TotalMiles)

	lng, lat := f64(cmd.Lng), f64(cmd.Lat)
	gcjLat, gcjLng := geo.Wgs84ToGcj02(lat, lng)
	gps := map[string]interface{}{
		"wgs84Lat": lat,
		"wgs84Lng": lng,
		"lng":      gcjLng,
		"lat":      gcjLat,
	}
	putIntPtr(gps, "speed", cmd.Speed)
	putIntPtr(gps, "course", cmd.Course)
	if cmd.Timestamp != nil {
		gps["timestamp"] = *cmd.Timestamp
	}
	result["gps"] = gps
	return result
}

// dtoBleDeviceInfo is the minimal view of BleDeviceInfoReportCmd used by c34DeviceInfo
// (kept local to avoid an import cycle with the api/dto package).
type dtoBleDeviceInfo struct {
	Acc          *int
	Defend       *int
	Gsm          *int
	Voltage      *int
	Helmet6Lock  *int
	Helmet6React *int
	TotalMiles   *int
	Speed        *int
	Course       *int
	Timestamp    *int64
	Lng          *float64
	Lat          *float64
}

func putIntPtr(m map[string]interface{}, key string, v *int) {
	if v != nil {
		m[key] = *v
	}
}

func f64(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

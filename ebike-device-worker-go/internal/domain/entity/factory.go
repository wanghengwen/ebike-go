package entity

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ebike-device-worker-go/internal/domain/message"
	"ebike-device-worker-go/internal/pkg/utils"
)

func StructEBikeGps(msg message.MessageInfoDTO) (imei string, metricJSON string, ts time.Time, geometryWKT string, createTime time.Time, valid bool) {
	metric := msg.Data
	tsSec := int64(0)
	if v, ok := metric["timestamp"].(float64); ok {
		tsSec = int64(v)
	}
	if tsSec == 0 {
		tsSec = msg.ReceiveDataTime.Unix()
	}
	ts = time.Unix(tsSec, 0)
	lng, lat := metricFloat(metric, "wgs84Lng"), metricFloat(metric, "wgs84Lat")
	if lng == 0 && lat == 0 {
		return "", "", time.Time{}, "", time.Time{}, false
	}
	metricBytes, _ := json.Marshal(metric)
	geometryWKT = utils.FormatPointZM(lng, lat, tsSec)
	return msg.DeviceID, string(metricBytes), ts, geometryWKT, msg.ReceiveDataTime, true
}

func StructEBikePing(msg message.MessageInfoDTO) (imei, metricJSON string, ts time.Time, valid bool) {
	if msg.DeviceID == "" || msg.Data == nil {
		return "", "", time.Time{}, false
	}
	metricBytes, _ := json.Marshal(msg.Data)
	return msg.DeviceID, string(metricBytes), epochSecond(msg.ReceiveDataTime), true
}

func StructEBikeBms(msg message.MessageInfoDTO) (imei, metricJSON string, ts time.Time, valid bool) {
	if msg.DeviceID == "" || msg.Data == nil {
		return "", "", time.Time{}, false
	}
	m := copyMap(msg.Data)
	delete(m, "sn")
	metricBytes, _ := json.Marshal(m)
	return msg.DeviceID, string(metricBytes), epochSecond(msg.ReceiveDataTime), true
}

func StructEBikeCmd(msg message.MessageInfoDTO) (imei, msgID, dir, cmdJSON string, ts time.Time, valid bool) {
	if msg.DeviceID == "" || msg.Data == nil {
		return "", "", "", "", time.Time{}, false
	}
	dir = strVal(msg.Data["dir"])
	msgID = strVal(msg.Data["msgId"])
	if dir == "" || msgID == "" {
		return "", "", "", "", time.Time{}, false
	}
	cmdBytes, _ := json.Marshal(msg.Data)
	return msg.DeviceID, msgID, dir, string(cmdBytes), epochSecond(msg.ReceiveDataTime), true
}

func StructEBikeAlarm(msg message.MessageInfoDTO) (imei string, alarmType int, ts time.Time, valid bool) {
	if msg.DeviceID == "" {
		return "", 0, time.Time{}, false
	}
	alarmType = int(metricFloat(msg.Data, "type"))
	return msg.DeviceID, alarmType, epochSecond(msg.ReceiveDataTime), true
}

func StructEBikeOnOffLine(msg message.MessageInfoDTO) (imei, lineState string, ts time.Time, valid bool) {
	if msg.DeviceID == "" {
		return "", "", time.Time{}, false
	}
	tsSec := int64(0)
	if v, ok := msg.Data["timestamp"].(float64); ok {
		tsSec = int64(v)
	}
	if tsSec == 0 {
		tsSec = msg.ReceiveDataTime.Unix()
	}
	lineState = strVal(msg.Data["lineState"])
	if lineState == "" {
		return "", "", time.Time{}, false
	}
	return msg.DeviceID, lineState, time.Unix(tsSec, 0), true
}

func StructEBikeFault(msg message.MessageInfoDTO) (imei string, etc, bms, ecu int, bitJSON string, ts time.Time, valid bool) {
	if msg.DeviceID == "" || msg.Data == nil {
		return "", 0, 0, 0, "", time.Time{}, false
	}
	etc = int(metricFloat(msg.Data, "etcFault"))
	bms = int(metricFloat(msg.Data, "bmsFault"))
	ecu = int(metricFloat(msg.Data, "ecuFault"))
	bitBytes, _ := json.Marshal(msg.Data)
	return msg.DeviceID, etc, bms, ecu, string(bitBytes), epochSecond(msg.ReceiveDataTime), true
}

// epochSecond mirrors Java EntityFactory: Instant.ofEpochSecond(receiveDataTime.getTime() / 1000).
func epochSecond(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return time.Unix(t.Unix(), 0)
}

func metricFloat(m map[string]interface{}, key string) float64 {
	if m == nil {
		return 0
	}
	switch v := m[key].(type) {
	case float64:
		return v
	case json.Number:
		f, _ := v.Float64()
		return f
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}

func strVal(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return strings.TrimSpace(strings.ReplaceAll(s, "\u0080", ""))
	default:
		return fmt.Sprintf("%v", s)
	}
}

func copyMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// GPSValidTimestamp mirrors Java EBikeGpsServiceImpl 1-hour window.
func GPSValidTimestamp(ts time.Time) bool {
	if ts.IsZero() || ts.Unix() == 0 {
		return false
	}
	diff := time.Since(ts)
	if diff < 0 {
		diff = -diff
	}
	return diff <= time.Hour
}

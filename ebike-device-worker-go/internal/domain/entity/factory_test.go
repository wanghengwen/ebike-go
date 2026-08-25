package entity_test

import (
	"strings"
	"testing"
	"time"

	"ebike-device-worker-go/internal/domain/entity"
	"ebike-device-worker-go/internal/domain/message"
)

func TestGPSValidTimestamp(t *testing.T) {
	if !entity.GPSValidTimestamp(time.Now()) {
		t.Fatal("now should be valid")
	}
	if entity.GPSValidTimestamp(time.Unix(0, 0)) {
		t.Fatal("zero should be invalid")
	}
}

func TestStructEBikeGps(t *testing.T) {
	msg := message.MessageInfoDTO{
		DeviceID: "860123456789012",
		Data: map[string]interface{}{
			"timestamp": float64(1700000000),
			"wgs84Lng":  114.1,
			"wgs84Lat":  22.2,
		},
		ReceiveDataTime: time.Unix(1700000000, 0),
	}
	imei, metric, ts, geom, _, ok := entity.StructEBikeGps(msg)
	if !ok || imei == "" || metric == "" || geom == "" || ts.IsZero() {
		t.Fatalf("unexpected gps struct: ok=%v imei=%s", ok, imei)
	}
	if !strings.Contains(geom, "POINT ZM") {
		t.Fatalf("geometry should be POINT ZM, got %s", geom)
	}
	if !strings.Contains(geom, "1700000000") {
		t.Fatalf("geometry M should use epoch seconds, got %s", geom)
	}
	if strings.Contains(geom, "1700000000000") {
		t.Fatalf("geometry M must not use milliseconds: %s", geom)
	}
}

func TestReceiveDataTimeTruncatedToSecond(t *testing.T) {
	recv := time.Unix(1700000000, 123456789) // reconsume may carry sub-second precision
	msg := message.MessageInfoDTO{
		DeviceID:        "860123456789012",
		Data:            map[string]interface{}{"cmd": float64(2)},
		ReceiveDataTime: recv,
	}

	_, _, ts, ok := entity.StructEBikePing(msg)
	if !ok || ts.Unix() != 1700000000 || ts.Nanosecond() != 0 {
		t.Fatalf("ping timestamp=%v want second-truncated", ts)
	}

	msg.Data = map[string]interface{}{"soc": float64(80)}
	_, _, ts, ok = entity.StructEBikeBms(msg)
	if !ok || ts.Unix() != 1700000000 || ts.Nanosecond() != 0 {
		t.Fatalf("bms timestamp=%v want second-truncated", ts)
	}

	msg.Data = map[string]interface{}{"dir": "up", "msgId": strings.Repeat("a", 32)}
	_, _, _, _, ts, ok = entity.StructEBikeCmd(msg)
	if !ok || ts.Unix() != 1700000000 || ts.Nanosecond() != 0 {
		t.Fatalf("cmd timestamp=%v want second-truncated", ts)
	}

	msg.Data = map[string]interface{}{"type": float64(1)}
	_, _, ts, ok = entity.StructEBikeAlarm(msg)
	if !ok || ts.Unix() != 1700000000 || ts.Nanosecond() != 0 {
		t.Fatalf("alarm timestamp=%v want second-truncated", ts)
	}

	msg.Data = map[string]interface{}{"etcFault": float64(1)}
	_, _, _, _, _, ts, ok = entity.StructEBikeFault(msg)
	if !ok || ts.Unix() != 1700000000 || ts.Nanosecond() != 0 {
		t.Fatalf("fault timestamp=%v want second-truncated", ts)
	}
}

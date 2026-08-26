package service

import (
	"encoding/json"
	"testing"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/client"
	"ebike-device-paas-go/internal/pkg/config"
)

func TestBuildCommandResultInnerParam(t *testing.T) {
	async := true
	jobID, payload := "job-1", "pl"
	do := &client.CommandResultDo{
		JobID:   &jobID,
		Async:   &async,
		Payload: &payload,
		Result:  `{"code":"0","result":{"mode":1,"freq_norm":600,"freq_move":5,"isOverSpeedOn":1}}`,
	}
	cr := buildCommandResult(do, mapInnerParam)

	if cr.JobId == nil || *cr.JobId != "job-1" || cr.Payload == nil || *cr.Payload != "pl" || cr.EcuCodeValue() != "0" {
		t.Fatalf("envelope mismatch: %+v", cr)
	}
	co, ok := cr.Result.(dto.InnerParamQryCo)
	if !ok {
		t.Fatalf("result type = %T, want InnerParamQryCo", cr.Result)
	}
	if co.FreqNorm == nil || *co.FreqNorm != 600 {
		t.Errorf("freqNorm not mapped from freq_norm: %+v", co.FreqNorm)
	}
	if co.FreqMove == nil || *co.FreqMove != 5 {
		t.Errorf("freqMove not mapped from freq_move: %+v", co.FreqMove)
	}
	if co.Mode == nil || *co.Mode != 1 {
		t.Errorf("mode not mapped: %+v", co.Mode)
	}
}

func TestBuildCommandResultNonZeroEcuCode(t *testing.T) {
	do := &client.CommandResultDo{Result: `{"code":"17002","result":null}`}
	cr := buildCommandResult(do, mapBleHelmet)
	if cr.EcuCodeValue() != "17002" {
		t.Fatalf("ecuCode = %q, want 17002", cr.EcuCodeValue())
	}
	if cr.Result != nil {
		t.Fatalf("result should be nil for null result, got %+v", cr.Result)
	}
}

func TestMapDeviceInfoCoordinateWGS84(t *testing.T) {
	orig := config.GlobalConfig.Coordinate.Type
	config.GlobalConfig.Coordinate.Type = 0 // WGS-84: lat/lng <- wgs84*
	defer func() { config.GlobalConfig.Coordinate.Type = orig }()

	raw := json.RawMessage(`{"imei":"x","gps":{"lat":1.1,"lng":2.2,"wgs84Lat":3.3,"wgs84Lng":4.4}}`)
	out, ok := mapDeviceInfo(raw).(map[string]interface{})
	if !ok {
		t.Fatalf("mapDeviceInfo type = %T", mapDeviceInfo(raw))
	}
	gps := out["gps"].(map[string]interface{})
	if gps["lat"] != 3.3 || gps["lng"] != 4.4 {
		t.Fatalf("coordinate not swapped to WGS84: lat=%v lng=%v", gps["lat"], gps["lng"])
	}
}

func TestMapDeviceInfoGCJ02Untouched(t *testing.T) {
	orig := config.GlobalConfig.Coordinate.Type
	config.GlobalConfig.Coordinate.Type = 1 // GCJ-02: leave lat/lng as-is
	defer func() { config.GlobalConfig.Coordinate.Type = orig }()

	raw := json.RawMessage(`{"gps":{"lat":1.1,"lng":2.2,"wgs84Lat":3.3,"wgs84Lng":4.4}}`)
	out := mapDeviceInfo(raw).(map[string]interface{})
	gps := out["gps"].(map[string]interface{})
	if gps["lat"] != 1.1 || gps["lng"] != 2.2 {
		t.Fatalf("GCJ-02 coords should be untouched: lat=%v lng=%v", gps["lat"], gps["lng"])
	}
}

func TestExtractVoltageMv(t *testing.T) {
	if v := extractVoltageMv(map[string]interface{}{"voltageMv": float64(49130)}); v == nil || *v != 49130 {
		t.Fatalf("voltageMv=%v", v)
	}
	if extractVoltageMv(nil) != nil {
		t.Fatal("nil result must yield nil voltageMv")
	}
	if extractVoltageMv(map[string]interface{}{"voltageMv": nil}) != nil {
		t.Fatal("null voltageMv must yield nil")
	}
	if extractVoltageMv(map[string]interface{}{}) != nil {
		t.Fatal("missing voltageMv must yield nil")
	}
}

func TestRestBatteryFallback(t *testing.T) {
	if got := restBatteryFallback(map[string]interface{}{"restBattery": float64(86)}); got != 86 {
		t.Fatalf("got %d, want 86", got)
	}
	if got := restBatteryFallback(map[string]interface{}{}); got != 0 {
		t.Fatalf("got %d, want 0", got)
	}
	if got := restBatteryFallback(nil); got != 0 {
		t.Fatalf("nil device got %d, want 0", got)
	}
}

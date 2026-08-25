package service

import (
	"errors"
	"testing"
)

// TestGPSPointsRejectsBadRangeBeforeCallingUpstream: both checks run before any
// upstream call, so an unbounded window is refused here instead of walking a
// months-long ZSET on the shared worker.
func TestGPSPointsRejectsBadRangeBeforeCallingUpstream(t *testing.T) {
	const day = int64(24 * 3600)
	cases := []struct {
		name       string
		start, end int64
		want       error
	}{
		{"end before start", 2000, 1000, ErrRangeInvalid},
		{"zero width", 1000, 1000, ErrRangeInvalid},
		{"eight days", 0, 8 * day, ErrRangeTooLarge},
		{"one year", 0, 365 * day, ErrRangeTooLarge},
	}
	for _, tc := range cases {
		_, err := GPSPoints("t1", "863488061234567", tc.start, tc.end)
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: GPSPoints(%d, %d) error = %v, want %v", tc.name, tc.start, tc.end, err, tc.want)
		}
	}
}

// TestGPSPointsAllowsExactlyMaxRange pins the boundary, so a caller asking for
// the documented seven days is not refused by an off-by-one.
func TestGPSPointsAllowsExactlyMaxRange(t *testing.T) {
	const start = int64(1_700_000_000)
	end := start + int64(MaxTrajectoryRange.Seconds())
	_, err := GPSPoints("t1", "863488061234567", start, end)
	if errors.Is(err, ErrRangeTooLarge) || errors.Is(err, ErrRangeInvalid) {
		t.Errorf("a %v window was rejected as an invalid range: %v", MaxTrajectoryRange, err)
	}
}

// TestProjectRealtimeResultDropsInternalFields: the c34 result carries fields
// that only mean something inside our platform, and forwarding them verbatim
// both leaked them and skipped the battery object the spec documents.
func TestProjectRealtimeResultDropsInternalFields(t *testing.T) {
	got := projectRealtimeResult(map[string]interface{}{
		"acc":         float64(1),
		"defend":      float64(0),
		"gsm":         float64(23),
		"bmsSoc":      float64(87),
		"bmsVoltage":  float64(48200),
		"helmet6Lock": float64(1),
		"bmsComm":     float64(1),
		"etcSpeed":    float64(12),
		"kickStand":   float64(0),
		"rfid":        "abc",
	})

	for _, leaked := range []string{"helmet6Lock", "bmsComm", "etcSpeed", "kickStand", "rfid"} {
		if _, ok := got[leaked]; ok {
			t.Errorf("projectRealtimeResult kept internal field %q", leaked)
		}
	}
	if got["acc"] != float64(1) || got["defend"] != float64(0) {
		t.Errorf("projectRealtimeResult dropped documented fields: %#v", got)
	}
	if got["GSMSignal"] != float64(23) {
		t.Errorf("GSMSignal = %#v, want 23 — the spec publishes signal under both names", got["GSMSignal"])
	}
	battery, ok := got["battery"].(map[string]interface{})
	if !ok {
		t.Fatalf("battery = %#v, want an object built from the bms fields", got["battery"])
	}
	if battery["percent"] != float64(87) || battery["voltage"] != float64(48200) {
		t.Errorf("battery = %#v, want percent 87 and voltage 48200", battery)
	}
}

// TestProjectRealtimeResultOmitsUnreported: the upstream serializes absent fields
// as null, and the spec says which fields exist depends on the device type.
func TestProjectRealtimeResultOmitsUnreported(t *testing.T) {
	got := projectRealtimeResult(map[string]interface{}{
		"acc":        float64(1),
		"defend":     nil,
		"bmsSoc":     nil,
		"bmsVoltage": nil,
	})
	if _, ok := got["defend"]; ok {
		t.Error("defend was null upstream but appeared in the response")
	}
	if _, ok := got["battery"]; ok {
		t.Error("battery object emitted with no bms values reported")
	}

	if len(projectRealtimeResult(nil)) != 0 {
		t.Error("projectRealtimeResult(nil) produced fields")
	}
}

// TestProjectRealtimeGPSUsesWGS84: the internal result carries both coordinate
// pairs and this endpoint specifies WGS84, so picking the GCJ02 pair would offset
// every point by a few hundred metres.
func TestProjectRealtimeGPSUsesWGS84(t *testing.T) {
	got := projectRealtimeGPS(map[string]interface{}{
		"lat":       float64(30.5),
		"lng":       float64(114.3),
		"wgs84Lat":  float64(30.497),
		"wgs84Lng":  float64(114.295),
		"timestamp": float64(1_700_000_000),
		"speed":     float64(12),
		"course":    float64(180),
	})
	if got["lat"] != float64(30.497) || got["lng"] != float64(114.295) {
		t.Errorf("gps = %#v, want the wgs84 pair", got)
	}
}

// TestProjectRealtimeGPSNeedsBothCoordinates avoids publishing a half-fix, which
// a caller would read as a position.
func TestProjectRealtimeGPSNeedsBothCoordinates(t *testing.T) {
	for _, in := range []map[string]interface{}{
		{"wgs84Lat": float64(30.5)},
		{"wgs84Lng": float64(114.3)},
		{"wgs84Lat": nil, "wgs84Lng": nil, "speed": float64(0)},
		{},
	} {
		if got := projectRealtimeGPS(in); len(got) != 0 {
			t.Errorf("projectRealtimeGPS(%#v) = %#v, want nothing", in, got)
		}
	}
}

// TestC34ResultAcceptsObjectOrString covers paas returning the payload either
// way depending on the upstream path.
func TestC34ResultAcceptsObjectOrString(t *testing.T) {
	if got := c34Result(map[string]interface{}{"acc": float64(1)}); got["acc"] != float64(1) {
		t.Errorf("c34Result(object) = %#v", got)
	}
	if got := c34Result(`{"acc":1}`); got["acc"] != float64(1) {
		t.Errorf("c34Result(string) = %#v", got)
	}
	for _, in := range []interface{}{nil, "", "not json", float64(3)} {
		if got := c34Result(in); got != nil {
			t.Errorf("c34Result(%#v) = %#v, want nil", in, got)
		}
	}
}

package service

import (
	"encoding/json"
	"testing"
)

// TestNormalizeImeiRejectsWrongShapes matters because the imei is interpolated
// into Redis key names and forwarded to four upstream services; anything that is
// not 15 digits has to fail here rather than become a lookup that cannot hit.
func TestNormalizeImeiRejectsWrongShapes(t *testing.T) {
	cases := map[string]interface{}{
		"nil":            nil,
		"empty":          "",
		"blank":          "   ",
		"too short":      "12345678901234",
		"too long":       "1234567890123456",
		"letters":        "12345678901234a",
		"leading plus":   "+12345678901234",
		"dashes":         "1234-567890-123",
		"redis wildcard": "12345678901234*",
	}
	for name, in := range cases {
		if got, err := NormalizeImei(in); err == nil {
			t.Errorf("%s: NormalizeImei(%#v) = %q, want an error", name, in, got)
		}
	}
}

func TestNormalizeImeiAcceptsNumberOrString(t *testing.T) {
	const want = "863488061234567"
	cases := map[string]interface{}{
		"string":      want,
		"padded":      "  " + want + "  ",
		"float64":     float64(863488061234567),
		"int64":       int64(863488061234567),
		"json.Number": json.Number(want),
	}
	for name, in := range cases {
		got, err := NormalizeImei(in)
		if err != nil {
			t.Errorf("%s: NormalizeImei(%#v) returned %v", name, in, err)
			continue
		}
		if got != want {
			t.Errorf("%s: NormalizeImei(%#v) = %q, want %q", name, in, got, want)
		}
	}
}

// TestOptHelpersDistinguishAbsentFromZero is the whole point of the Opt* helpers:
// the shadow serializes every field, so a value no report has filled in arrives
// as null, and reading it with a default turns "unknown" into a confident zero.
func TestOptHelpersDistinguishAbsentFromZero(t *testing.T) {
	raw := map[string]interface{}{
		"reportedZero": float64(0),
		"reportedOne":  float64(1),
		"neverSet":     nil,
	}

	if got := OptInt(raw, "reportedZero"); got == nil || *got != 0 {
		t.Errorf("OptInt(reportedZero) = %v, want a pointer to 0", got)
	}
	if got := OptInt(raw, "reportedOne"); got == nil || *got != 1 {
		t.Errorf("OptInt(reportedOne) = %v, want a pointer to 1", got)
	}
	if got := OptInt(raw, "neverSet"); got != nil {
		t.Errorf("OptInt(neverSet) = %v, want nil for an explicit JSON null", *got)
	}
	if got := OptInt(raw, "notInMap"); got != nil {
		t.Errorf("OptInt(notInMap) = %v, want nil for a missing key", *got)
	}

	if got := OptFloat(raw, "reportedZero"); got == nil || *got != 0 {
		t.Errorf("OptFloat(reportedZero) = %v, want a pointer to 0", got)
	}
	if got := OptFloat(raw, "neverSet"); got != nil {
		t.Errorf("OptFloat(neverSet) = %v, want nil", *got)
	}
	if got := OptInt64(raw, "neverSet"); got != nil {
		t.Errorf("OptInt64(neverSet) = %v, want nil", *got)
	}
}

// TestDeviceDetailOmitsUnreportedFields is the serialized consequence: a device
// that has never reported must not appear in the response as unlocked, offline
// and sitting at (0,0).
func TestDeviceDetailOmitsUnreportedFields(t *testing.T) {
	body, err := json.Marshal(&DeviceDetail{Imei: "863488061234567"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got, want := string(body), `{"imei":"863488061234567"}`; got != want {
		t.Errorf("marshalled empty detail = %s, want %s", got, want)
	}
}

// TestDeviceDetailKeepsReportedZero is the other half: a device that reported
// batteryLock=0 must say so, which omitempty on a plain int would have dropped.
func TestDeviceDetailKeepsReportedZero(t *testing.T) {
	zero := 0
	body, err := json.Marshal(&DeviceDetail{Imei: "863488061234567", BatteryLock: &zero})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got, want := string(body), `{"imei":"863488061234567","batteryLock":0}`; got != want {
		t.Errorf("marshalled detail = %s, want %s", got, want)
	}
}

func TestCoordinatesRequiresBothValues(t *testing.T) {
	lat, lng := 30.5, 114.3
	if _, _, ok := (&DeviceDetail{}).Coordinates(); ok {
		t.Error("Coordinates() on an unreported device returned ok=true")
	}
	if _, _, ok := (&DeviceDetail{Lat: &lat}).Coordinates(); ok {
		t.Error("Coordinates() with only lat returned ok=true")
	}
	gotLng, gotLat, ok := (&DeviceDetail{Lat: &lat, Lng: &lng}).Coordinates()
	if !ok || gotLat != lat || gotLng != lng {
		t.Errorf("Coordinates() = (%v, %v, %v), want (%v, %v, true)", gotLng, gotLat, ok, lng, lat)
	}
}

// TestParseVersionOmitsNonNumeric protects against reporting version 0, which
// reads as a real firmware revision a caller would compare against.
func TestParseVersionOmitsNonNumeric(t *testing.T) {
	if got := parseVersion("460905"); got == nil || *got != 460905 {
		t.Errorf("parseVersion(\"460905\") = %v, want 460905", got)
	}
	for _, in := range []interface{}{nil, "", "v1.2.3", "unknown"} {
		if got := parseVersion(in); got != nil {
			t.Errorf("parseVersion(%#v) = %d, want nil", in, *got)
		}
	}
}

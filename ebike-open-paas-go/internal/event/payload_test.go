package event

import (
	"encoding/json"
	"testing"
)

// gpsRecord is a saas_0 GPS record shaped like ebike-device-worker's output: the
// Bin68 decoder fields merged with worker's appId/deviceDataType/imei.
const gpsRecord = `{
  "appId": 1000,
  "appName": "tenant",
  "deviceDataType": "gps",
  "imei": "865067022403441",
  "cmd": 3,
  "msgType": "data",
  "sw": 86,
  "gsm": 23,
  "voltage": 50684,
  "timestamp": 1783650862,
  "wgs84Lng": 115.40901,
  "wgs84Lat": 22.987791,
  "lng": 115.413979,
  "lat": 22.985251,
  "speed": 0,
  "course": 11,
  "hdop": 186,
  "satellite": 7,
  "isFenceEnable": 1,
  "isOutofServAera": 0
}`

func TestGPSPayloadUsesWgs84AndRawSw(t *testing.T) {
	r, err := Parse([]byte(gpsRecord))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if r.TenantID() != "1000" {
		t.Fatalf("TenantID = %q, want \"1000\"", r.TenantID())
	}

	got := GPSPayload(r)

	// Xiaoan expects WGS84; forwarding the decoder's GCJ02 lng/lat would offset
	// every point by a few hundred metres.
	if got["longitude"] != 115.40901 {
		t.Errorf("longitude = %v, want the wgs84 value 115.40901", got["longitude"])
	}
	if got["latitude"] != 22.987791 {
		t.Errorf("latitude = %v, want the wgs84 value 22.987791", got["latitude"])
	}
	if got["sw"] != float64(86) {
		t.Errorf("sw = %v, want the raw switch word 86", got["sw"])
	}
	if got["satellite"] != float64(7) {
		t.Errorf("satellite = %v, want 7", got["satellite"])
	}
	if got["hdop"] != 1.86 {
		t.Errorf("hdop = %v, want 186/100 = 1.86", got["hdop"])
	}
	// gsm on a GPS report, not gsmSignal.
	if got["gsm"] != float64(23) {
		t.Errorf("gsm = %v, want 23", got["gsm"])
	}
	if _, present := got["gsmSignal"]; present {
		t.Error("gsmSignal must not appear in a GPS payload")
	}
}

func TestGPSPayloadMarshalsWholeNumbersAsIntegers(t *testing.T) {
	r, err := Parse([]byte(gpsRecord))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	raw, err := json.Marshal(GPSPayload(r))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var back map[string]json.Number
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	// The float64 record fields must not surface as "50684.0" to third parties.
	for _, key := range []string{"voltage", "timestamp", "speed", "course", "satellite", "sw", "gsm"} {
		if _, err := back[key].Int64(); err != nil {
			t.Errorf("%s = %s, want an integer encoding", key, back[key])
		}
	}
}

func TestPayloadOmitsAbsentFields(t *testing.T) {
	r, err := Parse([]byte(`{"appId":1000,"deviceDataType":"gps","imei":"x","wgs84Lng":1,"wgs84Lat":2}`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got := GPSPayload(r)
	// An absent field must be omitted, not sent as 0: a third party has to be
	// able to tell "not reported" from "reported zero".
	for _, key := range []string{"voltage", "speed", "hdop", "satellite", "sw"} {
		if _, present := got[key]; present {
			t.Errorf("%s present with value %v, want omitted", key, got[key])
		}
	}
}

// TestNormalizeHdopHandlesBothEncodings covers the reason the scaling is not a
// plain divide: worker labels several commands "gps" and rewrites cmd to 3, so a
// record no longer says whether Bin68 (centi-units, 186) or Bin41 (already
// scaled, 1.86) produced it. Dividing unconditionally made every Bin41 report a
// hundredth of its real value.
func TestNormalizeHdopHandlesBothEncodings(t *testing.T) {
	cases := []struct {
		name string
		raw  float64
		want float64
	}{
		{"bin68 centi-units", 186, 1.86},
		{"bin68 best fix", 50, 0.5},
		{"bin68 poor fix", 1500, 15},
		{"bin41 already scaled", 1.86, 1.86},
		{"bin41 best fix", 0.5, 0.5},
		{"bin41 upper plausible", 20, 20},
		{"no fix reported", 0, 0},
	}
	for _, tc := range cases {
		if got := normalizeHdop(tc.raw); got != tc.want {
			t.Errorf("%s: normalizeHdop(%v) = %v, want %v", tc.name, tc.raw, got, tc.want)
		}
	}
}

// TestGPSPayloadKeepsScaledHdop is the same case through the payload builder,
// since that is where the bug was visible to third parties.
func TestGPSPayloadKeepsScaledHdop(t *testing.T) {
	r, err := Parse([]byte(`{"appId":1000,"deviceDataType":"gps","imei":"x","wgs84Lng":1,"wgs84Lat":2,"hdop":1.86}`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := GPSPayload(r)["hdop"]; got != 1.86 {
		t.Errorf("hdop = %v, want 1.86 left as reported", got)
	}
}

func TestBMSPayloadForwardsRawFault(t *testing.T) {
	r, err := Parse([]byte(`{
	  "appId": 1000, "deviceDataType": "bms", "imei": "x",
	  "sn": "BMS123", "soc": 47, "voltage": 4982, "current": -120,
	  "fault": 12, "soh": 98, "cycleLifeCounter": 31, "mosTemperature": 25,
	  "isDischargeOverCurrent": 1, "isChargeOverCurrent": 0
	}`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got := BMSPayload(r)

	// fault carries bits 2 and 3 faithfully; the decoders' per-bit booleans do
	// not, so the raw field is the only correct source.
	if got["fault"] != float64(12) {
		t.Errorf("fault = %v, want the raw bitfield 12", got["fault"])
	}
	if got["sn"] != "BMS123" {
		t.Errorf("sn = %v, want BMS123", got["sn"])
	}
	// Xiaoan renames these; sending our internal names would break clients.
	for ours, theirs := range map[string]string{
		"soh":              "healthState",
		"cycleLifeCounter": "cycle",
		"mosTemperature":   "MOSTemp",
	} {
		if _, present := got[theirs]; !present {
			t.Errorf("%s must be published as %s", ours, theirs)
		}
		if _, present := got[ours]; present {
			t.Errorf("internal name %s must not be published", ours)
		}
	}
}

func TestPingPayloadRenamesGsmSignal(t *testing.T) {
	r, err := Parse([]byte(`{"appId":1000,"deviceDataType":"ping","imei":"x","gsmSignal":19,"voltage":48}`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got := PingPayload(r)
	if got["gsm"] != float64(19) {
		t.Errorf("gsm = %v, want 19", got["gsm"])
	}
	if got["voltage"] != float64(48) {
		t.Errorf("voltage = %v, want 48", got["voltage"])
	}
}

// TestParseToleratesFractionalIntegers guards the reason every numeric field is
// float64: saas_0 carries GPS from several protocols and one of them emitting
// 0.0 where another emits 0 must not fail the whole record.
func TestParseToleratesFractionalIntegers(t *testing.T) {
	r, err := Parse([]byte(`{"appId":1000,"deviceDataType":"gps","imei":"x","speed":0.0,"course":11.0,"voltage":50684.0}`))
	if err != nil {
		t.Fatalf("Parse rejected fractional encoding: %v", err)
	}
	if r.Speed == nil || *r.Speed != 0 {
		t.Errorf("Speed = %v, want 0", r.Speed)
	}
	if r.Course == nil || *r.Course != 11 {
		t.Errorf("Course = %v, want 11", r.Course)
	}
}

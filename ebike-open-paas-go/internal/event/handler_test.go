package event

import "testing"

// TestIsUnfixedGPSOnlyMatchesReportedZeroPair: Xiaoan v1.1.7 says a report at
// (0,0) is a device without a fix and must not be forwarded, but a record that
// omits the coordinates is an unrecognised decoder shape rather than a device in
// the Gulf of Guinea — treating the two the same would silently drop reports.
func TestIsUnfixedGPSOnlyMatchesReportedZeroPair(t *testing.T) {
	f := func(v float64) *float64 { return &v }

	cases := []struct {
		name     string
		lat, lng *float64
		want     bool
	}{
		{"reported (0,0)", f(0), f(0), true},
		{"real fix", f(22.987791), f(115.40901), false},
		{"zero latitude on the equator", f(0), f(115.40901), false},
		{"zero longitude on the meridian", f(51.4778), f(0), false},
		{"coordinates absent", nil, nil, false},
		{"latitude only", f(0), nil, false},
		{"longitude only", nil, f(0), false},
	}
	for _, tc := range cases {
		r := &Record{Wgs84Lat: tc.lat, Wgs84Lng: tc.lng}
		if got := isUnfixedGPS(r); got != tc.want {
			t.Errorf("%s: isUnfixedGPS = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestHandleSurvivesMalformedRecords: the consumer commits its batch either way,
// so a record that cannot be parsed must be counted and skipped rather than take
// the rest of the batch down with it.
func TestHandleSurvivesMalformedRecords(t *testing.T) {
	beforeSeen, _, beforeMalformed, _, beforePanicked := Stats()

	for _, raw := range [][]byte{
		[]byte(``),
		[]byte(`not json`),
		[]byte(`{`),
		[]byte(`{"appId":1000}`),             // no imei
		[]byte(`{"imei":"865067022403441"}`), // no tenant
		[]byte(`{"appId":1000,"imei":"865067022403441"}`), // no data type
		[]byte(`{"appId":"","imei":"","deviceDataType":"gps"}`),
	} {
		Handle(raw)
	}

	afterSeen, _, afterMalformed, _, afterPanicked := Stats()
	if got := afterSeen - beforeSeen; got != 7 {
		t.Errorf("seen advanced by %d, want 7", got)
	}
	if afterMalformed == beforeMalformed {
		t.Error("malformed counter did not advance; bad records went unnoticed")
	}
	if afterPanicked != beforePanicked {
		t.Errorf("panicked advanced by %d, want none", afterPanicked-beforePanicked)
	}
}

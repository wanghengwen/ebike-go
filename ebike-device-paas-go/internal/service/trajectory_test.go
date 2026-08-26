package service

import (
	"encoding/json"
	"testing"

	"ebike-device-paas-go/internal/api/dto"
)

func TestRoundHalfEven(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{48.125, 48.12}, // .5 -> down to even
		{48.135, 48.14}, // .5 -> up to even
		{48.121, 48.12},
		{48.126, 48.13},
		{0, 0},
	}
	for _, c := range cases {
		if got := roundHalfEven(c.in, 2); got != c.want {
			t.Errorf("roundHalfEven(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestReshapeMetric(t *testing.T) {
	raw := `[{"timestamp":1700000000,"course":90,"speed":12,"lng":116.3,"lat":39.9,"gsm":25,"voltage":48250}]`
	var list []map[string]json.Number
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		t.Fatal(err)
	}
	m := reshapeMetric(list)

	gps, ok := m["gps"].([]dto.MetricGps)
	if !ok || len(gps) != 1 {
		t.Fatalf("gps = %v, want 1 element", m["gps"])
	}
	if gps[0].Timestamp != 1700000000000 || gps[0].Course != 90 || gps[0].Speed != 12 {
		t.Fatalf("gps[0] = %+v", gps[0])
	}
	if gps[0].Lng != 116.3 || gps[0].Lat != 39.9 {
		t.Fatalf("gps[0] coords = %v,%v", gps[0].Lng, gps[0].Lat)
	}

	v := m["voltage"].(dto.MetricVoltage)
	if len(v.Voltage) != 1 || v.Voltage[0] != 48.25 {
		t.Fatalf("voltage = %v, want [48.25] (mV/1000)", v.Voltage)
	}
	if len(v.Time) != 1 {
		t.Fatalf("voltage.time = %v, want 1 formatted timestamp", v.Time)
	}

	course := m["course"].(dto.MetricCourse)
	if len(course.Course) != 1 || course.Course[0] != 90 {
		t.Fatalf("course = %v", course.Course)
	}
}

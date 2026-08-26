package service

import (
	"testing"
)

func TestFilterRealGps(t *testing.T) {
	now := int64(1_000_000)
	lng, lat := 116.397, 39.908

	devices := []map[string]interface{}{
		// fresh: timestamp + 15 > now
		{"imei": "fresh001", "lng": lng, "lat": lat, "timestamp": now - 5},
		// boundary stale: timestamp + 15 == now -> excluded (strict >)
		{"imei": "edge002", "lng": lng, "lat": lat, "timestamp": now - 15},
		// stale: well past freshness window
		{"imei": "stale003", "lng": lng, "lat": lat, "timestamp": now - 3600},
		// missing timestamp -> excluded
		{"imei": "nots004", "lng": lng, "lat": lat},
		// float64 timestamp (as decoded from JSON) should also work
		{"imei": "fresh005", "lng": lng, "lat": lat, "timestamp": float64(now - 1)},
	}

	got := FilterRealGps(devices, now)

	if len(got) != 2 {
		t.Fatalf("expected 2 fresh devices, got %d: %+v", len(got), got)
	}
	wantImeis := map[string]bool{"fresh001": true, "fresh005": true}
	for _, co := range got {
		if !wantImeis[co.Imei] {
			t.Errorf("unexpected imei in result: %s", co.Imei)
		}
		if co.Lng == nil || co.Lat == nil {
			t.Errorf("imei %s missing lng/lat", co.Imei)
			continue
		}
		if *co.Lng != lng || *co.Lat != lat {
			t.Errorf("imei %s wrong coords: %v,%v", co.Imei, *co.Lng, *co.Lat)
		}
	}
}

func TestFilterRealGpsEmpty(t *testing.T) {
	if got := FilterRealGps(nil, 1); len(got) != 0 {
		t.Fatalf("expected empty result, got %d", len(got))
	}
}

package geo

import (
	"math"
	"testing"
)

func TestWgs84Gcj02RoundTrip(t *testing.T) {
	// A point inside China (Beijing) should shift, then round-trip back closely.
	lat, lon := 39.9087, 116.3975
	gLat, gLon := Wgs84ToGcj02(lat, lon)
	if math.Abs(gLat-lat) < 1e-5 || math.Abs(gLon-lon) < 1e-5 {
		t.Fatalf("expected GCJ-02 shift, got (%f,%f) for (%f,%f)", gLat, gLon, lat, lon)
	}
	rLat, rLon := Gcj02ToWgs84(gLat, gLon)
	if math.Abs(rLat-lat) > 1e-4 || math.Abs(rLon-lon) > 1e-4 {
		t.Fatalf("round-trip drifted too far: got (%f,%f), want ~(%f,%f)", rLat, rLon, lat, lon)
	}
}

func TestOutOfChinaPassthrough(t *testing.T) {
	// Tokyo is outside the China bounds -> returned unchanged.
	lat, lon := 35.6895, 139.6917
	gLat, gLon := Wgs84ToGcj02(lat, lon)
	if gLat != lat || gLon != lon {
		t.Fatalf("out-of-China point must pass through, got (%f,%f)", gLat, gLon)
	}
}

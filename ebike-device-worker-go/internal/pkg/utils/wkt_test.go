package utils_test

import (
	"testing"

	"ebike-device-worker-go/internal/pkg/utils"
)

func TestFormatPointZM(t *testing.T) {
	wkt := utils.FormatPointZM(115.33446502685547, 22.972423553466797, 1781660822)
	if wkt != "POINT ZM (115.334465 22.972424 0 1781660822)" {
		t.Fatalf("wkt=%s", wkt)
	}
}

func TestParsePointWKTWithM(t *testing.T) {
	p, err := utils.ParsePointWKT("POINT(114.12 22.45 1700000000000)")
	if err != nil {
		t.Fatal(err)
	}
	if p.Lng != 114.12 || p.Lat != 22.45 || p.M != 1700000000000 {
		t.Fatalf("unexpected point: %+v", p)
	}
}

func TestFormatLineStringM(t *testing.T) {
	wkt, err := utils.FormatLineStringM([]utils.Point{
		{Lng: 114.1, Lat: 22.1, M: 1000},
		{Lng: 114.2, Lat: 22.2, M: 2000},
	})
	if err != nil {
		t.Fatal(err)
	}
	if wkt != "LINESTRING(114.1 22.1 1000, 114.2 22.2 2000)" {
		t.Fatalf("wkt=%s", wkt)
	}
}

func TestFormatLineStringMUnfilteredAllowsZeroCoords(t *testing.T) {
	wkt, err := utils.FormatLineStringMUnfiltered([]utils.Point{
		{Lng: 0, Lat: 0, M: 1000},
		{Lng: 114.2, Lat: 22.2, M: 2000},
	})
	if err != nil {
		t.Fatal(err)
	}
	if wkt != "LINESTRING(0 0 1000, 114.2 22.2 2000)" {
		t.Fatalf("wkt=%s", wkt)
	}
}

func TestValidWGS84RejectsZero(t *testing.T) {
	if utils.ValidWGS84(0, 0) {
		t.Fatal("zero coordinate should be invalid")
	}
}

func TestParseLineStringWKTWithM(t *testing.T) {
	points, err := utils.ParseLineStringWKT("LINESTRING(114.1 22.1 1000, 114.2 22.2 2000)")
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 2 {
		t.Fatalf("expected 2 points, got %d", len(points))
	}
	if points[0].M != 1000 || points[1].M != 2000 {
		t.Fatalf("unexpected M values: %+v", points)
	}
}

func TestGetDistanceVincentyVsSphere(t *testing.T) {
	// Same endpoints; Vincenty (type 0) and Sphere (type 1) should differ slightly but stay close.
	lng1, lat1 := 114.05, 22.55
	lng2, lat2 := 114.06, 22.56
	wgs := utils.GetDistance(lng1, lat1, lng2, lat2, 0)
	sph := utils.GetDistance(lng1, lat1, lng2, lat2, 1)
	if wgs <= 0 || sph <= 0 {
		t.Fatalf("distances should be positive: wgs=%v sph=%v", wgs, sph)
	}
	if diff := wgs - sph; diff < -50 || diff > 50 {
		t.Fatalf("distance algorithms diverged too much: wgs=%v sph=%v", wgs, sph)
	}
}

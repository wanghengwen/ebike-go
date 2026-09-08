package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ebike-fence-go/internal/api/dto"
	domainconfig "ebike-fence-go/internal/domain/config"
	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/geo"
)

// Production log fixtures from ebike-fence-1.log.

func loadLogYangheServiceArea(t *testing.T) *gateway.FenceE {
	t.Helper()
	path := filepath.Join("..", "geo", "testdata", "yanghe_service_area.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	pointList := strings.TrimSpace(strings.TrimPrefix(string(raw), "\ufeff"))
	poly, err := geo.ParsePolygon(pointList)
	if err != nil {
		t.Fatal(err)
	}
	return &gateway.FenceE{
		Id:            338362359727786125,
		Name:          "洋河镇",
		ParsedPolygon: poly,
		PointList:     pointList,
	}
}

func TestLogParityParseComputePointGeoJSON(t *testing.T) {
	var cmd dto.ComputeDistanceCmd
	if err := json.Unmarshal([]byte(`{"imei":"862551059864149","point":"{\"type\":\"Point\",\"coordinates\":[118.399564,33.768009]}"}`), &cmd); err != nil {
		t.Fatal(err)
	}
	loc, err := parseComputePoint(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if loc.Lng != 118.399564 || loc.Lat != 33.768009 {
		t.Fatalf("got %+v", loc)
	}
}

func TestLogParityIzInService(t *testing.T) {
	svc := loadLogYangheServiceArea(t)

	inside := geo.Location{Lng: 118.399564, Lat: 33.768009}
	if !izInService(inside, geo.Location{}, svc) {
		t.Fatal("118.399564,33.768009 should be inside 洋河镇")
	}

	outside := geo.Location{Lng: 118.381109, Lat: 33.753969}
	if izInService(outside, geo.Location{}, svc) {
		t.Fatal("118.381109,33.753969 should be outside 洋河镇 (OUT_SERVICE_AREA)")
	}
}

func TestLogParityIzCanReturnFromLog(t *testing.T) {
	yangheCfg := &domainconfig.ConfigBackcarCO{
		AllowOutofService: boolPtr(true),
		AllowOutofParking: boolPtr(true),
		AllowInNostop:     boolPtr(false),
		IzHelmetReign:     boolPtr(true),
	}

	cases := []struct {
		name       string
		returnType FenceRelation
		cfg        *domainconfig.ConfigBackcarCO
		want       bool
	}{
		{
			name:       "IN_PARKING_LAT tenant1003",
			returnType: IN_PARKING_LAT,
			cfg:        yangheCfg,
			want:       true,
		},
		{
			name:       "OUT_SERVICE_AREA tenant1003 allowOutofService=true",
			returnType: OUT_SERVICE_AREA,
			cfg:        yangheCfg,
			want:       true,
		},
		{
			name:       "OUT_PARKING_LAT tenant1003 allowOutofParking=true",
			returnType: OUT_PARKING_LAT,
			cfg:        yangheCfg,
			want:       true,
		},
		{
			name:       "OUT_PARKING_LAT tenant1004 allowOutofParking=false",
			returnType: OUT_PARKING_LAT,
			cfg: &domainconfig.ConfigBackcarCO{
				AllowOutofService: boolPtr(false),
				AllowOutofParking: boolPtr(false),
				AllowInNostop:     boolPtr(false),
				IzHelmetReign:     boolPtr(true),
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := *izCanReturn(tc.returnType, tc.cfg)
			if got != tc.want {
				t.Fatalf("izCanReturn(%s)=%v want %v", tc.returnType, got, tc.want)
			}
		})
	}
}

func TestLogParityComputeDistanceCloseLine(t *testing.T) {
	svc := loadLogYangheServiceArea(t)
	loc := geo.Location{Lng: 115.400423, Lat: 22.976168}
	dist := geo.PointToPolygonDistanceJava(loc, svc.ParsedPolygon)
	if dist != 0 {
		t.Fatalf("distance=%v want 0", dist)
	}
	izCloseLine := dist <= 200
	if !izCloseLine {
		t.Fatal("izCloseLine should be true when distance=0")
	}
}

func TestLogParityComputeDistanceInside(t *testing.T) {
	svc := loadLogYangheServiceArea(t)
	loc := geo.Location{Lng: 118.399564, Lat: 33.768009}
	dist := geo.PointToPolygonDistanceJava(loc, svc.ParsedPolygon)
	const javaWant = 1618.0
	const tolerance = 50.0
	if dist < javaWant-tolerance || dist > javaWant+tolerance {
		t.Fatalf("distance=%v want ~%v (±%v)", dist, javaWant, tolerance)
	}
	if dist <= 200 {
		t.Fatalf("izCloseLine should be false for distance=%v", dist)
	}
}

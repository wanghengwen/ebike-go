package geo

import (
	"os"
	"strings"
	"testing"
)

// Production log fixtures from ebike-fence-1.log (Java parity golden cases).

const (
	logParkingYanglao = "[[118.399436,33.768123],[118.399624,33.768018],[118.399584,33.767969],[118.399398,33.768074]]"
	logParkingZhengfu = "[[118.379182,33.765695],[118.379289,33.765667],[118.37915,33.765305],[118.379049,33.765331]]"
	logParkingErHuan  = "[[115.33663,22.982641],[115.336555,22.983349],[115.336708,22.983363],[115.336783,22.982656]]"
)

func loadLogServiceAreaYanghe(t *testing.T) []Location {
	t.Helper()
	raw, err := os.ReadFile("testdata/yanghe_service_area.json")
	if err != nil {
		t.Fatal(err)
	}
	s := strings.TrimSpace(strings.TrimPrefix(string(raw), "\ufeff"))
	poly, err := ParsePolygon(s)
	if err != nil {
		t.Fatal(err)
	}
	return poly
}

func mustParseLogPolygon(t *testing.T, s string) []Location {
	t.Helper()
	poly, err := ParsePolygon(s)
	if err != nil {
		t.Fatal(err)
	}
	return poly
}

func TestLogParityBindParkingYanglaoIntersect(t *testing.T) {
	// bindParking: lng=118.399564 lat=33.768009 buffer=11 → parkingId=185442561702239423
	loc := Location{Lng: 118.399564, Lat: 33.768009}
	poly := mustParseLogPolygon(t, logParkingYanglao)
	if !Intersect(loc, poly, 11) {
		t.Fatal("敬老院 bind point should intersect with 11m buffer")
	}
	if !IsPointInParsedPolygon(loc, poly) {
		t.Fatal("bind point should be inside 敬老院 polygon")
	}
}

func TestLogParityBindParkingNearErHuan(t *testing.T) {
	// bindParking near: lng=115.336815 lat=22.982905 buffer=6+10=16 → nearParkingId=274866299795413082
	loc := Location{Lng: 115.336815, Lat: 22.982905}
	poly := mustParseLogPolygon(t, logParkingErHuan)
	if Intersect(loc, poly, 6) {
		t.Fatal("near point should not intersect with parking buffer 6m alone")
	}
	if !Intersect(loc, poly, 16) {
		t.Fatal("near point should intersect with bufferDistance+10 (16m)")
	}
}

func TestLogParityComputeDistanceInsideServiceArea(t *testing.T) {
	// computeOutServiceDistance: [118.399564,33.768009] in 洋河镇 → distance=1618 izCloseLine=false
	loc := Location{Lng: 118.399564, Lat: 33.768009}
	poly := loadLogServiceAreaYanghe(t)
	if !IsPointInParsedPolygon(loc, poly) {
		t.Fatal("point should be inside 洋河镇 service area")
	}
	dist := PointToPolygonDistanceJava(loc, poly)
	const javaWant = 1618.0
	const tolerance = 2.0
	if dist < javaWant-tolerance || dist > javaWant+tolerance {
		t.Fatalf("distance=%v want ~%v (±%v)", dist, javaWant, tolerance)
	}
	izCloseLine := dist <= 200
	if izCloseLine {
		t.Fatalf("izCloseLine should be false for distance=%v", dist)
	}
}

func TestLogParityComputeDistanceOutsideServiceArea(t *testing.T) {
	// computeOutServiceDistance: [115.400423,22.976168] tenant 1000 → distance=0 izCloseLine=true
	loc := Location{Lng: 115.400423, Lat: 22.976168}
	poly := loadLogServiceAreaYanghe(t)
	if IsPointInParsedPolygon(loc, poly) {
		t.Fatal("汕尾 point should be outside 洋河镇 service area")
	}
	dist := PointToPolygonDistanceJava(loc, poly)
	if dist != 0 {
		t.Fatalf("outside service area distance=%v want 0", dist)
	}
}

func TestLogParityGetFenceRelationZhengfuParking(t *testing.T) {
	// getFenceRelation: car 33.765658,118.379253 → IN_PARKING_LAT 镇政府
	loc := Location{Lng: 118.379253, Lat: 33.765658}
	poly := mustParseLogPolygon(t, logParkingZhengfu)
	if !Intersect(loc, poly, 11) {
		t.Fatal("镇政府 parking point should intersect with 11m buffer")
	}
}

func TestLogParityGetFenceRelationOutOfService(t *testing.T) {
	// getFenceRelation: car 33.753969,118.381109 → OUT_SERVICE_AREA
	loc := Location{Lng: 118.381109, Lat: 33.753969}
	poly := loadLogServiceAreaYanghe(t)
	if IsPointInParsedPolygon(loc, poly) {
		t.Fatal("point should be outside 洋河镇 service area")
	}
}

func TestLogParityGeoJSONPointParse(t *testing.T) {
	raw := `{"type":"Point","coordinates":[118.399564,33.768009]}`
	loc, err := ParsePointString(raw)
	if err != nil {
		t.Fatal(err)
	}
	if loc.Lng != 118.399564 || loc.Lat != 33.768009 {
		t.Fatalf("got %+v", loc)
	}
}

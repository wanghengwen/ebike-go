package geo

import "testing"

func TestParseAndPointInPolygon(t *testing.T) {
	// unit square (0,0)-(10,10)
	ring := ParsePolygon("[[0,0],[10,0],[10,10],[0,10]]")
	if len(ring) != 4 {
		t.Fatalf("expected 4 points, got %d", len(ring))
	}
	if !PointInPolygon(5, 5, ring) {
		t.Fatal("center should be inside")
	}
	if PointInPolygon(15, 5, ring) {
		t.Fatal("point to the right should be outside")
	}
}

func TestRandomPointsCount(t *testing.T) {
	pts := RandomPointsInPolygon("[[0,0],[10,0],[10,10],[0,10]]", 7)
	if len(pts) != 7 {
		t.Fatalf("expected 7 random points, got %d", len(pts))
	}
	for _, p := range pts {
		if p.Lng < 0 || p.Lng > 10 || p.Lat < 0 || p.Lat > 10 {
			t.Fatalf("point out of bbox: %+v", p)
		}
	}
}

func TestParseEmpty(t *testing.T) {
	if ParsePolygon("") != nil {
		t.Fatal("empty string should parse to nil")
	}
}

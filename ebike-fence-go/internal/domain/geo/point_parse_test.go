package geo

import "testing"

func TestParsePointJSONGeoJSONObject(t *testing.T) {
	raw := []byte(`{"type":"Point","coordinates":[118.399564,33.768009]}`)
	loc, err := ParsePointJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if loc.Lng != 118.399564 || loc.Lat != 33.768009 {
		t.Fatalf("unexpected loc: %+v", loc)
	}
}

func TestParsePointJSONGeoJSONString(t *testing.T) {
	raw := []byte(`"{\"type\":\"Point\",\"coordinates\":[118.399564,33.768009]}"`)
	loc, err := ParsePointJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if loc.Lng != 118.399564 || loc.Lat != 33.768009 {
		t.Fatalf("unexpected loc: %+v", loc)
	}
}

func TestParsePointJSONLngLatObject(t *testing.T) {
	raw := []byte(`{"lng":118.399564,"lat":33.768009}`)
	loc, err := ParsePointJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if loc.Lng != 118.399564 || loc.Lat != 33.768009 {
		t.Fatalf("unexpected loc: %+v", loc)
	}
}

func TestCircleFullyInsideUsesDegreeSpace(t *testing.T) {
	center := Location{Lng: 116.40, Lat: 39.90}
	// Small radius fully inside testSquare should pass with degree-space circle.
	if !circleFullyInsidePolygon(center, 50, testSquare) {
		t.Fatal("50m degree circle should be inside test square at center")
	}
	// Large radius should fail containment.
	if circleFullyInsidePolygon(center, 5000, testSquare) {
		t.Fatal("5000m circle should not be fully inside test square")
	}
}

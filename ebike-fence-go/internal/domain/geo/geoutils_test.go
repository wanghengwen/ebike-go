package geo

import "testing"

// Square around (116.40, 39.90) — roughly 1km box for parity checks.
var testSquare = []Location{
	{Lng: 116.395, Lat: 39.895},
	{Lng: 116.405, Lat: 39.895},
	{Lng: 116.405, Lat: 39.905},
	{Lng: 116.395, Lat: 39.905},
	{Lng: 116.395, Lat: 39.895},
}

func TestIsPointInParsedPolygonInside(t *testing.T) {
	inside := Location{Lng: 116.40, Lat: 39.90}
	if !IsPointInParsedPolygon(inside, testSquare) {
		t.Fatal("point should be inside test square")
	}
}

func TestIsPointInParsedPolygonOutside(t *testing.T) {
	outside := Location{Lng: 116.50, Lat: 39.90}
	if IsPointInParsedPolygon(outside, testSquare) {
		t.Fatal("point should be outside test square")
	}
}

func TestJavaBufferIncludesNearBoundaryPoint(t *testing.T) {
	near := Location{Lng: 116.394, Lat: 39.90}
	if !IsPointInParsedPolygonWithJavaBuffer(near, testSquare, 200) {
		t.Fatal("point within 200m buffer should be considered inside")
	}
}

func TestPointToPolygonDistanceJavaInside(t *testing.T) {
	inside := Location{Lng: 116.40, Lat: 39.90}
	dist := PointToPolygonDistanceJava(inside, testSquare)
	if dist <= 0 {
		t.Fatalf("inside point should have positive in-circle distance, got %v", dist)
	}
}

func TestPointToPointDistanceJavaFence4Parity(t *testing.T) {
	// fence-4 getNearParking first item: 南街人家西门
	got := PointToPointDistanceJava(118.38108642578125, 33.75399197048611, 118.382454, 33.763386)
	want := 1049.6420944330168
	if diff := got - want; diff < -0.01 || diff > 0.01 {
		t.Fatalf("distance=%v want~%v", got, want)
	}
}

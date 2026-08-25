package geo

import "testing"

func TestIntersectCenterInside(t *testing.T) {
	inside := Location{Lng: 116.40, Lat: 39.90}
	if !Intersect(inside, testSquare, 5) {
		t.Fatal("center inside polygon should intersect with any positive buffer")
	}
}

func TestIntersectNearEdgeWithinBuffer(t *testing.T) {
	near := Location{Lng: 116.3945, Lat: 39.90}
	if !Intersect(near, testSquare, 100) {
		t.Fatal("point near west edge should intersect with 100m buffer")
	}
}

func TestIntersectFarOutside(t *testing.T) {
	far := Location{Lng: 116.50, Lat: 39.90}
	if Intersect(far, testSquare, 5) {
		t.Fatal("point far outside should not intersect with 5m buffer")
	}
}

func TestIntersectZeroRadiusOutside(t *testing.T) {
	outside := Location{Lng: 116.50, Lat: 39.90}
	if Intersect(outside, testSquare, 0) {
		t.Fatal("zero radius outside should not intersect")
	}
}

func TestDegreeCircleContains(t *testing.T) {
	center := Location{Lng: 116.40, Lat: 39.90}
	r := 0.001
	inside := Location{Lng: center.Lng + r*0.5, Lat: center.Lat}
	if !degreeCircleContains(center, r, inside) {
		t.Fatal("point inside degree circle should be contained")
	}
	outside := Location{Lng: center.Lng + r*2, Lat: center.Lat}
	if degreeCircleContains(center, r, outside) {
		t.Fatal("point outside degree circle should not be contained")
	}
}

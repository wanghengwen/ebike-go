package geo

import "math"

const jtsQuadrantSegments = 32

// Intersect mirrors Java GeoUtils.intersect: JTS point.buffer(radius * 180 / PI / 6371000, 32)
// intersects polygon in degree space (not geodesic).
func Intersect(loc Location, polygon []Location, radiusMeters float64) bool {
	if len(polygon) < 3 {
		return false
	}
	if radiusMeters <= 0 {
		return IsPointInParsedPolygon(loc, polygon)
	}
	if IsPointInParsedPolygon(loc, polygon) {
		return true
	}

	degreeRadius := radiusMeters * 180 / math.Pi / EarthRadius
	segments := jtsQuadrantSegments * 4
	circle := createDegreeCircle(loc, degreeRadius, segments)

	for _, p := range polygon {
		if degreeCircleContains(loc, degreeRadius, p) {
			return true
		}
	}
	for _, c := range circle {
		if IsPointInParsedPolygon(c, polygon) {
			return true
		}
	}
	n := len(polygon)
	for i := 0; i < len(circle); i++ {
		c1 := circle[i]
		c2 := circle[(i+1)%len(circle)]
		for j := 0; j < n-1; j++ {
			if segmentsIntersect(c1, c2, polygon[j], polygon[j+1]) {
				return true
			}
		}
	}
	return false
}

func createDegreeCircle(center Location, degreeRadius float64, segments int) []Location {
	if segments < 4 {
		segments = 4
	}
	pts := make([]Location, segments)
	for i := 0; i < segments; i++ {
		angle := 2 * math.Pi * float64(i) / float64(segments)
		pts[i] = Location{
			Lng: center.Lng + degreeRadius*math.Cos(angle),
			Lat: center.Lat + degreeRadius*math.Sin(angle),
		}
	}
	return pts
}

func degreeCircleContains(center Location, degreeRadius float64, p Location) bool {
	dx := p.Lng - center.Lng
	dy := p.Lat - center.Lat
	return dx*dx+dy*dy <= degreeRadius*degreeRadius
}

func segmentsIntersect(a, b, c, d Location) bool {
	return ccw(a, c, d) != ccw(b, c, d) && ccw(a, b, c) != ccw(a, b, d)
}

func ccw(a, b, c Location) bool {
	return (c.Lat-a.Lat)*(b.Lng-a.Lng) > (b.Lat-a.Lat)*(c.Lng-a.Lng)
}

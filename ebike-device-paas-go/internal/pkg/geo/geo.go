// Package geo ports the polygon helpers used by genDeviceMapFake
// (GEOJsonUtils): parse a "[[lng,lat],...]" ring, point-in-polygon, and random
// points within the polygon's bounding box.
package geo

import (
	"encoding/json"
	"math/rand"
)

// Point is an [lng, lat] coordinate.
type Point struct {
	Lng float64
	Lat float64
}

// ParsePolygon decodes a "[[lng,lat],...]" JSON ring into points. Returns nil on
// error / empty input.
func ParsePolygon(pointStr string) []Point {
	if pointStr == "" {
		return nil
	}
	var raw [][]float64
	if err := json.Unmarshal([]byte(pointStr), &raw); err != nil {
		return nil
	}
	pts := make([]Point, 0, len(raw))
	for _, p := range raw {
		if len(p) >= 2 {
			pts = append(pts, Point{Lng: p[0], Lat: p[1]})
		}
	}
	if len(pts) == 0 {
		return nil
	}
	return pts
}

// PointInPolygon reports whether (lng,lat) is inside the ring via ray casting.
func PointInPolygon(lng, lat float64, ring []Point) bool {
	n := len(ring)
	if n < 3 {
		return false
	}
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		xi, yi := ring[i].Lng, ring[i].Lat
		xj, yj := ring[j].Lng, ring[j].Lat
		if (yi > lat) != (yj > lat) &&
			lng < (xj-xi)*(lat-yi)/(yj-yi)+xi {
			inside = !inside
		}
		j = i
	}
	return inside
}

// RandomPointsInPolygon mirrors GEOJsonUtils.romPointFromPolygon: sample within
// the bounding box, retry up to 5 times to land inside the polygon, otherwise
// keep the last sample. Returns nil when the polygon cannot be parsed.
func RandomPointsInPolygon(pointStr string, count int) []Point {
	ring := ParsePolygon(pointStr)
	if ring == nil || count <= 0 {
		return nil
	}
	minLng, maxLng := ring[0].Lng, ring[0].Lng
	minLat, maxLat := ring[0].Lat, ring[0].Lat
	for _, p := range ring {
		if p.Lng < minLng {
			minLng = p.Lng
		}
		if p.Lng > maxLng {
			maxLng = p.Lng
		}
		if p.Lat < minLat {
			minLat = p.Lat
		}
		if p.Lat > maxLat {
			maxLat = p.Lat
		}
	}
	out := make([]Point, 0, count)
	for i := 0; i < count; i++ {
		for j := 0; j < 5; j++ {
			lng := rand.Float64()*(maxLng-minLng) + minLng
			lat := rand.Float64()*(maxLat-minLat) + minLat
			if PointInPolygon(lng, lat, ring) || j == 4 {
				out = append(out, Point{Lng: lng, Lat: lat})
				break
			}
		}
	}
	return out
}

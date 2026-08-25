package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// Point holds WGS84 coordinates; M is the PostGIS measure (epoch seconds for GPS/itinerary parity with Java).
type Point struct {
	Lng float64
	Lat float64
	M   float64
}

func parseCoords(parts []string) (Point, error) {
	if len(parts) < 2 {
		return Point{}, fmt.Errorf("invalid coordinate: %v", parts)
	}
	lng, err1 := strconv.ParseFloat(parts[0], 64)
	lat, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		return Point{}, fmt.Errorf("failed to parse floats: %v", parts)
	}
	p := Point{Lng: lng, Lat: lat}
	switch len(parts) {
	case 3:
		if m, err := strconv.ParseFloat(parts[2], 64); err == nil {
			p.M = m
		}
	case 4:
		if m, err := strconv.ParseFloat(parts[3], 64); err == nil {
			p.M = m
		}
	}
	return p, nil
}

// FormatPointZM builds POINT ZM WKT matching Java org.postgis.Point(lng,lat,0)+setM(ts).
func FormatPointZM(lng, lat float64, m int64) string {
	lng32 := float32(lng)
	lat32 := float32(lat)
	return fmt.Sprintf("POINT ZM (%g %g 0 %d)", lng32, lat32, m)
}

func ParsePointWKT(wkt string) (Point, error) {
	wkt = strings.TrimSpace(wkt)
	upper := strings.ToUpper(wkt)
	if strings.HasPrefix(upper, "POINT") {
		start := strings.Index(wkt, "(")
		end := strings.LastIndex(wkt, ")")
		if start < 0 || end <= start {
			return Point{}, fmt.Errorf("invalid POINT wkt: %s", wkt)
		}
		content := strings.TrimSpace(wkt[start+1 : end])
		return parseCoords(strings.Fields(content))
	}
	return Point{}, fmt.Errorf("invalid POINT wkt: %s", wkt)
}

// ValidWGS84 reports whether lng/lat are in range and non-zero.
func ValidWGS84(lng, lat float64) bool {
	if lng == 0 && lat == 0 {
		return false
	}
	return lng >= -180 && lng <= 180 && lat >= -90 && lat <= 90
}

// FormatLineStringM builds a LINESTRING WKT with M measure (epoch seconds) per vertex.
func FormatLineStringM(points []Point) (string, error) {
	if len(points) < 2 {
		return "", fmt.Errorf("linestring needs at least 2 points")
	}
	parts := make([]string, len(points))
	for i, p := range points {
		if !ValidWGS84(p.Lng, p.Lat) {
			return "", fmt.Errorf("invalid coordinate at index %d", i)
		}
		parts[i] = fmt.Sprintf("%g %g %d", p.Lng, p.Lat, int64(p.M))
	}
	return "LINESTRING(" + strings.Join(parts, ", ") + ")", nil
}

// FormatLineStringMUnfiltered builds LINESTRING WKT without coordinate validation,
// matching Java ItineraryDTO.addPoint for all GPS vertices.
func FormatLineStringMUnfiltered(points []Point) (string, error) {
	if len(points) < 2 {
		return "", fmt.Errorf("linestring needs at least 2 points")
	}
	parts := make([]string, len(points))
	for i, p := range points {
		parts[i] = fmt.Sprintf("%g %g %d", p.Lng, p.Lat, int64(p.M))
	}
	return "LINESTRING(" + strings.Join(parts, ", ") + ")", nil
}

// ParseLineStringWKT parses LINESTRING WKT with optional M coordinates per point.
func ParseLineStringWKT(wkt string) ([]Point, error) {
	wkt = strings.TrimSpace(wkt)
	upper := strings.ToUpper(wkt)
	if !strings.HasPrefix(upper, "LINESTRING") {
		return nil, fmt.Errorf("invalid LINESTRING wkt: %s", wkt)
	}
	start := strings.Index(wkt, "(")
	end := strings.LastIndex(wkt, ")")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("invalid LINESTRING wkt: %s", wkt)
	}
	content := strings.TrimSpace(wkt[start+1 : end])
	if content == "" {
		return []Point{}, nil
	}

	pairs := strings.Split(content, ",")
	points := make([]Point, 0, len(pairs))
	for _, pairStr := range pairs {
		p, err := parseCoords(strings.Fields(strings.TrimSpace(pairStr)))
		if err != nil {
			continue
		}
		points = append(points, p)
	}
	return points, nil
}

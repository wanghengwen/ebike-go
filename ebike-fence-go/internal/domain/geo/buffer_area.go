package geo

import (
	"encoding/json"
	"errors"
	"math"

	"github.com/peterstace/simplefeatures/geom"
)

// javaBufferQuadSegments mirrors Java BufferOp.bufferOp(..., 2, CAP_ROUND).
const javaBufferQuadSegments = 2

// Meter2DegreeJava mirrors Java GeoUtils.meter2Degree, including the historical
// (EARTH_RADIUS - π) bug and BigDecimal scale-5 HALF_UP rounding.
func Meter2DegreeJava(meter float64) float64 {
	raw := (meter * 180) / (EarthRadius - math.Pi)
	return roundHalfUp(raw, 5)
}

func roundHalfUp(v float64, scale int) float64 {
	pow := math.Pow(10, float64(scale))
	return math.Floor(v*pow+0.5) / pow
}

// CreatePolygonBufferArea mirrors Java GeoUtils.createPolygonBufferArea:
// distance = bufferDistance * (1 - coefficientOfDifficult), then JTS-like buffer
// in degree space with quadrantSegments=2. Returns JSON "[[lng,lat],...]" without
// the closing duplicate vertex.
func CreatePolygonBufferArea(pointStr string, bufferDistance, coefficientOfDifficult float64) (string, error) {
	distance := bufferDistance * (1 - coefficientOfDifficult)
	return createPolygonBufferAreaMeters(pointStr, distance)
}

func createPolygonBufferAreaMeters(pointStr string, distanceMeters float64) (string, error) {
	poly, err := ParsePolygon(pointStr)
	if err != nil {
		return "", err
	}
	area, err := polygonFromLocations(poly)
	if err != nil {
		return "", err
	}
	degreeDist := Meter2DegreeJava(distanceMeters)
	buffered, err := geom.Buffer(area.AsGeometry(), degreeDist, geom.BufferQuadSegments(javaBufferQuadSegments))
	if err != nil {
		return "", err
	}
	if buffered.IsEmpty() {
		return "", errors.New("empty buffer geometry")
	}

	ring, ok := exteriorRingSeq(buffered)
	if !ok || ring.Length() < 4 {
		return "", errors.New("buffer geometry has no exterior ring")
	}
	// JTS getCoordinates includes closing vertex; Java drops the last one.
	n := ring.Length() - 1
	pts := make([][2]float64, 0, n)
	for i := 0; i < n; i++ {
		xy := ring.GetXY(i)
		pts = append(pts, [2]float64{xy.X, xy.Y})
	}
	b, err := json.Marshal(pts)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func exteriorRingSeq(g geom.Geometry) (geom.Sequence, bool) {
	if g.IsPolygon() {
		return g.MustAsPolygon().ExteriorRing().Coordinates(), true
	}
	if g.IsMultiPolygon() {
		mp := g.MustAsMultiPolygon()
		if mp.NumPolygons() == 0 {
			return geom.Sequence{}, false
		}
		return mp.PolygonN(0).ExteriorRing().Coordinates(), true
	}
	return geom.Sequence{}, false
}

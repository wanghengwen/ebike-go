package geo

import "math"

// Coordinate transforms ported 1:1 from GEOJsonUtils (WGS-84 <-> GCJ-02). Used
// by the BLE c34 reports to fill both raw (wgs84) and Mars (gcj02) coordinates.

const (
	coordPi = 3.1415926535897932384626
	coordA  = 6378245.0
	coordEE = 0.00669342162296594323
)

// Wgs84ToGcj02 converts a WGS-84 (lat, lon) to GCJ-02, returning (lat, lon).
// Points outside China are returned unchanged (matches GEOJsonUtils).
func Wgs84ToGcj02(lat, lon float64) (float64, float64) {
	if outOfChina(lat, lon) {
		return lat, lon
	}
	dLat, dLon := deltaLatLon(lat, lon)
	return lat + dLat, lon + dLon
}

// Gcj02ToWgs84 reverses Wgs84ToGcj02 using the same one-step approximation as
// GEOJsonUtils.gcj02_To_Wgs84, returning (lat, lon).
func Gcj02ToWgs84(lat, lon float64) (float64, float64) {
	mLat, mLon := transform(lat, lon)
	return lat*2 - mLat, lon*2 - mLon
}

// transform mirrors GEOJsonUtils.transform: the GCJ-02 image of a WGS-84 point.
func transform(lat, lon float64) (float64, float64) {
	if outOfChina(lat, lon) {
		return lat, lon
	}
	dLat, dLon := deltaLatLon(lat, lon)
	return lat + dLat, lon + dLon
}

func deltaLatLon(lat, lon float64) (float64, float64) {
	dLat := transformLat(lon-105.0, lat-35.0)
	dLon := transformLon(lon-105.0, lat-35.0)
	radLat := lat / 180.0 * coordPi
	magic := math.Sin(radLat)
	magic = 1 - coordEE*magic*magic
	sqrtMagic := math.Sqrt(magic)
	dLat = (dLat * 180.0) / ((coordA * (1 - coordEE)) / (magic * sqrtMagic) * coordPi)
	dLon = (dLon * 180.0) / (coordA / sqrtMagic * math.Cos(radLat) * coordPi)
	return dLat, dLon
}

func outOfChina(lat, lon float64) bool {
	if lon < 72.004 || lon > 137.8347 {
		return true
	}
	if lat < 0.8293 || lat > 55.8271 {
		return true
	}
	return false
}

func transformLat(x, y float64) float64 {
	ret := -100.0 + 2.0*x + 3.0*y + 0.2*y*y + 0.1*x*y + 0.2*math.Sqrt(math.Abs(x))
	ret += (20.0*math.Sin(6.0*x*coordPi) + 20.0*math.Sin(2.0*x*coordPi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(y*coordPi) + 40.0*math.Sin(y/3.0*coordPi)) * 2.0 / 3.0
	ret += (160.0*math.Sin(y/12.0*coordPi) + 320*math.Sin(y*coordPi/30.0)) * 2.0 / 3.0
	return ret
}

func transformLon(x, y float64) float64 {
	ret := 300.0 + x + 2.0*y + 0.1*x*x + 0.1*x*y + 0.1*math.Sqrt(math.Abs(x))
	ret += (20.0*math.Sin(6.0*x*coordPi) + 20.0*math.Sin(2.0*x*coordPi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(x*coordPi) + 40.0*math.Sin(x/3.0*coordPi)) * 2.0 / 3.0
	ret += (150.0*math.Sin(x/12.0*coordPi) + 300.0*math.Sin(x/30.0*coordPi)) * 2.0 / 3.0
	return ret
}

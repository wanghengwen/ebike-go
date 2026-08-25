package utils

import (
	"math"
)

const (
	xPI = 3.14159265358979324 * 3000.0 / 180.0
	pi  = 3.1415926535897932384626
	a   = 6378245.0
	ee  = 0.00669342162296594323
)

// TransformBD09ToGCJ02 百度坐标（BD09）转 GCJ02
func TransformBD09ToGCJ02(lng, lat float64) (float64, float64) {
	x := lng - 0.0065
	y := lat - 0.006
	z := math.Sqrt(x*x+y*y) - 0.00002*math.Sin(y*xPI)
	theta := math.Atan2(y, x) - 0.000003*math.Cos(x*xPI)
	gcjLng := z * math.Cos(theta)
	gcjLat := z * math.Sin(theta)
	return gcjLng, gcjLat
}

// TransformGCJ02ToBD09 GCJ02 转百度坐标
func TransformGCJ02ToBD09(lng, lat float64) (float64, float64) {
	z := math.Sqrt(lng*lng+lat*lat) + 0.00002*math.Sin(lat*xPI)
	theta := math.Atan2(lat, lng) + 0.000003*math.Cos(lng*xPI)
	bdLng := z*math.Cos(theta) + 0.0065
	bdLat := z*math.Sin(theta) + 0.006
	return bdLng, bdLat
}

// TransformGCJ02ToWGS84 GCJ02 转 WGS84
func TransformGCJ02ToWGS84(lng, lat float64) (float64, float64) {
	if outOfChina(lng, lat) {
		return lng, lat
	}
	dLat := transformLat(lng-105.0, lat-35.0)
	dLng := transformLng(lng-105.0, lat-35.0)
	radLat := lat / 180.0 * pi
	magic := math.Sin(radLat)
	magic = 1 - ee*magic*magic
	sqrtMagic := math.Sqrt(magic)
	dLat = (dLat * 180.0) / ((a * (1 - ee)) / (magic * sqrtMagic) * pi)
	dLng = (dLng * 180.0) / (a / sqrtMagic * math.Cos(radLat) * pi)
	mgLat := lat + dLat
	mgLng := lng + dLng
	return lng*2 - mgLng, lat*2 - mgLat
}

// TransformWGS84ToGCJ02 WGS84 坐标 转 GCJ02
func TransformWGS84ToGCJ02(lng, lat float64) (float64, float64) {
	if outOfChina(lng, lat) {
		return lng, lat
	}
	dLat := transformLat(lng-105.0, lat-35.0)
	dLng := transformLng(lng-105.0, lat-35.0)
	redLat := lat / 180.0 * pi
	magic := math.Sin(redLat)
	magic = 1 - ee*magic*magic
	sqrtMagic := math.Sqrt(magic)
	dLat = (dLat * 180.0) / ((a * (1 - ee)) / (magic * sqrtMagic) * pi)
	dLng = (dLng * 180.0) / (a / sqrtMagic * math.Cos(redLat) * pi)
	mgLat := lat + dLat
	mgLng := lng + dLng
	return mgLng, mgLat
}

func transformLat(lng, lat float64) float64 {
	ret := -100.0 + 2.0*lng + 3.0*lat + 0.2*lat*lat + 0.1*lng*lat + 0.2*math.Sqrt(math.Abs(lng))
	ret += (20.0*math.Sin(6.0*lng*pi) + 20.0*math.Sin(2.0*lng*pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lat*pi) + 40.0*math.Sin(lat/3.0*pi)) * 2.0 / 3.0
	ret += (160.0*math.Sin(lat/12.0*pi) + 320*math.Sin(lat*pi/30.0)) * 2.0 / 3.0
	return ret
}

func transformLng(lng, lat float64) float64 {
	ret := 300.0 + lng + 2.0*lat + 0.1*lng*lng + 0.1*lng*lat + 0.1*math.Sqrt(math.Abs(lng))
	ret += (20.0*math.Sin(6.0*lng*pi) + 20.0*math.Sin(2.0*lng*pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lng*pi) + 40.0*math.Sin(lng/3.0*pi)) * 2.0 / 3.0
	ret += (150.0*math.Sin(lng/12.0*pi) + 300.0*math.Sin(lng/30.0*pi)) * 2.0 / 3.0
	return ret
}

func outOfChina(lng, lat float64) bool {
	return (lng < 72.004 || lng > 137.8347) || (lat < 0.8293 || lat > 55.8271)
}

const (
	wgs84A = 6378137.0
	wgs84F = 1 / 298.257223563
	wgs84B = wgs84A * (1 - wgs84F)
	// sphereRadius mirrors org.gavaghan.geodesy.Ellipsoid.Sphere semi-major axis.
	sphereRadius = 6371000.0
)

// GetDistance mirrors Java GeometryUtil.getDistance with the given ellipsoid type.
// coordType 0 -> Vincenty WGS84; 1 -> Sphere (gavaghan Ellipsoid.Sphere).
func GetDistance(lng1, lat1, lng2, lat2 float64, coordType int) float64 {
	if coordType == 0 {
		return vincentyDistance(lat1, lng1, lat2, lng2, wgs84A, wgs84B)
	}
	return sphereDistance(lat1, lng1, lat2, lng2, sphereRadius)
}

func sphereDistance(lat1, lng1, lat2, lng2, radius float64) float64 {
	radLat1 := lat1 * pi / 180.0
	radLat2 := lat2 * pi / 180.0
	dLat := radLat1 - radLat2
	dLng := (lng1 - lng2) * pi / 180.0
	s := 2 * math.Asin(math.Sqrt(math.Pow(math.Sin(dLat/2), 2)+
		math.Cos(radLat1)*math.Cos(radLat2)*math.Pow(math.Sin(dLng/2), 2)))
	return s * radius
}

// vincentyDistance implements the inverse Vincenty formula (WGS84).
func vincentyDistance(lat1, lng1, lat2, lng2, a, b float64) float64 {
	const maxIter = 200
	const tol = 1e-12

	if lat1 == lat2 && lng1 == lng2 {
		return 0
	}

	phi1 := lat1 * pi / 180.0
	phi2 := lat2 * pi / 180.0
	lambda := (lng2 - lng1) * pi / 180.0

	u1 := math.Atan((1 - wgs84F) * math.Tan(phi1))
	u2 := math.Atan((1 - wgs84F) * math.Tan(phi2))
	sinU1, cosU1 := math.Sincos(u1)
	sinU2, cosU2 := math.Sincos(u2)

	lambdaP := lambda
	var sinSigma, cosSigma, sigma, sinAlpha, cosSqAlpha, cos2SigmaM float64

	for i := 0; i < maxIter; i++ {
		sinLambda, cosLambda := math.Sincos(lambdaP)
		sinSigma = math.Sqrt((cosU2*sinLambda)*(cosU2*sinLambda) +
			(cosU1*sinU2-sinU1*cosU2*cosLambda)*(cosU1*sinU2-sinU1*cosU2*cosLambda))
		if sinSigma == 0 {
			return 0
		}
		cosSigma = sinU1*sinU2 + cosU1*cosU2*cosLambda
		sigma = math.Atan2(sinSigma, cosSigma)
		sinAlpha = cosU1 * cosU2 * sinLambda / sinSigma
		cosSqAlpha = 1 - sinAlpha*sinAlpha
		if cosSqAlpha == 0 {
			cos2SigmaM = 0
		} else {
			cos2SigmaM = cosSigma - 2*sinU1*sinU2/cosSqAlpha
		}
		c := wgs84F / 16 * cosSqAlpha * (4 + wgs84F*(4-3*cosSqAlpha))
		lambdaPrev := lambdaP
		lambdaP = lambda + (1-c)*wgs84F*sinAlpha*
			(sigma+c*sinSigma*(cos2SigmaM+c*cosSigma*(-1+2*cos2SigmaM*cos2SigmaM)))
		if math.Abs(lambdaP-lambdaPrev) < tol {
			break
		}
	}

	uSq := cosSqAlpha * (a*a - b*b) / (b * b)
	aCoeff := 1 + uSq/16384*(4096+uSq*(-768+uSq*(320-175*uSq)))
	bCoeff := uSq / 1024 * (256 + uSq*(-128+uSq*(74-47*uSq)))
	deltaSigma := bCoeff * sinSigma * (cos2SigmaM + bCoeff/4*(cosSigma*(-1+2*cos2SigmaM*cos2SigmaM)-
		bCoeff/6*cos2SigmaM*(-3+4*sinSigma*sinSigma)*(-3+4*cos2SigmaM*cos2SigmaM)))
	return b * aCoeff * (sigma - deltaSigma)
}


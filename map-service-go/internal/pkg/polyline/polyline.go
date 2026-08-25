package polyline

import (
	"fmt"
	"strconv"
	"strings"

	"map-service-go/internal/model"
)

// UnCompressPolyline decompresses Tencent Map polyline
// Java equivalent:
//
//	for (var i = 2; i < pd.length ; i++) {
//	    pd[i] = pd[i-2] + pd[i]/1000000;
//	}
func UnCompressPolyline(polyline []float64) []float64 {
	pd := make([]float64, len(polyline))
	copy(pd, polyline)

	for i := 2; i < len(pd); i++ {
		pd[i] = pd[i-2] + pd[i]/1000000.0
	}
	return pd
}

// PolylineLocation converts polyline array to Location objects
func PolylineLocation(polyline []float64) []model.Location {
	locations := make([]model.Location, 0)
	for i := 0; i < len(polyline)/2; i++ {
		lat := polyline[i]
		lng := polyline[i+1]
		locations = append(locations, model.Location{
			Latitude:  strconv.FormatFloat(lat, 'f', -1, 64),
			Longitude: strconv.FormatFloat(lng, 'f', -1, 64),
		})
	}
	return locations
}

// PolylineString converts polyline array to string separated by semicolon
func PolylineString(polyline []float64) string {
	var builder strings.Builder
	for i := 0; i < len(polyline)/2; i++ {
		lat := polyline[i*2]
		lng := polyline[i*2+1]

		// The Java code builds: lng + "," + lat
		builder.WriteString(strconv.FormatFloat(lng, 'f', -1, 64))
		builder.WriteString(",")
		builder.WriteString(strconv.FormatFloat(lat, 'f', -1, 64))

		if i < len(polyline)/2-1 {
			builder.WriteString(";")
		}
	}
	return builder.String()
}

func ExchangeLngLat(origin string) string {
	// The coordinate utils in Java splits by comma and swaps
	if origin == "" {
		return origin
	}
	parts := strings.Split(origin, ",")
	if len(parts) == 2 {
		return fmt.Sprintf("%s,%s", parts[1], parts[0])
	}
	return origin
}

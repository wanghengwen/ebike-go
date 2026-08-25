package geohash

import (
	"strings"
)

const (
	MinLat = -90.0
	MaxLat = 90.0
	MinLng = -180.0
	MaxLng = 180.0
)

var chars = []rune{'0', '1', '2', '3', '4', '5', '6', '7',
	'8', '9', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'k', 'm', 'n',
	'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}

type GeoHash struct {
	lat        float64
	lng        float64
	hashLength int
	latLength  int
	lngLength  int
	minLatUnit float64
	minLngUnit float64
}

func NewGeoHash(lat, lng float64) *GeoHash {
	g := &GeoHash{
		lat:        lat,
		lng:        lng,
		hashLength: 8,
		latLength:  20,
		lngLength:  20,
	}
	g.setMinLatLng()
	return g
}

func (g *GeoHash) SetHashLength(length int) bool {
	if length < 1 {
		return false
	}
	g.hashLength = length
	g.latLength = (length * 5) / 2
	if length%2 == 0 {
		g.lngLength = g.latLength
	} else {
		g.lngLength = g.latLength + 1
	}
	g.setMinLatLng()
	return true
}

func (g *GeoHash) setMinLatLng() {
	g.minLatUnit = MaxLat - MinLat
	for i := 0; i < g.latLength; i++ {
		g.minLatUnit /= 2.0
	}
	g.minLngUnit = MaxLng - MinLng
	for i := 0; i < g.lngLength; i++ {
		g.minLngUnit /= 2.0
	}
}

func (g *GeoHash) GetGeoHashBase32() string {
	return g.getGeoHashBase32(g.lat, g.lng)
}

func (g *GeoHash) GetGeoHashBase32For9() []string {
	leftLat := g.lat - g.minLatUnit
	rightLat := g.lat + g.minLatUnit
	upLng := g.lng - g.minLngUnit
	downLng := g.lng + g.minLngUnit

	var base32For9 []string

	// Left 3
	if leftUp := g.getGeoHashBase32(leftLat, upLng); leftUp != "" {
		base32For9 = append(base32For9, leftUp)
	}
	if leftMid := g.getGeoHashBase32(leftLat, g.lng); leftMid != "" {
		base32For9 = append(base32For9, leftMid)
	}
	if leftDown := g.getGeoHashBase32(leftLat, downLng); leftDown != "" {
		base32For9 = append(base32For9, leftDown)
	}

	// Mid 3
	if midUp := g.getGeoHashBase32(g.lat, upLng); midUp != "" {
		base32For9 = append(base32For9, midUp)
	}
	if midMid := g.getGeoHashBase32(g.lat, g.lng); midMid != "" {
		base32For9 = append(base32For9, midMid)
	}
	if midDown := g.getGeoHashBase32(g.lat, downLng); midDown != "" {
		base32For9 = append(base32For9, midDown)
	}

	// Right 3
	if rightUp := g.getGeoHashBase32(rightLat, upLng); rightUp != "" {
		base32For9 = append(base32For9, rightUp)
	}
	if rightMid := g.getGeoHashBase32(rightLat, g.lng); rightMid != "" {
		base32For9 = append(base32For9, rightMid)
	}
	if rightDown := g.getGeoHashBase32(rightLat, downLng); rightDown != "" {
		base32For9 = append(base32For9, rightDown)
	}

	return base32For9
}

func (g *GeoHash) getGeoHashBase32(lat, lng float64) string {
	bools := g.getGeoBinary(lat, lng)
	if bools == nil {
		return ""
	}

	var sb strings.Builder
	for i := 0; i < len(bools); i += 5 {
		base32 := make([]bool, 5)
		for j := 0; j < 5; j++ {
			if i+j < len(bools) {
				base32[j] = bools[i+j]
			}
		}
		cha := getBase32Char(base32)
		if cha == ' ' {
			return ""
		}
		sb.WriteRune(cha)
	}
	return sb.String()
}

func getBase32Char(base32 []bool) rune {
	if len(base32) != 5 {
		return ' '
	}
	num := 0
	for _, b := range base32 {
		num <<= 1
		if b {
			num += 1
		}
	}
	return chars[num%len(chars)]
}

func (g *GeoHash) getGeoBinary(lat, lng float64) []bool {
	latArray := getHashArray(lat, MinLat, MaxLat, g.latLength)
	lngArray := getHashArray(lng, MinLng, MaxLng, g.lngLength)
	return merge(latArray, lngArray)
}

func merge(latArray, lngArray []bool) []bool {
	if latArray == nil || lngArray == nil {
		return nil
	}
	result := make([]bool, len(lngArray)+len(latArray))
	for i := 0; i < len(lngArray); i++ {
		result[2*i] = lngArray[i]
	}
	for i := 0; i < len(latArray); i++ {
		result[2*i+1] = latArray[i]
	}
	return result
}

func getHashArray(value, min, max float64, length int) []bool {
	if value < min || value > max {
		return nil
	}
	if length < 1 {
		return nil
	}
	result := make([]bool, length)
	for i := 0; i < length; i++ {
		mid := (min + max) / 2.0
		if value > mid {
			result[i] = true
			min = mid
		} else {
			result[i] = false
			max = mid
		}
	}
	return result
}

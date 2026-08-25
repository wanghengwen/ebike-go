package geohash

import (
	"fmt"
	"strconv"
)

var base32Lookup = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	"b", "c", "d", "e", "f", "g", "h", "j", "k",
	"m", "n", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
}

const (
	minLat = -90.0
	maxLat = 90.0
	minLng = -180.0
	maxLng = 180.0
)

// GetSpaceCoordinate decodes a geoHash string to [lat, lng].
func GetSpaceCoordinate(geoHashCode string) [2]float64 {
	list := base32Decode(geoHashCode)
	binary := convertToIndex(list)
	latList, lngList := splitLatAndLng(binary)
	lat := revert(minLat, maxLat, latList)
	lng := revert(minLng, maxLng, lngList)
	return [2]float64{lat, lng}
}

func base32Decode(s string) []int {
	var list []int
	for _, ch := range s {
		c := string(ch)
		for j, v := range base32Lookup {
			if v == c {
				list = append(list, j)
				break
			}
		}
	}
	return list
}

func convertToIndex(nums []int) string {
	var str string
	for _, num := range nums {
		sb := strconv.FormatInt(int64(num), 2)
		for len(sb) < 5 {
			sb = "0" + sb
		}
		str += sb
	}
	return str
}

func splitLatAndLng(s string) (latList, lngList []string) {
	for i, ch := range s {
		c := string(ch)
		if i%2 == 1 {
			latList = append(latList, c)
		} else {
			lngList = append(lngList, c)
		}
	}
	return
}

func revert(min, max float64, list []string) float64 {
	if len(list) == 0 {
		return (max + min) / 2.0
	}
	value := 0.0
	for _, flag := range list {
		mid := (max + min) / 2
		if flag == "0" {
			max = mid
		}
		if flag == "1" {
			min = mid
		}
		value = (max + min) / 2
	}
	f, _ := strconv.ParseFloat(fmt.Sprintf("%.6f", value), 64)
	return f
}

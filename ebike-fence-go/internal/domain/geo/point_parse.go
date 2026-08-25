package geo

import (
	"encoding/json"
	"errors"
	"strings"
)

// ParsePointJSON accepts GeoJSON Point or {"lng","lat"} object bytes.
func ParsePointJSON(data []byte) (Location, error) {
	if len(data) == 0 {
		return Location{}, errors.New("invalid point")
	}
	if data[0] == '"' {
		var inner string
		if err := json.Unmarshal(data, &inner); err != nil {
			return Location{}, err
		}
		return ParsePointJSON([]byte(inner))
	}

	var geoJSON struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	}
	if err := json.Unmarshal(data, &geoJSON); err == nil &&
		strings.EqualFold(geoJSON.Type, "Point") && len(geoJSON.Coordinates) >= 2 {
		return Location{Lng: geoJSON.Coordinates[0], Lat: geoJSON.Coordinates[1]}, nil
	}

	var loc Location
	if err := json.Unmarshal(data, &loc); err != nil {
		return Location{}, errors.New("invalid point")
	}
	if loc.Lng == 0 && loc.Lat == 0 {
		return Location{}, errors.New("invalid point")
	}
	return loc, nil
}

// ParsePointString accepts GeoJSON text or a JSON-encoded GeoJSON string.
func ParsePointString(s string) (Location, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Location{}, errors.New("invalid point")
	}
	return ParsePointJSON([]byte(s))
}

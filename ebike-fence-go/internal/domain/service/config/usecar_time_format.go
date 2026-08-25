package configsvc

import (
	"strings"

	"ebike-fence-go/internal/api/dto"
)

func normalizeTimeHHMMSS(s string) string {
	if s == "" {
		return s
	}
	if strings.Count(s, ":") == 1 {
		return s + ":00"
	}
	return s
}

func normalizeUseCarTimeFields(co *dto.ConfigUseCarCO) {
	if co == nil {
		return
	}
	if co.StopTimeStart != nil {
		s := normalizeTimeHHMMSS(*co.StopTimeStart)
		co.StopTimeStart = &s
	}
	if co.StopTimeEnd != nil {
		s := normalizeTimeHHMMSS(*co.StopTimeEnd)
		co.StopTimeEnd = &s
	}
}

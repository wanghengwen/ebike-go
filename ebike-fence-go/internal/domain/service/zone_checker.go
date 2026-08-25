package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"ebike-fence-go/internal/domain/gateway"
	"ebike-fence-go/internal/domain/geo"
)

// ZoneCheckOption holds parameters for finding a zone
type ZoneCheckOption struct {
	FencePrefix          string
	GeoPrefix            string
	Radius               float64
	Count                int
	UseBuffer            bool
	GlobalBufferDistance float64 // Fallback 2
	FilterByService      bool
	ServiceAreaId        int64
}

// FindZoneByLocation is a generic function to find the matching zone (FenceE) based on location and options
func FindZoneByLocation(ctx context.Context, tenantId string, loc geo.Location, opt ZoneCheckOption) (*gateway.FenceE, error) {
	fences, err := gateway.GetFencesByLocation(ctx, opt.FencePrefix, opt.GeoPrefix+"_"+tenantId, tenantId, loc.Lng, loc.Lat, opt.Radius, opt.Count, 0)
	if err != nil {
		log.Printf("Error querying %s fences from Redis: %v", opt.FencePrefix, err)
		return nil, err
	}

	for _, fence := range fences {
		// Only check IzEnable and opening hours for non-service area fences.
		if opt.FencePrefix != ServiceAreaPrefix {
			if !fence.IzEnable {
				continue
			}
			if !isInOpeningHours(&fence) {
				continue
			}
		}

		// Cross-service area check
		if opt.FilterByService && fence.ServiceId != opt.ServiceAreaId {
			continue
		}

		var inPolygon bool
		if opt.UseBuffer {
			bufferDistance := fence.BufferDistance
			if bufferDistance <= 0 {
				bufferDistance = opt.GlobalBufferDistance // Fallback 2
			}
			if bufferDistance <= 0 {
				bufferDistance = 5.0 // Fallback 3
			}
			inPolygon = geo.IsPointInParsedPolygonWithJavaBuffer(loc, fence.ParsedPolygon, bufferDistance)
		} else {
			inPolygon = geo.IsPointInParsedPolygon(loc, fence.ParsedPolygon)
		}

		if inPolygon {
			return &fence, nil
		}
	}

	return nil, nil
}

// isInOpeningHours checks if the current time falls within the fence's operating hours.
func isInOpeningHours(fence *gateway.FenceE) bool {
	// Java getIzOpenAllDayNotNull(): defaults to true if null
	izOpenAllDay := true
	if fence.IzOpenAllDay != nil {
		izOpenAllDay = *fence.IzOpenAllDay
	}
	if izOpenAllDay {
		return true // If open all day is enabled, ignore times
	}

	if fence.OpeningHoursBegin == "" || fence.OpeningHoursEnd == "" {
		return true // If times are completely missing but izOpenAllDay was false, fallback to true to prevent blocking
	}

	now := time.Now()
	// Parse formats like "HH:MM", "HH:MM:SS"
	// time.Parse needs a dummy date for just time comparison.
	layout := "15:04:05"
	startStr := fence.OpeningHoursBegin
	if len(startStr) == 5 {
		startStr += ":00"
	}
	endStr := fence.OpeningHoursEnd
	if len(endStr) == 5 {
		endStr += ":00"
	}

	startTime, err := time.Parse(layout, startStr)
	if err != nil {
		return true // fallback
	}
	endTime, err := time.Parse(layout, endStr)
	if err != nil {
		return true // fallback
	}

	// Compare just the time component
	nowHour, nowMin, nowSec := now.Clock()
	nowTimeStr := fmt.Sprintf("%02d:%02d:%02d", nowHour, nowMin, nowSec)
	nowTime, _ := time.Parse(layout, nowTimeStr)

	return nowTime.After(startTime) && nowTime.Before(endTime)
}

package utils

import (
	"time"
)

// ClampedAddMonths adds (or subtracts) months to/from a time, matching Java Calendar.add(Calendar.MONTH, offset) clamping behavior.
// E.g., August 31 minus 6 months clamps to February 28 (or 29 in a leap year).
func ClampedAddMonths(t time.Time, months int) time.Time {
	y, m, d := t.Date()

	// Compute targeted year and month
	targetMonth := int(m) + months
	targetYear := y

	for targetMonth <= 0 {
		targetMonth += 12
		targetYear--
	}
	for targetMonth > 12 {
		targetMonth -= 12
		targetYear++
	}

	// Determine the last day of the target month
	daysInMonth := daysIn(time.Month(targetMonth), targetYear)

	// Clamp the day if it exceeds the last day of the target month
	if d > daysInMonth {
		d = daysInMonth
	}

	return time.Date(targetYear, time.Month(targetMonth), d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// FormatJavaGpsQueryTime formats epoch millis like Java SimpleDateFormat("yyyy-MM-dd HH:mm:ss") in local timezone.
func FormatJavaGpsQueryTime(epochMilli int64) string {
	return time.UnixMilli(epochMilli).Format("2006-01-02 15:04:05")
}

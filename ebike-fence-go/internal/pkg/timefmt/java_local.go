package timefmt

import "time"

var shanghai *time.Location

func init() {
	shanghai, _ = time.LoadLocation("Asia/Shanghai")
	if shanghai == nil {
		shanghai = time.FixedZone("CST", 8*3600)
	}
}

// FormatJavaLocal formats time like Java LocalDateTime JSON (no zone suffix).
func FormatJavaLocal(t time.Time) string {
	return t.In(shanghai).Format("2006-01-02T15:04:05")
}

// FormatJavaLocalSpace uses space separator (help config list endpoints).
func FormatJavaLocalSpace(t time.Time) string {
	return t.In(shanghai).Format("2006-01-02 15:04:05")
}

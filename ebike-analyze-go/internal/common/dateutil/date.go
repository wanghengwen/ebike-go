package dateutil

import (
	"time"
)

const DateTimeFmt = "2006-01-02 15:04:05"
const DateFmt = "2006-01-02"
const YyyyMmDd = "20060102"

func TimestampFormat(ts int64) string {
	if ts < 0 {
		return ""
	}
	return time.UnixMilli(ts).In(time.Local).Format(DateTimeFmt)
}

func LocalDateTimeFormat(t time.Time) string {
	return t.In(time.Local).Format(DateTimeFmt)
}

func NowFormatted() string {
	return LocalDateTimeFormat(time.Now())
}

func ParseDateTime(s string) (time.Time, error) {
	return time.ParseInLocation(DateTimeFmt, s, time.Local)
}

func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation(DateFmt, s, time.Local)
}

func FormatYyyyMmDd(t time.Time) string {
	return t.Format(YyyyMmDd)
}

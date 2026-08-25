package javacompat

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DateTimeLayout is the Java global LocalDateTimeSerializer pattern.
const DateTimeLayout = "2006-01-02 15:04:05"

// LongStr mirrors a Java Long under the global ToStringSerializer.
type LongStr int64

func (l *LongStr) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "null" {
		return nil
	}
	s = strings.Trim(s, `"`)
	if s == "" || s == "null" {
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid long value %q", s)
	}
	*l = LongStr(v)
	return nil
}

func (l LongStr) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(strconv.FormatInt(int64(l), 10))), nil
}

// DateTime mirrors a Java LocalDateTime under the global serializer.
type DateTime struct {
	time.Time
}

func (d *DateTime) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "null" || s == "" {
		return nil
	}
	if strings.HasPrefix(s, "[") {
		var parts []int
		if err := json.Unmarshal(b, &parts); err != nil {
			return err
		}
		if len(parts) < 3 {
			return fmt.Errorf("invalid datetime array %s", s)
		}
		get := func(i int) int {
			if i < len(parts) {
				return parts[i]
			}
			return 0
		}
		d.Time = time.Date(parts[0], time.Month(parts[1]), parts[2], get(3), get(4), get(5), get(6), time.Local)
		return nil
	}
	s = strings.Trim(s, `"`)
	for _, layout := range []string{
		DateTimeLayout,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02",
	} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			d.Time = t
			return nil
		}
	}
	return fmt.Errorf("cannot parse datetime %q", s)
}

func (d DateTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(d.Time.Format(DateTimeLayout))), nil
}

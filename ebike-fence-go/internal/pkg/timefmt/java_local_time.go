package timefmt

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// JavaLocalTime stores a local datetime compatible with Java LocalDateTime JSON
// (no timezone suffix) and MySQL datetime columns.
type JavaLocalTime struct {
	Time time.Time
}

func JavaLocalNow() JavaLocalTime {
	return JavaLocalTime{Time: time.Now()}
}

func FromTime(t time.Time) JavaLocalTime {
	return JavaLocalTime{Time: t}
}

func (t JavaLocalTime) IsZero() bool {
	return t.Time.IsZero()
}

func (t JavaLocalTime) After(u JavaLocalTime) bool {
	return t.Time.After(u.Time)
}

func (t *JavaLocalTime) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" {
		t.Time = time.Time{}
		return nil
	}
	s = strings.Trim(s, `"`)
	parsed, err := ParseJavaLocalString(s)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

func (t JavaLocalTime) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(FormatJavaLocal(t.Time))
}

func (t *JavaLocalTime) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		t.Time = v
		return nil
	case []byte:
		parsed, err := ParseJavaLocalString(string(v))
		if err != nil {
			return err
		}
		t.Time = parsed
		return nil
	case string:
		parsed, err := ParseJavaLocalString(v)
		if err != nil {
			return err
		}
		t.Time = parsed
		return nil
	default:
		return fmt.Errorf("timefmt.JavaLocalTime: unsupported Scan type %T", value)
	}
}

func (t JavaLocalTime) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}

// ParseJavaLocalString parses Java LocalDateTime strings from Redis/JSON/MySQL.
func ParseJavaLocalString(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, nil
	}
	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, s, shanghai); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported java local time %q", s)
}

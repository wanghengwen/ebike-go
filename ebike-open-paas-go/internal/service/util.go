package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// imeiLength is the length every Xiaoan endpoint documents ("设备的15位唯一标识")
// and the length of every IMEI in the device shadow.
const imeiLength = 15

// NormalizeImei coerces Xiaoan's number-or-string imei into a digit string.
//
// The check is exact rather than a length range: the imei is interpolated into
// Redis key names and forwarded to four upstream services, so anything that is
// not 15 digits has to be rejected here instead of becoming a lookup for a key
// that cannot exist. A caller sending a wrong length gets IMEI_ILLEGAL, which is
// the answer the spec defines for it.
func NormalizeImei(v interface{}) (string, error) {
	var s string
	switch t := v.(type) {
	case nil:
		return "", fmt.Errorf("imei missing")
	case string:
		s = strings.TrimSpace(t)
	case float64:
		s = strconv.FormatInt(int64(t), 10)
	case int:
		s = strconv.Itoa(t)
	case int64:
		s = strconv.FormatInt(t, 10)
	case json.Number:
		s = t.String()
	default:
		s = strings.TrimSpace(fmt.Sprint(t))
	}
	if s == "" {
		return "", fmt.Errorf("imei missing")
	}
	if len(s) != imeiLength {
		return "", fmt.Errorf("imei illegal: expected %d digits, got %d", imeiLength, len(s))
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("imei illegal: must be digits only")
		}
	}
	return s, nil
}

// AsInt extracts an int from a loosely-typed JSON value.
func AsInt(v interface{}, def int) int {
	switch t := v.(type) {
	case nil:
		return def
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	case json.Number:
		n, err := t.Int64()
		if err != nil {
			return def
		}
		return int(n)
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(t))
		if err != nil {
			return def
		}
		return n
	default:
		return def
	}
}

// AsInt64 extracts an int64 from a loosely-typed JSON value.
func AsInt64(v interface{}, def int64) int64 {
	switch t := v.(type) {
	case nil:
		return def
	case float64:
		return int64(t)
	case int:
		return int64(t)
	case int64:
		return t
	case json.Number:
		n, err := t.Int64()
		if err != nil {
			return def
		}
		return n
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		if err != nil {
			return def
		}
		return n
	default:
		return def
	}
}

// OptInt returns a pointer to raw[key] as an int, or nil when the key is absent
// or JSON null.
//
// The upstream shadow serializes every field, so a value the device has not
// reported arrives as null rather than being missing. Reading that through
// AsInt(v, default) is what turns "unknown" into a confident zero or one, which
// is the single most common way these mappings mislead a caller.
func OptInt(raw map[string]interface{}, key string) *int {
	v, present := raw[key]
	if !present || v == nil {
		return nil
	}
	n := AsInt(v, 0)
	return &n
}

// OptInt64 is OptInt for 64-bit values.
func OptInt64(raw map[string]interface{}, key string) *int64 {
	v, present := raw[key]
	if !present || v == nil {
		return nil
	}
	n := AsInt64(v, 0)
	return &n
}

// OptFloat is OptInt for floating-point values.
func OptFloat(raw map[string]interface{}, key string) *float64 {
	v, present := raw[key]
	if !present || v == nil {
		return nil
	}
	f := AsFloat(v)
	return &f
}

// OptString returns raw[key] as a string, or "" when absent or null.
func OptString(raw map[string]interface{}, key string) string {
	v, present := raw[key]
	if !present || v == nil {
		return ""
	}
	return AsString(v)
}

// AsString coerces a JSON value to string.
func AsString(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case json.Number:
		return t.String()
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

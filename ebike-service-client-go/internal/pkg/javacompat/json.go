// Package javacompat provides helpers shared across modules that port Java
// gateway/service behavior (null checks, Long-as-string, RPC result handling).
package javacompat

import (
	"encoding/json"
	"strconv"
	"strings"
)

// IsNullJSON reports whether a JSON value is absent or JSON null.
func IsNullJSON(raw json.RawMessage) bool {
	s := strings.TrimSpace(string(raw))
	return s == "" || s == "null"
}

// LongStrPtr converts a downstream Long (JSON number or string) into the
// string form used in Java responses (global ToStringSerializer). nil for null.
func LongStrPtr(raw json.RawMessage) *string {
	if IsNullJSON(raw) {
		return nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return &s
	}
	v := strings.TrimSpace(string(raw))
	return &v
}

// StringValueOf mirrors Java String.valueOf(Long): null becomes "null".
func StringValueOf(raw json.RawMessage) string {
	if p := LongStrPtr(raw); p != nil {
		return *p
	}
	return "null"
}

// RawInt extracts an integer from JSON that may be a number or stringified number.
func RawInt(raw json.RawMessage) (int64, bool) {
	if IsNullJSON(raw) {
		return 0, false
	}
	var n int64
	if err := json.Unmarshal(raw, &n); err == nil {
		return n, true
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if v, err2 := strconv.ParseInt(s, 10, 64); err2 == nil {
			return v, true
		}
	}
	return 0, false
}

// RawInt64Ptr is RawInt returning a pointer (nil for null).
func RawInt64Ptr(raw json.RawMessage) *int64 {
	if v, ok := RawInt(raw); ok {
		return &v
	}
	return nil
}

// RawStringPtr extracts a JSON string value (nil for null/non-string).
func RawStringPtr(raw json.RawMessage) *string {
	if IsNullJSON(raw) {
		return nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return &s
	}
	return nil
}

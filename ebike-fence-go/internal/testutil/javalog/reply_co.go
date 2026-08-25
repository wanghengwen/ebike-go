package javalog

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	coFloatField  = regexp.MustCompile(`([A-Za-z0-9_]+)=(-?\d+(?:\.\d+)?)`)
	coIntField    = regexp.MustCompile(`([A-Za-z0-9_]+)=(\d+)(?:,|\))`)
	coNullField   = regexp.MustCompile(`([A-Za-z0-9_]+)=null`)
	coBoolField   = regexp.MustCompile(`([A-Za-z0-9_]+)=(true|false)`)
	coStringField = regexp.MustCompile(`([A-Za-z0-9_]+)=([^,()]+)`)
)

// ParseJavaCO extracts fields from Java toString like ComputeDistanceCO(distance=1618.0, izCloseLine=false).
func ParseJavaCO(rep, typeName string) map[string]any {
	prefix := typeName + "("
	idx := strings.Index(rep, prefix)
	if idx < 0 {
		return nil
	}
	start := idx + len(prefix)
	end := strings.Index(rep[start:], ")")
	if end < 0 {
		return nil
	}
	body := rep[start : start+end]
	out := map[string]any{}
	for _, m := range coNullField.FindAllStringSubmatch(body, -1) {
		out[m[1]] = nil
	}
	for _, m := range coBoolField.FindAllStringSubmatch(body, -1) {
		out[m[1]] = m[2] == "true"
	}
	for _, m := range coIntField.FindAllStringSubmatch(body, -1) {
		if _, exists := out[m[1]]; exists {
			continue
		}
		out[m[1]] = m[2]
	}
	for _, m := range coFloatField.FindAllStringSubmatch(body, -1) {
		if _, exists := out[m[1]]; exists {
			continue
		}
		if !strings.Contains(m[2], ".") {
			continue
		}
		if f, err := strconv.ParseFloat(m[2], 64); err == nil {
			out[m[1]] = f
		}
	}
	return out
}

// ParseJavaCOInt64Field reads a snowflake-style id from Java toString CO output.
func ParseJavaCOInt64Field(rep, typeName, field string) (int64, bool) {
	co := ParseJavaCO(rep, typeName)
	if co == nil {
		return 0, false
	}
	raw, ok := co[field]
	if !ok || raw == nil {
		return 0, false
	}
	switch v := raw.(type) {
	case string:
		n, err := strconv.ParseInt(v, 10, 64)
		return n, err == nil
	case float64:
		return int64(v), true
	default:
		return 0, false
	}
}

// DataTypeName returns the Java CO class name embedded in rep data, e.g. ComputeDistanceCO.
func DataTypeName(rep string) string {
	const marker = `"data":`
	idx := strings.Index(rep, marker)
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(rep[idx+len(marker):])
	if rest == "null" {
		return ""
	}
	paren := strings.Index(rest, "(")
	if paren <= 0 {
		return ""
	}
	return rest[:paren]
}

package javacompat

import (
	"encoding/json"
	"strconv"
)

// ParseJSONArrayField mirrors Java setters that parse a JSON string into an array
// (e.g. CBillingConfigCmd.setLadderItem, BSpecialTipsCO.setBgUrl).
func ParseJSONArrayField(raw json.RawMessage) json.RawMessage {
	if IsNullJSON(raw) {
		return nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if s == "" {
			return nil
		}
		var v interface{}
		if json.Unmarshal([]byte(s), &v) == nil {
			b, _ := json.Marshal(v)
			return b
		}
		return raw
	}
	return raw
}

// CoalesceJSONArray mirrors Java setters that default null/blank JSON strings to
// an empty array (e.g. AdConfigDTO.setIzOn).
func CoalesceJSONArray(raw json.RawMessage) json.RawMessage {
	if IsNullJSON(raw) {
		return mustMarshalJSON([]interface{}{})
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil && (s == "" || s == "null") {
		return mustMarshalJSON([]interface{}{})
	}
	parsed := ParseJSONArrayField(raw)
	if IsNullJSON(parsed) {
		return mustMarshalJSON([]interface{}{})
	}
	return parsed
}

// ParseStringArrayField parses a field that may be a native JSON string array or
// a JSON string containing a serialized string array (e.g. CInvoiceCO.setOrderIds).
func ParseStringArrayField(raw json.RawMessage) []string {
	parsed := ParseJSONArrayField(raw)
	if IsNullJSON(parsed) {
		if !IsNullJSON(raw) {
			var s string
			if json.Unmarshal(raw, &s) == nil && (s == "" || s == "null") {
				return []string{}
			}
		}
		if IsNullJSON(raw) {
			return []string{}
		}
		return []string{}
	}
	var arr []string
	if json.Unmarshal(parsed, &arr) != nil {
		return []string{}
	}
	return arr
}

func mustMarshalJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

// NormalizeJSONArrayLongFields converts named Long fields from JSON numbers to
// strings inside a JSON array of objects (Java ToStringSerializer on responses).
func NormalizeJSONArrayLongFields(raw json.RawMessage, fields ...string) json.RawMessage {
	if IsNullJSON(raw) {
		return raw
	}
	var items []map[string]interface{}
	if json.Unmarshal(raw, &items) != nil {
		return raw
	}
	fieldSet := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		fieldSet[f] = struct{}{}
	}
	for _, it := range items {
		for f := range fieldSet {
			if v, ok := it[f]; ok {
				if s := longFieldToString(v); s != nil {
					it[f] = *s
				}
			}
		}
	}
	b, _ := json.Marshal(items)
	return b
}

// NormalizeJSONObjectLongFields converts named Long fields inside one object.
func NormalizeJSONObjectLongFields(raw json.RawMessage, fields ...string) json.RawMessage {
	if IsNullJSON(raw) {
		return raw
	}
	var obj map[string]interface{}
	if json.Unmarshal(raw, &obj) != nil {
		return raw
	}
	for _, f := range fields {
		if v, ok := obj[f]; ok {
			if s := longFieldToString(v); s != nil {
				obj[f] = *s
			}
		}
	}
	b, _ := json.Marshal(obj)
	return b
}

func longFieldToString(v interface{}) *string {
	switch t := v.(type) {
	case string:
		return &t
	case float64:
		s := strconv.FormatInt(int64(t), 10)
		return &s
	case json.Number:
		s := t.String()
		return &s
	default:
		return nil
	}
}

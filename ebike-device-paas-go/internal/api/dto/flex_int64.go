package dto

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexInt64 unmarshals a JSON number or numeric string into an optional int64,
// mirroring Jackson's coercion for snowflake ids / timestamps from mobile clients.
type FlexInt64 struct {
	V *int64
}

func (f *FlexInt64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		f.V = nil
		return nil
	}
	v, err := parseFlexInt64(data)
	if err != nil {
		return err
	}
	f.V = &v
	return nil
}

func (f FlexInt64) Ptr() *int64 { return f.V }

func (f FlexInt64) Int64() int64 {
	if f.V == nil {
		return 0
	}
	return *f.V
}

// FlexInt64Value unmarshals a required int64 from JSON number or numeric string.
type FlexInt64Value int64

func (f *FlexInt64Value) UnmarshalJSON(data []byte) error {
	v, err := parseFlexInt64(data)
	if err != nil {
		return err
	}
	*f = FlexInt64Value(v)
	return nil
}

func (f FlexInt64Value) Int64() int64 { return int64(f) }

// Int64Slice unmarshals JSON numbers or numeric strings into []int64.
type Int64Slice []int64

func (s *Int64Slice) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = nil
		return nil
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	out := make([]int64, len(raw))
	for i, item := range raw {
		v, err := parseFlexInt64(item)
		if err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
		out[i] = v
	}
	*s = out
	return nil
}

func parseFlexInt64(raw json.RawMessage) (int64, error) {
	var n json.Number
	if err := json.Unmarshal(raw, &n); err == nil {
		return n.Int64()
	}
	var str string
	if err := json.Unmarshal(raw, &str); err != nil {
		return 0, err
	}
	return strconv.ParseInt(str, 10, 64)
}

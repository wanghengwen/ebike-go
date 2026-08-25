package rpc

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexCode accepts Java Result codes as either JSON number or string ("0", "200").
type FlexCode int

func (c *FlexCode) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*c = 0
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		if s == "" {
			*c = 0
			return nil
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("invalid code string %q: %w", s, err)
		}
		*c = FlexCode(n)
		return nil
	}
	var n int
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*c = FlexCode(n)
	return nil
}

// Result mirrors upstream Java/Microservice API envelopes.
type Result[T any] struct {
	Success bool     `json:"success"`
	Code    FlexCode `json:"code"`
	Msg     string   `json:"msg"`
	Data    T        `json:"data"`
}

// OK reports whether the upstream call succeeded (Java "0" or HTTP-style 200).
func (r Result[T]) OK() bool {
	if r.Success && (r.Code == 0 || r.Msg == "成功") {
		return true
	}
	return r.Code == 0 || r.Code == 200
}

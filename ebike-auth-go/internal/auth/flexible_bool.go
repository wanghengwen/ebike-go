package auth

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type FlexibleBool bool

func (b *FlexibleBool) UnmarshalJSON(data []byte) error {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	switch value := raw.(type) {
	case bool:
		*b = FlexibleBool(value)
	case string:
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool string %q: %w", value, err)
		}
		*b = FlexibleBool(parsed)
	default:
		*b = false
	}
	return nil
}

func (b FlexibleBool) Bool() bool {
	return bool(b)
}

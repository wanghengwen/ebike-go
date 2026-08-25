package cache

import (
	"encoding/json"
	"strings"

	"github.com/guregu/null/v5"
)

// flexBool unmarshals Java Redis booleans and Go sql.NullBool JSON objects written by mistake.
type flexBool null.Bool

func (b flexBool) toNull() null.Bool { return null.Bool(b) }

func (b *flexBool) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		*b = flexBool(null.Bool{})
		return nil
	}
	if trimmed[0] == '{' {
		var obj struct {
			Bool  bool `json:"Bool"`
			Valid bool `json:"Valid"`
		}
		if err := json.Unmarshal(data, &obj); err != nil {
			return err
		}
		if obj.Valid {
			*b = flexBool(null.BoolFrom(obj.Bool))
		} else {
			*b = flexBool(null.Bool{})
		}
		return nil
	}
	var nb null.Bool
	if err := json.Unmarshal(data, &nb); err != nil {
		return err
	}
	*b = flexBool(nb)
	return nil
}

func (b flexBool) MarshalJSON() ([]byte, error) {
	return null.Bool(b).MarshalJSON()
}

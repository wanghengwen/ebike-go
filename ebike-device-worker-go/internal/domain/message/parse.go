package message

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseData mirrors Java DeviceMessageDTO.jsonData() / JSON.parseObject(data.toString()).
// Openapi publishes data as a JSON string inside the Kafka envelope; some producers use a JSON object.
func ParseData(raw json.RawMessage) (map[string]interface{}, error) {
	raw = json.RawMessage(strings.TrimSpace(string(raw)))
	if len(raw) == 0 || string(raw) == "null" {
		return nil, fmt.Errorf("data is empty")
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(raw, &obj); err == nil {
		return obj, nil
	}

	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("data is neither object nor json-string: %w", err)
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("data json-string is empty")
	}
	if err := json.Unmarshal([]byte(s), &obj); err != nil {
		return nil, fmt.Errorf("data json-string inner unmarshal failed: %w", err)
	}
	return obj, nil
}

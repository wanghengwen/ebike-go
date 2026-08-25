package message

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ReceiveMillisTime parses receiveDataTime from Kafka JSON (epoch milliseconds).
// Java worker uses Timestamp.getTime(); openapi-go publishes int64 millis.
type ReceiveMillisTime struct {
	t time.Time
}

func (r ReceiveMillisTime) Time() time.Time {
	return r.t
}

// TruncateToSecond mirrors Java Instant.ofEpochSecond(receiveDataTime.getTime() / 1000).
func (r ReceiveMillisTime) TruncateToSecond() time.Time {
	if r.t.IsZero() {
		return r.t
	}
	return time.Unix(r.t.Unix(), 0).UTC()
}

func (r *ReceiveMillisTime) UnmarshalJSON(data []byte) error {
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	ms, err := parseEpochMillis(data)
	if err != nil {
		return err
	}
	r.t = time.UnixMilli(ms).UTC()
	return nil
}

func parseEpochMillis(data []byte) (int64, error) {
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return 0, err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return 0, fmt.Errorf("receiveDataTime string is empty")
		}
		if ms, err := strconv.ParseInt(s, 10, 64); err == nil {
			return ms, nil
		}
		t, err := time.Parse(time.RFC3339Nano, s)
		if err != nil {
			t, err = time.Parse(time.RFC3339, s)
		}
		if err != nil {
			return 0, fmt.Errorf("receiveDataTime: %w", err)
		}
		return t.UnixMilli(), nil
	}
	var n json.Number
	if err := json.Unmarshal(data, &n); err != nil {
		return 0, err
	}
	ms, err := n.Int64()
	if err != nil {
		f, err2 := n.Float64()
		if err2 != nil {
			return 0, err
		}
		ms = int64(f)
	}
	return ms, nil
}

package dto_test

import (
	"encoding/json"
	"testing"

	"ebike-device-worker-go/internal/api/dto"
)

func TestTrajectoryCoSerializesNullTotalMiles(t *testing.T) {
	raw, err := json.Marshal(dto.TrajectoryCo{
		Lng: 115.35, Lat: 22.97, Timestamp: 1781587449,
	})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["totalMiles"] != nil {
		t.Fatalf("totalMiles=%v want null", m["totalMiles"])
	}
}

func TestTrajectoryCoSerializesZeroTotalMiles(t *testing.T) {
	zero := 0
	raw, err := json.Marshal(dto.TrajectoryCo{
		Lng: 115.35, Lat: 22.97, Timestamp: 1781587449, TotalMiles: &zero,
	})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["totalMiles"].(float64) != 0 {
		t.Fatalf("totalMiles=%v", m["totalMiles"])
	}
}

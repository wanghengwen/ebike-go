package service

import (
	"encoding/json"
	"testing"

	"ebike-fence-go/internal/api/dto"
)

func TestBindParkingCOResponseOmitsBanRidingId(t *testing.T) {
	id := int64(42)
	co := dto.BindParkingCO{
		ParkingId:   &id,
		NoParkingId: &id,
	}
	raw, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["banRidingId"]; ok {
		t.Fatalf("banRidingId should be omitted from response JSON, got %s", raw)
	}
}

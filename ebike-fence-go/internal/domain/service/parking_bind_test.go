package service

import (
	"encoding/json"
	"testing"

	"ebike-fence-go/internal/api/dto"
)

// Java BindParkingCO declares banRidingId but bindParking never sets it, and Jackson keeps
// null-valued properties, so the key must be present and null.
func TestBindParkingCOResponseKeepsNullBanRidingId(t *testing.T) {
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
	got, ok := m["banRidingId"]
	if !ok {
		t.Fatalf("banRidingId must be present to match Java, got %s", raw)
	}
	if string(got) != "null" {
		t.Fatalf("banRidingId should be null, got %s", got)
	}
}

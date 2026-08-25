package fenceadmin

import (
	"encoding/json"
	"testing"
)

func TestEmptyBanRidingNearestDistanceIsNull(t *testing.T) {
	raw, err := json.Marshal(emptyBanRidingNearest())
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if string(m["distance"]) != "null" {
		t.Fatalf("distance=%s want null (consume maps null→100; 0 triggers accOff)", m["distance"])
	}
	if string(m["type"]) != "9" {
		t.Fatalf("type=%s want 9 (Java BanRidingCO ctor)", m["type"])
	}
	if string(m["id"]) != "null" {
		t.Fatalf("id=%s want null", m["id"])
	}
}

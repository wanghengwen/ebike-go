package dto

import (
	"encoding/json"
	"testing"
)

func TestComputeDistanceCmdGeoJSONString(t *testing.T) {
	raw := `{"imei":"862551059864149","point":"{\"type\":\"Point\",\"coordinates\":[118.399564,33.768009]}"}`
	var cmd ComputeDistanceCmd
	if err := json.Unmarshal([]byte(raw), &cmd); err != nil {
		t.Fatal(err)
	}
	if cmd.Point == nil {
		t.Fatal("expected parsed Point")
	}
	if cmd.Point.Lng != 118.399564 || cmd.Point.Lat != 33.768009 {
		t.Fatalf("unexpected point: %+v", cmd.Point)
	}
}

func TestComputeDistanceCmdGeoJSONObject(t *testing.T) {
	raw := `{"imei":"862551059864149","point":{"type":"Point","coordinates":[115.400423,22.976168]}}`
	var cmd ComputeDistanceCmd
	if err := json.Unmarshal([]byte(raw), &cmd); err != nil {
		t.Fatal(err)
	}
	if cmd.Point == nil || cmd.Point.Lng != 115.400423 || cmd.Point.Lat != 22.976168 {
		t.Fatalf("unexpected point: %+v", cmd.Point)
	}
}

package service

import (
	"testing"

	"ebike-device-worker-go/internal/api/dto"
)

func TestParseTotalMilesJavaCompactJSONFails(t *testing.T) {
	// Java Integer.parseInt("10}") throws; totalMiles stays 0.
	metric := `{"speed":1.0,"course":0.0,"totalMiles":10}`
	if got := parseTotalMiles(metric); got != 0 {
		t.Fatalf("expected 0 for compact JSON totalMiles, got %d", got)
	}
}

func TestParseTotalMilesJavaSpacedJSONSucceeds(t *testing.T) {
	// totalMiles must not be the last JSON field (trailing `}` breaks Java parseInt).
	metric := `{"totalMiles":10,"speed":1.0}`
	if got := parseTotalMiles(metric); got != 10 {
		t.Fatalf("expected 10, got %d", got)
	}
}

func TestParseTotalMilesJavaNullSkipped(t *testing.T) {
	metric := `{"totalMiles":null}`
	if got := parseTotalMiles(metric); got != 0 {
		t.Fatalf("expected 0 for null totalMiles, got %d", got)
	}
}

func TestApplyGpsTotalMilesAbsentLeavesNil(t *testing.T) {
	var traj dto.TrajectoryCo
	applyGpsTotalMiles(&traj, `{"speed":1.0,"course":0.0}`)
	if traj.TotalMiles != nil {
		t.Fatalf("expected nil totalMiles, got %v", *traj.TotalMiles)
	}
}

func TestApplyGpsTotalMilesPresentSetsValue(t *testing.T) {
	var traj dto.TrajectoryCo
	applyGpsTotalMiles(&traj, `{"totalMiles":10,"speed":1.0}`)
	if traj.TotalMiles == nil || *traj.TotalMiles != 10 {
		t.Fatalf("expected totalMiles=10, got %v", traj.TotalMiles)
	}
}

package service

import "testing"

func TestEndGpsGeometryMUsesSeconds(t *testing.T) {
	const endMs int64 = 1_700_000_123_456
	got := endGpsGeometryM(endMs)
	want := float64(1_700_000_123)
	if got != want {
		t.Fatalf("endGpsGeometryM(%d)=%v want %v", endMs, got, want)
	}
}

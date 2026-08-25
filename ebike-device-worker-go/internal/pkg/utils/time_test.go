package utils

import (
	"testing"
	"time"
)

func TestClampedAddMonths(t *testing.T) {
	tests := []struct {
		name   string
		input  time.Time
		months int
		want   time.Time
	}{
		{
			name:   "Aug 31 to Feb 28 (non-leap year)",
			input:  time.Date(2026, 8, 31, 10, 20, 30, 40, time.Local),
			months: -6,
			want:   time.Date(2026, 2, 28, 10, 20, 30, 40, time.Local),
		},
		{
			name:   "Aug 31 to Feb 29 (leap year)",
			input:  time.Date(2028, 8, 31, 10, 20, 30, 40, time.Local),
			months: -6,
			want:   time.Date(2028, 2, 29, 10, 20, 30, 40, time.Local),
		},
		{
			name:   "Oct 31 to Sep 30",
			input:  time.Date(2026, 10, 31, 12, 0, 0, 0, time.Local),
			months: -1,
			want:   time.Date(2026, 9, 30, 12, 0, 0, 0, time.Local),
		},
		{
			name:   "Jan 31 to Dec 31",
			input:  time.Date(2026, 1, 31, 0, 0, 0, 0, time.Local),
			months: -1,
			want:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.Local),
		},
		{
			name:   "Normal addition without clamping",
			input:  time.Date(2026, 1, 15, 0, 0, 0, 0, time.Local),
			months: 2,
			want:   time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClampedAddMonths(tt.input, tt.months)
			if !got.Equal(tt.want) {
				t.Errorf("ClampedAddMonths(%v, %d) = %v; want %v", tt.input, tt.months, got, tt.want)
			}
		})
	}
}

func TestFormatJavaGpsQueryTime(t *testing.T) {
	// 2026-06-17 10:30:45.678 local
	input := time.Date(2026, 6, 17, 10, 30, 45, 678_000_000, time.Local)
	got := FormatJavaGpsQueryTime(input.UnixMilli())
	want := input.Format("2006-01-02 15:04:05")
	if got != want {
		t.Fatalf("FormatJavaGpsQueryTime() = %q; want %q", got, want)
	}
}

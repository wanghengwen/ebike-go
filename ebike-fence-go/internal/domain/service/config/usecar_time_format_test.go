package configsvc

import (
	"testing"

	"ebike-fence-go/internal/api/dto"
)

func TestNormalizeTimeHHMMSS(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"00:00", "00:00:00"},
		{"23:59:59", "23:59:59"},
		{"", ""},
	}
	for _, tc := range tests {
		if got := normalizeTimeHHMMSS(tc.in); got != tc.want {
			t.Fatalf("normalizeTimeHHMMSS(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeUseCarTimeFields(t *testing.T) {
	start := "00:00"
	end := "23:59:59"
	co := &dto.ConfigUseCarCO{StopTimeStart: &start, StopTimeEnd: &end}
	normalizeUseCarTimeFields(co)
	if *co.StopTimeStart != "00:00:00" {
		t.Fatalf("stopTimeStart = %q", *co.StopTimeStart)
	}
	if *co.StopTimeEnd != "23:59:59" {
		t.Fatalf("stopTimeEnd = %q", *co.StopTimeEnd)
	}
}

func TestReplaceProtocolPlaceholders(t *testing.T) {
	in := "<h3>${tenantName}用户协议</h3>公司与${companyName}之间"
	got := replaceProtocolPlaceholders(in, "腾山出行", "腾山公司")
	want := "<h3>腾山出行用户协议</h3>公司与腾山公司之间"
	if got != want {
		t.Fatalf("replaceProtocolPlaceholders() = %q, want %q", got, want)
	}
}

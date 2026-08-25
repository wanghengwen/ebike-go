package service

import (
	"testing"
	"time"
)

func TestShouldUsePresenceCatchup(t *testing.T) {
	cases := []struct {
		name         string
		enabled      bool
		disconnected string
		want         bool
	}{
		{name: "enabled+disconnected", enabled: true, disconnected: "$SYS/.../disconnected", want: true},
		{name: "enabled+empty disconnected", enabled: true, want: false},
		// MQTT disabled → must NOT use catch-up cadence even if topic filled.
		{name: "disabled+disconnected filled", enabled: false, disconnected: "$SYS/.../disconnected", want: false},
		{name: "disabled+empty", enabled: false, want: false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := shouldUsePresenceCatchup(c.enabled, c.disconnected)
			if got != c.want {
				t.Fatalf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestSweepIntervalConstants(t *testing.T) {
	if luopingSweepIntervalPresence != 15*time.Minute {
		t.Fatalf("presence catch-up interval want 15m, got %s", luopingSweepIntervalPresence)
	}
	if luopingSweepIntervalFallback != 2*time.Minute {
		t.Fatalf("fallback interval want 2m, got %s", luopingSweepIntervalFallback)
	}
}

package env_test

import (
	"testing"

	"ebike-device-worker-go/internal/pkg/env"
)

func TestIsDryRun(t *testing.T) {
	cases := []struct {
		val  string
		want bool
	}{
		{"true", true},
		{`"true"`, true},
		{"'true'", true},
		{" TRUE ", true},
		{"1", true},
		{"false", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Setenv("DRY_RUN", tc.val)
		if got := env.IsDryRun(); got != tc.want {
			t.Fatalf("DRY_RUN=%q got=%v want=%v", tc.val, got, tc.want)
		}
	}
}

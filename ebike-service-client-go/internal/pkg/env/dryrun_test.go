package env

import "testing"

func TestIsDryRun(t *testing.T) {
	cases := []struct {
		val  string
		want bool
	}{
		{"true", true},
		{"TRUE", true},
		{"1", true},
		{"yes", true},
		{"\"true\"", true},
		{"false", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.val, func(t *testing.T) {
			t.Setenv("DRY_RUN", tc.val)
			if got := IsDryRun(); got != tc.want {
				t.Fatalf("IsDryRun()=%v want %v", got, tc.want)
			}
		})
	}
}

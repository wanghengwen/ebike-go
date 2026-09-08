package redis

import "testing"

func TestParseAuthCacheValue(t *testing.T) {
	cases := []struct {
		in     string
		want   bool
		wantOK bool
	}{
		{"true", true, true},
		{"false", false, true},
		{"1", true, true},
		{"0", false, true},
		{"invalid", false, false},
	}
	for _, tc := range cases {
		got, ok := parseAuthCacheValue(tc.in)
		if ok != tc.wantOK || got != tc.want {
			t.Fatalf("parseAuthCacheValue(%q) = (%v, %v), want (%v, %v)", tc.in, got, ok, tc.want, tc.wantOK)
		}
	}
}

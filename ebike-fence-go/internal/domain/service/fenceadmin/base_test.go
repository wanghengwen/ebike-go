package fenceadmin

import "testing"

func TestNextDedupedName(t *testing.T) {
	tests := []struct {
		name       string
		base       string
		suffixRows []string
		want       string
	}{
		{
			name:       "base only conflict uses suffix 2",
			base:       "禁停A",
			suffixRows: nil,
			want:       "禁停A2",
		},
		{
			name:       "contiguous suffix chain",
			base:       "禁停A",
			suffixRows: []string{"禁停A2", "禁停A3"},
			want:       "禁停A4",
		},
		{
			name:       "gap in suffix chain picks first gap",
			base:       "禁停A",
			suffixRows: []string{"禁停A2", "禁停A4"},
			want:       "禁停A3",
		},
		{
			name:       "db order stops at first mismatch",
			base:       "foo",
			suffixRows: []string{"foo4", "foo2"},
			want:       "foo2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nextDedupedName(tt.base, tt.suffixRows); got != tt.want {
				t.Fatalf("nextDedupedName() = %q, want %q", got, tt.want)
			}
		})
	}
}

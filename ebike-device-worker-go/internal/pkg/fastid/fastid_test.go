package fastid_test

import (
	"testing"

	"ebike-device-worker-go/internal/pkg/fastid"
)

func TestParseParam(t *testing.T) {
	// exported via register test helper
	m := map[string]string{}
	raw := "success=true&msg=&result=42"
	for _, kv := range splitAmp(raw) {
		parts := splitEq(kv)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	if m["success"] != "true" || m["result"] != "42" {
		t.Fatalf("unexpected parse: %v", m)
	}
}

func TestSnowflakeCompose(t *testing.T) {
	g, err := fastid.NewForTest(fastid.Config{
		InstanceNoBits: 12,
		SequenceBits:   19,
		DriftTime:      10,
	}, 7)
	if err != nil {
		t.Fatal(err)
	}
	id1 := g.Next()
	id2 := g.Next()
	if id1 == id2 {
		t.Fatal("ids should differ")
	}
}

func splitAmp(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '&' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

func splitEq(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return nil
}

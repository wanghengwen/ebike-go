package javacompat

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCoalesceJSONArrayParsesStringField(t *testing.T) {
	raw := json.RawMessage(`"[{\"type\":\"banner\",\"enable\":true}]"`)
	out := CoalesceJSONArray(raw)
	s := string(out)
	if !strings.HasPrefix(s, "[") || strings.Contains(s, `\"type\"`) {
		t.Fatalf("expected parsed array, got %s", s)
	}
}

func TestCoalesceJSONArrayDefaultsEmpty(t *testing.T) {
	for _, raw := range []json.RawMessage{nil, json.RawMessage("null"), json.RawMessage(`""`)} {
		out := CoalesceJSONArray(raw)
		if string(out) != "[]" {
			t.Fatalf("expected [], got %s for %q", out, raw)
		}
	}
}

func TestParseStringArrayFieldFromString(t *testing.T) {
	raw := json.RawMessage(`"[\"331089826894843934\",\"331098350257444022\"]"`)
	arr := ParseStringArrayField(raw)
	if len(arr) != 2 || arr[0] != "331089826894843934" {
		t.Fatalf("unexpected arr: %#v", arr)
	}
}

func TestParseStringArrayFieldFromNativeArray(t *testing.T) {
	raw := json.RawMessage(`["a","b"]`)
	arr := ParseStringArrayField(raw)
	if len(arr) != 2 || arr[1] != "b" {
		t.Fatalf("unexpected arr: %#v", arr)
	}
}

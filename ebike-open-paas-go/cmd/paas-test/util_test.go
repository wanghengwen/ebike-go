package main

import "testing"

func TestMustHTTPS(t *testing.T) {
	if err := mustHTTPS("https://hook.example.com/paas"); err != nil {
		t.Fatalf("https url rejected: %v", err)
	}
	for _, raw := range []string{"", "http://hook.example.com/paas", "ftp://x", "https:///nohost"} {
		if err := mustHTTPS(raw); err == nil {
			t.Errorf("mustHTTPS(%q) = nil, want error", raw)
		}
	}
}

func TestParseEvents(t *testing.T) {
	got, err := parseEvents("1,2, 3,4")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 || got[0] != 1 || got[3] != 4 {
		t.Fatalf("got %v", got)
	}
	if _, err := parseEvents("1,5"); err != nil {
		// 5 is syntactically valid for parseEvents; deliverability is enforced at register.
		t.Fatalf("unexpected: %v", err)
	}
	if _, err := parseEvents("0,1"); err == nil {
		t.Fatal("event 0 should be rejected")
	}
	if _, err := parseEvents(""); err == nil {
		t.Fatal("empty events should be rejected")
	}
}

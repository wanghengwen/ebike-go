package client

import (
	"testing"
	"time"
)

func TestIsConsoleAuthError(t *testing.T) {
	cases := []struct {
		code string
		want bool
	}{
		{"5001", true},
		{"5002", true},
		{"5003", true},
		{"0", false},
		{"500", false},
	}
	for _, tc := range cases {
		got := isConsoleAuthError(&GatewayError{Code: tc.code, Msg: "x"})
		if got != tc.want {
			t.Errorf("isConsoleAuthError(%q) = %v, want %v", tc.code, got, tc.want)
		}
	}
}

func TestFormatBearer(t *testing.T) {
	if got := formatBearer("abc"); got != "Bearer abc" {
		t.Fatalf("formatBearer = %q", got)
	}
	if got := formatBearer("Bearer abc"); got != "Bearer abc" {
		t.Fatalf("formatBearer preserved = %q", got)
	}
}

func TestConsoleTokenStillValid(t *testing.T) {
	consoleExpMs = -1
	consoleFetched = time.Now()
	if !consoleTokenStillValid() {
		t.Fatal("exp=-1 should stay valid")
	}
	consoleExpMs = time.Now().Add(2 * time.Minute).UnixMilli()
	if !consoleTokenStillValid() {
		t.Fatal("future exp should stay valid")
	}
	consoleExpMs = time.Now().Add(-time.Minute).UnixMilli()
	if consoleTokenStillValid() {
		t.Fatal("past exp should be invalid")
	}
}

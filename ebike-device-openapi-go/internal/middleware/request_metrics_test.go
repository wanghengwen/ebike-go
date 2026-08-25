package middleware

import "testing"

func TestShouldSkipRequestMetrics_decode(t *testing.T) {
	if !shouldSkipRequestMetrics("/device-gateway/xiaoan/decode") {
		t.Fatal("decode path should skip request metrics logging")
	}
}

func TestShouldSkipRequestMetrics_cmd(t *testing.T) {
	if shouldSkipRequestMetrics("/ebike/cmd/lock") {
		t.Fatal("cmd path should not skip request metrics logging")
	}
}

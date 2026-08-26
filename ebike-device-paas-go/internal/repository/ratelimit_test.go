package repository

import "testing"

// When Redis is unavailable (Client == nil in tests), the limiter must degrade
// to "allowed", mirroring the Java aspect which swallows non-BizException errors
// and proceeds with the call.
func TestVoiceFindCarAllowedDegradesOpen(t *testing.T) {
	ok, msg := VoiceFindCarAllowed("100", "user-1")
	if !ok || msg != "" {
		t.Fatalf("expected allow on redis-unavailable, got ok=%v msg=%q", ok, msg)
	}
}

func TestVoiceRateLimitMessages(t *testing.T) {
	if VoicePreUnitMsg != "请勿频繁点击寻车铃" {
		t.Fatalf("preUnit msg mismatch: %q", VoicePreUnitMsg)
	}
	if VoiceWindowMsg != "寻车铃功能将禁用24小时" {
		t.Fatalf("window msg mismatch: %q", VoiceWindowMsg)
	}
}

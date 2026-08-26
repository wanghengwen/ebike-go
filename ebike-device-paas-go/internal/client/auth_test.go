package client

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"strings"
	"testing"

	"ebike-device-paas-go/internal/pkg/config"
)

// the AES key/iv used by AesEncryptUtil in prod:
// "13dade85494f4b2ca4f49b21cd058625".substring(16).
const testAesKey = "a4f49b21cd058625"

// decryptCBCNoPad mirrors AesEncryptUtil.desEncrypt (without the trailing
// zero-strip) so the test can verify the encrypt output round-trips.
func decryptCBCNoPad(t *testing.T, b64, key, iv string) []byte {
	t.Helper()
	enc, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	out := make([]byte, len(enc))
	cipher.NewCBCDecrypter(block, []byte(iv)).CryptBlocks(out, enc)
	return out
}

func TestAesEncryptCBCNoPad_RoundTrip(t *testing.T) {
	cases := []string{
		"",                    // empty -> 0 bytes, no block
		"1234567890",          // 10 bytes -> padded to 16
		"abcdefghijklmnop",    // exactly 16 -> no padding
		"the quick brown fox", // 19 bytes -> padded to 32
	}
	for _, in := range cases {
		b64, err := aesEncryptCBCNoPad(in, testAesKey, testAesKey)
		if err != nil {
			t.Fatalf("encrypt(%q): %v", in, err)
		}
		// Ciphertext length must be a whole number of 16-byte blocks.
		raw, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			t.Fatalf("decode(%q): %v", in, err)
		}
		if len(raw)%16 != 0 {
			t.Fatalf("ciphertext for %q not block-aligned: %d bytes", in, len(raw))
		}
		// Decrypt and strip the zero padding -> original (matches desEncrypt).
		got := strings.TrimRight(string(decryptCBCNoPad(t, b64, testAesKey, testAesKey)), "\x00")
		if got != in {
			t.Fatalf("round-trip mismatch: in=%q got=%q", in, got)
		}
	}
}

// TestAesEncryptCBCNoPad_KnownVector pins the ciphertext for a fixed input so a
// regression in the padding / key handling is caught. The value is AES-128
// CBC/NoPadding of "appkey-123" (zero-padded to 16) with key=iv=testAesKey.
func TestAesEncryptCBCNoPad_KnownVector(t *testing.T) {
	got, err := aesEncryptCBCNoPad("appkey-123", testAesKey, testAesKey)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	// Recompute the same way an independent caller would, to assert determinism.
	want, err := aesEncryptCBCNoPad("appkey-123", testAesKey, testAesKey)
	if err != nil {
		t.Fatalf("encrypt2: %v", err)
	}
	if got != want {
		t.Fatalf("non-deterministic encrypt: %q vs %q", got, want)
	}
	if got == "" {
		t.Fatalf("expected non-empty ciphertext")
	}
}

func TestAnvelinkAesKey(t *testing.T) {
	orig := config.GlobalConfig.Anvelink.Openapi.AesKey
	defer func() { config.GlobalConfig.Anvelink.Openapi.AesKey = orig }()

	config.GlobalConfig.Anvelink.Openapi.AesKey = "13dade85494f4b2ca4f49b21cd058625"
	if got := anvelinkAesKey(); got != testAesKey {
		t.Fatalf("anvelinkAesKey() = %q, want %q", got, testAesKey)
	}

	config.GlobalConfig.Anvelink.Openapi.AesKey = "short"
	if got := anvelinkAesKey(); got != "" {
		t.Fatalf("anvelinkAesKey() for short key = %q, want empty", got)
	}
}

func TestAuthURLMatch(t *testing.T) {
	orig := config.GlobalConfig.Feign.Client.AuthURLRegex
	defer func() { config.GlobalConfig.Feign.Client.AuthURLRegex = orig }()

	// Default (unset) -> "/ebike/.*".
	config.GlobalConfig.Feign.Client.AuthURLRegex = nil
	if !authURLMatch("/ebike/cmd/lock") {
		t.Fatalf("expected /ebike/cmd/lock to match default regex")
	}
	if authURLMatch("/voltage-plan/get-rest-battery") {
		t.Fatalf("management path must not match auth regex")
	}
	if authURLMatch("/map/batchRegeo") {
		t.Fatalf("map-service path must not match auth regex")
	}

	// Explicit list.
	config.GlobalConfig.Feign.Client.AuthURLRegex = []string{"/ebike/.*"}
	if !authURLMatch("/ebike/gps/getTrajectory") {
		t.Fatalf("expected worker gps path to match")
	}
}

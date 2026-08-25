package auth

import (
	"os"
	"testing"
)

func TestLoadLocalKeys(t *testing.T) {
	// Change working directory to the project root if running from a subdirectory
	// and restore it after the test.
	origWd, err := os.Getwd()
	if err == nil {
		// Tests run from internal/auth; climb to module root for public-keys-*.json
		for i := 0; i < 3; i++ {
			if _, err := os.Stat("public-keys-apple.json"); err == nil {
				break
			}
			_ = os.Chdir("..")
		}
	}
	defer func() {
		if origWd != "" {
			_ = os.Chdir(origWd)
		}
	}()

	loadLocalKeys()

	// Check if keys are loaded
	pubKeyApple := GetApplePublicKey("eXaunmL")
	if pubKeyApple == nil {
		t.Errorf("expected to find Apple public key 'eXaunmL', got nil")
	}

	// Note: kid list might differ, but let's check if the count is non-zero
	keysMu.RLock()
	googleCount := len(googleKeys)
	appleCount := len(appleKeys)
	keysMu.RUnlock()

	if appleCount == 0 {
		t.Errorf("expected appleKeys to be non-empty")
	}
	if googleCount == 0 {
		t.Errorf("expected googleKeys to be non-empty")
	}
}

func TestParseJWK(t *testing.T) {
	// Mock n and e from standard values
	nStr := "4dGQ7bQK8LgILOdLsYzfZjkEAoQeVC_aqyc8GC6RX7dq_KvRAQAWPvkam8VQv4GK5T4ogklEKEvj5ISBamdDNq1n52TpxQwI2EqxSk7I9fKPKhRt4F8-2yETlYvye-2s6NeWJim0KBtOVrk0gWvEDgd6WOqJl_yt5WBISvILNyVg1qAAM8JeX6dRPosahRVDjA52G2X-Tip84wqwyRpUlq2ybzcLh3zyhCitBOebiRWDQfG26EH9lTlJhll-p_Dg8vAXxJLIJ4SNLcqgFeZe4OfHLgdzMvxXZJnPp_VgmkcpUdRotazKZumj6dBPcXI_XID4Z4Z3OM1KrZPJNdUhxw"
	eStr := "AQAB"

	pubKey, err := parseJWK(nStr, eStr)
	if err != nil {
		t.Fatalf("parseJWK failed: %v", err)
	}

	if pubKey == nil {
		t.Fatalf("parseJWK returned nil public key")
	}

	if pubKey.E != 65537 {
		t.Errorf("expected exponent 65537, got %d", pubKey.E)
	}
}

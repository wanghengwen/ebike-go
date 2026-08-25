package repository

import (
	"errors"
	"strings"
	"testing"
)

// TestValidateCallbackURLRequiresHTTPS: scheme is the only hard rule — private
// IPs, cluster DNS and credential-bearing URLs are accepted as long as they are
// https.
func TestValidateCallbackURLRequiresHTTPS(t *testing.T) {
	for _, raw := range []string{
		"http://open.example.com/hook",
		"http://10.0.0.5:8848/hook",
		"ftp://open.example.com/hook",
		"file:///etc/passwd",
		"open.example.com/hook",
	} {
		if err := ValidateCallbackURL(raw); err == nil {
			t.Errorf("ValidateCallbackURL(%q) = nil, want rejection", raw)
		} else if !errors.Is(err, ErrInvalidCallbackURL) {
			t.Errorf("ValidateCallbackURL(%q) error = %v, want it to wrap ErrInvalidCallbackURL", raw, err)
		}
	}
}

func TestValidateCallbackURLRejectsMalformedInput(t *testing.T) {
	cases := map[string]string{
		"empty":         "",
		"blank":         "   ",
		"no host":       "https:///hook",
		"overlong path": "https://example.com/" + strings.Repeat("a", 1100),
	}
	for name, raw := range cases {
		if err := ValidateCallbackURL(raw); err == nil {
			t.Errorf("%s: ValidateCallbackURL(%q) = nil, want rejection", name, raw)
		}
	}
}

func TestValidateCallbackURLAcceptsHTTPSEndpoints(t *testing.T) {
	for _, raw := range []string{
		"https://open.example.com/xiaoan/callback",
		"https://open.example.com:9443/callback?agent=1",
		"https://203.0.113.10/callback",
		"https://10.0.0.5:8443/hook",
		"https://ebike-device-paas.prod.svc.cluster.local:8080/hook",
		"https://user:secret@open.example.com/hook",
		"  https://open.example.com/callback  ",
	} {
		if err := ValidateCallbackURL(raw); err != nil {
			t.Errorf("ValidateCallbackURL(%q) = %v, want nil", raw, err)
		}
	}
}

package middleware

import (
	"testing"

	"ebike-gateway-go/sign"
)

func TestMultipartSignUsesEmptyJSONObject(t *testing.T) {
	timestamp := "1710000000000"
	secret := "test-secret"
	expected := sign.SignJSON("{}", sign.TimestampHeaderKey, timestamp, secret)

	// PC request.js: sha256(JSON.stringify(formData) + `_t=${timestamp}${secret}`)
	// JSON.stringify(FormData) === "{}"
	got := sign.SignJSON("{}", sign.TimestampHeaderKey, timestamp, secret)
	if got != expected {
		t.Fatalf("multipart sign mismatch: got %s want %s", got, expected)
	}
}

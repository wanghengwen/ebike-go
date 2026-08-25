package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRedactSecretsHidesTokens matters because access logs outlive credential
// rotation: a token logged once is readable for as long as the logs are kept.
func TestRedactSecretsHidesTokens(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		leaks string
	}{
		{
			name:  "json body",
			in:    `{"agentId":"a1","agentToken":"s3cr3t-value","imei":"123456789012345"}`,
			leaks: "s3cr3t-value",
		},
		{
			name:  "json body with spaces",
			in:    `{"token" : "s3cr3t-value"}`,
			leaks: "s3cr3t-value",
		},
		{
			name:  "query string",
			in:    "?imei=123456789012345&token=s3cr3t-value&event=1",
			leaks: "s3cr3t-value",
		},
		{
			name:  "aes headers",
			in:    `{"authKey":"k","authSecret":"s3cr3t-value","appId":"1"}`,
			leaks: "s3cr3t-value",
		},
	}
	for _, tc := range cases {
		got := redactSecrets(tc.in)
		if strings.Contains(got, tc.leaks) {
			t.Errorf("%s: redactSecrets(%q) = %q, still contains the secret", tc.name, tc.in, got)
		}
		if !strings.Contains(got, "redacted") {
			t.Errorf("%s: redactSecrets(%q) = %q, want a redaction marker", tc.name, tc.in, got)
		}
	}
}

// TestRedactSecretsKeepsDiagnostics guards against the redaction eating the parts
// of the request an operator actually needs.
func TestRedactSecretsKeepsDiagnostics(t *testing.T) {
	in := `{"agentId":"a1","agentToken":"s3cr3t","imei":"123456789012345","event":3}`
	got := redactSecrets(in)
	for _, want := range []string{`"agentId":"a1"`, `"imei":"123456789012345"`, `"event":3`} {
		if !strings.Contains(got, want) {
			t.Errorf("redactSecrets(%q) = %q, want it to keep %s", in, got, want)
		}
	}
}

// TestRedactQueryValueMatchesWholeParamName checks that a key which is a suffix
// of another parameter is not mistaken for the real one.
func TestRedactQueryValueMatchesWholeParamName(t *testing.T) {
	got := redactQueryValue("?mytoken=keep-me&token=hide-me", "token")
	if !strings.Contains(got, "mytoken=keep-me") {
		t.Errorf("got %q, want mytoken left alone", got)
	}
	if strings.Contains(got, "hide-me") {
		t.Errorf("got %q, want token value redacted", got)
	}
}

// TestBodyLogWriterTruncates covers allDevices and GPSPoints, whose answers are
// large enough that logging them whole costs more than the diagnostics are worth.
func TestBodyLogWriterTruncates(t *testing.T) {
	rec := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(rec)
	w := &bodyLogWriter{ResponseWriter: c.Writer}

	full := strings.Repeat("x", maxLoggedBody*3)
	if _, err := w.WriteString(full); err != nil {
		t.Fatalf("WriteString: %v", err)
	}

	logged := w.logged()
	if len(logged) > maxLoggedBody+64 {
		t.Errorf("logged body is %d bytes, want it capped near %d", len(logged), maxLoggedBody)
	}
	if !strings.Contains(logged, "truncated") {
		t.Errorf("logged body = %q, want a truncation note", logged)
	}
	if rec.Body.Len() != len(full) {
		t.Errorf("client received %d bytes, want the full %d — truncation must only affect the log",
			rec.Body.Len(), len(full))
	}
}

func TestBodyLogWriterKeepsShortBodyIntact(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	w := &bodyLogWriter{ResponseWriter: c.Writer}

	body := `{"code":0,"msg":"success"}`
	if _, err := w.WriteString(body); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if got := w.logged(); got != body {
		t.Errorf("logged() = %q, want %q", got, body)
	}
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ebike-open-paas-go/internal/pkg/config"

	"github.com/gin-gonic/gin"
)

// TestInternalAuthDeniesWhenTokenUnset is the fail-closed case: these routes
// expose callback subscriptions and consumer internals on a public ingress, so a
// missing token has to deny rather than open them to anyone who finds the path.
func TestInternalAuthDeniesWhenTokenUnset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	withInternalToken(t, "")

	rec := serveInternal(t, map[string]string{"X-Internal-Token": "anything"})
	if rec.Code == http.StatusOK && strings.Contains(rec.Body.String(), "reached handler") {
		t.Error("an unset internal token allowed the request through")
	}
	if !strings.Contains(rec.Body.String(), "internalToken") {
		t.Errorf("body = %s, want it to name the setting an operator has to fill in", rec.Body.String())
	}
}

func TestInternalAuthDeniesWrongOrMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	withInternalToken(t, "s3cr3t")

	cases := map[string]map[string]string{
		"no header":    {},
		"empty header": {"X-Internal-Token": ""},
		"wrong token":  {"X-Internal-Token": "wrong"},
		"prefix only":  {"X-Internal-Token": "s3cr"},
		"with suffix":  {"X-Internal-Token": "s3cr3t-extra"},
		"agent header": {"xc-access-token": "wrong"},
	}
	for name, headers := range cases {
		rec := serveInternal(t, headers)
		if strings.Contains(rec.Body.String(), "reached handler") {
			t.Errorf("%s: request reached the handler", name)
		}
	}
}

func TestInternalAuthAcceptsEitherHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	withInternalToken(t, "s3cr3t")

	for name, headers := range map[string]map[string]string{
		"dedicated header": {"X-Internal-Token": "s3cr3t"},
		"agent header":     {"xc-access-token": "s3cr3t"},
		"padded":           {"X-Internal-Token": "  s3cr3t  "},
	} {
		rec := serveInternal(t, headers)
		if !strings.Contains(rec.Body.String(), "reached handler") {
			t.Errorf("%s: a correct token was refused (body %s)", name, rec.Body.String())
		}
	}
}

// TestExtractAgentIDReadsQueryOrBody: Xiaoan sends agentId in the query on GETs
// and in the JSON body on POSTs, and the body has to stay readable for the
// handler that follows.
func TestExtractAgentIDReadsQueryOrBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/ebike/v1/deviceInfo?agentId=a1&imei=1", nil)
	if got := extractAgentID(c); got != "a1" {
		t.Errorf("extractAgentID from query = %q, want a1", got)
	}

	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/ebike/v1/lock", strings.NewReader(`{"agentId":"a2","imei":"1"}`))
	if got := extractAgentID(c); got != "a2" {
		t.Errorf("extractAgentID from body = %q, want a2", got)
	}
	body := make([]byte, 64)
	n, _ := c.Request.Body.Read(body)
	if !strings.Contains(string(body[:n]), `"imei"`) {
		t.Error("the request body was consumed; the handler after Auth would see nothing")
	}
}

// TestExtractAgentIDAcceptsNumericIDs: the spec types agentId as a number, and
// several customers send it unquoted.
func TestExtractAgentIDAcceptsNumericIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/ebike/v1/lock", strings.NewReader(`{"agentId":10086}`))
	if got := extractAgentID(c); got != "10086" {
		t.Errorf("extractAgentID = %q, want 10086", got)
	}
}

func TestExtractAgentIDToleratesMalformedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, body := range []string{``, `not json`, `{`, `{"agentId":null}`, `{"agentId":{"nested":1}}`, `[]`} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/ebike/v1/lock", strings.NewReader(body))
		if got := extractAgentID(c); got != "" {
			t.Errorf("extractAgentID(%q) = %q, want empty", body, got)
		}
	}
}

// withInternalToken sets the token the way production does, through a Nacos push.
func withInternalToken(t *testing.T, token string) {
	t.Helper()
	prev := config.GlobalConfig().Open.InternalToken
	config.MergeNacosAppConfig("open:\n  internalToken: \"" + token + "\"\n")
	t.Cleanup(func() {
		config.MergeNacosAppConfig("open:\n  internalToken: \"" + prev + "\"\n")
	})
}

func serveInternal(t *testing.T, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	r := gin.New()
	r.GET("/internal/callback/stats", InternalAuth(), func(c *gin.Context) {
		c.String(http.StatusOK, "reached handler")
	})

	req := httptest.NewRequest(http.MethodGet, "/internal/callback/stats", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

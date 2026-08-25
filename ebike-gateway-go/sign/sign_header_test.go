package sign

import (
	"net/http"
	"testing"
)

func TestParseRequestHeadersWeChat(t *testing.T) {
	h := make(http.Header)
	h.Set("_t", "1783049694813")
	h.Set("_s", "abc123")

	got := ParseRequestHeaders(h)
	if got.Convention != ConventionWeChat {
		t.Fatalf("convention = %+v, want WeChat", got.Convention)
	}
	if got.Timestamp != "1783049694813" || got.Sign != "abc123" {
		t.Fatalf("timestamp/sign = %q / %q", got.Timestamp, got.Sign)
	}
	if got.TimestampPayloadKey != TimestampHeaderKey {
		t.Fatalf("payload key = %q, want %q", got.TimestampPayloadKey, TimestampHeaderKey)
	}
}

func TestParseRequestHeadersThirdParty(t *testing.T) {
	h := make(http.Header)
	h.Set("T", "1783049694813")
	h.Set("S", "abc123")

	got := ParseRequestHeaders(h)
	if got.Convention != ConventionThirdParty {
		t.Fatalf("convention = %+v, want ThirdParty", got.Convention)
	}
	if got.Timestamp != "1783049694813" || got.Sign != "abc123" {
		t.Fatalf("timestamp/sign = %q / %q", got.Timestamp, got.Sign)
	}
	if got.TimestampPayloadKey != ThirdPartyTimestampHeaderKey {
		t.Fatalf("payload key = %q, want %q", got.TimestampPayloadKey, ThirdPartyTimestampHeaderKey)
	}
}

func TestParseRequestHeadersWeChatWinsWhenBothPresent(t *testing.T) {
	h := make(http.Header)
	h.Set("_t", "111")
	h.Set("_s", "from-wechat")
	h.Set("T", "222")
	h.Set("S", "from-third")

	got := ParseRequestHeaders(h)
	if got.Convention != ConventionWeChat {
		t.Fatal("expected WeChat convention when _t is present")
	}
	if got.Timestamp != "111" || got.Sign != "from-wechat" {
		t.Fatalf("got %q / %q", got.Timestamp, got.Sign)
	}
}

func TestSignJSONUsesDifferentPayloadKeys(t *testing.T) {
	body := `{"traceId":"1","tenantId":"2"}`
	ts := "1783050136361"
	secret := "test-secret"
	withUnderscore := SignJSON(body, TimestampHeaderKey, ts, secret)
	withThirdParty := SignJSON(body, ThirdPartyTimestampHeaderKey, ts, secret)
	if withUnderscore == withThirdParty {
		t.Fatal("expected different signatures for _t= vs t=")
	}
}

func TestGetTimestampAndSignWrappers(t *testing.T) {
	h := make(http.Header)
	h.Set("T", "1783049694813")
	h.Set("S", "abc123")
	if got := GetTimestamp(h); got != "1783049694813" {
		t.Fatalf("GetTimestamp() = %q", got)
	}
	if got := GetSign(h); got != "abc123" {
		t.Fatalf("GetSign() = %q", got)
	}
	if got := ResolveTimestampSignKey(h); got != ThirdPartyTimestampHeaderKey {
		t.Fatalf("ResolveTimestampSignKey() = %q", got)
	}
}

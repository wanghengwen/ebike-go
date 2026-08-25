package sign

import "net/http"

// Convention describes one client family's sign header names and payload format.
type Convention struct {
	timestampHeader     string // inbound header for timestamp
	signHeader          string // inbound header for signature
	payloadTimestampKey string // field name in sign payload, e.g. "_t=" or "t="
}

var (
	// ConventionWeChat: _s/_t headers, payload uses "_t=".
	// WeChat mini-program and most default clients.
	ConventionWeChat = Convention{
		timestampHeader:     "_t",
		signHeader:          "_s",
		payloadTimestampKey: "_t",
	}

	// ConventionThirdParty: s/t headers, payload uses "t=".
	// Third-party integration; aligns with ebike-gateway-client branch
	// feature/Bin201Decode-20240531 (client chose this naming, not proxy rewriting).
	ConventionThirdParty = Convention{
		timestampHeader:     "t",
		signHeader:          "s",
		payloadTimestampKey: "t",
	}
)

// Legacy constants kept for sign builders and Java parity tests.
const (
	TimestampHeaderKey           = "_t"
	SignHeaderKey                = "_s"
	ThirdPartyTimestampHeaderKey = "t"
)

// RequestHeaders holds sign-related values parsed from an inbound request.
type RequestHeaders struct {
	Convention          Convention
	Timestamp           string
	Sign                string
	TimestampPayloadKey string
}

// ParseRequestHeaders detects the client convention and reads sign headers.
// If "_t" is present, WeChat convention wins; otherwise third-party (s/t).
func ParseRequestHeaders(h http.Header) RequestHeaders {
	convention := ConventionThirdParty
	if h.Get(ConventionWeChat.timestampHeader) != "" {
		convention = ConventionWeChat
	}
	return RequestHeaders{
		Convention:          convention,
		Timestamp:           h.Get(convention.timestampHeader),
		Sign:                h.Get(convention.signHeader),
		TimestampPayloadKey: convention.payloadTimestampKey,
	}
}

// GetTimestamp returns the request timestamp header value.
func GetTimestamp(h http.Header) string {
	return ParseRequestHeaders(h).Timestamp
}

// GetSign returns the request signature header value.
func GetSign(h http.Header) string {
	return ParseRequestHeaders(h).Sign
}

// ResolveTimestampSignKey returns the timestamp field name used in sign payload.
func ResolveTimestampSignKey(h http.Header) string {
	return ParseRequestHeaders(h).TimestampPayloadKey
}

package sign

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

// FormParseMode controls form body parsing differences between gateways.
type FormParseMode int

const (
	FormParseBusiness FormParseMode = iota
	FormParseClient
)

// ComputeSHA256 returns lowercase hex SHA-256 of UTF-8 data, matching Java Guava Hashing.sha256().
func ComputeSHA256(data string) string {
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

func appendTimestampAndSecret(data *strings.Builder, timestampKey, timestamp, secret string) {
	data.WriteString(timestampKey)
	data.WriteString("=")
	data.WriteString(timestamp)
	data.WriteString(secret)
}

// BuildQuerySignData builds sign payload for GET requests.
func BuildQuerySignData(params url.Values, timestampKey, timestamp, secret string) string {
	var data strings.Builder
	if len(params) > 0 {
		keys := make([]string, 0, len(params))
		for key := range params {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			values := append([]string(nil), params[key]...)
			sort.Strings(values)
			for _, val := range values {
				data.WriteString(key)
				data.WriteString("=")
				data.WriteString(val)
				data.WriteString("&")
			}
		}
	}
	appendTimestampAndSecret(&data, timestampKey, timestamp, secret)
	return data.String()
}

// SignQuery computes the expected signature for GET query parameters.
func SignQuery(params url.Values, timestampKey, timestamp, secret string) string {
	return ComputeSHA256(BuildQuerySignData(params, timestampKey, timestamp, secret))
}

// BuildJSONSignData builds sign payload for JSON request bodies.
func BuildJSONSignData(requestBody, timestampKey, timestamp, secret string) string {
	var data strings.Builder
	if strings.TrimSpace(requestBody) != "" {
		data.WriteString(TrimJSON(requestBody))
	}
	appendTimestampAndSecret(&data, timestampKey, timestamp, secret)
	return data.String()
}

// SignJSON computes the expected signature for JSON request bodies.
func SignJSON(requestBody, timestampKey, timestamp, secret string) string {
	return ComputeSHA256(BuildJSONSignData(requestBody, timestampKey, timestamp, secret))
}

// ParseFormBody parses application/x-www-form-urlencoded body into url.Values.
func ParseFormBody(body string, mode FormParseMode) url.Values {
	result := make(url.Values)
	if strings.TrimSpace(body) == "" {
		return result
	}

	for _, kv := range strings.Split(body, "&") {
		if kv == "" {
			continue
		}
		pair := strings.Split(kv, "=")
		if len(pair) == 0 {
			continue
		}
		switch mode {
		case FormParseBusiness:
			if len(pair) == 1 {
				result.Add(pair[0], "")
			} else {
				result.Add(pair[0], pair[1])
			}
		case FormParseClient:
			if len(pair) != 2 {
				continue
			}
			result.Add(pair[0], pair[1])
		}
	}
	return result
}

// BuildFormSignData builds sign payload for form request bodies.
func BuildFormSignData(body string, mode FormParseMode, timestampKey, timestamp, secret string) string {
	return BuildQuerySignData(ParseFormBody(body, mode), timestampKey, timestamp, secret)
}

// SignForm computes the expected signature for form request bodies.
func SignForm(body string, mode FormParseMode, timestampKey, timestamp, secret string) string {
	return ComputeSHA256(BuildFormSignData(body, mode, timestampKey, timestamp, secret))
}

// Verify compares computed and provided signatures.
func Verify(expectedData, providedSign string) bool {
	return ComputeSHA256(expectedData) == providedSign
}

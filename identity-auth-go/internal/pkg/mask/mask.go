package mask

import (
	"encoding/json"
	"strings"
)

var idJSONKeys = []string{
	"idCardNum",
	"identityNo",
	"identity_no",
}

// MaskIDCard replaces the middle 6 characters with asterisks for log output.
func MaskIDCard(id string) string {
	if id == "" {
		return ""
	}
	n := len(id)
	if n <= 6 {
		return strings.Repeat("*", n)
	}
	start := (n - 6) / 2
	return id[:start] + strings.Repeat("*", 6) + id[start+6:]
}

// MaskOSSPath masks the ID segment in paths like match/{idNo}/{name}.jpg.
func MaskOSSPath(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if (p == "match" || p == "no_match") && i+1 < len(parts) {
			parts[i+1] = MaskIDCard(parts[i+1])
			break
		}
	}
	return strings.Join(parts, "/")
}

// MaskURL masks the ID segment in OSS URLs that contain match/ or no_match/.
func MaskURL(rawURL string) string {
	parts := strings.Split(rawURL, "/")
	for i, p := range parts {
		if (p == "match" || p == "no_match") && i+1 < len(parts) {
			parts[i+1] = MaskIDCard(parts[i+1])
			break
		}
	}
	return strings.Join(parts, "/")
}

// JSONForLog returns JSON with known ID fields masked for safe logging.
func JSONForLog(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		return string(body)
	}
	maskValue(v)
	out, err := json.Marshal(v)
	if err != nil {
		return string(body)
	}
	return string(out)
}

func maskValue(v interface{}) {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, val := range t {
			if isIDKey(k) {
				if s, ok := val.(string); ok {
					t[k] = MaskIDCard(s)
				}
				continue
			}
			maskValue(val)
		}
	case []interface{}:
		for _, item := range t {
			maskValue(item)
		}
	}
}

func isIDKey(key string) bool {
	for _, k := range idJSONKeys {
		if key == k {
			return true
		}
	}
	return false
}

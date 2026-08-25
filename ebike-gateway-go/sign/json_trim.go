package sign

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

// TrimJSON replicates Java JsonUtils.trim:
// JSON.parseObject(jsonString, Feature.OrderedField).toString(SerializerFeature.WriteMapNullValue)
//
// Key order is preserved from the original JSON input; null fields are included.
func TrimJSON(jsonString string) string {
	if strings.TrimSpace(jsonString) == "" {
		return ""
	}
	return serializeGJSON(gjson.Parse(jsonString))
}

func quoteJSONString(s string) string {
	return `"` + escapeFastJSONString(s) + `"`
}

// escapeFastJSONString matches fastjson default string serialization.
// Unlike gjson.Escape, fastjson does not escape '/', '+', or '=' in string values.
func escapeFastJSONString(s string) string {
	var sb strings.Builder
	sb.Grow(len(s) + 16)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		case '\b':
			sb.WriteString(`\b`)
		case '\f':
			sb.WriteString(`\f`)
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		case '\t':
			sb.WriteString(`\t`)
		default:
			if c < 0x20 {
				sb.WriteString(fmt.Sprintf(`\u%04x`, c))
			} else {
				sb.WriteByte(c)
			}
		}
	}
	return sb.String()
}

func serializeGJSON(r gjson.Result) string {
	switch r.Type {
	case gjson.Null:
		return "null"
	case gjson.False:
		return "false"
	case gjson.True:
		return "true"
	case gjson.Number:
		return r.Raw
	case gjson.String:
		return quoteJSONString(r.String())
	default:
		if r.IsArray() {
			var sb strings.Builder
			sb.WriteByte('[')
			first := true
			r.ForEach(func(_, value gjson.Result) bool {
				if !first {
					sb.WriteByte(',')
				}
				first = false
				sb.WriteString(serializeGJSON(value))
				return true
			})
			sb.WriteByte(']')
			return sb.String()
		}

		var sb strings.Builder
		sb.WriteByte('{')
		first := true
		r.ForEach(func(key, value gjson.Result) bool {
			if !first {
				sb.WriteByte(',')
			}
			first = false
			sb.WriteString(quoteJSONString(key.String()))
			sb.WriteByte(':')
			sb.WriteString(serializeGJSON(value))
			return true
		})
		sb.WriteByte('}')
		return sb.String()
	}
}

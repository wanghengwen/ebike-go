package user

import (
	"strconv"
	"strings"

	"ebike-service-client-go/internal/pkg/javacompat"
	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// Re-export Java serialization types from javacompat for module-local use.
type (
	LongStr  = javacompat.LongStr
	DateTime = javacompat.DateTime
)

// withBase merges the ClientDTO base validation messages (@NotEmpty on
// traceId/tenantId) with endpoint-specific messages.
func withBase(extra map[string]string) map[string]string {
	m := map[string]string{
		"traceId":  "must not be empty",
		"tenantId": "must not be empty",
	}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

var baseMsgs = withBase(nil)

// javaDoubleString mirrors Java's String.valueOf(Double): the literal "null" for
// null, otherwise Double.toString (which always keeps a decimal point for plain
// values, e.g. 113.0 -> "113.0"). Coordinate magnitudes never reach Java's
// scientific-notation thresholds, so the plain form is sufficient.
func javaDoubleString(d *float64) string {
	if d == nil {
		return "null"
	}
	s := strconv.FormatFloat(*d, 'g', -1, 64)
	if !strings.ContainsAny(s, ".eE") {
		s += ".0"
	}
	return s
}

var (
	isNullJSON   = javacompat.IsNullJSON
	callData     = javacompat.CallData
	callDataQuiet = javacompat.CallDataQuiet
	writeNPE     = javacompat.WriteNPE
)

// notBlankMsg enforces Java @NotBlank semantics with a custom message
// (the Java response is "<field> <message>", even when message is empty).
func notBlankMsg(c *gin.Context, field, value, msg string) bool {
	if strings.TrimSpace(value) == "" {
		web.WriteParamError(c, field+" "+msg)
		return false
	}
	return true
}

func boolVal(b *bool) bool { return b != nil && *b }

// ---------------------------------------------------------------------------
// DesensitizedUtil port (ebike-service-client-common, backed by hutool)
// ---------------------------------------------------------------------------

// hideRunes ports hutool StrUtil.hide / CharSequenceUtil.replace(str, start, end, '*').
func hideRunes(s string, startInclude, endExclude int) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	n := len(r)
	if startInclude > n {
		return s
	}
	if endExclude > n {
		endExclude = n
	}
	if startInclude > endExclude {
		return s
	}
	out := make([]rune, n)
	for i := 0; i < n; i++ {
		if i >= startInclude && i < endExclude {
			out[i] = '*'
		} else {
			out[i] = r[i]
		}
	}
	return string(out)
}

// javaSplitLen mirrors Java String.split(" ").length (trailing empty strings removed).
func javaSplitLen(s string) int {
	parts := strings.Split(s, " ")
	n := len(parts)
	for n > 0 && parts[n-1] == "" {
		n--
	}
	return n
}

// desensitizedName ports DesensitizedUtil.name: blank -> "", English names keep
// the part after the first space (or the last letter), Chinese names keep the
// last character (single characters are fully masked).
func desensitizedName(fullName *string) *string {
	empty := ""
	if fullName == nil || strings.TrimSpace(*fullName) == "" {
		return &empty
	}
	s := *fullName
	r := []rune(s)
	first := r[0]
	var out string
	if (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') {
		if javaSplitLen(s) > 1 {
			spaceIdx := 0
			for i, c := range r {
				if c == ' ' {
					spaceIdx = i
					break
				}
			}
			out = hideRunes(s, 0, spaceIdx)
		} else {
			out = hideRunes(s, 0, len(r)-1)
		}
	} else {
		endEx := len(r) - 1
		if len(r) == 1 {
			endEx = 1
		}
		out = hideRunes(s, 0, endEx)
	}
	return &out
}

// desensitizedIdCard ports hutool DesensitizedUtil.idCardNum(str, front, end):
// keeps the first `front` and last `end` characters, masking the middle.
func desensitizedIdCard(s *string, front, end int) *string {
	empty := ""
	if s == nil || strings.TrimSpace(*s) == "" {
		return &empty
	}
	r := []rune(*s)
	if front+end > len(r) || front < 0 || end < 0 {
		return &empty
	}
	out := hideRunes(*s, front, len(r)-end)
	return &out
}

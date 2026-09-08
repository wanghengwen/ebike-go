package management

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// Java-semantics helpers
// ---------------------------------------------------------------------------

// javaSplit replicates Java String.split(regexless separator): trailing empty
// strings are removed from the result (e.g. "+86-" -> ["+86"]).
func javaSplit(s, sep string) []string {
	parts := strings.Split(s, sep)
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

// javaDoubleToString replicates Java Double.toString(): decimal form for
// 1e-3 <= |d| < 1e7 (always with a fractional part, e.g. "114.0"),
// otherwise computerized scientific notation like "1.0E7".
func javaDoubleToString(d float64) string {
	if math.IsNaN(d) {
		return "NaN"
	}
	if math.IsInf(d, 1) {
		return "Infinity"
	}
	if math.IsInf(d, -1) {
		return "-Infinity"
	}
	if d == 0 {
		if math.Signbit(d) {
			return "-0.0"
		}
		return "0.0"
	}
	abs := math.Abs(d)
	if abs >= 1e-3 && abs < 1e7 {
		s := strconv.FormatFloat(d, 'f', -1, 64)
		if !strings.Contains(s, ".") {
			s += ".0"
		}
		return s
	}
	// scientific form: Go "1.234E+08" -> Java "1.234E8"
	s := strconv.FormatFloat(d, 'E', -1, 64)
	idx := strings.IndexByte(s, 'E')
	mant := s[:idx]
	if !strings.Contains(mant, ".") {
		mant += ".0"
	}
	exp := s[idx+1:]
	neg := strings.HasPrefix(exp, "-")
	exp = strings.TrimLeft(strings.TrimPrefix(strings.TrimPrefix(exp, "+"), "-"), "0")
	if exp == "" {
		exp = "0"
	}
	if neg {
		exp = "-" + exp
	}
	return mant + "E" + exp
}

// parseDoubleOrFail mirrors Java Double.parseDouble + GlobalExceptionHandler:
// on failure it writes the code-00001 response Java would produce for the
// uncaught NumberFormatException and returns ok=false.
func parseDoubleOrFail(c *gin.Context, s string) (float64, bool) {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		web.WriteException(c, fmt.Sprintf("NumberFormatException:For input string: %q", s))
		return 0, false
	}
	return f, true
}

// randomUUID mirrors UuidUtils.randomUUID(): UUID v4 without dashes.
func randomUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return hex.EncodeToString(b)
}

// clientMsgs maps the ClientDTO bean-validation annotations
// (traceId/tenantId are @NotEmpty) to their hibernate default messages.
var clientMsgs = map[string]string{
	"traceId":  "must not be empty",
	"tenantId": "must not be empty",
}

// mergeMsgs combines clientMsgs with module-specific field messages.
func mergeMsgs(extra map[string]string) map[string]string {
	m := make(map[string]string, len(clientMsgs)+len(extra))
	for k, v := range clientMsgs {
		m[k] = v
	}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

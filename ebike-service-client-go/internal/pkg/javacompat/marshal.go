package javacompat

import (
	"bytes"
	"encoding/json"
)

// MarshalJSONNoHTMLEscape mirrors Java Jackson default for HTML in JSON strings
// (Go's encoding/json escapes < > & by default).
func MarshalJSONNoHTMLEscape(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	b := buf.Bytes()
	if n := len(b); n > 0 && b[n-1] == '\n' {
		b = b[:n-1]
	}
	return b, nil
}

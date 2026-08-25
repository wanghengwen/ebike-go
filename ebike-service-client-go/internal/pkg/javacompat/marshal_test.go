package javacompat

import (
	"strings"
	"testing"
)

func TestMarshalJSONNoHTMLEscape(t *testing.T) {
	type payload struct {
		Content string `json:"content"`
	}
	b, err := MarshalJSONNoHTMLEscape(payload{Content: "<h3>ok</h3>"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), `\u003c`) {
		t.Fatalf("unexpected unicode escape: %s", b)
	}
	if string(b) != `{"content":"<h3>ok</h3>"}` {
		t.Fatalf("got %s", b)
	}
}

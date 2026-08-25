package dto

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestOrderLocationAnalyzeCO_JSONFieldNamesMatchJava(t *testing.T) {
	co := EmptyOrderLocationAnalyze()
	b, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	for _, key := range []string{`"olist"`, `"pcount"`, `"ncount"`, `"total"`} {
		if !strings.Contains(got, key) {
			t.Fatalf("json=%s, missing %s", got, key)
		}
	}
	if strings.Contains(got, `"oList"`) || strings.Contains(got, `"pCount"`) || strings.Contains(got, `"nCount"`) {
		t.Fatalf("json=%s, should use lowercase field names", got)
	}
}

package fence

import (
	"strings"
	"testing"

	"ebike-service-client-go/internal/pkg/jsondiff"
)

func TestConvertConfigProtocolCO(t *testing.T) {
	raw := []byte(`{"id":240063977932464092,"serviceId":239659900966803395,"title":"充值协议","type":4,"content":"<h3>腾山出行充值协议</h3><p>测试&quot;引号&quot;</p>"}`)
	out := convertConfigProtocolCO(raw)
	if strings.Contains(string(out), `\u003c`) {
		t.Fatalf("content should not unicode-escape HTML: %s", out)
	}
	if !strings.Contains(string(out), `"id":"240063977932464092"`) {
		t.Fatalf("id should be string long: %s", out)
	}
	java := `{"id":"240063977932464092","serviceId":"239659900966803395","title":"充值协议","type":4,"content":"<h3>腾山出行充值协议</h3><p>测试&quot;引号&quot;</p>","updatedPin":null,"updatedAt":null}`
	if !jsondiff.Equal([]byte(java), out) {
		t.Fatalf("protocol CO mismatch:\n%s", string(out))
	}
}

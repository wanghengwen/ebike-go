package mask

import (
	"strings"
	"testing"
)

func TestMaskIDCard(t *testing.T) {
	got := MaskIDCard("110105199003071234")
	want := "110105******071234"
	if got != want {
		t.Fatalf("MaskIDCard() = %q, want %q", got, want)
	}
	if MaskIDCard("123456") != "******" {
		t.Fatalf("short id should be fully masked")
	}
}

func TestMaskOSSPath(t *testing.T) {
	got := MaskOSSPath("match/110105199003071234/张三.jpg")
	if strings.Contains(got, "199003") {
		t.Fatalf("path still contains raw id segment: %s", got)
	}
}

func TestJSONForLog(t *testing.T) {
	body := []byte(`{"traceId":"t1","idCardNum":"110105199003071234","name":"张三"}`)
	out := JSONForLog(body)
	if strings.Contains(out, "199003071234") {
		t.Fatalf("json log still contains raw id: %s", out)
	}
	if !strings.Contains(out, "******") {
		t.Fatalf("json log should contain mask: %s", out)
	}
}

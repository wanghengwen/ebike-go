package javacompat

import (
	"encoding/base64"
	"testing"
)

func TestDataFieldBoolFromShadowResult(t *testing.T) {
	java := `{"success":true,"code":"0","msg":"成功","data":{"izPopup":true}}`
	header := base64.StdEncoding.EncodeToString([]byte(java))

	v, ok := DataFieldBoolFromShadowResult(header, "izPopup")
	if !ok || !v {
		t.Fatalf("expected izPopup=true, got ok=%v v=%v", ok, v)
	}

	if _, ok := DataFieldBoolFromShadowResult("", "izPopup"); ok {
		t.Fatal("expected false for empty header")
	}
}

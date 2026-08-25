package clientconfig

import (
	"encoding/json"
	"strings"
	"testing"

	"ebike-service-client-go/internal/pkg/jsondiff"
)

func TestConvertGuidePageConfigCOPreservesLargeIDs(t *testing.T) {
	raw := []byte(`{"id":365068182906475221,"serviceId":364881328207300695,"guidePages":[{"chainType":3,"linkUrl":"","picUrl":"https://example.com/a.jpg","linkTitle":"","appId":"","param":""}],"allowSuperEsc":false,"pageNumEsc":1,"visibleRange":0,"frequency":1,"izOn":true,"byRegister":true,"byTags":false,"tagIds":""}`)
	out := convertGuidePageConfigCO(raw)
	if strings.Contains(string(out), "365068182906475200") {
		t.Fatalf("precision lost on id: %s", out)
	}
	if !strings.Contains(string(out), `"id":"365068182906475221"`) || !strings.Contains(string(out), `"serviceId":"364881328207300695"`) {
		t.Fatalf("expected exact long strings: %s", out)
	}
}

func TestConvertGuidePageConfigCOMatchesJavaSample(t *testing.T) {
	java := `{"success":true,"code":"0","msg":"成功","data":{"id":"365068182906475221","serviceId":"364881328207300695","guidePages":[{"chainType":3,"linkUrl":"","picUrl":"https://luoping-upload.oss-cn-shanghai.aliyuncs.com/download/backend/1007/type/20260522/23363655-27f4-458c-aada-24b18b368aff.jpg","linkTitle":"","appId":"","param":""}],"allowSuperEsc":false,"pageNumEsc":1,"visibleRange":0,"frequency":1,"izOn":true,"byRegister":true,"byTags":false,"tagIds":""}}`
	raw := []byte(`{"id":365068182906475221,"serviceId":364881328207300695,"guidePages":[{"chainType":3,"linkUrl":"","picUrl":"https://luoping-upload.oss-cn-shanghai.aliyuncs.com/download/backend/1007/type/20260522/23363655-27f4-458c-aada-24b18b368aff.jpg","linkTitle":"","appId":"","param":""}],"allowSuperEsc":false,"pageNumEsc":1,"visibleRange":0,"frequency":1,"izOn":true,"byRegister":true,"byTags":false,"tagIds":""}`)
	goBytes, _ := json.Marshal(map[string]interface{}{
		"success": true, "code": "0", "msg": "成功", "data": json.RawMessage(convertGuidePageConfigCO(raw)),
	})
	if !jsondiff.Equal([]byte(java), goBytes) {
		t.Fatalf("Go guide page response should match Java:\n%s", string(goBytes))
	}
}

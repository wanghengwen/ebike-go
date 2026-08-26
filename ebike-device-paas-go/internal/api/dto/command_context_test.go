package dto

import (
	"encoding/json"
	"testing"
)

func TestCommandContext_AuditFields(t *testing.T) {
	raw := `{"tenantId":"1003","traceId":"178366799353996054730945","pin":"a_2fb92d500680c33","ip":"172.16.1.12","platform":"wechat","deviceId":"deviceId_174460457002689502293982","source":"ebike-fence","stressTesting":false,"name":"测试用户"}`
	var cc CommandContext
	if err := json.Unmarshal([]byte(raw), &cc); err != nil {
		t.Fatal(err)
	}
	if cc.TenantID != "1003" || cc.TraceID == "" || cc.Pin == "" {
		t.Fatalf("core fields: %+v", cc)
	}
	if cc.IP == nil || *cc.IP != "172.16.1.12" {
		t.Fatalf("ip: %+v", cc.IP)
	}
	if cc.DeviceID == nil || *cc.DeviceID != "deviceId_174460457002689502293982" {
		t.Fatalf("deviceId: %+v", cc.DeviceID)
	}
	if cc.Source == nil || *cc.Source != "ebike-fence" {
		t.Fatalf("source: %+v", cc.Source)
	}
	if cc.Name == nil || *cc.Name != "测试用户" {
		t.Fatalf("name: %+v", cc.Name)
	}
	if cc.StressTesting == nil || *cc.StressTesting {
		t.Fatalf("stressTesting: %+v", cc.StressTesting)
	}
}

func TestCommandContext_NullableAuditFields(t *testing.T) {
	raw := `{"tenantId":"1000","traceId":"abc","pin":"x","ip":null,"deviceId":null,"source":null,"name":null,"stressTesting":false}`
	var cc CommandContext
	if err := json.Unmarshal([]byte(raw), &cc); err != nil {
		t.Fatal(err)
	}
	if cc.IP != nil || cc.DeviceID != nil || cc.Source != nil || cc.Name != nil {
		t.Fatalf("expected nil audit pointers, got %+v", cc)
	}
}

package dto

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestErrorResultEmitsNullDataAndMsg(t *testing.T) {
	b, _ := json.Marshal(NewErrorResult(CodeIllegalArgument, "carId must not be blank"))
	s := string(b)
	if !strings.Contains(s, `"data":null`) {
		t.Errorf("expected data:null, got %s", s)
	}
	if !strings.Contains(s, `"code":"00004"`) {
		t.Errorf("expected code 00004, got %s", s)
	}

	var r Result
	_ = json.Unmarshal([]byte(`{"success":true,"code":"0","msg":null,"data":{"count":"5"}}`), &r)
	b2, _ := json.Marshal(r)
	if !strings.Contains(string(b2), `"msg":null`) {
		t.Errorf("expected msg:null passthrough, got %s", string(b2))
	}
	if !strings.Contains(string(b2), `"count":"5"`) {
		t.Errorf("expected Long-as-string passthrough, got %s", string(b2))
	}
}

func TestCommandEmbedsCommandContext(t *testing.T) {
	var cmd ReturnCarCmd
	raw := `{
		"commandContext":{"tenantId":"t1","traceId":"tr1","pin":"u1"},
		"serviceAreaId":1,
		"carCmd":{"carId":"c1","imei":"i1","carLocation":{"lng":1,"lat":2}},
		"userCmd":{"userLocation":{"lng":1,"lat":2}}
	}`
	if err := json.Unmarshal([]byte(raw), &cmd); err != nil {
		t.Fatal(err)
	}
	if cmd.CommandContext == nil || cmd.CommandContext.TenantId != "t1" || cmd.CommandContext.TraceId != "tr1" {
		t.Fatalf("commandContext not parsed: %+v", cmd.CommandContext)
	}
}

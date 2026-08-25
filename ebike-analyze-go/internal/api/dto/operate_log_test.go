package dto

import (
	"encoding/json"
	"testing"
)

func TestOperateLogCmd_UnmarshalPageQueryFields(t *testing.T) {
	raw := `{
		"commandContext":{"tenantId":"1003","traceId":"t-1"},
		"pageNum":2,
		"pageSize":20,
		"startTime":1700000000,
		"endTime":1700003600,
		"pins":["b_001"],
		"eventType":"login,logout",
		"platform":"pc",
		"traceId":"trace-1",
		"carId":"100600100",
		"imei":"861881055224233"
	}`
	var cmd OperateLogCmd
	if err := json.Unmarshal([]byte(raw), &cmd); err != nil {
		t.Fatal(err)
	}
	if cmd.PageNum != 2 || cmd.PageSize != 20 || len(cmd.Pins) != 1 || cmd.EventType != "login,logout" {
		t.Fatalf("cmd=%+v", cmd)
	}
	if cmd.CommandContext == nil || cmd.CommandContext.TenantId != "1003" {
		t.Fatalf("commandContext=%+v", cmd.CommandContext)
	}
}

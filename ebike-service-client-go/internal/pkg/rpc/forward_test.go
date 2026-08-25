package rpc

import (
	"encoding/json"
	"testing"

	"ebike-service-client-go/internal/api/dto"
)

func TestMergeCommandContextInjectsCommandContext(t *testing.T) {
	svcID := int64(100)
	dtoObj := map[string]interface{}{
		"pin":       "13800000000",
		"serviceId": svcID,
	}
	cmdCtx := &dto.CommandContext{
		TenantId: "tenant-1",
		Pin:      "13800000000",
		TraceId:  "trace-abc",
		Source:   "ebike-service-client",
		Platform: "wechat",
	}

	body, err := mergeCommandContext(dtoObj, cmdCtx)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := body["commandContext"]; !ok {
		t.Fatal("commandContext must be present in downstream body")
	}

	var merged map[string]interface{}
	raw, _ := json.Marshal(body)
	if err := json.Unmarshal(raw, &merged); err != nil {
		t.Fatal(err)
	}
	ctx, ok := merged["commandContext"].(map[string]interface{})
	if !ok {
		t.Fatalf("commandContext type = %T", merged["commandContext"])
	}
	if ctx["tenantId"] != "tenant-1" || ctx["pin"] != "13800000000" || ctx["traceId"] != "trace-abc" {
		t.Fatalf("commandContext fields mismatch: %+v", ctx)
	}
	if merged["pin"] != "13800000000" {
		t.Fatalf("dto pin not preserved: %+v", merged)
	}
}

func TestMergeCommandContextEmptyDTO(t *testing.T) {
	cmdCtx := &dto.CommandContext{TenantId: "t1", Pin: "p1"}
	body, err := mergeCommandContext(struct{}{}, cmdCtx)
	if err != nil {
		t.Fatal(err)
	}
	if len(body) != 1 {
		t.Fatalf("expected only commandContext, got keys %d", len(body))
	}
}

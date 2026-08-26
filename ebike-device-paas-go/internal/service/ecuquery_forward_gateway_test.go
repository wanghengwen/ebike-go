package service

import (
	"encoding/json"
	"testing"

	"ebike-device-paas-go/internal/api/dto"
)

func TestReqFrom_IncludesCommandContext(t *testing.T) {
	cc := &dto.CommandContext{TenantID: "1000", TraceID: "trace-1"}
	q := dto.EcuQuery{
		Imei:           "866940070054481",
		CommandContext: cc,
	}
	body := reqFrom(q, `{"type":"c34"}`, nil)
	if body.CommandContext == nil || body.CommandContext.TenantID != "1000" {
		t.Fatalf("commandContext not forwarded: %+v", body.CommandContext)
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(raw) {
		t.Fatalf("invalid json: %s", raw)
	}
}

package app

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"ebike-analyze-go/internal/api/dto"
)

func TestMapHitToOperationLogCo_NullFieldsAndTimestampFallback(t *testing.T) {
	src := map[string]interface{}{
		"tenantId":   "1003",
		"eventType":  "login",
		"@timestamp": "2024-08-23T10:11:12.000Z",
		"name":       nil,
	}
	co, err := mapHitToOperationLogCo(src)
	if err != nil {
		t.Fatal(err)
	}
	if co.TenantID == nil || *co.TenantID != "1003" {
		t.Fatalf("tenantId=%v", co.TenantID)
	}
	if co.Name != nil {
		t.Fatalf("name should stay null, got %v", co.Name)
	}
	if co.Time == nil {
		t.Fatal("time should be filled from @timestamp")
	}

	b, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	raw := string(b)
	if !strings.Contains(raw, `"name":null`) {
		t.Fatalf("json=%s, expected name:null", raw)
	}
}

func TestMapHitToOperationLogCo_EpochSeconds(t *testing.T) {
	src := map[string]interface{}{
		"time": float64(1700000000),
	}
	co, err := mapHitToOperationLogCo(src)
	if err != nil {
		t.Fatal(err)
	}
	if co.Time == nil {
		t.Fatal("time nil")
	}
	want := time.Unix(1700000000, 0).In(time.Local).Format("2006-01-02T15:04:05")
	if co.Time.AsTime().Format("2006-01-02T15:04:05") != want {
		t.Fatalf("got=%s want=%s", co.Time.AsTime(), want)
	}
}

func TestOperateLogPage_NilClientReturnsNil(t *testing.T) {
	var svc OperateLogService
	if got := svc.Page(t.Context(), &dto.OperateLogCmd{PageNum: 1, PageSize: 10}, "1003"); got != nil {
		t.Fatalf("got=%v", got)
	}
}

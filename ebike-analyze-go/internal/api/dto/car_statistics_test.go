package dto

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCarStatisticsCO_ImeiNullWhenUnset(t *testing.T) {
	co := CarStatisticsCO{CarID: "100600100", OrderCount: intPtr(0)}
	b, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if !strings.Contains(got, `"imei":null`) {
		t.Fatalf("json=%s, want imei null", got)
	}
}

func TestCarStatisticsCO_PageOmitsExtraFieldsWhenUnset(t *testing.T) {
	orderCount := 2
	co := CarStatisticsCO{
		CarID: "100600021", OrderCount: &orderCount,
		OrderCost: int64Ptr(0), RidingDistance: int64Ptr(38), RidingTime: int64Ptr(16034),
	}
	if co.DdMissOrder != nil || co.OperationMissOrderCount != nil {
		t.Fatalf("page-style CO should leave aggregated counters unset, got %+v", co)
	}
}

func intPtr(v int) *int       { return &v }
func int64Ptr(v int64) *int64 { return &v }

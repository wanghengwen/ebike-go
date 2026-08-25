package app

import (
	"encoding/json"
	"strings"
	"testing"

	"ebike-analyze-go/internal/api/dto"
	"ebike-analyze-go/internal/infrastructure/persistence/model"
)

func TestMapCarStatPage_OmitsAggregatedFieldsLikeJavaEntity(t *testing.T) {
	row := model.CarStatistic{
		CarID: "100600100", OrderCount: 1, OrderCost: 100,
		RidingDistance: 38, RidingTime: 16034,
		DdMissOrder: 5, OperationMissOrderCount: 3,
	}
	co := mapCarStatPage(row)
	if co.DdMissOrder != nil || co.OperationMissOrderCount != nil || co.ChangeBatteryCount != nil {
		t.Fatalf("mapCarStatPage should not populate aggregated counters, got %+v", co)
	}
	b, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if !strings.Contains(got, `"carId":"100600100"`) || !strings.Contains(got, `"orderCount":1`) {
		t.Fatalf("json=%s", got)
	}
}

func TestHotPointCmd_UnmarshalJavaDateTimeFormat(t *testing.T) {
	raw := `{
		"commandContext":{"tenantId":"2"},
		"start":"2026-06-01 00:00:00",
		"end":"2026-06-03 00:00:00",
		"serviceId":227587649840878077
	}`
	var cmd dto.HotPointCmd
	if err := json.Unmarshal([]byte(raw), &cmd); err != nil {
		t.Fatal(err)
	}
	if cmd.Start.AsTime().Format("2006-01-02 15:04:05") != "2026-06-01 00:00:00" {
		t.Fatalf("start=%v", cmd.Start.AsTime())
	}
}

func TestCarStatisticsQuery_UnmarshalPageQueryFields(t *testing.T) {
	raw := `{
		"commandContext":{"tenantId":"1003"},
		"carId":"100600000",
		"pageNum":2,
		"pageSize":20,
		"searchCount":false
	}`
	var q dto.CarStatisticsQuery
	if err := json.Unmarshal([]byte(raw), &q); err != nil {
		t.Fatal(err)
	}
	if q.PageNum != 2 || q.PageSize != 20 || q.SearchCount == nil || *q.SearchCount {
		t.Fatalf("query=%+v", q)
	}
}

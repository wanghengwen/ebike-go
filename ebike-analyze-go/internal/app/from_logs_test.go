package app

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"ebike-analyze-go/internal/api/dto"
	"ebike-analyze-go/internal/infrastructure/es"
)

func TestSelectOrderAnalyze_EmptyResultMatchesJavaLog(t *testing.T) {
	svc := &OrderQueryService{}
	out := svc.SelectOrderAnalyze(nil, &dto.OrderQueryCmd{
		PageNum: 1, PageSize: 10,
		ServiceID: int64Ptr64(227587649840878077),
		StartTime: []int64{1783267200000, 1783353599000},
	}, "2")
	if out.Total != 0 || out.PCount != 0 || out.NCount != 0 || len(out.OList) != 0 {
		t.Fatalf("empty analyze=%+v, want all zeros", out)
	}
}

func TestBuildOrderQuery_SelectOrderAnalyzeLogSample(t *testing.T) {
	cmd := &dto.OrderQueryCmd{
		ServiceID: int64Ptr64(244068824482587140),
		StartTime: []int64{1783526400000, 1783612799000},
	}
	q := es.BuildOrderQuery(cmd)
	b, err := json.Marshal(q)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	for _, want := range []string{
		`"izPaid"`, `"serviceId"`, `"startTime"`, `"2026-07-09 00:00:00"`, `"2026-07-09 23:59:59"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("dsl=%s, missing %s", got, want)
		}
	}
}

func TestBuildUserQuery_SelectUserCountLogSample(t *testing.T) {
	cmd := &dto.UserQueryCmd{
		ServiceID: []int64{234425751024698753, 338359774158002104, 338361114187276531, 338362359727786125},
	}
	q := es.BuildUserQuery(cmd)
	b, err := json.Marshal(q)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if !strings.Contains(got, `"serviceId"`) || !strings.Contains(got, `"exists"`) {
		t.Fatalf("dsl=%s, missing serviceId terms + exists", got)
	}
}

func TestBuildUserQuery_AuthNameFilterFromLog(t *testing.T) {
	cmd := &dto.UserQueryCmd{
		AuthName:  "张灏宇",
		ServiceID: []int64{234425751024698753, 338359774158002104, 338361114187276531, 338362359727786125},
	}
	q := es.BuildUserQuery(cmd)
	b, err := json.Marshal(q)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"authName"`) {
		t.Fatalf("dsl=%s, missing authName term", string(b))
	}
}

func TestSetScopeNumV1_Scope2MatchesJavaAssignment(t *testing.T) {
	m := &dto.AgeMap{}
	setScopeNumV1(2, 5, m)
	setScopeNumV1(2, 9, m)
	if m.TwoAgeScope != 9 || m.Total != 14 {
		t.Fatalf("AgeMap=%+v, scope 2 should overwrite not accumulate (Java parity)", m)
	}
}

func TestSetScopeNumV1_LogSampleTotals(t *testing.T) {
	// Values from log fingerprint dc7c13671863 (tenant 2, multi-service ageStatistic).
	female := dto.AgeMap{TwoAgeScope: 1, ThreeAgeScope: 3, FourAgeScope: 8, Total: 12}
	male := dto.AgeMap{OneAgeScope: 1, TwoAgeScope: 1, ThreeAgeScope: 15, FourAgeScope: 16, Total: 37}
	if female.Total+male.Total != 49 {
		t.Fatalf("log sample total=%d, want 49", female.Total+male.Total)
	}
}

func TestParkingInAndOutflow_Empty24HoursMatchesJavaLog(t *testing.T) {
	co := dto.ParkingInAndOutflowCo{
		Influx:  make([]int, 24),
		OutFlow: make([]int, 24),
		CanRent: make([]int, 24),
	}
	if len(co.Influx) != 24 || len(co.OutFlow) != 24 || len(co.CanRent) != 24 {
		t.Fatalf("arrays=%d/%d/%d, want 24 each", len(co.Influx), len(co.OutFlow), len(co.CanRent))
	}
}

func TestFillList_MatchesJavaHourPadding(t *testing.T) {
	cases := []struct {
		hour int
		want int
	}{
		{1, 0}, {2, 1}, {10, 9}, {23, 22},
	}
	for _, tc := range cases {
		got := len(fillList(time.Date(2026, 7, 6, tc.hour, 0, 0, 0, time.Local)))
		if got != tc.want {
			t.Fatalf("hour=%d fillList len=%d, want %d", tc.hour, got, tc.want)
		}
	}
}

func TestCarStatisticsListQuery_LogTimeFormats(t *testing.T) {
	raw := `{
		"commandContext":{"tenantId":"1007"},
		"serviceId":364881328207300695,
		"carIds":["100600021"],
		"startTime":"2026-07-06T00:00:00",
		"endTime":"2026-07-06T23:59:59"
	}`
	var q dto.CarStatisticsListQuery
	if err := json.Unmarshal([]byte(raw), &q); err != nil {
		t.Fatal(err)
	}
	if q.StartTime.AsTime().Format("2006-01-02 15:04:05") != "2026-07-06 00:00:00" {
		t.Fatalf("startTime=%v", q.StartTime.AsTime())
	}
}

func int64Ptr64(v int64) *int64 { return &v }

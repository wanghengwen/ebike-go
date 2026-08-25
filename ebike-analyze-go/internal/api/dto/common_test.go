package dto

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestDateOnlyValue_UnmarshalJSONUsesLocalMidnight(t *testing.T) {
	var d DateOnlyValue
	if err := json.Unmarshal([]byte(`"2026-07-08"`), &d); err != nil {
		t.Fatal(err)
	}
	got := d.LocalDate()
	want := time.Date(2026, 7, 8, 0, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Fatalf("LocalDate=%v, want %v (loc=%v)", got, want, got.Location())
	}
}

func TestNormalizeLocalDate_FromUTCMidnight(t *testing.T) {
	utc := time.Date(2026, 6, 26, 0, 0, 0, 0, time.UTC)
	got := NormalizeLocalDate(utc)
	want := time.Date(2026, 6, 26, 0, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Fatalf("NormalizeLocalDate=%v, want %v", got, want)
	}
}

func TestCarStatisticsListQuery_UnmarshalFlexibleFields(t *testing.T) {
	raw := `{
		"commandContext":{"tenantId":"1000"},
		"serviceId":"229631747183616067",
		"carIds":["100600000"],
		"startTime":"2026-07-09 00:00:00",
		"endTime":"2026-07-09 23:59:59"
	}`
	var q CarStatisticsListQuery
	if err := json.Unmarshal([]byte(raw), &q); err != nil {
		t.Fatal(err)
	}
	if q.ServiceID.Int64() != 229631747183616067 {
		t.Fatalf("serviceId=%d", q.ServiceID.Int64())
	}
	if q.StartTime.AsTime().Format("2006-01-02 15:04:05") != "2026-07-09 00:00:00" {
		t.Fatalf("startTime=%v", q.StartTime.AsTime())
	}
}

func TestParkingOrderStatisticListQry_UnmarshalFlexibleFields(t *testing.T) {
	raw := `{
		"commandContext":{"tenantId":"1000"},
		"serviceId":"229631747183616067",
		"parkingIds":["229631903949920355","229632002734164629"],
		"startTime":"2026-07-09 00:00:00",
		"endTime":"2026-07-09 23:59:59"
	}`
	var q ParkingOrderStatisticListQry
	if err := json.Unmarshal([]byte(raw), &q); err != nil {
		t.Fatal(err)
	}
	if q.ServiceID.Int64() != 229631747183616067 {
		t.Fatalf("serviceId=%d", q.ServiceID.Int64())
	}
	if len(q.ParkingIds) != 2 || q.ParkingIds[0] != 229631903949920355 {
		t.Fatalf("parkingIds=%v", []int64(q.ParkingIds))
	}
}

func TestParkingOrderStatisticCO_ListResponseOmitsNameAndAreaSize(t *testing.T) {
	co := ParkingOrderStatisticCO{ParkingID: 338330675754571245}
	b, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if !strings.Contains(got, `"name":null`) || !strings.Contains(got, `"areaSize":null`) {
		t.Fatalf("json=%s, want null name and areaSize", got)
	}
}

package es

import (
	"encoding/json"
	"strings"
	"testing"

	"ebike-analyze-go/internal/api/dto"
)

func TestBuildOperateLogQuery_MatchesJavaFilters(t *testing.T) {
	start := int64(1700000000)
	end := int64(1700003600)
	cmd := &dto.OperateLogCmd{
		Pins:      []string{"b_001", "a_002"},
		EventType: "login,logout",
		Platform:  "pc,ios",
		TraceID:   "trace-1",
		CarID:     "100600100",
		Imei:      "861881055224233",
		StartTime: &start,
		EndTime:   &end,
	}

	q := BuildOperateLogQuery(cmd, "1003")

	b, err := json.Marshal(q)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	for _, want := range []string{
		`"tenantId":"1003"`,
		`"pin":["b_001","a_002"]`,
		`"eventType":["login","logout"]`,
		`"platform":["pc","ios"]`,
		`"traceId":"trace-1"`,
		`"carId":"100600100"`,
		`"imei":"861881055224233"`,
		`"@timestamp"`,
		`"gte":1700000000`,
		`"lte":1700003600`,
		`"constant_score"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("json=%s, missing %s", got, want)
		}
	}
}

func TestBuildOperateLogQuery_OmitsNullRangeBounds(t *testing.T) {
	start := int64(1700000000)
	onlyStart := BuildOperateLogQuery(&dto.OperateLogCmd{StartTime: &start}, "1003")
	b, err := json.Marshal(onlyStart)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if !strings.Contains(got, `"gte":1700000000`) {
		t.Fatalf("json=%s, missing gte", got)
	}
	if strings.Contains(got, `"lte"`) || strings.Contains(got, "null") {
		t.Fatalf("json=%s, should omit null lte", got)
	}

	noTime := BuildOperateLogQuery(&dto.OperateLogCmd{}, "1003")
	b, err = json.Marshal(noTime)
	if err != nil {
		t.Fatal(err)
	}
	got = string(b)
	if strings.Contains(got, `"@timestamp"`) || strings.Contains(got, "null") {
		t.Fatalf("json=%s, should omit range when both bounds missing", got)
	}
}

func TestSplitCSV(t *testing.T) {
	if got := splitCSV(""); got != nil {
		t.Fatalf("got=%v", got)
	}
	if got := splitCSV("a,b, c"); len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("got=%v", got)
	}
}

func TestSearchAll_EmitsCorrectFromOffset(t *testing.T) {
	body := buildSearchAllBody(M{"match_all": M{}}, 3, 20, []M{{"endTime": M{"order": "desc"}}}, 0)
	from, ok := body["from"].(int)
	if !ok || from != 40 {
		t.Fatalf("from=%v body=%v", body["from"], body)
	}
	size, ok := body["size"].(int)
	if !ok || size != 20 {
		t.Fatalf("size=%v body=%v", body["size"], body)
	}

	// Also assert the JSON payload that would be sent to ES.
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)
	if !strings.Contains(got, `"from":40`) || !strings.Contains(got, `"size":20`) {
		t.Fatalf("json=%s", got)
	}
}

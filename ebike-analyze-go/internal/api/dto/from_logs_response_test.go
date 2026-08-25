package dto

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCarStatisticsCO_MatchesJavaLogShape(t *testing.T) {
	// From log: CarStatisticsCO(carId=100600021, imei=null, orderCount=2, ...)
	orderCount := 2
	orderCost := int64(0)
	ridingDistance := int64(38)
	ridingTime := int64(16034)
	zero := 0
	co := CarStatisticsCO{
		CarID: "100600021", Imei: nil,
		OrderCount: &orderCount, OrderCost: &orderCost,
		RidingDistance: &ridingDistance, RidingTime: &ridingTime,
		DdMissOrder: &zero, OperationMissOrderCount: &zero,
		ChangeBatteryCount: &zero, RepairCount: &zero, MoveCarCount: &zero,
	}
	b, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	for _, key := range []string{
		`"carId":"100600021"`, `"imei":null`, `"orderCount":2`,
		`"ridingDistance":38`, `"ridingTime":16034`,
	} {
		if !strings.Contains(got, key) {
			t.Fatalf("json=%s, missing %s", got, key)
		}
	}
}

func TestOrderLocationAnalyzeCO_EmptyMatchesJavaLog(t *testing.T) {
	co := EmptyOrderLocationAnalyze()
	b, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	for _, key := range []string{`"total":0`, `"pcount":0`, `"ncount":0`, `"olist":[]`} {
		if !strings.Contains(got, key) {
			t.Fatalf("json=%s, missing %s", got, key)
		}
	}
}

func TestRidingCardPageDTO_MatchesJavaLogShape(t *testing.T) {
	page := NewPageDTO(1, 10, 0, []RidingCardCO{})
	b, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	for _, key := range []string{
		`"count":0`, `"pageNum":1`, `"pageSize":10`,
		`"orders":null`, `"searchCount":true`, `"list":[]`,
	} {
		if !strings.Contains(got, key) {
			t.Fatalf("json=%s, missing %s", got, key)
		}
	}
}

func TestParkingInAndOutflowCo_MatchesJavaLogShape(t *testing.T) {
	zeros := make([]int, 24)
	co := ParkingInAndOutflowCo{Influx: zeros, OutFlow: zeros, CanRent: zeros}
	b, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	for _, key := range []string{`"influx":[`, `"outFlow":[`, `"canRent":[`} {
		if !strings.Contains(got, key) {
			t.Fatalf("json=%s, missing %s", got, key)
		}
	}
	var decoded ParkingInAndOutflowCo
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Influx) != 24 || len(decoded.OutFlow) != 24 || len(decoded.CanRent) != 24 {
		t.Fatalf("arrays=%d/%d/%d, want 24 each", len(decoded.Influx), len(decoded.OutFlow), len(decoded.CanRent))
	}
}

func TestAgeStatisticCO_MatchesJavaLogFieldNames(t *testing.T) {
	co := AgeStatisticCO{
		HaveRidingQualification: RidingQualification{
			Female: AgeMap{TwoAgeScope: 1, ThreeAgeScope: 3, FourAgeScope: 8, Total: 12},
			Male:   AgeMap{OneAgeScope: 1, TwoAgeScope: 1, ThreeAgeScope: 15, FourAgeScope: 16, Total: 37},
			Total:  49,
		},
		NoRidingQualification: RidingQualification{},
	}
	b, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	for _, key := range []string{
		`"have_riding_qualification"`, `"no_riding_qualification"`,
		`"16_to_22"`, `"23_to_28"`, `"29_to_40"`, `"41_to_60"`,
		`"female"`, `"male"`, `"unknown"`,
	} {
		if !strings.Contains(got, key) {
			t.Fatalf("json=%s, missing %s", got, key)
		}
	}
}

func TestLogFixtures_JavaSuccessCode(t *testing.T) {
	for _, fx := range loadLogFixtures(t) {
		for _, sample := range fx.Samples {
			if sample.Code != "0" {
				t.Fatalf("%s sample %s code=%q, want 0", fx.Path, sample.Fingerprint, sample.Code)
			}
		}
	}
}

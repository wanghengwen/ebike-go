package service

import (
	"testing"

	domainconfig "ebike-fence-go/internal/domain/config"
)

func boolPtr(v bool) *bool { return &v }
func intPtr(v int) *int    { return &v }

func TestIzCanReturnParity(t *testing.T) {
	cfg := &domainconfig.ConfigBackcarCO{
		AllowOutofParking: boolPtr(true),
		AllowInNostop:     boolPtr(false),
		AllowOutofService: boolPtr(true),
		IzHelmetReign:     boolPtr(true),
	}

	cases := []struct {
		returnType FenceRelation
		want       bool
	}{
		{IN_PARKING_LAT, true},
		{OUT_PARKING_LAT, true},
		{IN_PARKING_RFID, false},
		{IN_PARKING_HELMET, false},
		{IN_NO_PARKING, false},
		{OUT_SERVICE_AREA, true},
		{IN_BAN_RIDING, false},
	}

	for _, tc := range cases {
		got := *izCanReturn(tc.returnType, cfg)
		if got != tc.want {
			t.Fatalf("izCanReturn(%s)=%v want %v", tc.returnType, got, tc.want)
		}
	}
}

func TestProcessPartResultsHelmetMapping(t *testing.T) {
	partRes := []PartAnalysisResultCO{{Name: "helmet", Result: partResultBool(false)}}
	got := processPartResults(partRes, IN_PARKING_LAT)
	if got != IN_PARKING_HELMET {
		t.Fatalf("got %s want %s", got, IN_PARKING_HELMET)
	}
}

func TestProcessPartResultsRFIDMapping(t *testing.T) {
	partRes := []PartAnalysisResultCO{{Name: "rfid_beacon", Result: partResultBool(false)}}
	got := processPartResults(partRes, IN_PARKING_LAT)
	if got != IN_PARKING_RFID {
		t.Fatalf("got %s want %s", got, IN_PARKING_RFID)
	}
}

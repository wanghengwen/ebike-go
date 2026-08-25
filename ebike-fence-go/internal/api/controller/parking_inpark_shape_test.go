package controller

import (
	"encoding/json"
	"testing"

	"ebike-fence-go/internal/api/dto"
)

func TestInParkInfoJavaShapePreservesSnowflakeIDs(t *testing.T) {
	const (
		parkingID = int64(365824859507267002)
		serviceID = int64(227596718664324575)
	)
	co := &dto.ParkingCO{
		FenceCO: dto.FenceCO{
			Id:        parkingID,
			Type:      2,
			Name:      "一键还车测试",
			ShapeType: "polygon",
			CenterLat: 30.189486,
			CenterLng: 120.185647,
		},
		ServiceId:        serviceID,
		MaxParkingNumber: 1,
		IzEnable:         true,
	}

	shaped := inParkInfoJavaShape(co)
	if shaped == nil {
		t.Fatal("expected shaped map")
	}

	idNum, ok := shaped["id"].(json.Number)
	if !ok {
		t.Fatalf("id type=%T want json.Number", shaped["id"])
	}
	id, err := idNum.Int64()
	if err != nil || id != parkingID {
		t.Fatalf("id=%v err=%v want %d", idNum, err, parkingID)
	}

	sidNum, ok := shaped["serviceId"].(json.Number)
	if !ok {
		t.Fatalf("serviceId type=%T want json.Number", shaped["serviceId"])
	}
	sid, err := sidNum.Int64()
	if err != nil || sid != serviceID {
		t.Fatalf("serviceId=%v err=%v want %d", sidNum, err, serviceID)
	}

	for _, k := range []string{"pics", "distance", "tenantId", "carCount"} {
		if shaped[k] != nil {
			t.Fatalf("%s=%v want null", k, shaped[k])
		}
	}

	// Round-trip through JSON must keep exact decimal digits.
	out, err := json.Marshal(shaped)
	if err != nil {
		t.Fatal(err)
	}
	var again map[string]json.RawMessage
	if err := json.Unmarshal(out, &again); err != nil {
		t.Fatal(err)
	}
	if string(again["id"]) != "365824859507267002" {
		t.Fatalf("marshaled id=%s", again["id"])
	}
	if string(again["serviceId"]) != "227596718664324575" {
		t.Fatalf("marshaled serviceId=%s", again["serviceId"])
	}
}

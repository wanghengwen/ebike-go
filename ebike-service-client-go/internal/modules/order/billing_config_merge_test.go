package order

import (
	"encoding/json"
	"testing"

	"ebike-service-client-go/internal/pkg/jsondiff"
)

func TestEnsureCBillingConfigNullFields(t *testing.T) {
	merged := map[string]interface{}{
		"id":       "339618167360852753",
		"discount": float64(1),
	}
	ensureCBillingConfigNullFields(merged)
	if _, ok := merged["maxPenaltyOutofParking"]; !ok {
		t.Fatal("expected maxPenaltyOutofParking key")
	}
	if merged["maxPenaltyOutofParking"] != nil {
		t.Fatalf("expected null maxPenaltyOutofParking, got %v", merged["maxPenaltyOutofParking"])
	}
}

func TestBillingConfigShadowDiffWithJavaSample(t *testing.T) {
	java := `{"success":true,"code":"0","msg":"成功","data":{"traceId":null,"tenantId":null,"platform":null,"deviceId":null,"version":null,"ip":null,"longitude":null,"latitude":null,"source":null,"stressTesting":false,"id":"339618167360852753","serviceId":"338362359727786125","freeDistance":5,"freeTime":60000,"freeTimes":3,"startingDistance":0,"startingTime":900000,"startingPrice":200,"discount":1.0,"overDistanceCostPerMter":0,"timeOutCostPerMin":100,"allowOutofService":true,"allowInNostop":false,"allowOutofParking":true,"timeUnit":600000,"dispatchCost":1000,"maxPenaltyOutofParking":null,"penaltyInNostop":0,"penaltyOutofService":2000,"mostMoneyOneday":6000,"type":0,"allowInBanRiding":false,"penaltyInBanRiding":0,"izPopup":true,"izEnable":true,"izAccumulate":false,"ladderItem":null,"updatedAt":"2026-01-05 09:43:54","updatedPin":"b_4b360bb024811db"}}`

	merged := map[string]interface{}{
		"traceId": nil, "tenantId": nil, "platform": nil, "deviceId": nil, "version": nil,
		"ip": nil, "longitude": nil, "latitude": nil, "source": nil, "stressTesting": false,
		"id": "339618167360852753", "serviceId": "338362359727786125",
		"freeDistance": 5, "freeTime": 60000, "freeTimes": 3,
		"startingDistance": 0, "startingTime": 900000, "startingPrice": 200,
		"discount": float64(1), "overDistanceCostPerMter": 0, "timeOutCostPerMin": 100,
		"allowOutofService": true, "allowInNostop": false, "allowOutofParking": true,
		"timeUnit": 600000, "dispatchCost": 1000,
		"penaltyInNostop": 0, "penaltyOutofService": 2000, "mostMoneyOneday": 6000,
		"type": 0, "allowInBanRiding": false, "penaltyInBanRiding": 0,
		"izPopup": true, "izEnable": true, "izAccumulate": false, "ladderItem": nil,
		"updatedAt": "2026-01-05 09:43:54", "updatedPin": "b_4b360bb024811db",
	}
	ensureCBillingConfigNullFields(merged)
	goBytes, err := json.Marshal(map[string]interface{}{
		"success": true, "code": "0", "msg": "成功", "data": merged,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !jsondiff.Equal([]byte(java), goBytes) {
		t.Fatalf("Go response should match Java sample after null-field fix:\n%s", string(goBytes))
	}
}

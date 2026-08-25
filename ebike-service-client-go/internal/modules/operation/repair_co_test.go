package operation

import (
	"strings"
	"testing"

	"ebike-service-client-go/internal/pkg/jsondiff"
)

func TestConvertRepairRecordPageDTO(t *testing.T) {
	raw := []byte(`{"count":1,"pageNum":1,"pageSize":10,"orders":null,"searchCount":true,"list":[{"carId":"100600136","state":1,"bikeType":0,"reportName":"test","reportPhone":"+86-15218125585","reportTime":"2026-01-31 15:39:17","reportDesc":"desc","repairPart":[69139954125511978,69140282690511876],"repairName":null,"reportPhoto":["https://example.com/a.jpg"],"reportLat":22.9748,"reportLng":115.353,"reportAddress":null}]}`)
	out := convertRepairRecordPageDTO(raw)
	if !strings.Contains(string(out), `"count":"1"`) {
		t.Fatalf("count should be string long: %s", out)
	}
	if strings.Contains(string(out), `"repairPart":[69139954125511978`) {
		t.Fatalf("repairPart should be string array: %s", out)
	}
	java := `{"count":"1","pageNum":1,"pageSize":10,"orders":null,"searchCount":true,"list":[{"carId":"100600136","state":1,"bikeType":0,"reportName":"test","reportPhone":"+86-15218125585","reportTime":"2026-01-31 15:39:17","reportDesc":"desc","repairPart":["69139954125511978","69140282690511876"],"repairName":null,"reportPhoto":["https://example.com/a.jpg"],"reportLat":22.9748,"reportLng":115.353,"reportAddress":null}]}`
	if !jsondiff.Equal([]byte(java), out) {
		t.Fatalf("repair page mismatch:\n%s", out)
	}
}

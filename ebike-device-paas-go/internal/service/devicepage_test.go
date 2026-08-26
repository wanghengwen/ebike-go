package service

import (
	"testing"

	"ebike-device-paas-go/internal/api/dto"
)

func devOps(carID string, riding int, ops, alarm []int, reportTime int64) map[string]interface{} {
	return map[string]interface{}{
		"carId": carID, "imei": carID, "ridingState": riding,
		"operationState": ops, "alarmState": alarm, "reportTime": reportTime,
	}
}

func TestBusinessFilterDefaultDropsOnShelf(t *testing.T) {
	devices := []map[string]interface{}{
		devOps("c1", 1, []int{}, []int{}, 1),
		devOps("c2", 1, []int{1}, []int{}, 2), // on-shelf -> dropped
	}
	got := businessFilterDevices(devices, dto.DevicePageBusQry{})
	if len(got) != 1 || got[0]["carId"] != "c1" {
		t.Fatalf("expected only c1, got %v", got)
	}
}

func TestBusinessFilterUnion(t *testing.T) {
	devices := []map[string]interface{}{
		devOps("c1", 2, []int{}, []int{}, 1),  // riding match
		devOps("c2", 1, []int{3}, []int{}, 2), // ope match (3)
		devOps("c3", 1, []int{}, []int{}, 3),  // no match
	}
	q := dto.DevicePageBusQry{RidingStates: []int{2}, OperationStates: []int{3}}
	got := businessFilterDevices(devices, q)
	if len(got) != 2 {
		t.Fatalf("union expected 2, got %d (%v)", len(got), got)
	}
}

func TestBusinessFilterIntersection(t *testing.T) {
	devices := []map[string]interface{}{
		devOps("c1", 2, []int{3}, []int{}, 1), // riding=2 AND ope contains 3
		devOps("c2", 2, []int{}, []int{}, 2),  // missing ope 3
	}
	q := dto.DevicePageBusQry{IzStateAnd: true, RidingStates: []int{2}, OperationStates: []int{3}}
	got := businessFilterDevices(devices, q)
	if len(got) != 1 || got[0]["carId"] != "c1" {
		t.Fatalf("intersection expected only c1, got %v", got)
	}
}

func TestBusinessFilterCarIdRange(t *testing.T) {
	devices := []map[string]interface{}{
		devOps("100", 1, []int{}, []int{}, 1),
		devOps("200", 1, []int{}, []int{}, 2),
		devOps("300", 1, []int{}, []int{}, 3),
	}
	q := dto.DevicePageBusQry{MinCarId: "150", MaxCarId: "250"}
	got := businessFilterDevices(devices, q)
	if len(got) != 1 || got[0]["carId"] != "200" {
		t.Fatalf("carId range expected only 200, got %v", got)
	}
}

func TestPaginateDevicePageSortAndSlice(t *testing.T) {
	devices := []map[string]interface{}{
		devOps("a", 1, []int{}, []int{}, 10),
		devOps("b", 1, []int{}, []int{}, 30),
		devOps("c", 1, []int{}, []int{}, 20),
	}
	page := paginateDevicePage(devices, 1, 2, true)
	if page.Count != 3 || len(page.List) != 2 {
		t.Fatalf("count=%d list=%d", page.Count, len(page.List))
	}
	// reportTime desc -> b(30), c(20)
	if *page.List[0].Imei != "b" || *page.List[1].Imei != "c" {
		t.Fatalf("order wrong: %v %v", *page.List[0].Imei, *page.List[1].Imei)
	}
	if page.List[0].CarHelmetState == nil {
		t.Fatalf("carHelmetState should be set when withHelmetState=true")
	}
	page2 := paginateDevicePage(devices, 2, 2, false)
	if page2.List[0].CarHelmetState != nil {
		t.Fatalf("carHelmetState should be nil when withHelmetState=false")
	}
	if len(page2.List) != 1 || *page2.List[0].Imei != "a" {
		t.Fatalf("page2 wrong: %v", page2.List)
	}
}

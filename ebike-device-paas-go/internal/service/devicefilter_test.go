package service

import (
	"testing"

	"ebike-device-paas-go/internal/api/dto"
)

func iptr(v int) *int { return &v }

func TestFilterDevicesOperationStateDefaultDropsOnShelf(t *testing.T) {
	devices := []map[string]interface{}{
		{"imei": "a", "operationState": []int{}},     // kept (no op-state 1)
		{"imei": "b", "operationState": []int{1}},    // dropped (on-shelf)
		{"imei": "c", "operationState": []int{2}},    // kept
		{"imei": "d", "operationState": []int{1, 2}}, // dropped (contains 1)
	}
	got := filterDevices(devices, dto.DevicePageQry{})
	if len(got) != 2 {
		t.Fatalf("kept %d, want 2", len(got))
	}
}

func TestFilterDevicesOperationStateExplicit(t *testing.T) {
	devices := []map[string]interface{}{
		{"imei": "a", "operationState": []int{2}},
		{"imei": "b", "operationState": []int{2, 1}}, // dropped: contains 1
		{"imei": "c", "operationState": []int{3}},
	}
	got := filterDevices(devices, dto.DevicePageQry{OperationState: iptr(2)})
	if len(got) != 1 || got[0]["imei"] != "a" {
		t.Fatalf("got %+v, want only a", got)
	}
}

func TestFilterDevicesRidingAndBatteryAndTag(t *testing.T) {
	devices := []map[string]interface{}{
		{"imei": "a", "ridingState": int64(1), "restBattery": int64(10), "carTagTypeIds": []int{5}},
		{"imei": "b", "ridingState": int64(2), "restBattery": int64(10), "carTagTypeIds": []int{5}}, // wrong riding
		{"imei": "c", "ridingState": int64(1), "restBattery": int64(50), "carTagTypeIds": []int{5}}, // battery too high
		{"imei": "d", "ridingState": int64(1), "restBattery": int64(10), "carTagTypeIds": []int{9}}, // tag mismatch
	}
	q := dto.DevicePageQry{RidingState: iptr(1), RestBattery: iptr(20), CarTagTypeIds: []int{5}}
	got := filterDevices(devices, q)
	if len(got) != 1 || got[0]["imei"] != "a" {
		t.Fatalf("got %+v, want only a", got)
	}
}

func TestFilterDevicesAlarmState(t *testing.T) {
	devices := []map[string]interface{}{
		{"imei": "a", "operationState": []int{}, "alarmState": []int{3}},
		{"imei": "b", "operationState": []int{}, "alarmState": []int{4}},
	}
	got := filterDevices(devices, dto.DevicePageQry{AlarmState: iptr(3)})
	if len(got) != 1 || got[0]["imei"] != "a" {
		t.Fatalf("got %+v, want only a", got)
	}
}

func TestContainsAndIntersects(t *testing.T) {
	if !containsInt([]int{1, 2, 3}, 2) || containsInt([]int{1, 2}, 9) {
		t.Fatal("containsInt wrong")
	}
	if !intersectsInt([]int{1, 2}, []int{2, 3}) || intersectsInt([]int{1}, []int{2}) {
		t.Fatal("intersectsInt wrong")
	}
}

func TestStrOrNil(t *testing.T) {
	if strOrNil("") != nil {
		t.Fatal("empty should be nil")
	}
	if p := strOrNil("x"); p == nil || *p != "x" {
		t.Fatal("non-empty should pass through")
	}
}

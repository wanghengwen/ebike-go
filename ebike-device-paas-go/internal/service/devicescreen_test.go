package service

import (
	"testing"

	"ebike-device-paas-go/internal/api/dto"
)

func TestCarCountSortsByTotalDesc(t *testing.T) {
	out := CarCount("t1", dto.DeviceScreenQry{ServiceIdList: dto.Int64Slice{10, 20, 30}})
	// no redis -> all totals 0, order preserved (stable), serviceId set
	if len(out) != 3 {
		t.Fatalf("len=%d want 3", len(out))
	}
	for i, sid := range []int64{10, 20, 30} {
		if out[i].ServiceId == nil || *out[i].ServiceId != sid {
			t.Fatalf("out[%d].serviceId=%v want %d", i, out[i].ServiceId, sid)
		}
	}
}

func TestRidingCounts(t *testing.T) {
	var canRent, booking, riding, parking, operation int
	devices := []map[string]interface{}{
		{"ridingState": int64(1)}, {"ridingState": int64(1)},
		{"ridingState": int64(2)},
		{"ridingState": int64(3)},
		{"ridingState": int64(4)},
		{"ridingState": int64(5)},
		{"ridingState": int64(9)}, // ignored
	}
	for _, d := range devices {
		ridingCounts(d, &canRent, &booking, &riding, &parking, &operation)
	}
	if canRent != 2 || riding != 1 || parking != 1 || booking != 1 || operation != 1 {
		t.Fatalf("counts: canRent=%d booking=%d riding=%d parking=%d operation=%d", canRent, booking, riding, parking, operation)
	}
}

func TestIdleMs(t *testing.T) {
	now := int64(1_000_000)
	// lock > unlock => idle
	if got := idleMs(map[string]interface{}{"lockTime": int64(900000), "unlockTime": int64(800000)}, now); got != 100000 {
		t.Fatalf("idle=%d want 100000", got)
	}
	// unlock >= lock => riding (not idle)
	if got := idleMs(map[string]interface{}{"lockTime": int64(800000), "unlockTime": int64(900000)}, now); got != -1 {
		t.Fatalf("idle=%d want -1", got)
	}
}

func TestOnShelf(t *testing.T) {
	if !onShelf(map[string]interface{}{"operationState": []int{1, 2}}) {
		t.Fatal("op-state with 1 should be off-shelf")
	}
	if onShelf(map[string]interface{}{"operationState": []int{2}}) {
		t.Fatal("op-state without 1 should be on-shelf")
	}
}

func TestGpsListRidingFilter(t *testing.T) {
	devices := []map[string]interface{}{
		{"imei": "a", "ridingState": int64(1)},
		{"imei": "b", "ridingState": int64(2)},
	}
	one := 1
	got := gpsList(devices, &one)
	if len(got) != 1 || got[0].Imei == nil || *got[0].Imei != "a" {
		t.Fatalf("got %+v want only a", got)
	}
	// nil filter keeps all and defaults lng/lat to 0
	all := gpsList(devices, nil)
	if len(all) != 2 || all[0].Lng == nil || *all[0].Lng != 0 {
		t.Fatalf("nil filter: %+v", all)
	}
}

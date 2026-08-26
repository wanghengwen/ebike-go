package service

import "testing"

func TestRetainAll(t *testing.T) {
	got := retainAll([]string{"a", "b", "c", "b"}, []string{"b", "c", "x"})
	want := []string{"b", "c", "b"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestFilterNearbyKeepsOnlyAvailableOnlineInArea(t *testing.T) {
	devices := []map[string]interface{}{
		// kept: ridingState 1, online, in-area, no no-ride park
		{"imei": "i1", "carId": "c1", "ridingState": int64(1), "isOnline": int64(1)},
		// dropped: riding (ridingState 2)
		{"imei": "i2", "carId": "c2", "ridingState": int64(2), "isOnline": int64(1)},
		// dropped: offline
		{"imei": "i3", "carId": "c3", "ridingState": int64(1), "isOnline": int64(0)},
		// dropped: in a no-ride park
		{"imei": "i4", "carId": "c4", "ridingState": int64(1), "isOnline": int64(1), "noRideParkId": int64(7)},
		// dropped: out of service area
		{"imei": "i5", "carId": "c5", "ridingState": int64(1), "isOnline": int64(1), "isOutofServAera": int64(1)},
		// kept
		{"imei": "i6", "carId": "c6", "ridingState": int64(1), "isOnline": int64(1)},
	}
	locations := map[string]float64{"i1": 50, "i2": 1, "i3": 1, "i4": 1, "i5": 1, "i6": 10}

	got := filterNearby(devices, locations, defaultNearbyLimit)
	if len(got) != 2 {
		t.Fatalf("kept %d, want 2: %+v", len(got), got)
	}
	// sorted by distance ascending: i6 (10) before i1 (50)
	if got[0].Imei == nil || *got[0].Imei != "i6" {
		t.Fatalf("first = %v, want i6", got[0].Imei)
	}
	if got[1].Imei == nil || *got[1].Imei != "i1" {
		t.Fatalf("second = %v, want i1", got[1].Imei)
	}
	if got[0].Distance == nil || *got[0].Distance != 10 {
		t.Fatalf("i6 distance = %v, want 10", got[0].Distance)
	}
}

func TestFilterNearbyAppliesLimit(t *testing.T) {
	devices := []map[string]interface{}{
		{"imei": "a", "ridingState": int64(1), "isOnline": int64(1)},
		{"imei": "b", "ridingState": int64(1), "isOnline": int64(1)},
		{"imei": "c", "ridingState": int64(1), "isOnline": int64(1)},
	}
	locations := map[string]float64{"a": 1, "b": 2, "c": 3}
	got := filterNearby(devices, locations, 2)
	if len(got) != 2 {
		t.Fatalf("limit not applied: got %d, want 2", len(got))
	}
}

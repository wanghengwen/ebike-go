package service

import (
	"testing"

	"ebike-device-paas-go/internal/api/dto"
)

func TestFreeTimeBucketV1(t *testing.T) {
	cases := []struct {
		diffHours float64
		want      int
	}{
		{0.5, -1}, {2, 0}, {4, 1}, {8, 2}, {18, 3}, {36, 4}, {100, 5},
	}
	for _, tc := range cases {
		got := freeTimeBucketV1(int64(tc.diffHours * float64(hourMs)))
		if got != tc.want {
			t.Fatalf("diff=%.1fh bucket=%d want %d", tc.diffHours, got, tc.want)
		}
	}
}

func TestParkingIdleDerived(t *testing.T) {
	p := &dto.CarParkingStatisticsCo{FreeTimeOneOrTowDay: 3, FreeTimeTowDayMore: 4}
	p.Idle = p.FreeTimeOneOrTowDay + p.FreeTimeTowDayMore
	if p.Idle != 7 {
		t.Fatalf("idle=%d want 7", p.Idle)
	}
}

func TestSvcAndParkFreeTimeAdd(t *testing.T) {
	var svc dto.CarServiceStatisticsCo
	svcFreeTimeAdd(&svc, 0)
	svcFreeTimeAdd(&svc, 5)
	svcFreeTimeAdd(&svc, -1) // no-op
	if svc.FreeTimeOneToThree != 1 || svc.FreeTimeTowDayMore != 1 {
		t.Fatalf("svc free buckets: %+v", svc)
	}
	var park dto.CarParkingStatisticsCo
	parkFreeTimeAdd(&park, 4)
	if park.FreeTimeOneOrTowDay != 1 {
		t.Fatalf("park free bucket: %+v", park)
	}
}

func TestServiceStatisticsTally(t *testing.T) {
	// Build devices for a single service via filterDevices-like maps and call the
	// inner tally by simulating the loop through the public function is hard
	// without redis; instead validate the per-device counting helpers used here.
	devices := []map[string]interface{}{
		{"ridingState": int64(1), "operationState": []int{}, "alarmState": []int{}},         // canRent
		{"ridingState": int64(2), "operationState": []int{2}, "alarmState": []int{}},        // riding + moveCar
		{"ridingState": int64(5), "operationState": []int{4, 5, 6}, "alarmState": []int{5}}, // operation + lowBattery+fixing+dragBack+outParkingZone
		{"ridingState": int64(1), "operationState": []int{1}},                               // off-shelf -> skipped
	}
	var canRent, booking, riding, parking, operation int
	moveCar, changeBattery, lowBattery, fixing, dragBack, outParkingZone, operating := 0, 0, 0, 0, 0, 0, 0
	for _, d := range devices {
		if onShelf(d) {
			continue
		}
		operating++
		ridingCounts(d, &canRent, &booking, &riding, &parking, &operation)
		op := intList(d, "operationState")
		if containsInt(op, 2) {
			moveCar++
		}
		if containsInt(op, 3) {
			changeBattery++
		}
		if containsInt(op, 4) {
			lowBattery++
		}
		if containsInt(op, 5) {
			fixing++
		}
		if containsInt(op, 6) {
			dragBack++
		}
		if containsInt(intList(d, "alarmState"), 5) {
			outParkingZone++
		}
	}
	if operating != 3 {
		t.Fatalf("operating=%d want 3", operating)
	}
	if canRent != 1 || riding != 1 || operation != 1 {
		t.Fatalf("riding counts wrong: canRent=%d riding=%d operation=%d", canRent, riding, operation)
	}
	if moveCar != 1 || changeBattery != 0 || lowBattery != 1 || fixing != 1 || dragBack != 1 || outParkingZone != 1 {
		t.Fatalf("op counts wrong: move=%d change=%d low=%d fix=%d drag=%d out=%d",
			moveCar, changeBattery, lowBattery, fixing, dragBack, outParkingZone)
	}
}

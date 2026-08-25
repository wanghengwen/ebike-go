package event

import (
	"encoding/json"
	"strconv"
	"testing"

	"ebike-open-paas-go/internal/pkg/config"
)

var defaultSteps = []config.SocStep{{Percent: 50, Notify: 21}, {Percent: 30, Notify: 22}}

// socSequence replays SOC readings through socEdge, threading the persisted
// state the way DerivedNotifies does, and returns the notifies that fired.
func socSequence(t *testing.T, socs ...int) []int {
	t.Helper()
	state := map[string]string{}
	var fired []int
	for _, soc := range socs {
		e := socEdge(soc, defaultSteps, state)
		if e.hasNotify {
			fired = append(fired, e.notify)
		}
		if e.write {
			state[e.field] = e.value
		}
	}
	return fired
}

func TestSocFirstReportOnlyRecordsState(t *testing.T) {
	// A device discovered already below 30% must not immediately emit both
	// steps; there is no observed transition to report.
	if fired := socSequence(t, 25); len(fired) != 0 {
		t.Errorf("first report fired %v, want no notify", fired)
	}
}

func TestSocFiresEachStepOnceWhileDischarging(t *testing.T) {
	fired := socSequence(t, 80, 60, 49, 45, 31, 29, 20)
	want := []int{21, 22}
	if len(fired) != len(want) {
		t.Fatalf("fired %v, want %v", fired, want)
	}
	for i := range want {
		if fired[i] != want[i] {
			t.Fatalf("fired %v, want %v", fired, want)
		}
	}
}

func TestSocRechargeRearmsWithoutNotifying(t *testing.T) {
	// Down to 30, back to full, down again: the steps must fire a second time on
	// the new discharge but never on the way up.
	fired := socSequence(t, 80, 45, 25, 100, 45, 25)
	want := []int{21, 22, 21, 22}
	if len(fired) != len(want) {
		t.Fatalf("fired %v, want %v", fired, want)
	}
	for i := range want {
		if fired[i] != want[i] {
			t.Fatalf("fired %v, want %v", fired, want)
		}
	}
}

func TestSocJumpReportsOnlyDeepestStep(t *testing.T) {
	// A reading that skips a threshold reports where the battery actually is,
	// not a synthetic event for each level passed.
	fired := socSequence(t, 80, 20)
	if len(fired) != 1 || fired[0] != 22 {
		t.Errorf("fired %v, want just [22]", fired)
	}
}

// fenceSequence replays (isFenceEnable, isOutofServAera) pairs through fenceEdge.
func fenceSequence(t *testing.T, pairs ...[2]int) []int {
	t.Helper()
	state := map[string]string{}
	var fired []int
	for _, p := range pairs {
		enable, out := float64(p[0]), float64(p[1])
		r := &Record{IsFenceEnable: &enable, IsOutofServAera: &out}
		e := fenceEdge(r, state)
		if e.hasNotify {
			fired = append(fired, e.notify)
		}
		if e.write {
			state[e.field] = e.value
		}
	}
	return fired
}

func TestFenceFirstReportOnlyRecordsState(t *testing.T) {
	// Treating an unknown previous state as "inside" would announce an exit to
	// every third party whenever this service restarts with a cold Redis key.
	if fired := fenceSequence(t, [2]int{1, 1}); len(fired) != 0 {
		t.Errorf("first report fired %v, want no notify", fired)
	}
}

func TestFenceFiresOnCrossingsOnly(t *testing.T) {
	fired := fenceSequence(t,
		[2]int{1, 0}, // first report: inside, recorded
		[2]int{1, 0}, // still inside
		[2]int{1, 1}, // exit
		[2]int{1, 1}, // still outside
		[2]int{1, 0}, // enter
	)
	want := []int{NotifyFenceExit, NotifyFenceEnter}
	if len(fired) != len(want) || fired[0] != want[0] || fired[1] != want[1] {
		t.Errorf("fired %v, want %v", fired, want)
	}
}

func TestFenceIgnoredWhenFencingDisabled(t *testing.T) {
	// The device only maintains the out-of-area bit while fencing is on, so a
	// disabled device must not be read as "inside".
	if fired := fenceSequence(t, [2]int{0, 1}, [2]int{0, 0}); len(fired) != 0 {
		t.Errorf("fired %v with fencing disabled, want none", fired)
	}
}

func TestCurrentSocStepBoundaries(t *testing.T) {
	cases := map[int]int{
		100: socStepAboveAll,
		51:  socStepAboveAll,
		50:  50, // inclusive: "dropped to 50%"
		31:  50,
		30:  30,
		0:   30,
	}
	for soc, want := range cases {
		if got := currentSocStep(soc, defaultSteps); got != want {
			t.Errorf("currentSocStep(%d) = %d, want %d", soc, got, want)
		}
	}
}

func TestNotifyForSocStep(t *testing.T) {
	for percent, want := range map[int]int{50: 21, 30: 22} {
		got, ok := notifyForSocStep(percent, defaultSteps)
		if !ok || got != want {
			t.Errorf("notifyForSocStep(%d) = %d,%v want %d,true", percent, got, ok, want)
		}
	}
	if _, ok := notifyForSocStep(socStepAboveAll, defaultSteps); ok {
		t.Error("socStepAboveAll must not resolve to a notify code")
	}
}

func TestRecordSocPrefersBmsSocOnGpsReports(t *testing.T) {
	soc, bmsSoc := 47.0, 62.0
	gps := &Record{DeviceDataType: typeGPS, Soc: &soc, BmsSoc: &bmsSoc}
	if got := recordSoc(gps); got == nil || *got != 62 {
		t.Errorf("gps soc = %v, want the TLV bmsSoc 62", got)
	}
	bms := &Record{DeviceDataType: typeBMS, Soc: &soc}
	if got := recordSoc(bms); got == nil || *got != 47 {
		t.Errorf("bms soc = %v, want 47", got)
	}
	if got := recordSoc(&Record{DeviceDataType: typePing}); got != nil {
		t.Errorf("ping soc = %v, want nil", got)
	}
}

func TestNotifyForAlarmDropsUnmappedCodes(t *testing.T) {
	if n, ok := NotifyForAlarm(3); !ok || n != NotifyMoveAlarm {
		t.Errorf("alarm 3 = %d,%v want %d,true", n, ok, NotifyMoveAlarm)
	}
	// The two schemes overlap with different meanings, so an unknown internal
	// alarm must be dropped rather than forwarded as a plausible-looking code.
	if _, ok := NotifyForAlarm(999); ok {
		t.Error("alarm 999 must not map to a notify code")
	}
}

// TestDropDisabledStateClearsStaleFields covers the one-feature-off case: the
// surviving feature keeps refreshing the hash TTL, so a field left behind by a
// disabled feature would never expire and a later re-enable would compare against
// a value from before the gap.
func TestDropDisabledStateClearsStaleFields(t *testing.T) {
	prev := map[string]string{fieldOutOfFence: "1", fieldSocStep: "30"}
	dropDisabledState("k", prev, config.NotifyConfig{FenceEnabled: false, SocStepsEnabled: true})
	if _, still := prev[fieldOutOfFence]; still {
		t.Error("the fence field survived with fencing disabled")
	}
	if _, gone := prev[fieldSocStep]; !gone {
		t.Error("the soc field was dropped even though soc steps are enabled")
	}

	prev = map[string]string{fieldOutOfFence: "1", fieldSocStep: "30"}
	dropDisabledState("k", prev, config.NotifyConfig{FenceEnabled: true, SocStepsEnabled: false})
	if _, still := prev[fieldSocStep]; still {
		t.Error("the soc field survived with soc steps disabled")
	}
	if _, gone := prev[fieldOutOfFence]; !gone {
		t.Error("the fence field was dropped even though fencing is enabled")
	}
}

// TestDropDisabledStateLeavesEnabledFeaturesAlone is the common path, where the
// function must not touch Redis at all.
func TestDropDisabledStateLeavesEnabledFeaturesAlone(t *testing.T) {
	prev := map[string]string{fieldOutOfFence: "0", fieldSocStep: "50"}
	dropDisabledState("k", prev, config.NotifyConfig{FenceEnabled: true, SocStepsEnabled: true})
	if len(prev) != 2 {
		t.Errorf("state = %v, want both fields kept", prev)
	}
}

// TestDerivedNotifiesReturnsNothingWithoutRedis: a failed read is not an empty
// state. Treating it as one would make every device look like a first report, and
// the report after Redis recovers would compare against a stale value and
// announce a crossing that never happened.
func TestDerivedNotifiesReturnsNothingWithoutRedis(t *testing.T) {
	enable, out := 1.0, 1.0
	r := &Record{
		DeviceDataType:  typeGPS,
		AppID:           json.Number("1000"),
		Imei:            "865067022403441",
		IsFenceEnable:   &enable,
		IsOutofServAera: &out,
	}
	if got := DerivedNotifies(r); len(got) != 0 {
		t.Errorf("DerivedNotifies with no reachable Redis returned %v, want nothing", got)
	}
}

func TestSocStepStateValueIsParseable(t *testing.T) {
	// The stored value round-trips through Redis as a string; a format change
	// here would silently reset every device's step on deploy.
	e := socEdge(45, defaultSteps, map[string]string{})
	if _, err := strconv.Atoi(e.value); err != nil {
		t.Errorf("socStep value %q is not an integer: %v", e.value, err)
	}
}

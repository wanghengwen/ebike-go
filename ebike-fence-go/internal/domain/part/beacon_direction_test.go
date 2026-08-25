package part

import "testing"

func TestBeaconReturnResult(t *testing.T) {
	ev := 3
	if !BeaconReturnResult(&ev, "addr") {
		t.Fatal("expected pass for event=3 with addr")
	}
	if BeaconReturnResult(&ev, "") {
		t.Fatal("expected fail without addr")
	}
}

func TestDirectionReturnResult(t *testing.T) {
	fence := 90.0
	form := 30.0
	heading := 95.0
	if !DirectionReturnResult(&heading, &fence, &form, false) {
		t.Fatal("expected direction match")
	}
}

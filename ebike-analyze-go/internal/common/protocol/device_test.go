package protocol_test

import (
	"testing"

	"ebike-analyze-go/internal/common/protocol"
)

func TestDecodeDevice_MinimalFields(t *testing.T) {
	raw := make([]byte, 446)
	for i := range raw {
		raw[i] = ' '
	}
	copy(raw[0:15], "867567042500892")
	copy(raw[44:59], "CAR001         ")
	copy(raw[59:78], "1001               ")
	copy(raw[137:147], "30.123456 ")
	copy(raw[147:158], "120.123456 ")
	copy(raw[354:357], "80 ")
	copy(raw[383:385], "1 ")

	d := protocol.DecodeDevice(string(raw))
	if d == nil {
		t.Fatal("expected device")
	}
	if d.Imei != "867567042500892" {
		t.Fatalf("imei=%q", d.Imei)
	}
	if d.CarID != "CAR001" {
		t.Fatalf("carId=%q", d.CarID)
	}
	if d.RidingState == nil || *d.RidingState != 1 {
		t.Fatalf("ridingState=%v", d.RidingState)
	}
	if d.RestBattery == nil || *d.RestBattery != 80 {
		t.Fatalf("restBattery=%v", d.RestBattery)
	}
}

func TestContainsOpState(t *testing.T) {
	states := []int{1, 11}
	if !protocol.ContainsOpState(states, 1) {
		t.Fatal("expected op state 1")
	}
	if protocol.ContainsOpState(states, 2) {
		t.Fatal("did not expect op state 2")
	}
}

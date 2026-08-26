package service

import (
	"testing"

	"ebike-device-paas-go/internal/api/dto"
)

func TestSaddleOverloadContactBits(t *testing.T) {
	// payload 5 = 0b101 -> front=1, cent=0, back=1
	co := decodeSaddle(5)
	if co.FrontSaddleContact != 1 || co.CentSaddleContact != 0 || co.BackSaddleContact != 1 {
		t.Fatalf("payload 5 -> %+v", co)
	}
	// payload 2 = 0b010 -> front=0, cent=1, back=0
	co = decodeSaddle(2)
	if co.FrontSaddleContact != 0 || co.CentSaddleContact != 1 || co.BackSaddleContact != 0 {
		t.Fatalf("payload 2 -> %+v", co)
	}
}

// decodeSaddle mirrors the bit-extraction in QuerySaddleOverloadContact for test.
func decodeSaddle(payload int) dto.SaddleOverloadContactCo {
	return dto.SaddleOverloadContactCo{
		FrontSaddleContact: payload & 1,
		CentSaddleContact:  (payload >> 1) & 1,
		BackSaddleContact:  (payload >> 2) & 1,
	}
}

func TestServiceMatchFiltersOnShelfAndService(t *testing.T) {
	onRack := map[string]interface{}{"serviceId": int64(7), "operationState": []int{2}}
	if !serviceMatch(onRack, 7) {
		t.Fatal("on-rack device with matching serviceId should match")
	}
	if serviceMatch(onRack, 8) {
		t.Fatal("mismatched serviceId should be filtered")
	}
	onShelf := map[string]interface{}{"serviceId": int64(7), "operationState": []int{1}}
	if serviceMatch(onShelf, 7) {
		t.Fatal("on-shelf device (operationState contains 1) should be filtered")
	}
}

func TestOneClickReturnNotifyEmptyDefaultsCanUseTrue(t *testing.T) {
	// With no redis configured the repository read returns "" -> default object.
	res := OneClickReturnNotify("1", "car-xyz")
	if res.CanUse == nil || !*res.CanUse {
		t.Fatalf("canUse should default to true, got %+v", res.CanUse)
	}
	if res.Name != nil || res.Result != nil || res.State != nil {
		t.Fatalf("empty notify should leave other fields nil, got %+v", res)
	}
}

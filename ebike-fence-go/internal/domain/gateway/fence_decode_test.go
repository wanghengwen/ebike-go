package gateway

import "testing"

func TestDecodeFenceJSONNullFieldsNotMarkedSet(t *testing.T) {
	raw := `{
		"id": 1,
		"name": "test",
		"shapeType": "polygon",
		"centerLat": 31.1,
		"centerLng": 118.1,
		"pointList": "[[118.1,31.1]]",
		"type": 2,
		"serviceId": 100,
		"izEnable": true,
		"areaSize": null,
		"minAmount": null,
		"maxAmount": null,
		"bufferDistance": null
	}`
	fe, err := decodeFenceJSON("k", raw)
	if err != nil {
		t.Fatal(err)
	}
	if fe.AreaSizeSet {
		t.Fatal("areaSize null should not mark AreaSizeSet")
	}
	if fe.MinAmountSet || fe.MaxAmountSet {
		t.Fatal("min/max null should not mark amount set flags")
	}
	if fe.BufferDistanceSet {
		t.Fatal("bufferDistance null should not mark BufferDistanceSet")
	}
	if fe.AreaSize != 0 || fe.MinAmount != nil || fe.MaxAmount != nil {
		t.Fatalf("unexpected values: area=%v min=%v max=%v", fe.AreaSize, fe.MinAmount, fe.MaxAmount)
	}
}

func TestDecodeFenceJSONZeroFieldsMarkedSet(t *testing.T) {
	raw := `{
		"id": 1,
		"name": "test",
		"shapeType": "polygon",
		"centerLat": 31.1,
		"centerLng": 118.1,
		"pointList": "[[118.1,31.1]]",
		"type": 2,
		"serviceId": 100,
		"izEnable": true,
		"areaSize": 0,
		"minAmount": 0,
		"maxAmount": 0,
		"bufferDistance": 0
	}`
	fe, err := decodeFenceJSON("k", raw)
	if err != nil {
		t.Fatal(err)
	}
	if !fe.AreaSizeSet || !fe.MinAmountSet || !fe.MaxAmountSet || !fe.BufferDistanceSet {
		t.Fatalf("zero fields should be marked set: area=%v min=%v max=%v buf=%v",
			fe.AreaSizeSet, fe.MinAmountSet, fe.MaxAmountSet, fe.BufferDistanceSet)
	}
}

package service

import (
	"testing"

	"ebike-device-paas-go/internal/api/dto"
)

func TestBuildJobCommandResult_EcuCodeOnly(t *testing.T) {
	cr := buildJobCommandResult(`{"code":"0","result":null}`, nil)
	if cr.EcuCodeValue() != "0" {
		t.Fatalf("ecuCode = %q, want 0", cr.EcuCodeValue())
	}
	if cr.Result != nil {
		t.Fatalf("result = %v, want nil", cr.Result)
	}
	if cr.JobId != nil || cr.Async != nil {
		t.Fatalf("job-state result must clear jobId/async, got jobId=%v async=%v", cr.JobId, cr.Async)
	}
}

func TestBuildJobCommandResult_EmptyResult(t *testing.T) {
	cr := buildJobCommandResult("", mapDefend)
	if cr.EcuCode != nil || cr.Result != nil {
		t.Fatalf("empty result must yield zero CommandResult, got %+v", cr)
	}
}

func TestMapDefend_LonCopiedToLng(t *testing.T) {
	cr := buildJobCommandResult(`{"code":"0","result":{"defend":1,"lon":116.3,"lat":39.9}}`, mapDefend)
	co, ok := cr.Result.(dto.DefendCo)
	if !ok {
		t.Fatalf("result type = %T, want dto.DefendCo", cr.Result)
	}
	if co.Lng == nil || *co.Lng != 116.3 {
		t.Fatalf("lng = %v, want 116.3 (copied from lon)", co.Lng)
	}
	if co.Lat == nil || *co.Lat != 39.9 {
		t.Fatalf("lat = %v, want 39.9", co.Lat)
	}
	if co.Defend == nil || *co.Defend != 1 {
		t.Fatalf("defend = %v, want 1", co.Defend)
	}
}

func TestMapBluetooth(t *testing.T) {
	cr := buildJobCommandResult(`{"code":"0","result":{"token":123,"name":"BLE-1"}}`, mapBluetooth)
	co, ok := cr.Result.(dto.BluetoothCo)
	if !ok {
		t.Fatalf("result type = %T, want dto.BluetoothCo", cr.Result)
	}
	if co.Token == nil || *co.Token != 123 || co.Name == nil || *co.Name != "BLE-1" {
		t.Fatalf("bluetooth co = %+v, want token=123 name=BLE-1", co)
	}
}

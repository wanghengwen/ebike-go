package validator

import "testing"

func TestValidateCmdParams_lockMissingAcc(t *testing.T) {
	r := ValidateCmdParams(33, map[string]interface{}{})
	if r == nil || r.Code != "10002" || r.Msg != "acc不能为空" {
		t.Fatalf("expected acc不能为空, got %+v", r)
	}
}

func TestValidateCmdParams_lockInvalidAcc(t *testing.T) {
	r := ValidateCmdParams(33, map[string]interface{}{"acc": float64(2)})
	if r == nil || r.Code != "10002" || r.Msg != "acc应为0或1" {
		t.Fatalf("expected acc应为0或1, got %+v", r)
	}
}

func TestValidateCmdParams_defendOK(t *testing.T) {
	if r := ValidateCmdParams(4, map[string]interface{}{"defend": float64(1)}); r != nil {
		t.Fatalf("expected nil, got %+v", r)
	}
}

func TestValidateCmdParams_broadcastVoiceMissingIdx(t *testing.T) {
	r := ValidateCmdParams(14, map[string]interface{}{})
	if r == nil || r.Code != "10002" {
		t.Fatalf("expected idx validation error, got %+v", r)
	}
}

func TestValidateCmdParams_restartNoParams(t *testing.T) {
	if r := ValidateCmdParams(21, nil); r != nil {
		t.Fatalf("restart should not require params, got %+v", r)
	}
}

func TestValidateCmdParams_feignLockCloseCar(t *testing.T) {
	// Mini-program 关锁: business -> paas -> openapi /ebike/cmd/lock (acc=0)
	m := map[string]interface{}{
		"acc":   float64(0),
		"carId": "CAR001",
	}
	if r := ValidateCmdParams(33, m); r != nil {
		t.Fatalf("feign lock close payload should pass, got %+v", r)
	}
}

func TestValidateCmdParams_defendEmptyOptionalFields(t *testing.T) {
	m := map[string]interface{}{
		"defend": float64(1),
		"url":    "",
		"type":   "",
	}
	if r := ValidateCmdParams(4, m); r != nil {
		t.Fatalf("blank optional url/type should pass like Java, got %+v", r)
	}
}

func TestValidateCmdParams_lockAccString(t *testing.T) {
	m := map[string]interface{}{"acc": "0"}
	if r := ValidateCmdParams(33, m); r != nil {
		t.Fatalf("string acc should pass, got %+v", r)
	}
}

func TestValidateCmdParams_defendVolumeZero(t *testing.T) {
	m := map[string]interface{}{
		"defend": float64(1),
		"volume": float64(0),
	}
	if r := ValidateCmdParams(4, m); r != nil {
		t.Fatalf("defend volume=0 should pass, got %+v", r)
	}
}

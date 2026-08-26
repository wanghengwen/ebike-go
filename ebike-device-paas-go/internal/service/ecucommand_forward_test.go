package service

import (
	"testing"

	"ebike-device-paas-go/internal/api/dto"
)

func TestPrepCommandBody_DropsContextAddsTrace(t *testing.T) {
	raw := map[string]interface{}{
		"imei":           "860000000000001",
		"acc":            float64(1),
		"commandContext": map[string]interface{}{"tenantId": "9", "traceId": "T-1"},
	}
	cc := &dto.CommandContext{TenantID: "9", TraceID: "T-1"}
	body := prepCommandBody(raw, cc)

	if _, ok := body["commandContext"]; ok {
		t.Fatal("commandContext must be dropped from the gateway body")
	}
	if body["traceId"] != "T-1" {
		t.Fatalf("traceId = %v, want T-1", body["traceId"])
	}
	if body["imei"] != "860000000000001" {
		t.Fatalf("imei = %v", body["imei"])
	}
}

func TestPrepCommandBody_NoTraceWhenEmpty(t *testing.T) {
	body := prepCommandBody(map[string]interface{}{"imei": "x"}, nil)
	if _, ok := body["traceId"]; ok {
		t.Fatal("traceId must be omitted when commandContext is nil")
	}
}

func TestInnerParamBody_SnakeCaseFreq(t *testing.T) {
	raw := map[string]interface{}{
		"imei":     "860000000000001",
		"freqMove": float64(30),
		"freqNorm": float64(60),
	}
	body := innerParamBody(raw, nil)
	if body["freq_move"] != float64(30) {
		t.Fatalf("freq_move = %v, want 30", body["freq_move"])
	}
	if body["freq_norm"] != float64(60) {
		t.Fatalf("freq_norm = %v, want 60", body["freq_norm"])
	}
}

func TestBodyHelpers(t *testing.T) {
	m := map[string]interface{}{
		"acc":      float64(1),
		"second":   float64(120),
		"izAccOn":  true,
		"imei":     "860000000000001",
		"izRiskCo": false,
	}
	if bodyInt(m, "acc") != 1 {
		t.Fatalf("bodyInt acc = %d", bodyInt(m, "acc"))
	}
	if bodyInt64(m, "second") != 120 {
		t.Fatalf("bodyInt64 second = %d", bodyInt64(m, "second"))
	}
	if !bodyBool(m, "izAccOn") {
		t.Fatal("bodyBool izAccOn = false")
	}
	if bodyStr(m, "imei") != "860000000000001" {
		t.Fatalf("bodyStr imei = %q", bodyStr(m, "imei"))
	}
	if bodyInt(m, "missing") != 0 || bodyStr(m, "missing") != "" {
		t.Fatal("missing keys must yield zero values")
	}
}

func TestTempUnLockIzAccOn_DefaultsTrue(t *testing.T) {
	if !tempUnLockIzAccOn(map[string]interface{}{}) {
		t.Fatal("missing izAccOn must default to true (Java field initializer)")
	}
	if !tempUnLockIzAccOn(map[string]interface{}{"izAccOn": nil}) {
		t.Fatal("null izAccOn must default to true")
	}
	if !tempUnLockIzAccOn(map[string]interface{}{"izAccOn": true}) {
		t.Fatal("explicit true")
	}
	if tempUnLockIzAccOn(map[string]interface{}{"izAccOn": false}) {
		t.Fatal("explicit false must stay false")
	}
}

func TestTempUnLockLockRaw_InjectsJavaDefaults(t *testing.T) {
	// ebike-rent only sets imei/second/izAccOn — acc/isIgnoreFence come from Java defaults.
	raw := map[string]interface{}{
		"imei":    "860000000000001",
		"second":  float64(120),
		"izAccOn": true,
	}
	body := tempUnLockLockRaw(raw)
	if body["acc"] != 1 {
		t.Fatalf("acc = %v, want 1", body["acc"])
	}
	if body["isIgnoreFence"] != 1 {
		t.Fatalf("isIgnoreFence = %v, want 1", body["isIgnoreFence"])
	}
	if body["imei"] != "860000000000001" {
		t.Fatalf("imei = %v", body["imei"])
	}
	// Must not mutate the caller's map.
	if _, ok := raw["acc"]; ok {
		t.Fatal("tempUnLockLockRaw must not mutate raw")
	}
}

func TestTempUnLockLockRaw_PreservesExplicitValues(t *testing.T) {
	raw := map[string]interface{}{
		"acc":           float64(0),
		"isIgnoreFence": float64(0),
	}
	body := tempUnLockLockRaw(raw)
	if body["acc"] != float64(0) {
		t.Fatalf("explicit acc=0 must be preserved, got %v", body["acc"])
	}
	if body["isIgnoreFence"] != float64(0) {
		t.Fatalf("explicit isIgnoreFence=0 must be preserved, got %v", body["isIgnoreFence"])
	}
}

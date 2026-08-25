package tps

import (
	"encoding/json"
	"testing"
)

func TestParseConfigFromJSONString(t *testing.T) {
	raw := `{"appId":"id1","appKey":"key1","idCardAuthUrl":"http://example.com","faceMatchUrl":"http://example.com/face"}`
	var cfg ChuangLanAuthParam
	if err := parseConfig(raw, &cfg); err != nil {
		t.Fatalf("parseConfig failed: %v", err)
	}
	if cfg.AppId != "id1" || cfg.AppKey != "key1" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestParseConfigFromMap(t *testing.T) {
	raw := map[string]interface{}{
		"regionId":   "cn-north-4",
		"apiKey":     "ak",
		"secretKey":  "sk",
		"matchScore": 80.0,
	}
	var cfg HuaWeiAuthParam
	if err := parseConfig(raw, &cfg); err != nil {
		t.Fatalf("parseConfig failed: %v", err)
	}
	if cfg.ApiKey != "ak" || cfg.SecretKey != "sk" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestAuthResponseSerialization(t *testing.T) {
	// Case 1: Empty Score and Info -> Score defaults to "0", Info omitted
	r1 := AuthResponse{IzSame: true}
	b1, err := json.Marshal(r1)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	expected1 := `{"score":"0","izSame":true}`
	if string(b1) != expected1 {
		t.Fatalf("expected %s, got %s", expected1, string(b1))
	}

	// Case 2: Populated Score and empty Info -> Score preserved, Info omitted
	r2 := AuthResponse{IzSame: false, Score: "85.4"}
	b2, err := json.Marshal(r2)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	expected2 := `{"score":"85.4","izSame":false}`
	if string(b2) != expected2 {
		t.Fatalf("expected %s, got %s", expected2, string(b2))
	}

	// Case 3: Populated Info and empty Score -> Score defaults to "0", Info preserved
	r3 := AuthResponse{IzSame: false, Info: "failed"}
	b3, err := json.Marshal(r3)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	expected3 := `{"score":"0","izSame":false,"info":"failed"}`
	if string(b3) != expected3 {
		t.Fatalf("expected %s, got %s", expected3, string(b3))
	}
}

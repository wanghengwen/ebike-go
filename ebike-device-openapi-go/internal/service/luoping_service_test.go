package service

import (
	"testing"

	"ebike-device-openapi-go/internal/api/dto"
	"ebike-device-openapi-go/internal/decode/xiaoan"
)

func TestIsLuopingDeviceClientID(t *testing.T) {
	cases := []struct {
		id   string
		want bool
	}{
		{"1234567890", true},
		{"0000000001", true},
		{"123456789", false},
		{"12345678901", false},
		{"123456789a", false},
		{"ebike-device-openapi-x", false},
		{"", false},
	}
	for _, c := range cases {
		if got := isLuopingDeviceClientID(c.id); got != c.want {
			t.Fatalf("isLuopingDeviceClientID(%q)=%v, want %v", c.id, got, c.want)
		}
	}
}

func TestParseBrptTopic(t *testing.T) {
	deviceId, imei, err := parseBrptTopic("ecu/brpt/ebike/2604000274/868480084776843")
	if err != nil {
		t.Fatal(err)
	}
	if deviceId != "2604000274" || imei != "868480084776843" {
		t.Fatalf("got deviceId=%s imei=%s", deviceId, imei)
	}
}

func TestParseBrptTopicInvalid(t *testing.T) {
	if _, _, err := parseBrptTopic("ecu/pass/ecu/2604000274"); err == nil {
		t.Fatal("expected error for legacy topic")
	}
	if _, _, err := parseBrptTopic("ecu/brpt/ebike/only"); err == nil {
		t.Fatal("expected error for short topic")
	}
}

func TestParseLuopingTopicBrpt(t *testing.T) {
	kind, group, deviceId, err := parseLuopingTopic("ecu/brpt/ebike/10001/861553052509151")
	if err != nil {
		t.Fatal(err)
	}
	if kind != "brpt" || group != "ebike" || deviceId != "10001" {
		t.Fatalf("got %s %s %s", kind, group, deviceId)
	}
}

func TestParamString(t *testing.T) {
	params := map[string]interface{}{"IMEI": "861553052509151"}
	if got := paramString(params, "imei", "IMEI"); got != "861553052509151" {
		t.Fatalf("paramString IMEI fallback failed: %q", got)
	}
	params2 := map[string]interface{}{"imei": "  123  "}
	if got := paramString(params2, "imei", "IMEI"); got != "123" {
		t.Fatalf("paramString trim failed: %q", got)
	}
	if got := paramString(map[string]interface{}{}, "imei"); got != "" {
		t.Fatalf("paramString empty want '', got %q", got)
	}
}

func TestNormalizeToMillis(t *testing.T) {
	cases := []struct {
		in   int64
		want int64
	}{
		{in: 1_700_000_000, want: 1_700_000_000_000},
		{in: 1_700_000_000_000, want: 1_700_000_000_000},
		{in: 0, want: 0},
	}
	for _, c := range cases {
		if got := normalizeToMillis(c.in); got != c.want {
			t.Fatalf("normalizeToMillis(%d)=%d, want %d", c.in, got, c.want)
		}
	}
}

func TestPresenceMillis(t *testing.T) {
	if got := presenceMillis(0, 1_700_000_000, 1_800_000_000); got != 1_700_000_000_000 {
		t.Fatalf("presenceMillis picked wrong candidate: %d", got)
	}
	if got := presenceMillis(0, 0); got <= 0 {
		t.Fatalf("presenceMillis fallback should be positive, got %d", got)
	}
}

func TestDeviceTypeConstants(t *testing.T) {
	if dto.DeviceTypeLuoping != "luoping" {
		t.Fatalf("unexpected type %s", dto.DeviceTypeLuoping)
	}
}

func TestExtractLuopingLoginInfoFromDecoded(t *testing.T) {
	decoded := map[string]interface{}{
		"imsi":       "460081944102200",
		"version":    float64(0x0003000B),
		"deviceType": float64(100),
	}
	info := extractLuopingLoginInfo(nil, decoded)
	if info.Imsi != "460081944102200" {
		t.Fatalf("imsi=%s", info.Imsi)
	}
	if info.Version == nil || *info.Version != 0x0003000B {
		t.Fatalf("version=%v", info.Version)
	}
	if info.DeviceType == nil || *info.DeviceType != 100 {
		t.Fatalf("deviceType=%v", info.DeviceType)
	}
}

func TestParseXiaoanFirmwareVer(t *testing.T) {
	v := parseXiaoanFirmwareVer("20.4.1")
	if v == nil || *v != 0x00140401 {
		t.Fatalf("got %v", v)
	}
	if parseXiaoanFirmwareVer("1.2") != nil {
		t.Fatal("expected nil for incomplete ver")
	}
}

func TestLuopingPendingKey(t *testing.T) {
	if got := luopingPendingKey("imei1", 9); got != "mqtt:pending:imei1:9" {
		t.Fatalf("got %s", got)
	}
}

func TestLoginInfoFromBin35(t *testing.T) {
	msg := &xiaoan.Bin35LoginMessage{
		Imei:       "864423069952637",
		Imsi:       "460081944102200",
		Version:    0, // zero must still be kept (not treated as missing)
		DeviceType: 100,
	}
	info := loginInfoFromBin35(msg)
	if info.Version == nil || *info.Version != 0 {
		t.Fatalf("version=%v", info.Version)
	}
	if info.DeviceType == nil || *info.DeviceType != 100 {
		t.Fatalf("deviceType=%v", info.DeviceType)
	}
	if info.Imsi != "460081944102200" {
		t.Fatalf("imsi=%s", info.Imsi)
	}
}

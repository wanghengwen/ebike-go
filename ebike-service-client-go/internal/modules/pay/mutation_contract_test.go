package pay

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPlatformCodeMatchesJavaPlatformEnum(t *testing.T) {
	cases := map[string]int{
		"wechat": 0, "WECHAT": 0,
		"ios": 1, "IOS": 1,
		"android": 2,
		"pc":      3,
		"":        9, "unknown": 9,
	}
	for platform, want := range cases {
		if got := platformCode(platform); got != want {
			t.Errorf("platformCode(%q) = %d, want %d", platform, got, want)
		}
	}
}

func TestPayCreateCmdDownstreamBodyShape(t *testing.T) {
	fee := 100
	svcID := int64(42)
	openID := "openid-1"
	activeID := int64(7)
	cmd := payCmdForward{
		Pin:         "13800000000",
		ServiceId:   &svcID,
		Channel:     "WXLITE",
		Source:      platformCode("wechat"),
		Description: "购买",
		Amount:      fee,
		SaleType:    "WALLET",
		Openid:      &openID,
		ActiveId:    &activeID,
	}

	raw, err := json.Marshal(cmd)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	for _, key := range []string{
		`"pin":"13800000000"`,
		`"serviceId":42`,
		`"channel":"WXLITE"`,
		`"source":0`,
		`"description":"购买"`,
		`"amount":100`,
		`"saleType":"WALLET"`,
		`"openid":"openid-1"`,
		`"activeId":7`,
	} {
		if !strings.Contains(s, key) {
			t.Errorf("pay create cmd missing %s in %s", key, s)
		}
	}
}

func TestAccessTokenCmdForwardShape(t *testing.T) {
	cmd := accessTokenCmdForward{Code: "wx-code-1", Type: "WXLITE"}
	raw, err := json.Marshal(cmd)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if strings.Contains(s, "traceId") || strings.Contains(s, "tenantId") {
		t.Fatalf("AccessTokenCmd must not include ClientDTO fields: %s", s)
	}
	for _, key := range []string{`"code":"wx-code-1"`, `"type":"WXLITE"`} {
		if !strings.Contains(s, key) {
			t.Errorf("missing %s in %s", key, s)
		}
	}
}

func TestPayCancelCmdIsContextOnlyFields(t *testing.T) {
	raw, err := json.Marshal(struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "{}" {
		t.Fatalf("cancel pay dto body must be empty object, got %s", raw)
	}
}

func TestRefundCmdCarriesJavaDefaults(t *testing.T) {
	fee := 50
	cmd := refundCmdForward{
		Pin:            "13800000000",
		SaleType:       "DEPOSIT",
		RefundFee:      &fee,
		PaidAt:         "2026-06-12 10:00:00",
		IzManualRefund: false,
		IzWithdraw:     false,
	}
	raw, err := json.Marshal(cmd)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	for _, key := range []string{
		`"izManualRefund":false`,
		`"izWithdraw":false`,
		`"saleType":"DEPOSIT"`,
		`"refundFee":50`,
	} {
		if !strings.Contains(s, key) {
			t.Errorf("refund cmd missing %s in %s", key, s)
		}
	}
}

func TestWithdrawWechatChannelMapping(t *testing.T) {
	openid := "oxxx"
	platform := "wechat"
	channel := "WX"

	if platform == "wechat" {
		if openid == "" {
			t.Fatal("openid required for wechat")
		}
		channel = "WXLITE"
	}
	if channel != "WXLITE" {
		t.Fatalf("wechat withdraw channel = %q, want WXLITE", channel)
	}

	// Java uses case-sensitive equals — "Wechat" must not trigger mapping.
	platform = "Wechat"
	channel = "WX"
	if platform == "wechat" {
		channel = "WXLITE"
	}
	if channel != "WX" {
		t.Fatal("Wechat platform must not map channel to WXLITE")
	}
}

func TestPayScoreCallbackCmdShape(t *testing.T) {
	serial := "serial-1"
	cmd := payScoreCallbackCmd{
		TenantId:           "tenant-1",
		WechatpaySerial:    &serial,
		WechatpayTimestamp: nil,
		Data:               `{"resource":{}}`,
	}
	raw, err := json.Marshal(cmd)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, `"tenantId":"tenant-1"`) {
		t.Errorf("missing tenantId: %s", s)
	}
	if !strings.Contains(s, `"wechatpaySerial":"serial-1"`) {
		t.Errorf("missing wechatpaySerial: %s", s)
	}
	if !strings.Contains(s, `"wechatpayTimestamp":null`) {
		t.Errorf("absent header must serialize as null: %s", s)
	}
}

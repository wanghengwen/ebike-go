package rediskeys

import "testing"

// Production Redis samples (2026-07) — must round-trip through Go helpers and match Java FenceRedisKey.format.

func TestProductionSampleKeysMatchJava(t *testing.T) {
	type sample struct {
		key  string
		want string
	}
	tests := []sample{
		{"fence_parking_1000_247495733790908225", Parking("1000", 247495733790908225)},
		{"fence_parking_1000_264124897973247394", Parking("1000", 264124897973247394)},
		{"fence_parking_1001_244314271126133454", Parking("1001", 244314271126133454)},
		{"fence_parking_1001_235793921446779582", Parking("1001", 235793921446779582)},
		{"fence_parking_1001_232445390996644035", Parking("1001", 232445390996644035)},
		{"fence_parking_1000_261927137438078863", Parking("1000", 261927137438078863)},
		{"fence_parking_1001_296766662307743979", Parking("1001", 296766662307743979)},
		{"fence_parking_1000_274866299795412760", Parking("1000", 274866299795412760)},
		{"fence_parking_1001_235470201171745886", Parking("1001", 235470201171745886)},
		{"fence_parking_1000_274866299795413086", Parking("1000", 274866299795413086)},
		{"fence_parking_1000_274866299795412133", Parking("1000", 274866299795412133)},
		{"fence_parking_1001_235462251187282089", Parking("1001", 235462251187282089)},
		{"fence_parking_1000_274866299795413910", Parking("1000", 274866299795413910)},
		{"fence_noParking_1000_274872774458608393", NoParking("1000", 274872774458608393)},
		{"fence_noParking_1000_246515184167291707", NoParking("1000", 246515184167291707)},
		{"parking_deletail_1000_100600155", ParkingDetail("1000", "100600155")},
		{"parking_deletail_1001_100600183", ParkingDetail("1001", "100600183")},
		{"parking_deletail_1001_100600974", ParkingDetail("1001", "100600974")},
		{"parking_deletail_1004_100600061", ParkingDetail("1004", "100600061")},
		{"config_customer_service_2_227596718664324575", CustomerService("2", 227596718664324575)},
		{"config_customer_service_1003_338359774158002104", CustomerService("1003", 338359774158002104)},
		{"config_home_scroll_msg_1001_229631747183616067", HomeScrollMsg("1001", 229631747183616067)},
		{"config_home_scroll_msg_1002_234817857916508739", HomeScrollMsg("1002", 234817857916508739)},
		{"ad_config_1001_170792134044032003", AdConfig("1001", 170792134044032003)},
	}
	for _, tt := range tests {
		if tt.key != tt.want {
			t.Fatalf("key %q helper = %q", tt.key, tt.want)
		}
	}
}

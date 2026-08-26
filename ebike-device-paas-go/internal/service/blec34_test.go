package service

import "testing"

func ip(v int) *int { return &v }

func TestC34LockAccOn(t *testing.T) {
	r := c34Lock(ip(1), 116.4, 39.9)
	if r["acc"] != 1 {
		t.Fatalf("acc = %v, want 1", r["acc"])
	}
	if r["defend"] != 0 || r["wheelLock"] != 0 {
		t.Fatalf("acc==1 should zero defend/wheelLock, got defend=%v wheelLock=%v", r["defend"], r["wheelLock"])
	}
	gps, ok := r["gps"].(map[string]interface{})
	if !ok || gps["lng"] != 116.4 || gps["lat"] != 39.9 {
		t.Fatalf("gps lng/lat must echo the phone fix, got %v", r["gps"])
	}
}

func TestC34LockAccOffNoDefend(t *testing.T) {
	r := c34Lock(ip(0), 116.4, 39.9)
	if _, ok := r["defend"]; ok {
		t.Fatalf("acc==0 must not set defend, got %v", r["defend"])
	}
}

func TestC34DefendOn(t *testing.T) {
	r := c34Defend(ip(1), 116.4, 39.9)
	if r["defend"] != 1 || r["acc"] != 0 || r["wheelLock"] != 1 {
		t.Fatalf("defend==1 mapping wrong: %v", r)
	}
}

func TestC34DeviceInfoVoltageMv(t *testing.T) {
	r := c34DeviceInfo(&dtoBleDeviceInfo{Voltage: ip(72), Acc: ip(1)})
	if r["voltageMv"] != 720 {
		t.Fatalf("voltageMv = %v, want 720 (voltage*10)", r["voltageMv"])
	}
	if r["acc"] != 1 {
		t.Fatalf("acc = %v, want 1", r["acc"])
	}
	if _, ok := r["gps"]; !ok {
		t.Fatalf("gps block missing")
	}
}

func TestC34LocationKeys(t *testing.T) {
	r := c34Location(116.4, 39.9)
	gps, ok := r["gps"].(map[string]interface{})
	if !ok {
		t.Fatalf("gps block missing: %v", r)
	}
	for _, k := range []string{"wgs84Lat", "wgs84Lng", "lng", "lat", "timestamp"} {
		if _, ok := gps[k]; !ok {
			t.Fatalf("gps missing key %q: %v", k, gps)
		}
	}
}

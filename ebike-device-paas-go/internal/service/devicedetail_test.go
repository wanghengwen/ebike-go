package service

import "testing"

func TestNumberToVersion(t *testing.T) {
	// 7.8.105 -> (7<<16)|(8<<8)|105 = 460905
	raw := int64(7<<16 | 8<<8 | 105)
	got := numberToVersion(strconvItoa(raw))
	if got == nil || *got != "7.8.105" {
		t.Fatalf("numberToVersion(%d) = %v, want 7.8.105", raw, got)
	}
	if numberToVersion("not-a-number") != nil {
		t.Errorf("non-numeric version should yield nil")
	}
	if numberToVersion(nil) != nil {
		t.Errorf("nil version should yield nil")
	}
}

func TestCarHelmetState(t *testing.T) {
	cases := []struct {
		name string
		d    map[string]interface{}
		want int
	}{
		{"not bind", map[string]interface{}{"isCarBindHelmet": 0}, helmetNotBind},
		{"lock warn", map[string]interface{}{"isCarBindHelmet": 1, "helmetLock": 1, "helmetReact": 0}, helmetLockWarn},
		{"riding not wear", map[string]interface{}{"isCarBindHelmet": 1, "ridingState": 2, "helmetBind": 1, "helmetState": 0}, helmetRidingNotWear},
		{"riding wear (detected)", map[string]interface{}{"isCarBindHelmet": 1, "ridingState": 3, "helmetBind": 1, "helmetState": 1}, helmetWear},
		{"riding wear (no detect, both 0)", map[string]interface{}{"isCarBindHelmet": 1, "ridingState": 2, "helmetBind": 0, "helmetLock": 0, "helmetReact": 0}, helmetWear},
		{"not riding lost", map[string]interface{}{"isCarBindHelmet": 1, "ridingState": 1, "helmetLock": 0, "helmetReact": 0}, helmetLost},
		{"react default", map[string]interface{}{"isCarBindHelmet": 1, "ridingState": 1, "helmetLock": 1, "helmetReact": 1}, helmetReactState},
	}
	for _, c := range cases {
		if got := carHelmetState(c.d); got != c.want {
			t.Errorf("%s: carHelmetState = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestMapDeviceDetailDefaults(t *testing.T) {
	// Empty device (no fields) exercises the DO getter defaults.
	co := mapDeviceDetail(map[string]interface{}{})

	if co.Defend == nil || *co.Defend != 1 {
		t.Errorf("defend default = %v, want 1", co.Defend)
	}
	if co.Acc == nil || *co.Acc != 0 {
		t.Errorf("acc default = %v, want 0", co.Acc)
	}
	if co.BatteryLock == nil || *co.BatteryLock != 1 {
		t.Errorf("batteryLock default = %v, want 1", co.BatteryLock)
	}
	if co.BackWheelLock == nil || *co.BackWheelLock != 1 {
		t.Errorf("backWheelLock default = %v, want 1", co.BackWheelLock)
	}
	if co.HelmetLock == nil || *co.HelmetLock != 0 {
		t.Errorf("helmetLock default = %v, want 0", co.HelmetLock)
	}
	if co.Lat == nil || *co.Lat != 0 {
		t.Errorf("lat default = %v, want 0", co.Lat)
	}
	// list fields must be empty arrays, never nil
	if co.OperationState == nil || co.AlarmState == nil || co.CarTagTypeIds == nil {
		t.Errorf("list fields must be non-nil empty slices")
	}
	// absent non-defaulted field stays nil (serialized as null)
	if co.Course != nil {
		t.Errorf("course should be nil when absent")
	}
	// not bound helmet -> state 1
	if co.CarHelmetState == nil || *co.CarHelmetState != helmetNotBind {
		t.Errorf("carHelmetState = %v, want %d", co.CarHelmetState, helmetNotBind)
	}
}

func TestMapDeviceDetailHelmetSocCap(t *testing.T) {
	co := mapDeviceDetail(map[string]interface{}{"helmetSOC": 150})
	if co.HelmetSOC == nil || *co.HelmetSOC != 100 {
		t.Fatalf("helmetSOC cap = %v, want 100", co.HelmetSOC)
	}
}

// strconvItoa avoids importing strconv just for the test helper.
func strconvItoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

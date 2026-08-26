package protocol

import (
	"reflect"
	"testing"
)

// sampleDevice is a realistic record derived from a real production log line
// (DeviceInfoController.getDeviceDetail -> DeviceDetailCo). It is used to verify
// encode/decode round-trip symmetry and field-table offsets.
func sampleDevice() map[string]interface{} {
	return map[string]interface{}{
		"imei":           "862551059863125",
		"imsi":           "460087010307677",
		"version":        int64(460905), // packed Long -> "7.8.105"; raw stored value is numeric
		"deviceType":     100,
		"carId":          "100600251",
		"serviceId":      int64(338362359727786125),
		"tenantId":       "1003",
		"gsmSignal":      31,
		"defend":         0,
		"acc":            1,
		"wgs84Lat":       33.782431,
		"lat":            33.781187,
		"timestamp":      int64(1782442731),
		"speed":          27.0,
		"batteryLock":    1,
		"batteryConnect": 1,
		"backWheelLock":  0,
		"voltage":        47179,
		"isMoving":       1,
		"helmetType":     1,
		"helmetLock":     0,
		"isOnline":       1,
		"soc":            34,
		"restBattery":    95,
		"restMileage":    76,
		"ridingState":    2,
		"operationState": []int{},
		"alarmState":     []int{2, 6},
		"carTagTypeIds":  []int{},
		"helmetBind":     1,
		"rfidAck":        0,
		"batteryId":      int64(340361772253321111),
	}
}

func TestOffsetsContiguous(t *testing.T) {
	off := 0
	for _, f := range Layout() {
		if f.Offset != off {
			t.Fatalf("field %s offset=%d want=%d", f.Name, f.Offset, off)
		}
		off += f.Length
	}
	if TotalLength != off {
		t.Fatalf("TotalLength=%d want=%d", TotalLength, off)
	}
}

func TestBitmapRoundTrip(t *testing.T) {
	// Round-trip only covers indices < 31. Bit indices >= 31 are intentionally
	// excluded: Java's `int 1<<31` sign-extends into bits 32-63 of the Long, so
	// neither Java nor Go round-trips them (decode yields the extra high bits).
	// TestBitmapMatchesJava pins the exact encode output for those boundaries.
	cases := [][]int{
		{},
		{0},
		{2, 6},
		{0, 1, 4, 11},
	}
	for _, c := range cases {
		hex := getOneIndexHexString(c)
		got := getOneIndexes(hex)
		if len(c) == 0 && len(got) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, c) {
			t.Errorf("bitmap round-trip %v -> %q -> %v", c, hex, got)
		}
	}
}

// TestBitmapMatchesJava locks in the exact hex strings Java's
// ProtocolConvertor.getOneIndexHexString produces (int 1<<type, then unsigned
// Long.toHexString) for the boundary indices.
func TestBitmapMatchesJava(t *testing.T) {
	cases := []struct {
		idx  []int
		want string
	}{
		{[]int{0}, "1"},
		{[]int{2, 6}, "44"},
		{[]int{31}, "ffffffff80000000"}, // int 1<<31 sign-extended to Long, unsigned hex
		{[]int{0, 31}, "ffffffff80000001"},
		{[]int{32}, "1"}, // 1<<32 wraps to 1<<0 in Java's int shift
	}
	for _, c := range cases {
		if got := getOneIndexHexString(c.idx); got != c.want {
			t.Errorf("getOneIndexHexString(%v) = %q, want %q", c.idx, got, c.want)
		}
	}
}

// TestDoubleStringMatchesJava pins toStringValue for KDouble fields to Java's
// Double.toString output. The critical cases are whole numbers (Java keeps the
// ".0"; Go's %v would drop it) and zero.
func TestDoubleStringMatchesJava(t *testing.T) {
	speed := byName["speed"]
	cases := []struct {
		in   interface{}
		want string
	}{
		{0.0, "0.0"},
		{27.0, "27.0"},
		{33.782431, "33.782431"},
		{116.404, "116.404"},
		{-0.5, "-0.5"},
		{int64(5), "5.0"}, // defensive: integral value in a Double field
	}
	for _, c := range cases {
		if got := toStringValue(speed, c.in); got != c.want {
			t.Errorf("toStringValue(KDouble, %v) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestLongStringMatchesJava guards against float64-sourced Long/Int values ever
// rendering in scientific form (Java Long.toString is always plain decimal).
func TestLongStringMatchesJava(t *testing.T) {
	serviceID := byName["serviceId"]
	if got := toStringValue(serviceID, int64(338362359727786125)); got != "338362359727786125" {
		t.Errorf("serviceId int64 = %q", got)
	}
	if got := toStringValue(serviceID, float64(1234567890)); got != "1234567890" {
		t.Errorf("serviceId float64 = %q, want plain decimal", got)
	}
}

func TestInitAndDecodeRoundTrip(t *testing.T) {
	device := sampleDevice()
	encoded := InitProtocolData(device)
	if len(encoded) != TotalLength {
		t.Fatalf("encoded length=%d want=%d", len(encoded), TotalLength)
	}

	decoded := Decode(encoded)

	for k, want := range device {
		got, ok := decoded[k]
		if !ok {
			// Empty lists are emitted; everything else present must match.
			t.Errorf("field %s missing after decode", k)
			continue
		}
		if !valuesEqual(want, got) {
			t.Errorf("field %s: got %v (%T) want %v (%T)", k, got, got, want, want)
		}
	}
}

func TestEncodeNotNullSkipsEmpty(t *testing.T) {
	device := map[string]interface{}{
		"imei":  "862551059863125",
		"acc":   1,
		"carId": "", // empty string -> skipped
	}
	ops := EncodeNotNull(device)
	for _, op := range ops {
		if op.Offset == byName["carId"].Offset {
			t.Errorf("empty carId should be skipped, got op %+v", op)
		}
	}
	// imei op must be present and padded to field length.
	imei := byName["imei"]
	found := false
	for _, op := range ops {
		if op.Offset == imei.Offset {
			found = true
			if len(op.Value) != imei.Length {
				t.Errorf("imei value len=%d want=%d", len(op.Value), imei.Length)
			}
		}
	}
	if !found {
		t.Error("imei op missing")
	}
}

func valuesEqual(a, b interface{}) bool {
	if reflect.DeepEqual(a, b) {
		return true
	}
	// Tolerate int vs int64 differences from decode typing.
	af, aok := asFloat(a)
	bf, bok := asFloat(b)
	if aok && bok {
		return af == bf
	}
	return false
}

func asFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}

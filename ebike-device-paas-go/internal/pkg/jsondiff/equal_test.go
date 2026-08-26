package jsondiff

import "testing"

func TestUnorderedKeysArrayMultiset(t *testing.T) {
	a := []byte(`{"data":[{"serviceId":1,"count":2},{"serviceId":3,"count":4}]}`)
	b := []byte(`{"data":[{"serviceId":3,"count":4},{"serviceId":1,"count":2}]}`)

	// strict: different order -> not equal
	if EqualOpts(a, b, Options{}) {
		t.Fatal("strict order should differ")
	}
	// unordered "data": equal as multiset
	if !EqualOpts(a, b, Options{UnorderedKeys: map[string]bool{"data": true}}) {
		t.Fatal("unordered data should match")
	}
}

func TestUnorderedKeysDetectsRealDiff(t *testing.T) {
	a := []byte(`{"data":[{"serviceId":1,"count":2}]}`)
	b := []byte(`{"data":[{"serviceId":1,"count":99}]}`)
	if EqualOpts(a, b, Options{UnorderedKeys: map[string]bool{"data": true}}) {
		t.Fatal("different element values must still differ")
	}
}

func TestUnorderedKeysDifferentLength(t *testing.T) {
	a := []byte(`{"d":[1,2,3]}`)
	b := []byte(`{"d":[1,2]}`)
	if EqualOpts(a, b, Options{UnorderedKeys: map[string]bool{"d": true}}) {
		t.Fatal("different length must differ")
	}
}

func TestUnorderedKeysNestedMultiset(t *testing.T) {
	a := []byte(`{"serviceStatistics":[{"serviceId":1,"canRent":5},{"serviceId":2,"canRent":1}]}`)
	b := []byte(`{"serviceStatistics":[{"serviceId":2,"canRent":1},{"serviceId":1,"canRent":5}]}`)
	if !EqualOpts(a, b, Options{UnorderedKeys: map[string]bool{"serviceStatistics": true}}) {
		t.Fatal("nested unordered list should match")
	}
}

func TestEqual(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{`{"code":"0","data":{"id":123}}`, `{"code":"0","data":{"id":"123"}}`, true},
		{`{"count":10}`, `{"count":"10"}`, true},
		{`{"asc":true}`, `{"asc":false}`, false},
		{`{"x":1.5}`, `{"x":"1.5"}`, true},
		{`{"x":1.5}`, `{"x":"2"}`, false},
		{`[1,2,3]`, `[1,"2",3]`, true},
	}
	for _, tc := range cases {
		if got := Equal([]byte(tc.a), []byte(tc.b)); got != tc.want {
			t.Errorf("Equal(%s, %s) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestEqualOptsIgnoreKeys(t *testing.T) {
	opts := Options{IgnoreKeys: map[string]bool{"reportTime": true, "lat": true}}
	a := `{"imei":"123","reportTime":1000,"lat":33.78}`
	b := `{"imei":"123","reportTime":2000,"lat":33.79}`
	if !EqualOpts([]byte(a), []byte(b), opts) {
		t.Error("expected equal when volatile keys ignored")
	}
	// Without ignore, they differ.
	if Equal([]byte(a), []byte(b)) {
		t.Error("expected not-equal without ignore")
	}
}

func TestEqualOptsFloatTolerance(t *testing.T) {
	opts := Options{FloatTolerance: 1e-5}
	a := `{"lng":118.362729}`
	b := `{"lng":118.362730}`
	if !EqualOpts([]byte(a), []byte(b), opts) {
		t.Error("expected equal within float tolerance")
	}
}

// TestNullEqualsAbsent verifies the Jackson-ALWAYS (Java) vs omitempty (Go)
// representation difference is tolerated: a null-valued key equals an absent key.
func TestNullEqualsAbsent(t *testing.T) {
	// Java emits the null field, Go omits it -> must compare equal.
	java := []byte(`{"imei":"x","rfid":null,"event":1}`)
	goJSON := []byte(`{"imei":"x","event":1}`)
	if !EqualOpts(java, goJSON, Options{}) {
		t.Fatal("null key must equal absent key")
	}
	// Nested null vs absent.
	jn := []byte(`{"gps":{"lat":1.0,"course":null}}`)
	gn := []byte(`{"gps":{"lat":1.0}}`)
	if !EqualOpts(jn, gn, Options{}) {
		t.Fatal("nested null key must equal absent key")
	}
	// A concrete value vs absent is still a real diff.
	if EqualOpts([]byte(`{"a":1,"b":2}`), []byte(`{"a":1}`), Options{}) {
		t.Fatal("value vs absent must differ")
	}
	// A concrete value vs null is still a real diff.
	if EqualOpts([]byte(`{"a":1,"b":2}`), []byte(`{"a":1,"b":null}`), Options{}) {
		t.Fatal("value vs null must differ")
	}
}

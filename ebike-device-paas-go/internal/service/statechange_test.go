package service

import "testing"

func TestServiceIDOf(t *testing.T) {
	if _, ok := serviceIDOf(nil); ok {
		t.Fatalf("nil device must report missing serviceId")
	}
	if _, ok := serviceIDOf(map[string]interface{}{"imei": "1"}); ok {
		t.Fatalf("device without serviceId must report missing")
	}
	sid, ok := serviceIDOf(map[string]interface{}{"serviceId": int64(42)})
	if !ok || sid != 42 {
		t.Fatalf("serviceIDOf = (%d,%v), want (42,true)", sid, ok)
	}
}

func TestInt64Val(t *testing.T) {
	if int64Val(nil) != 0 {
		t.Fatalf("nil -> 0 expected")
	}
	v := int64(7)
	if int64Val(&v) != 7 {
		t.Fatalf("deref expected 7")
	}
}

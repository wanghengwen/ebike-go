package middleware

import (
	"testing"
)

func TestContainsExactMatchOnly(t *testing.T) {
	list := []string{
		"/helpConfig/getHomeScrollerMsgByServiceId/v2",
		"/parking/getById",
	}
	if !contains(list, "/parking/getById") {
		t.Fatal("exact path should match")
	}
	if contains(list, "/helpConfig/getHomeScrollerMsgByServiceId") {
		t.Fatal("non-v2 must not match v2 via prefix")
	}
	if contains(list, "/parking/getById/extra") {
		t.Fatal("longer path must not match via prefix")
	}
	if contains(list, "/parking") {
		t.Fatal("shorter path must not match via reverse prefix")
	}
}

func TestIsJSONEqNullVsMissing(t *testing.T) {
	withNull := []byte(`{"success":true,"data":{"izOn":null}}`)
	missing := []byte(`{"success":true,"data":{}}`)
	if isJSONEq(withNull, missing) {
		t.Fatal("explicit null must not equal missing field")
	}
	sameNull := []byte(`{"success":true,"data":{"izOn":null}}`)
	if !isJSONEq(withNull, sameNull) {
		t.Fatal("identical null fields should match")
	}
}

func TestFloatJSONEqualTight(t *testing.T) {
	if floatJSONEqual(1, 3.5) {
		t.Fatal("diff>1e-5 must not match")
	}
	if !floatJSONEqual(1.000000001, 1.000000002) {
		t.Fatal("tiny drift should match")
	}
}

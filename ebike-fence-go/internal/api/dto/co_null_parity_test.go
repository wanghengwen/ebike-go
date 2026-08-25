package dto

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestResponseCOEmitsNullNotOmit(t *testing.T) {
	co := HomeScrollerMsgCO{
		Id:        1,
		ServiceId: 2,
		IzOn:      nil,
	}
	b, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `"izOn":null`) {
		t.Fatalf("expected izOn:null after omitempty removal, got %s", s)
	}
	if !strings.Contains(s, `"appid":""`) {
		t.Fatalf("expected empty string fields present, got %s", s)
	}
	if !strings.Contains(s, `"updateName":null`) {
		t.Fatalf("expected updateName:null to match Java ConvertorHelper, got %s", s)
	}
}

func TestResourceManagementCONilPointers(t *testing.T) {
	co := ResourceManagementCO{}
	b, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, field := range []string{"id", "name", "startTime", "izLimitTime"} {
		needle := `"` + field + `":null`
		if !strings.Contains(s, needle) {
			t.Fatalf("expected %s in %s", needle, s)
		}
	}
}

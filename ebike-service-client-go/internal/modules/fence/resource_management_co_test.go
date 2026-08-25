package fence

import (
	"strings"
	"testing"
)

func TestConvertResourceManagementListJavaShape(t *testing.T) {
	raw := []byte(`[{"id":240368800351393938,"serviceId":239659900966803395,"startTime":"2024-07-19T00:00:00","endTime":"2099-12-31T23:59:59","exposureCount":43324,"clickCount":234,"clickPerson":105,"updatedAt":"2024-07-19 11:47:28","name":"骑行提示","pageCode":1,"type":4,"status":1,"sort":1}]`)
	out := convertResourceManagementList(raw)
	s := string(out)
	if strings.Contains(s, `"id":240368800351393938`) {
		t.Fatalf("id should be string: %s", s)
	}
	if !strings.Contains(s, `"id":"240368800351393938"`) {
		t.Fatalf("missing string id: %s", s)
	}
	if strings.Contains(s, "T00:00:00") {
		t.Fatalf("datetime should use space separator: %s", s)
	}
	if !strings.Contains(s, `"startTime":"2024-07-19 00:00:00"`) {
		t.Fatalf("bad startTime: %s", s)
	}
	if !strings.Contains(s, `"exposureCount":"43324"`) {
		t.Fatalf("long counters should be strings: %s", s)
	}
}

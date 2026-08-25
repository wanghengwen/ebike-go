package clientconfig

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestConvertHomeNavListPreservesDistinctLargeIDs(t *testing.T) {
	raw := []byte(`[
		{"id":239661638281077064,"serviceId":239659900966803395,"carType":0,"name":"a","icon":"i","jumpPage":"{\"linkUrl\":\"/a\",\"chainType\":0}","izOn":true},
		{"id":239661638281077065,"serviceId":239659900966803395,"carType":0,"name":"b","icon":"i","jumpPage":"{\"linkUrl\":\"/b\",\"chainType\":0}","izOn":true}
	]`)
	out := convertHomeNavList(raw)
	if strings.Contains(string(out), "239661638281077056") {
		t.Fatalf("precision lost: %s", out)
	}
	if !strings.Contains(string(out), "239661638281077064") || !strings.Contains(string(out), "239661638281077065") {
		t.Fatalf("expected distinct ids: %s", out)
	}
	var rows []bHomeNavCO
	if json.Unmarshal(out, &rows) != nil || len(rows) != 2 {
		t.Fatalf("rows = %#v", rows)
	}
	if rows[0].Id == nil || rows[1].Id == nil || *rows[0].Id == *rows[1].Id {
		t.Fatalf("ids collapsed: %#v", rows)
	}
}

func TestConvertHomeScrollerMsgPreservesLargeIDs(t *testing.T) {
	raw := []byte(`{"id":240355586884507544,"serviceId":239659900966803395,"content":"佩戴头盔，安全你我！包车100元/天。","type":1,"appid":null,"skipUrl":"/pagesSub/accountRules/accountRules","params":null,"title":null,"detailTitle":null,"detail":null,"izOn":true,"createdAt":"2024-07-19 10:04:52","updatedPin":"b_3461e478048216e","updatedAt":"2026-05-05 08:54:24","updateName":null}`)
	out := convertHomeScrollerMsg(raw)
	if strings.Contains(string(out), "240355586884507550") {
		t.Fatalf("precision lost on id: %s", out)
	}
	if !strings.Contains(string(out), `"id":"240355586884507544"`) || !strings.Contains(string(out), `"serviceId":"239659900966803395"`) {
		t.Fatalf("expected exact long strings: %s", out)
	}
}

func TestConvertSpecialTipsListPreservesLargeIDs(t *testing.T) {
	raw := []byte(`[{"id":240355586884507544,"serviceId":239659900966803395,"title":"tip","izOn":true}]`)
	out := convertSpecialTipsList(raw)
	if strings.Contains(string(out), "240355586884507550") {
		t.Fatalf("precision lost: %s", out)
	}
	if !strings.Contains(string(out), `"id":"240355586884507544"`) {
		t.Fatalf("expected string long id: %s", out)
	}
}

func TestConvertMainPushListPreservesLargeIDs(t *testing.T) {
	raw := []byte(`[{"cardId":240355586884507544,"ridingCardName":"card","openStartTime":1710000000000,"openEndTime":1710086400000,"createdAt":1710000000000,"state":1}]`)
	out := convertMainPushList(raw)
	if strings.Contains(string(out), "240355586884507550") {
		t.Fatalf("precision lost: %s", out)
	}
	if !strings.Contains(string(out), `"cardId":"240355586884507544"`) {
		t.Fatalf("expected string long cardId: %s", out)
	}
}

func TestFindRawItemByLongId(t *testing.T) {
	items := []byte(`[{"id":240355586884507544,"izOn":true},{"id":1,"izOn":true}]`)
	found := findRawItemByLongId(items, 240355586884507544)
	if found == nil {
		t.Fatal("expected item")
	}
	if !strings.Contains(string(found), "240355586884507544") {
		t.Fatalf("wrong item: %s", found)
	}
}

func TestFilterIzOnArrayPreservesLargeIDs(t *testing.T) {
	raw := []byte(`[{"id":240355586884507544,"serviceId":239659900966803395,"izOn":true},{"id":1,"izOn":false}]`)
	out := filterIzOnArray(raw)
	if strings.Contains(string(out), "240355586884507550") {
		t.Fatalf("precision lost: %s", out)
	}
	if !strings.Contains(string(out), "240355586884507544") {
		t.Fatalf("expected id preserved: %s", out)
	}
}

func TestFilterTimeWindowPreservesLargeIDs(t *testing.T) {
	raw := []byte(`[{"id":240355586884507544,"unlimited":true,"startTime":"2020-01-01 00:00:00","endTime":"2099-01-01 00:00:00"}]`)
	out := filterTimeWindow(raw)
	if !strings.Contains(string(out), "240355586884507544") {
		t.Fatalf("expected id preserved: %s", out)
	}
}

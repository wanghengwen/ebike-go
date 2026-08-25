package dto

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestErrorResultEmitsNullDataAndMsg(t *testing.T) {
	b, _ := json.Marshal(NewErrorResult(CodeIllegalArgument, "carId must not be blank"))
	s := string(b)
	if !strings.Contains(s, `"data":null`) {
		t.Errorf("expected data:null, got %s", s)
	}
	if !strings.Contains(s, `"code":"00004"`) {
		t.Errorf("expected code 00004, got %s", s)
	}

	// downstream msg null must stay null (not "")
	var r Result
	_ = json.Unmarshal([]byte(`{"success":true,"code":"0","msg":null,"data":{"count":"5"}}`), &r)
	b2, _ := json.Marshal(r)
	if !strings.Contains(string(b2), `"msg":null`) {
		t.Errorf("expected msg:null passthrough, got %s", string(b2))
	}
	if !strings.Contains(string(b2), `"count":"5"`) {
		t.Errorf("expected Long-as-string passthrough, got %s", string(b2))
	}
}

func TestOrderItemAscDefaultsTrue(t *testing.T) {
	var items []OrderItem
	_ = json.Unmarshal([]byte(`[{"column":"x"},{"column":"y","asc":false}]`), &items)
	if !items[0].Asc {
		t.Error("asc should default to true like Java OrderItem")
	}
	if items[1].Asc {
		t.Error("explicit asc=false must be kept")
	}
}

func TestBasePageCmdDefaults(t *testing.T) {
	// Simulate controller pre-set defaults + client omitting page fields
	req := PageClientDTO{PageNum: 1, PageSize: 10}
	_ = json.Unmarshal([]byte(`{"traceId":"t","tenantId":"1"}`), &req)
	cmd := ConvertToBasePageCmd(&req, nil)
	if cmd.PageNum != 1 || cmd.PageSize != 10 || !cmd.SearchCount {
		t.Errorf("defaults lost: %+v", cmd)
	}

	_ = json.Unmarshal([]byte(`{"traceId":"t","tenantId":"1","pageNum":3,"pageSize":20,"searchCount":false,"lastRecordId":"abc"}`), &req)
	cmd = ConvertToBasePageCmd(&req, nil)
	if cmd.PageNum != 3 || cmd.PageSize != 20 || cmd.SearchCount || cmd.LastRecordId != "abc" {
		t.Errorf("explicit values lost: %+v", cmd)
	}
}

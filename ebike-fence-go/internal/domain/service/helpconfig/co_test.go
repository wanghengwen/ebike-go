package helpconfig

import (
	"encoding/json"
	"strings"
	"testing"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/infrastructure/persistence/model"
)

func TestToGuidePageCOSerializesZeroFrequency(t *testing.T) {
	falseVal := false
	trueVal := true
	row := model.TConfigGuidePage{
		ID:           373641672081806660,
		ServiceID:    364848372922716182,
		PageNumEsc:   1,
		VisibleRange: 0,
		Frequency:    0,
		IzOn:         &falseVal,
		ByRegister:   &trueVal,
		ByTags:       &falseVal,
	}
	co := ToGuidePageCO(row)
	raw, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, `"frequency":0`) {
		t.Fatalf("expected frequency=0 in JSON, got %s", body)
	}
}

func TestNewSuccessResultGuidePageListKeepsZeroFrequency(t *testing.T) {
	falseVal := false
	rows := []dto.GuidePageConfigCO{{
		Id:           1,
		ServiceId:    2,
		PageNumEsc:   1,
		VisibleRange: 0,
		Frequency:    0,
		IzOn:         &falseVal,
	}}
	result := dto.NewSuccessResult(rows)
	if !strings.Contains(string(result.Data), `"frequency":0`) {
		t.Fatalf("expected frequency=0 in result data, got %s", string(result.Data))
	}
}

func TestToGuidePageCOSerializesZeroChainType(t *testing.T) {
	row := model.TConfigGuidePage{
		ID:        1,
		ServiceID: 2,
		GuidePages: `[{"chainType":0,"linkUrl":"/pagesSub/charge/charge","picUrl":"https://example.com/a.jpg","linkTitle":"","appId":"","param":""}]`,
	}
	co := ToGuidePageCO(row)
	if len(co.GuidePages) != 1 {
		t.Fatalf("expected 1 guide page, got %d", len(co.GuidePages))
	}
	if co.GuidePages[0].ChainType != 0 {
		t.Fatalf("expected chainType=0, got %d", co.GuidePages[0].ChainType)
	}
	raw, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"chainType":0`) {
		t.Fatalf("expected chainType=0 in JSON, got %s", string(raw))
	}
}

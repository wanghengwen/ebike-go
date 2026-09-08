package helpconfig

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	"ebike-fence-go/internal/pkg/timefmt"
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

// Java FaqCO.createdAt is copied from FaqDO and rendered as "yyyy-MM-dd HH:mm:ss";
// leaving it unset produced a zero-value timestamp instead.
func TestToFaqCOFormatsCreatedAt(t *testing.T) {
	created := time.Date(2024, 7, 15, 16, 19, 8, 0, time.FixedZone("CST", 8*3600))
	row := model.TConfigFaq{ID: 1, ServiceID: 2}
	row.CreatedAt = timefmt.FromTime(created)
	co := ToFaqCO(row)
	if co.CreatedAt == nil {
		t.Fatal("createdAt must be populated")
	}
	if *co.CreatedAt != "2024-07-15 16:19:08" {
		t.Fatalf("unexpected createdAt %q", *co.CreatedAt)
	}
	if ToFaqCO(model.TConfigFaq{ID: 1}).CreatedAt != nil {
		t.Fatal("missing createdAt must stay null, not a zero timestamp")
	}
}

// getHomeActivityEntranceByServiceId returns HomeActivityEntranceCO, so the response must
// carry no audit columns and must format the window with a space separator.
func TestToHomeActivityCODropsAuditFields(t *testing.T) {
	start := time.Date(2024, 7, 15, 16, 19, 8, 0, time.FixedZone("CST", 8*3600))
	izOn := true
	row := model.TConfigHomeActivityEntrance{
		ID:        1,
		ServiceID: 2,
		LinkTitle: "活动",
		IzOn:      &izOn,
		StartTime: &start,
	}
	co := ToHomeActivityCO(row)
	if co.StartTime == nil || *co.StartTime != "2024-07-15 16:19:08" {
		t.Fatalf("unexpected startTime %v", co.StartTime)
	}
	if co.EndTime != nil {
		t.Fatalf("expected null endTime, got %v", *co.EndTime)
	}
	raw, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	for _, leaked := range []string{"tenantId", "createdPin", "updatedPin", "izDel", "orderWeights"} {
		if strings.Contains(string(raw), `"`+leaked+`"`) {
			t.Fatalf("model field %q leaked into response: %s", leaked, raw)
		}
	}
}

func TestToGuidePageCOClampsPageNumEsc(t *testing.T) {
	cases := []struct {
		name       string
		guidePages string
		pageNumEsc int
		want       int
	}{
		{name: "empty pages force 1", guidePages: "", pageNumEsc: 0, want: 1},
		{name: "empty array force 1", guidePages: "[]", pageNumEsc: 5, want: 1},
		{name: "zero becomes 1", guidePages: `[{"chainType":0,"linkUrl":"","picUrl":"a","linkTitle":"","appId":"","param":""}]`, pageNumEsc: 0, want: 1},
		{name: "overflow clamps to size", guidePages: `[{"chainType":0,"linkUrl":"","picUrl":"a","linkTitle":"","appId":"","param":""},{"chainType":0,"linkUrl":"","picUrl":"b","linkTitle":"","appId":"","param":""}]`, pageNumEsc: 9, want: 2},
		{name: "in range kept", guidePages: `[{"chainType":0,"linkUrl":"","picUrl":"a","linkTitle":"","appId":"","param":""},{"chainType":0,"linkUrl":"","picUrl":"b","linkTitle":"","appId":"","param":""}]`, pageNumEsc: 2, want: 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			co := ToGuidePageCO(model.TConfigGuidePage{
				ID:         1,
				ServiceID:  2,
				PageNumEsc: tc.pageNumEsc,
				GuidePages: tc.guidePages,
			})
			if co.PageNumEsc != tc.want {
				t.Fatalf("pageNumEsc=%d want %d", co.PageNumEsc, tc.want)
			}
		})
	}
}

func TestToGuidePageCOSerializesZeroChainType(t *testing.T) {
	row := model.TConfigGuidePage{
		ID:         1,
		ServiceID:  2,
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

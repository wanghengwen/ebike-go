package controller

import (
	"testing"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/infrastructure/persistence/model"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

const specialTipsAddJSON = `{
    "id": "",
    "izOn": false,
    "bodyColor": "#333333",
    "buttonColor": "#FFFFFF",
    "title": "案说法",
    "body": "阿斯蒂芬撒",
    "popUpType": 1,
    "izCheckRead": false,
    "frequency": 0,
    "popUpTime": 1,
    "titleColor": "#333333",
    "izButton": true,
    "buttonTextColor": "#333333",
    "serviceId": "364848372922716182",
    "buttonText": "啊发射点发",
    "izSubtitle": false,
    "closePosition": 1,
    "clickEvent": 0,
    "jumpPage": {
        "chainType": "",
        "linkUrl": "",
        "linkTitle": ""
    },
    "subtitleColor": "#333333",
    "visibleRange": 0,
    "byRegister": true,
    "byTags": false,
    "tagIds": "",
    "bgUrl": [
        "https://luoping-upload.oss-cn-shanghai.aliyuncs.com/download/backend/1007/type/20260707/6ee40312-efbb-42ba-954b-a64ae09bed19.png"
    ],
    "bgColor": "custom",
    "commandContext": {"tenantId":"1007","pin":"p1","traceId":"t1"}
}`

func TestBindSpecialTipsAddRequestJSON(t *testing.T) {
	var req dto.SpecialTipsCmd
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(specialTipsAddJSON), &req); err != nil {
		t.Fatalf("bind failed: %v", err)
	}
	if req.ServiceId == nil || *req.ServiceId != 364848372922716182 {
		t.Fatalf("serviceId = %v", req.ServiceId)
	}
	if req.PopUpTime == nil || *req.PopUpTime != 1 {
		t.Fatalf("popUpTime = %v", req.PopUpTime)
	}
	if req.JumpPage == "" {
		t.Fatal("expected jumpPage JSON string")
	}
	if req.BgUrl == "" || !strings.Contains(req.BgUrl, "luoping-upload") {
		t.Fatalf("bgUrl = %q", req.BgUrl)
	}
}

func TestCmdToModelSpecialTipsWithObjectJumpPageAndArrayBgUrl(t *testing.T) {
	var req dto.SpecialTipsCmd
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(specialTipsAddJSON), &req); err != nil {
		t.Fatalf("bind failed: %v", err)
	}
	row, err := cmdToModel[model.TConfigSpecialTips](req)
	if err != nil {
		t.Fatalf("cmdToModel failed: %v", err)
	}
	if row == nil {
		t.Fatal("nil row")
	}
	if row.ServiceID != 364848372922716182 {
		t.Fatalf("serviceID=%d", row.ServiceID)
	}
	if row.PopUpTime != 1 {
		t.Fatalf("popUpTime=%d", row.PopUpTime)
	}
	if row.Frequency != 0 || row.VisibleRange != 0 || row.ClickEvent != 0 {
		t.Fatalf("zero ints not mapped: frequency=%d visibleRange=%d clickEvent=%d",
			row.Frequency, row.VisibleRange, row.ClickEvent)
	}
	if row.JumpPage == "" || row.BgUrl == "" {
		t.Fatalf("jumpPage=%q bgUrl=%q", row.JumpPage, row.BgUrl)
	}
}

func TestSpecialTipsCmdToModelWritesZeroInts(t *testing.T) {
	var req dto.SpecialTipsCmd
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(specialTipsAddJSON), &req); err != nil {
		t.Fatalf("bind failed: %v", err)
	}
	row := specialTipsCmdToModel(req)
	if row.Frequency != 0 || row.VisibleRange != 0 || row.ClickEvent != 0 {
		t.Fatalf("expected zero ints, got frequency=%d visibleRange=%d clickEvent=%d",
			row.Frequency, row.VisibleRange, row.ClickEvent)
	}
}

func TestSpecialTipsCmdToModelEditWithStringID(t *testing.T) {
	const wantID int64 = 987654321012345678
	raw := `{"id":"987654321012345678","serviceId":"364848372922716182","frequency":0,"popUpTime":0,"visibleRange":0,"clickEvent":0,"closePosition":0,"popUpType":1,"title":"t","commandContext":{"tenantId":"1007","pin":"p1","traceId":"t1"}}`
	var req dto.SpecialTipsCmd
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("bind failed: %v", err)
	}
	row := specialTipsCmdToModel(req)
	if row.ID != wantID {
		t.Fatalf("expected ID %d, got %d", wantID, row.ID)
	}
	if row.Frequency != 0 || row.PopUpTime != 0 || row.VisibleRange != 0 || row.ClickEvent != 0 || row.ClosePosition != 0 {
		t.Fatalf("zero ints not mapped: frequency=%d popUpTime=%d visibleRange=%d clickEvent=%d closePosition=%d",
			row.Frequency, row.PopUpTime, row.VisibleRange, row.ClickEvent, row.ClosePosition)
	}
}

func TestSpecialTipsCmdToModelForAddIgnoresClientID(t *testing.T) {
	var req dto.SpecialTipsCmd
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(specialTipsAddJSON), &req); err != nil {
		t.Fatalf("bind failed: %v", err)
	}
	row := specialTipsCmdToModelForAdd(req)
	if row.ID != 0 {
		t.Fatalf("expected ID 0 for add, got %d", row.ID)
	}
}

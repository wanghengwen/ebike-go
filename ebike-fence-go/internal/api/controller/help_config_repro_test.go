package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/infrastructure/persistence/model"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

func init() {
	extra.RegisterFuzzyDecoders()
}

const guidePageActualRequestJSON = `{"commandContext":{"tenantId":"1007","traceId":"178341244555414424055736","pin":"b_3289379801817ad","ip":"172.16.1.159","platform":"pc","deviceId":null,"source":"ebike-service-business","name":"赵坤鹏","stressTesting":false},"id":null,"serviceId":"364881328207300695","guidePages":[{"chainType":3,"linkUrl":"","picUrl":"https://luoping-upload.oss-cn-shanghai.aliyuncs.com/download/backend/1007/type/20260707/e694b2b7-fb76-446c-9aa2-37164e8aad84.png","linkTitle":"","appId":"","param":""}],"allowSuperEsc":false,"pageNumEsc":1,"visibleRange":0,"frequency":1,"izOn":false,"ids":null,"byRegister":true,"byTags":false,"tagIds":""}`

func bindGuidePageRequest(t *testing.T) dto.GuidePageConfigCmd {
	t.Helper()
	extra.RegisterFuzzyDecoders()
	var req dto.GuidePageConfigCmd
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(guidePageActualRequestJSON), &req); err != nil {
		t.Fatalf("bind JSON failed: %v", err)
	}
	if req.ServiceId == nil {
		t.Fatal("expected serviceId to be bound")
	}
	return req
}

func TestGuidePageCmdToModelFromActualRequestJSON(t *testing.T) {
	req := bindGuidePageRequest(t)
	row := guidePageCmdToModel(req)
	if row == nil {
		t.Fatal("guidePageCmdToModel returned nil")
	}
	if row.ServiceID != 364881328207300695 {
		t.Fatalf("expected ServiceID 364881328207300695, got %d", row.ServiceID)
	}
	if row.GuidePages == "" {
		t.Fatal("expected guidePages JSON to be populated")
	}
	if row.PageNumEsc != 1 || row.VisibleRange != 0 || row.Frequency != 1 {
		t.Fatalf("unexpected scalar fields: pageNumEsc=%d visibleRange=%d frequency=%d", row.PageNumEsc, row.VisibleRange, row.Frequency)
	}
	if row.AllowSuperEsc == nil || *row.AllowSuperEsc {
		t.Fatalf("expected allowSuperEsc=false, got %v", row.AllowSuperEsc)
	}
	if row.ByRegister == nil || !*row.ByRegister {
		t.Fatalf("expected byRegister=true, got %v", row.ByRegister)
	}
}

func TestGuidePageCmdToModelForAddIgnoresClientID(t *testing.T) {
	req := bindGuidePageRequest(t)
	badID := int64(12345)
	req.Id = &badID
	row := guidePageCmdToModelForAdd(req)
	if row.ID != 0 {
		t.Fatalf("expected add row ID cleared, got %d", row.ID)
	}
	if row.OrderWeights != 9 {
		t.Fatalf("expected default orderWeights=9, got %d", row.OrderWeights)
	}
}

func TestCmdToModelGuidePageFromActualRequestJSON(t *testing.T) {
	req := bindGuidePageRequest(t)
	row, err := cmdToModel[model.TConfigGuidePage](req)
	if err != nil {
		t.Fatalf("cmdToModel failed: %v", err)
	}
	if row == nil {
		t.Fatal("cmdToModel returned nil row")
	}
	if row.ServiceID != *req.ServiceId {
		t.Errorf("expected ServiceID %d, got %d", *req.ServiceId, row.ServiceID)
	}
}

func TestCmdToModelGuidePageStringServiceIDWithJsoniter(t *testing.T) {
	raw := `{"serviceId":"364881328207300695","pageNumEsc":1,"visibleRange":0,"frequency":1}`
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatal(err)
	}
	row, err := cmdToModel[model.TConfigGuidePage](payload)
	if err != nil {
		t.Fatalf("cmdToModel should accept Jackson string numbers via jsoniter: %v", err)
	}
	if row == nil {
		t.Fatal("cmdToModel returned nil row")
	}
	if row.ServiceID != 364881328207300695 {
		t.Fatalf("expected ServiceID 364881328207300695, got %d", row.ServiceID)
	}
}

func TestTConfigGuidePageUnmarshalAcceptsStringServiceID(t *testing.T) {
	// Custom UnmarshalJSON on TConfigGuidePage uses jsoniter (Jackson-compatible).
	raw := `{"serviceId":"364881328207300695","pageNumEsc":1,"visibleRange":0,"frequency":1}`
	var row model.TConfigGuidePage
	if err := json.Unmarshal([]byte(raw), &row); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if row.ServiceID != 364881328207300695 {
		t.Fatalf("expected ServiceID 364881328207300695, got %d", row.ServiceID)
	}
}

func TestGuidePageCmdToModelEditWithStringID(t *testing.T) {
	const wantID int64 = 987654321012345678
	raw := `{"id":"987654321012345678","serviceId":"364881328207300695","guidePages":[{"chainType":3,"picUrl":"https://example.com/a.png"}],"pageNumEsc":1,"visibleRange":0,"frequency":1,"allowSuperEsc":false,"izOn":false}`
	var req dto.GuidePageConfigCmd
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("bind JSON failed: %v", err)
	}
	if req.Id == nil {
		t.Fatal("expected id to be bound from Jackson string number")
	}
	row := guidePageCmdToModel(req)
	if row.ID != wantID {
		t.Fatalf("expected ID %d, got %d", wantID, row.ID)
	}
	if row.ServiceID != 364881328207300695 {
		t.Fatalf("unexpected serviceID %d", row.ServiceID)
	}
	if row.GuidePages == "" {
		t.Fatal("expected guidePages to be populated")
	}
}

func TestRequireHelpIDMessageFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/helpConfig/editGuidePage", nil)
	if requireHelpID(c, nil) {
		t.Fatal("expected false for nil id")
	}
	var res dto.Result
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Code == nil || *res.Code != dto.CodeIllegalArgument {
		t.Fatalf("code = %v", res.Code)
	}
	if res.Msg == nil || *res.Msg != "id 不能为null" {
		t.Fatalf("msg = %v", res.Msg)
	}
}

func TestBindGuidePageJSONDeleteWithStringID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	raw := `{"commandContext":{"tenantId":"1007","traceId":"t1","pin":"p1"},"id":"987654321012345678","serviceId":"364881328207300695"}`
	req := httptest.NewRequest(http.MethodPost, "/helpConfig/delGuidePage", strings.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	var cmd dto.GuidePageConfigCmd
	if !bindGuidePageJSON(c, &cmd, true) {
		t.Fatalf("bind failed, body: %s", w.Body.String())
	}
	if cmd.Id == nil || *cmd.Id != 987654321012345678 {
		t.Fatalf("id not bound, got %v", cmd.Id)
	}
	if cmd.ServiceId == nil || *cmd.ServiceId != 364881328207300695 {
		t.Fatalf("serviceId not bound, got %v", cmd.ServiceId)
	}
}

func TestBindGuidePageJSONRejectsNullID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	raw := `{"commandContext":{"tenantId":"1007","traceId":"t1","pin":"p1"},"id":null,"serviceId":"364881328207300695"}`
	req := httptest.NewRequest(http.MethodPost, "/helpConfig/delGuidePage", strings.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	var cmd dto.GuidePageConfigCmd
	if bindGuidePageJSON(c, &cmd, true) {
		t.Fatal("expected bind to fail for null id")
	}
	var res dto.Result
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Msg == nil || *res.Msg != "id 不能为null" {
		t.Fatalf("msg = %v", res.Msg)
	}
}

package order

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ebike-service-client-go/internal/api/dto"
	"ebike-service-client-go/internal/pkg/jsondiff"
	"github.com/gin-gonic/gin"
)

func TestBindUserBillConfigQueryRejectsNullServiceId(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/client/order/config/get",
		bytes.NewBufferString(`{"serviceId":null,"userPin":"u1","traceId":"t","tenantId":"1"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	var req userBillConfigQuery
	if bindUserBillConfigQuery(c, &req) {
		t.Fatal("expected binding failure")
	}
	java := `{"success":false,"code":"00004","msg":"serviceId 不能为null","data":null}`
	if !jsondiff.Equal([]byte(java), w.Body.Bytes()) {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
	var resp dto.Result
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Success {
		t.Fatal("expected success=false")
	}
}

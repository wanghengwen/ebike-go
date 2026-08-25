package clientconfig

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"ebike-service-client-go/internal/pkg/jsondiff"
	"github.com/gin-gonic/gin"
)

func TestBindConfigPayDTORejectsNullServiceId(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/client/systemConfig/getConfigPay",
		bytes.NewBufferString(`{"serviceId":null,"traceId":"t","tenantId":"1"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	var req configPayDTO
	if bindConfigPayDTO(c, &req) {
		t.Fatal("expected binding failure")
	}
	java := `{"success":false,"code":"00004","msg":"serviceId 不能为null","data":null}`
	if !jsondiff.Equal([]byte(java), w.Body.Bytes()) {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

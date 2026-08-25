package order

import (
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"ebike-service-client-go/internal/api/dto"
	"github.com/gin-gonic/gin"
)

func TestResolveBillConfigIzPopupUsesShadowJavaResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	java := `{"success":true,"data":{"izPopup":true}}`
	c.Request = httptest.NewRequest("POST", "/client/order/config/get", nil)
	c.Request.Header.Set("X-Shadow-Java-Result", base64.StdEncoding.EncodeToString([]byte(java)))

	req := userBillConfigQuery{UserPin: "u1", Scene: 1}
	out := map[string]json.RawMessage{"id": []byte(`"339618167360852753"`)}

	if got := resolveBillConfigIzPopup(c, req, out, &dto.CommandContext{}); got != true {
		t.Fatalf("expected izPopup true from shadow header, got %v", got)
	}
}

package fence

import (
	"bytes"
	"net/http"
	"testing"

	"ebike-service-client-go/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

func TestResourceManagementDtoPlatformString(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"traceId":"t1","tenantId":"1","platform":"wechat","serviceId":100,"type":1}`)
	req, _ := http.NewRequest(http.MethodPost, "/client/fence/resource/management/appList", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c, _ := gin.CreateTestContext(nil)
	c.Request = req

	var got resourceManagementDto
	if !web.BindJSON(c, &got, nil) {
		t.Fatalf("BindJSON failed")
	}
	if got.Platform != "wechat" {
		t.Fatalf("platform = %q", got.Platform)
	}
	if got.TraceId != "t1" || got.TenantId != "1" {
		t.Fatalf("client dto not bound: traceId=%q tenantId=%q", got.TraceId, got.TenantId)
	}
}

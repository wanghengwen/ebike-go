package clientconfig

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

func TestAdDetailHandlerRejectsInvalidServiceIdShapes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name string
		body string
	}{
		{"null", `{"serviceId":null,"traceId":"t","tenantId":"1"}`},
		{"emptyString", `{"serviceId":"","traceId":"t","tenantId":"1"}`},
		{"missing", `{"traceId":"t","tenantId":"1"}`},
		{"emptyObject", `{}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.POST("/client/ad_config/detail", adDetail)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/client/ad_config/detail", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			var resp dto.Result
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Success {
				t.Fatalf("expected validation error, body=%s", w.Body.String())
			}
		})
	}
}

func TestAdDetailHandlerRejectsNullServiceIdMatchesJava(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/client/ad_config/detail", adDetail)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/client/ad_config/detail", bytes.NewBufferString(`{"serviceId":null,"traceId":"t","tenantId":"1"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	java := `{"success":false,"code":"00004","msg":"serviceId 不能为null","data":null}`
	if !jsondiff.Equal([]byte(java), w.Body.Bytes()) {
		t.Fatalf("Go error response should match Java:\n%s", w.Body.String())
	}
}

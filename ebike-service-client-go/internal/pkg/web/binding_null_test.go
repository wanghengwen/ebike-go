package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ebike-service-client-go/internal/api/dto"
	"github.com/gin-gonic/gin"
)

func TestRequireJSONFieldNotNull(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name string
		body string
		ok   bool
	}{
		{"null", `{"serviceId":null,"traceId":"t","tenantId":"1"}`, false},
		{"emptyString", `{"serviceId":"","traceId":"t","tenantId":"1"}`, false},
		{"missing", `{"traceId":"t","tenantId":"1"}`, false},
		{"zero", `{"serviceId":0,"traceId":"t","tenantId":"1"}`, true},
		{"number", `{"serviceId":123,"traceId":"t","tenantId":"1"}`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/client/ad_config/detail", bytes.NewBufferString(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")
			ok := RequireJSONFieldNotNull(c, "serviceId", "不能为null")
			if ok != tc.ok {
				t.Fatalf("ok=%v want=%v body=%s", ok, tc.ok, w.Body.String())
			}
			if !tc.ok {
				var resp dto.Result
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatal(err)
				}
				if resp.Success || resp.Msg == nil || *resp.Msg != "serviceId 不能为null" {
					t.Fatalf("unexpected error body=%s", w.Body.String())
				}
			}
		})
	}
}

package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/middleware"
	"ebike-fence-go/internal/pkg/web"

	"github.com/gin-gonic/gin"
)

// Before fix: missing lat/lng → binding:"required" rejected with code 00004.
// After fix: missing lat/lng defaults to (0,0), meaning APP GPS unavailable.
// Java also accepts missing lat/lng and processes the request normally.
func TestGetServiceByLocationMissingLatLngAccepted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/serviceArea/getServiceByLocation", handleServiceAreaGetByLocation)

	body, _ := json.Marshal(map[string]interface{}{
		"commandContext": map[string]string{
			"tenantId": "1003",
			"pin":      "test",
			"traceId":  "t1",
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/serviceArea/getServiceByLocation", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var res dto.Result
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	// Should NOT return 00004 (param error) anymore
	if res.Code != nil && *res.Code == dto.CodeIllegalArgument {
		t.Fatalf("should not reject missing lat/lng as param error, body: %s", w.Body.String())
	}
}

func TestBindHelpServiceIDMessageFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/helpConfig/addFaq", nil)
	if requireHelpServiceID(c, nil) {
		t.Fatal("expected false for nil serviceId")
	}
	var res dto.Result
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Code == nil || *res.Code != dto.CodeIllegalArgument {
		t.Fatalf("code = %v", res.Code)
	}
}

func init() {
	gin.SetMode(gin.TestMode)
	_ = middleware.ApplyCommand // keep import if extended later
	_ = web.BindJSON
}

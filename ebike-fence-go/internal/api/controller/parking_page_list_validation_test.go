package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ebike-fence-go/internal/api/dto"

	"github.com/gin-gonic/gin"
)

func TestGetPageListByServiceIdFieldListBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/parking/getPageListByServiceId", handleParkingGetPageListByServiceID)

	// Test passing fieldList as []
	body, _ := json.Marshal(map[string]interface{}{
		"commandContext": map[string]string{
			"tenantId": "1007",
			"pin":      "test",
			"traceId":  "t1",
		},
		"id":        "364848372922716182",
		"pageNum":   1,
		"pageSize":  10,
		"fieldList": []interface{}{},
	})
	req := httptest.NewRequest(http.MethodPost, "/parking/getPageListByServiceId", bytes.NewReader(body))
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

	// It should NOT fail with JSON binding error (code 00002)
	if res.Code != nil && *res.Code == "00002" {
		t.Fatalf("JSON binding failed: %s", w.Body.String())
	}
}

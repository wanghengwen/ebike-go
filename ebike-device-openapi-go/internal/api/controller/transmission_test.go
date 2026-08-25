package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandleTransmission2RawBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/transmission2", handleTransmission2)

	body := map[string]interface{}{
		"imei": "123",
		"cmd":  map[string]interface{}{"c": 11},
	}
	raw, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/transmission2", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"imei应为15位数字"`)) {
		t.Fatalf("expected imei validation error, got %s", w.Body.String())
	}
}

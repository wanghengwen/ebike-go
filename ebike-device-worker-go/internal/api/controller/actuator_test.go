package controller_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ebike-device-worker-go/internal/api/controller"
	"ebike-device-worker-go/internal/config"

	"github.com/gin-gonic/gin"
)

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config.GlobalConfig = &config.Config{}
	config.GlobalConfig.Spring.Xyy.Fastid.Enabled = false
	config.GlobalConfig.Kafka.Enabled = false

	r := gin.New()
	r.GET("/actuator/health", controller.Health)

	req := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		Status string `json:"status"`
		Components map[string]struct {
			Status string `json:"status"`
		} `json:"components"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "UP" {
		t.Fatalf("status=%s", body.Status)
	}
	if body.Components["fastid"].Status != "UP" {
		t.Fatalf("fastid=%v", body.Components["fastid"])
	}
}

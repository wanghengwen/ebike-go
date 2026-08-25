package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMemstats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerActuatorDebug(r.Group(""))

	req := httptest.NewRequest(http.MethodGet, "/actuator/memstats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var resp memstatsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Goroutines <= 0 {
		t.Fatalf("expected goroutines > 0, got %d", resp.Goroutines)
	}
	if resp.HeapSysBytes == 0 {
		t.Fatal("expected non-zero heap sys")
	}
}

func TestPprofHeapRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerActuatorDebug(r.Group(""))

	req := httptest.NewRequest(http.MethodGet, "/actuator/debug/pprof/heap", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct == "" {
		t.Fatal("expected content-type")
	}
}

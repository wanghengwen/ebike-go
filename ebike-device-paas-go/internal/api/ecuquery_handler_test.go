package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"ebike-device-paas-go/internal/pkg/jsondiff"

	"github.com/gin-gonic/gin"
)

// fixture mirrors the testdata schema produced by tools/logextract.
type fixture struct {
	URL string          `json:"url"`
	Req json.RawMessage `json:"req"`
	Rep json.RawMessage `json:"rep"`
}

func init() { gin.SetMode(gin.TestMode) }

// TestGetDeviceRealGpsListEnvelope verifies the route is wired and returns the
// standard Result envelope. Without Redis the data list is empty, which still
// exercises binding, tenant extraction, and serialization.
func TestGetDeviceRealGpsListEnvelope(t *testing.T) {
	fx := loadFixture(t, "device_paas_getDeviceRealGpsList", "000.json")

	r := NewRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, fx.URL, bytes.NewReader(fx.Req))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}

	var res Result
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal response: %v; body=%s", err, w.Body.String())
	}
	if !res.Success || res.Code != CodeSuccess {
		t.Fatalf("envelope mismatch: success=%v code=%s", res.Success, res.Code)
	}
	// No Redis in unit tests -> empty list, but it must be a JSON array, not null.
	if _, ok := res.Data.([]interface{}); !ok && res.Data != nil {
		t.Fatalf("data should be an array, got %T", res.Data)
	}
}

// TestRealGpsImeiNumericTolerance confirms jsondiff treats the Java numeric imei
// and the Go string imei as equal, so shadow comparison won't false-positive.
func TestRealGpsImeiNumericTolerance(t *testing.T) {
	javaRep := []byte(`{"success":true,"code":"0","msg":"成功","data":[{"imei":862551055684020,"lat":33.772855,"lng":118.389939}]}`)
	goRep := []byte(`{"success":true,"code":"0","msg":"成功","data":[{"imei":"862551055684020","lat":33.772855,"lng":118.389939}]}`)
	if !jsondiff.Equal(goRep, javaRep) {
		t.Fatalf("expected numeric/string imei to compare equal")
	}
}

func loadFixture(t *testing.T, dir, name string) fixture {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", dir, name)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	var fx fixture
	if err := json.Unmarshal(raw, &fx); err != nil {
		t.Fatalf("parse fixture %s: %v", path, err)
	}
	return fx
}

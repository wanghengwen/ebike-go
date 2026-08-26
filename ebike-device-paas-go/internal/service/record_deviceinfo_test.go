package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ebike-device-paas-go/internal/api/dto"
	"ebike-device-paas-go/internal/pkg/config"
)

// TestRecordDeviceInfoFixtures_BindAndMap walks testdata/device_paas_deviceInfo
// (including prod RECORD imports) and checks: req binds to DeviceInfoQry, and
// when Java rep carries a result object, mapDeviceInfo accepts it.
func TestRecordDeviceInfoFixtures_BindAndMap(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "device_paas_deviceInfo")
	files, err := filepath.Glob(filepath.Join(root, "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("testdata deviceInfo: %v (n=%d)", err, len(files))
	}

	origCoord := config.GlobalConfig.Coordinate.Type
	config.GlobalConfig.Coordinate.Type = 0
	defer func() { config.GlobalConfig.Coordinate.Type = origCoord }()

	var withResult, bindOnly int
	for _, fp := range files {
		fp := fp
		t.Run(filepath.Base(fp), func(t *testing.T) {
			raw, err := os.ReadFile(fp)
			if err != nil {
				t.Fatal(err)
			}
			var fx struct {
				URL string          `json:"url"`
				Req json.RawMessage `json:"req"`
				Rep json.RawMessage `json:"rep"`
			}
			if err := json.Unmarshal(raw, &fx); err != nil {
				t.Fatal(err)
			}
			if fx.URL != "" && fx.URL != "/device/paas/deviceInfo" {
				t.Fatalf("url=%q", fx.URL)
			}

			var q dto.DeviceInfoQry
			if err := json.Unmarshal(fx.Req, &q); err != nil {
				t.Fatalf("bind req: %v", err)
			}
			if q.Imei == "" {
				t.Fatal("imei empty")
			}
			if q.CommandContext == nil || q.CommandContext.TenantID == "" {
				t.Fatalf("tenant missing: %+v", q.CommandContext)
			}

			var rep struct {
				Success bool `json:"success"`
				Data    *struct {
					EcuCode json.RawMessage `json:"ecuCode"`
					Result  json.RawMessage `json:"result"`
				} `json:"data"`
			}
			if len(fx.Rep) == 0 || json.Unmarshal(fx.Rep, &rep) != nil || rep.Data == nil {
				bindOnly++
				return
			}
			if len(rep.Data.Result) == 0 || string(rep.Data.Result) == "null" {
				bindOnly++
				return
			}

			out := mapDeviceInfo(rep.Data.Result)
			m, ok := out.(map[string]interface{})
			if !ok || m == nil {
				t.Fatalf("mapDeviceInfo type=%T", out)
			}
			if _, has := m["gps"]; has {
				gps, _ := m["gps"].(map[string]interface{})
				if gps != nil && config.GlobalConfig.Coordinate.Type == 0 {
					if _, ok := gps["wgs84Lat"]; ok {
						if gps["lat"] != gps["wgs84Lat"] {
							t.Fatalf("WGS84 lat not applied: lat=%v wgs84=%v", gps["lat"], gps["wgs84Lat"])
						}
					}
				}
			}
			withResult++
		})
	}
	t.Logf("fixtures=%d withResult=%d bindOnly=%d", len(files), withResult, bindOnly)
	if withResult == 0 {
		t.Fatal("expected at least one RECORD fixture with data.result")
	}
}

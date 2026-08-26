package dto

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type logFixture struct {
	URL string          `json:"url"`
	Req json.RawMessage `json:"req"`
}

// endpointDTO maps testdata directory names to the Go request DTO used by bindBody.
var endpointDTO = map[string]func() any{
	"device_paas_device_detail":              func() any { return &DeviceDetailQry{} },
	"device_paas_device_list":                func() any { return &DeviceListQry{} },
	"device_paas_device_listByImeiList":      func() any { return &ImeiListQry{} },
	"device_paas_device_eBikeLocation":       func() any { return &DeviceLocationQry{} },
	"device_paas_device_page":               func() any { return &DevicePageQry{} },
	"device_paas_device_pageBus":             func() any { return &DevicePageBusQry{} },
	"device_paas_device_queryDeviceOpeMap":   func() any { return &DeviceOpeMapQry{} },
	"device_paas_device_queryDeviceScreen":   func() any { return &DeviceScreenQry{} },
	"device_paas_device_queryDeviceScreen_v2": func() any { return &DeviceScreenQry{} },
	"device_paas_device_car_count":           func() any { return &DeviceScreenQry{} },
	"device_paas_device_carStatisticsByService": func() any { return &CarStatisticsByServiceQry{} },
	"device_paas_deviceInfo":                 func() any { return &DeviceInfoQry{} },
	"device_paas_getDeviceRealGpsList":       func() any { return &ImeiListQry{} },
	"device_scanLocation_change":             func() any { return &ScanLocationChangeCmd{} },
	"device_trajectory_history":              func() any { return &TrajectoryHistoryQry{} },
	"device_trajectory_history_batch":        func() any { return &TrajectoryHistoryBatchQry{} },
	"device_trajectory_distance":             func() any { return &MetricQry{} },
	"device_trajectory_saveDb":               func() any { return &TrajectoryDbCmd{} },
}

// ecuCommandDirs are RECORD/proxy paths bound via bindCommand (raw map + CommandContext).
var ecuCommandDirs = []string{
	"device_paas_defend",
	"device_paas_lock",
	"device_paas_bluetooth",
	"device_paas_helmetLock",
	"device_paas_voice",
	"device_paas_setInnerParam",
	"device_paas_replyStopMove",
	"device_paas_dashboard",
}

func TestTestdataReqBind_AllKnownEndpoints(t *testing.T) {
	root := filepath.Join("..", "..", "..", "testdata")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		name := ent.Name()
		newDTO, ok := endpointDTO[name]
		if !ok {
			continue
		}
		dir := filepath.Join(root, name)
		files, _ := filepath.Glob(filepath.Join(dir, "*.json"))
		for _, fp := range files {
			fp := fp
			t.Run(name+"/"+filepath.Base(fp), func(t *testing.T) {
				raw, err := os.ReadFile(fp)
				if err != nil {
					t.Fatal(err)
				}
				var fx logFixture
				if err := json.Unmarshal(raw, &fx); err != nil {
					t.Fatal(err)
				}
				if len(fx.Req) == 0 {
					t.Skip("no req")
				}
				v := newDTO()
				if err := json.Unmarshal(fx.Req, v); err != nil {
					t.Fatalf("bind %s req: %v", fx.URL, err)
				}
				// Every fixture should carry a tenant in commandContext when present.
				if strings.Contains(string(fx.Req), `"commandContext"`) {
					switch q := v.(type) {
					case *DeviceDetailQry:
						assertTenant(t, q.CommandContext)
					case *DeviceListQry:
						assertTenant(t, q.CommandContext)
					case *ImeiListQry:
						assertTenant(t, q.CommandContext)
					case *DeviceLocationQry:
						assertTenant(t, q.CommandContext)
					case *DevicePageQry:
						assertTenant(t, q.CommandContext)
					case *DevicePageBusQry:
						assertTenant(t, q.CommandContext)
					case *DeviceOpeMapQry:
						assertTenant(t, q.CommandContext)
					case *DeviceScreenQry:
						assertTenant(t, q.CommandContext)
					case *CarStatisticsByServiceQry:
						assertTenant(t, q.CommandContext)
					case *DeviceInfoQry:
						assertTenant(t, q.CommandContext)
					case *ScanLocationChangeCmd:
						assertTenant(t, q.CommandContext)
					case *TrajectoryHistoryQry:
						assertTenant(t, q.CommandContext)
					case *TrajectoryHistoryBatchQry:
						assertTenant(t, q.CommandContext)
					case *MetricQry:
						assertTenant(t, q.CommandContext)
					case *TrajectoryDbCmd:
						assertTenant(t, q.CommandContext)
					}
				}
			})
		}
	}
}

func assertTenant(t *testing.T, cc *CommandContext) {
	t.Helper()
	if cc == nil || cc.TenantID == "" {
		t.Fatal("commandContext.tenantId missing")
	}
}

func TestTestdataEcuCommandBind(t *testing.T) {
	root := filepath.Join("..", "..", "..", "testdata")
	for _, name := range ecuCommandDirs {
		dir := filepath.Join(root, name)
		files, err := filepath.Glob(filepath.Join(dir, "*.json"))
		if err != nil || len(files) == 0 {
			t.Fatalf("%s: %v", name, err)
		}
		for _, fp := range files {
			fp := fp
			t.Run(name+"/"+filepath.Base(fp), func(t *testing.T) {
				raw, err := os.ReadFile(fp)
				if err != nil {
					t.Fatal(err)
				}
				var fx logFixture
				if err := json.Unmarshal(raw, &fx); err != nil {
					t.Fatal(err)
				}
				if len(fx.Req) == 0 {
					t.Skip("no req")
				}
				body := map[string]interface{}{}
				if err := json.Unmarshal(fx.Req, &body); err != nil {
					t.Fatalf("bind %s raw: %v", fx.URL, err)
				}
				var ctx struct {
					CommandContext *CommandContext `json:"commandContext"`
				}
				if err := json.Unmarshal(fx.Req, &ctx); err != nil {
					t.Fatalf("bind %s commandContext: %v", fx.URL, err)
				}
				assertTenant(t, ctx.CommandContext)
			})
		}
	}
}

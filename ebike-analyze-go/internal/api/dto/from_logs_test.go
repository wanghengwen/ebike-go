package dto

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type logPathFixture struct {
	Path    string `json:"path"`
	Samples []struct {
		Fingerprint string          `json:"fingerprint"`
		Request     json.RawMessage `json:"request"`
		DataRaw     string          `json:"dataRaw"`
		Success     *bool           `json:"success"`
		Code        string          `json:"code"`
		Source      string          `json:"source"`
	} `json:"samples"`
}

func logFixturesDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Join(filepath.Dir(file), "..", "..", "..", "testdata", "from_logs", "by_path")
	return dir
}

func loadLogFixtures(t *testing.T) []logPathFixture {
	t.Helper()
	dir := logFixturesDir(t)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read fixtures dir: %v", err)
	}
	var out []logPathFixture
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		var fx logPathFixture
		if err := json.Unmarshal(b, &fx); err != nil {
			t.Fatalf("unmarshal %s: %v", e.Name(), err)
		}
		out = append(out, fx)
	}
	if len(out) == 0 {
		t.Fatal("no log fixtures found")
	}
	return out
}

func TestLogFixtures_RequestUnmarshal(t *testing.T) {
	for _, fx := range loadLogFixtures(t) {
		fx := fx
		t.Run(fx.Path, func(t *testing.T) {
			for _, sample := range fx.Samples {
				sample := sample
				t.Run(sample.Fingerprint, func(t *testing.T) {
					switch fx.Path {
					case "/car-statistics/queryByCarList":
						var cmd CarStatisticsListQuery
						if err := json.Unmarshal(sample.Request, &cmd); err != nil {
							t.Fatalf("unmarshal request from %s: %v", sample.Source, err)
						}
						if cmd.ServiceID.Int64() == 0 {
							t.Fatalf("serviceId missing in %s", sample.Source)
						}
						if len(cmd.CarIds) == 0 {
							t.Fatalf("carIds missing in %s", sample.Source)
						}
					case "/orderQuery/selectOrderAnalyze":
						var cmd OrderQueryCmd
						if err := json.Unmarshal(sample.Request, &cmd); err != nil {
							t.Fatalf("unmarshal request from %s: %v", sample.Source, err)
						}
						if cmd.ServiceID == nil || *cmd.ServiceID == 0 {
							t.Fatalf("serviceId missing in %s", sample.Source)
						}
						if len(cmd.StartTime) != 2 {
							t.Fatalf("startTime range missing in %s", sample.Source)
						}
					case "/parking/getParkingInAndOutflow":
						var cmd ParkingInAndOutflowQry
						if err := json.Unmarshal(sample.Request, &cmd); err != nil {
							t.Fatalf("unmarshal request from %s: %v", sample.Source, err)
						}
						if cmd.ParkingID == 0 {
							t.Fatalf("parkingId missing in %s", sample.Source)
						}
					case "/parking/statistics":
						var cmd ParkingStatisticalPageQuery
						if err := json.Unmarshal(sample.Request, &cmd); err != nil {
							t.Fatalf("unmarshal request from %s: %v", sample.Source, err)
						}
						if cmd.ServiceID == 0 || len(cmd.AreaIds) == 0 {
							t.Fatalf("serviceId/areaIds missing in %s", sample.Source)
						}
					case "/riding_card_order/page":
						var cmd OrderQueryPageCmd
						if err := json.Unmarshal(sample.Request, &cmd); err != nil {
							t.Fatalf("unmarshal request from %s: %v", sample.Source, err)
						}
						if cmd.ServiceID == 0 {
							t.Fatalf("serviceId missing in %s", sample.Source)
						}
						if cmd.SearchCount == nil || !*cmd.SearchCount {
							t.Fatalf("searchCount=true expected in %s", sample.Source)
						}
					case "/user/ageStatistic":
						var cmd AgeStatisticCmd
						if err := json.Unmarshal(sample.Request, &cmd); err != nil {
							t.Fatalf("unmarshal request from %s: %v", sample.Source, err)
						}
						if len(cmd.ServiceIds) == 0 {
							t.Fatalf("serviceIds missing in %s", sample.Source)
						}
					case "/userQuery/selectUserCount":
						var cmd UserQueryCmd
						if err := json.Unmarshal(sample.Request, &cmd); err != nil {
							t.Fatalf("unmarshal request from %s: %v", sample.Source, err)
						}
						if len(cmd.ServiceID) == 0 {
							t.Fatalf("serviceId missing in %s", sample.Source)
						}
					default:
						t.Fatalf("unexpected path %s", fx.Path)
					}
				})
			}
		})
	}
}

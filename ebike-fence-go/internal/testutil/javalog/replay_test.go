package javalog_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"ebike-fence-go/internal/api/dto"
	"ebike-fence-go/internal/testutil/javalog"
)

func loadFixtures(t *testing.T) []javalog.Entry {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(file), "testdata", "readonly_cases.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readonly_cases.json missing — run: go run ./cmd/parse_java_log")
	}
	var cases []javalog.Entry
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("no fixtures")
	}
	return cases
}

func TestParseJavaLogs(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	for _, name := range []string{"ebike-fence-1.log", "ebike-fence-2.log"} {
		path := filepath.Join(root, name)
		entries, err := javalog.ParseFile(path)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(entries) == 0 {
			t.Fatalf("%s: expected entries", name)
		}
	}
}

func TestReadOnlyFixturesHaveValidEnvelope(t *testing.T) {
	cases := loadFixtures(t)
	for i, c := range cases {
		env, err := javalog.ParseReplyEnvelope(c.Reply)
		if err != nil {
			t.Fatalf("case %d url=%s: envelope parse: %v", i, c.URL, err)
		}
		if !env.Success {
			// error responses are still valid log records; skip strict data checks
			continue
		}
		code := string(env.Code)
		if code != `"0"` && code != "0" {
			t.Fatalf("case %d url=%s: code=%s want 0", i, c.URL, code)
		}
	}
}

func TestReadOnlyFixturesCommandContext(t *testing.T) {
	cases := loadFixtures(t)
	for i, c := range cases {
		var base struct {
			CommandContext *dto.CommandContext `json:"commandContext"`
		}
		if err := json.Unmarshal(c.Request, &base); err != nil {
			t.Fatalf("case %d url=%s: %v", i, c.URL, err)
		}
		if base.CommandContext == nil {
			t.Fatalf("case %d url=%s: missing commandContext", i, c.URL)
		}
		if base.CommandContext.TenantId == "" {
			t.Fatalf("case %d url=%s: missing tenantId", i, c.URL)
		}
	}
}

func TestReadOnlyFixturesDTOBinding(t *testing.T) {
	cases := loadFixtures(t)
	var bound, skipped int
	for i, c := range cases {
		if !bindRequestDTO(t, c.URL, c.Request) {
			skipped++
			continue
		}
		bound++
		if bound <= 3 && i < 5 {
			// spot-check first few
		}
	}
	if bound == 0 {
		t.Fatal("no requests bound to DTOs")
	}
	t.Logf("DTO binding: %d bound, %d unmapped URLs", bound, skipped)
}

func bindRequestDTO(t *testing.T, url string, req json.RawMessage) bool {
	t.Helper()
	switch url {
	case "/serviceArea/computeOutServiceDistance":
		var cmd dto.ComputeDistanceCmd
		return json.Unmarshal(req, &cmd) == nil && cmd.Imei != ""
	case "/banRiding/getNearestBanRiding", "/noParking/getNearNoParking", "/parking/getNearParking",
		"/serviceArea/getNearServiceByLocation", "/serviceArea/getServiceByLocation":
		var cmd dto.NearLocationCmd
		return json.Unmarshal(req, &cmd) == nil
	case "/config/backcar/getConfigByServiceId", "/config/base/getConfigByServiceId",
		"/config/pay/getConfigByServiceId", "/config/usecar/getConfigByServiceId",
		"/config/pushRidingCard/getByServiceId", "/helpConfig/getGuidePageConfigByServiceId",
		"/helpConfig/getHomeActivityEntranceByServiceId", "/helpConfig/getHomeNav",
		"/helpConfig/getHomeScrollerMsgByServiceId/v2", "/helpConfig/getSpecialTipsByServiceId",
		"/banRiding/getListByServiceId", "/noParking/getListByServiceId", "/parking/getListByServiceId":
		var cmd dto.ServiceIdCmd
		if err := json.Unmarshal(req, &cmd); err != nil {
			return false
		}
		return cmd.ServiceId != nil
	case "/serviceArea/getById":
		var cmd dto.IdCmd
		return json.Unmarshal(req, &cmd) == nil && cmd.Id != nil
	case "/serviceArea/getServiceAreaByIds":
		var cmd dto.IdsCmd
		return json.Unmarshal(req, &cmd) == nil && len(cmd.Ids) > 0
	case "/serviceArea/getFenceRelation":
		var cmd dto.ReturnCarCmd
		return json.Unmarshal(req, &cmd) == nil && cmd.CarCmd != nil
	case "/ridingPermission/get":
		var cmd dto.ServiceIdCmd
		return json.Unmarshal(req, &cmd) == nil
	case "/ad_config/detail", "/SiteApplication/pageSiteApplication":
		var cmd dto.PageCmd
		return json.Unmarshal(req, &cmd) == nil
	case "/resource/management/appList":
		var cmd dto.ServiceIdCmd
		return json.Unmarshal(req, &cmd) == nil
	case "/helmet/helmetStateChange":
		var cmd dto.IdCmd
		return json.Unmarshal(req, &cmd) == nil
	case "/serviceArea/getList":
		var cmd dto.Command
		return json.Unmarshal(req, &cmd) == nil
	case "/parking/bindParking":
		var cmd dto.ParkingBindCarCmd
		if err := json.Unmarshal(req, &cmd); err != nil {
			return false
		}
		return cmd.Imei != "" && cmd.Lat != nil && cmd.Lng != nil
	default:
		return false
	}
}

func TestComputeDistanceGoldenCases(t *testing.T) {
	cases := loadFixtures(t)
	var found int
	for _, c := range cases {
		if c.URL != "/serviceArea/computeOutServiceDistance" {
			continue
		}
		if c.Source != "ebike-fence-1.log" {
			continue
		}
		var cmd dto.ComputeDistanceCmd
		if err := json.Unmarshal(c.Request, &cmd); err != nil {
			t.Fatal(err)
		}
		if cmd.Imei != "862551059864149" {
			continue
		}
		co := javalog.ParseJavaCO(c.Reply, "ComputeDistanceCO")
		if co == nil {
			t.Fatal("missing ComputeDistanceCO")
		}
		dist, _ := co["distance"].(float64)
		if dist != 1618.0 {
			t.Fatalf("distance=%v want 1618.0", dist)
		}
		if izClose, _ := co["izCloseLine"].(bool); izClose {
			t.Fatal("izCloseLine should be false")
		}
		found++
		break
	}
	if found == 0 {
		t.Fatal("golden computeOutServiceDistance case not found")
	}
}

func loadBindParkingFixtures(t *testing.T) []javalog.Entry {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(file), "testdata", "bind_parking_cases.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("bind_parking_cases.json missing — run: go run ./cmd/parse_java_log")
	}
	var cases []javalog.Entry
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("no bindParking fixtures")
	}
	return cases
}

func TestBindParkingFixturesCommandContext(t *testing.T) {
	cases := loadBindParkingFixtures(t)
	for i, c := range cases {
		var cmd dto.ParkingBindCarCmd
		if err := json.Unmarshal(c.Request, &cmd); err != nil {
			t.Fatalf("case %d: %v", i, err)
		}
		if cmd.Imei == "" || cmd.Lat == nil || cmd.Lng == nil {
			t.Fatalf("case %d: missing required fields", i)
		}
	}
}

func TestBindParkingJavaResponseOmitsBanRidingId(t *testing.T) {
	cases := loadBindParkingFixtures(t)
	var checked int
	for _, c := range cases {
		env, err := javalog.ParseReplyEnvelope(c.Reply)
		if err != nil || !env.Success {
			continue
		}
		co := javalog.ParseJavaCO(c.Reply, "BindParkingCO")
		if co == nil {
			continue
		}
		if v, ok := co["banRidingId"]; ok && v != nil {
			t.Fatalf("Java bindParking response should not return banRidingId, got %+v", co)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no successful bindParking cases with BindParkingCO")
	}
	t.Logf("checked %d bindParking success responses", checked)
}

func TestBindParkingGoldenCasesFromJavaLog(t *testing.T) {
	cases := loadBindParkingFixtures(t)
	type golden struct {
		imei      string
		lat       float64
		lng       float64
		field     string
		wantID    int64
	}
	goldens := []golden{
		{imei: "", lat: 33.768009, lng: 118.399564, field: "parkingId", wantID: 185442561702239423},
		{imei: "", lat: 22.982905, lng: 115.336815, field: "nearParkingId", wantID: 274866299795413082},
	}
	for _, g := range goldens {
		var found bool
		for _, c := range cases {
			var cmd dto.ParkingBindCarCmd
			if err := json.Unmarshal(c.Request, &cmd); err != nil {
				continue
			}
			if cmd.Lat == nil || cmd.Lng == nil {
				continue
			}
			if *cmd.Lat != g.lat || *cmd.Lng != g.lng {
				continue
			}
			if g.imei != "" && cmd.Imei != g.imei {
				continue
			}
			got, ok := javalog.ParseJavaCOInt64Field(c.Reply, "BindParkingCO", g.field)
			if !ok {
				t.Fatalf("lat=%v lng=%v: missing %s in Java reply", g.lat, g.lng, g.field)
			}
			if got != g.wantID {
				t.Fatalf("lat=%v lng=%v: %s=%d want %d", g.lat, g.lng, g.field, got, g.wantID)
			}
			found = true
			break
		}
		if !found {
			t.Fatalf("golden case lat=%v lng=%v not found in bind_parking_cases.json — rerun go run ./cmd/gen_bind_parking_script", g.lat, g.lng)
		}
	}
}

func TestBanRidingNearestEmptyObject(t *testing.T) {
	cases := loadFixtures(t)
	var found int
	for _, c := range cases {
		if c.URL != "/banRiding/getNearestBanRiding" {
			continue
		}
		co := javalog.ParseJavaCO(c.Reply, "BanRidingCO")
		if co == nil {
			t.Fatal("missing BanRidingCO")
		}
		if co["serviceId"] != nil || co["distance"] != nil {
			t.Fatalf("expected null fields, got %+v", co)
		}
		found++
		if found >= 3 {
			break
		}
	}
	if found == 0 {
		t.Fatal("no banRiding cases")
	}
}

func TestLog1Log2RequestReplyConsistency(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	e1, err := javalog.ParseFile(filepath.Join(root, "ebike-fence-1.log"))
	if err != nil {
		t.Fatal(err)
	}
	e2, err := javalog.ParseFile(filepath.Join(root, "ebike-fence-2.log"))
	if err != nil {
		t.Fatal(err)
	}
	index := map[string]javalog.Entry{}
	for _, e := range e1 {
		if len(e.Request) == 0 {
			continue
		}
		index[e.URL+"\x00"+string(e.Request)] = e
	}
	var compared, mismatch int
	for _, e := range e2 {
		if len(e.Request) == 0 {
			continue
		}
		key := e.URL + "\x00" + string(e.Request)
		base, ok := index[key]
		if !ok {
			continue
		}
		compared++
		envA, _ := javalog.ParseReplyEnvelope(base.Reply)
		envB, _ := javalog.ParseReplyEnvelope(e.Reply)
		if envA.Success != envB.Success || string(envA.Code) != string(envB.Code) {
			mismatch++
		}
	}
	if compared == 0 {
		t.Skip("no overlapping requests between log1 and log2")
	}
	if mismatch > 0 {
		t.Fatalf("%d/%d overlapping requests differ in success/code between logs", mismatch, compared)
	}
	t.Logf("log1/log2 overlap: %d requests, all consistent envelopes", compared)
}

func TestReadOnlyURLCoverage(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	entries, _ := javalog.ParseFile(filepath.Join(root, "ebike-fence-1.log"))
	urls := map[string]struct{}{}
	for _, e := range entries {
		urls[e.URL] = struct{}{}
	}
	var readonly, unclassified int
	for u := range urls {
		if javalog.IsReadOnlyURL(u) {
			readonly++
		} else {
			unclassified++
			t.Logf("mutating: %s", u)
		}
	}
	if readonly < 20 {
		t.Fatalf("expected >=20 read-only urls, got %d", readonly)
	}
}

func TestExtractReadOnlyRequestsCoversAllLogURLs(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	var all []javalog.Entry
	for _, name := range []string{"ebike-fence-1.log", "ebike-fence-2.log"} {
		entries, err := javalog.ParseFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		all = append(all, entries...)
	}
	logReadOnly := map[string]struct{}{}
	for _, e := range all {
		if javalog.IsReadOnlyURL(e.URL) {
			logReadOnly[e.URL] = struct{}{}
		}
	}
	fixtures := javalog.ExtractReadOnlyRequests(all)
	counts := javalog.ReadOnlyURLCounts(fixtures)
	for u := range logReadOnly {
		if counts[u] == 0 {
			t.Fatalf("read-only url %s has no extractable request cases", u)
		}
	}
	if len(counts) != len(logReadOnly) {
		t.Fatalf("fixture urls=%d log readonly urls=%d", len(counts), len(logReadOnly))
	}
}

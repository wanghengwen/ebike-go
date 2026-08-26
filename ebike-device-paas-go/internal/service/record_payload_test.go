package service

import (
	"encoding/json"
	"testing"

	"ebike-device-paas-go/internal/api/dto"
)

// Production RECORD samples (unique shapes) from prod log dfdbf8cbd-j444z.
var recordSamples = map[string]string{
	"/device/paas/helmetLock":     `{"commandContext":{"tenantId":"1003","traceId":"178366799353996054730945","pin":"a_2fb92d500680c33","ip":"172.16.1.12","platform":"wechat","deviceId":"deviceId_174460457002689502293982","source":"ebike-fence","stressTesting":false},"imei":"862551059864305","carId":"100600147","sw":0}`,
	"/device/paas/deviceInfo":     `{"commandContext":{"tenantId":"1003","traceId":"178366799572180858573316","pin":"a_2fb92d500680c33","ip":"172.16.1.12","platform":"wechat","deviceId":"deviceId_174460457002689502293982","source":"ebike-fence","stressTesting":false},"imei":"862551059864305"}`,
	"/device/paas/bluetooth":      `{"commandContext":{"tenantId":"1003","traceId":"266c257d-50a1-41d3-b84f-51fd42f1b288","pin":"ebike-device-consume","ip":null,"platform":null,"deviceId":null,"source":null,"name":null,"stressTesting":false},"async":false,"carId":null,"imei":"862551055691447","payload":null,"token":144725592,"name":null}`,
	"/device/scanLocation/change": `{"commandContext":{"tenantId":"1003","traceId":"178366820516325743980373","pin":"a_4ed373201080f8a","ip":"172.16.1.12","platform":"wechat","deviceId":"deviceId_177414085755635146469770","source":"ebike-rent","name":null,"stressTesting":false},"carId":"100600167","lng":118.42223795572917,"lat":33.76476752387153}`,
	"/device/paas/lock":           `{"commandContext":{"tenantId":"1003","traceId":"61915413-9793-4538-8271-c8c85172423f","pin":"ebike-device-consume","ip":null,"platform":null,"deviceId":null,"source":null,"name":null,"stressTesting":false},"async":true,"carId":"100600222","imei":"862551059850064","payload":null,"acc":1,"izRiskControl":true,"dt":null,"idx":null,"volume":null,"isIgnoreFence":null,"isTBeacon":null,"isKickstand":null,"singleSpeedLimit":null}`,
	"/device/paas/defend":         `{"commandContext":{"tenantId":"1003","traceId":"178366841936267915643166","pin":"a_530366481c804dd","ip":"172.16.1.12","platform":"wechat","deviceId":"deviceId_178349593273337512545126","source":"ebike-rent","name":"WeChat User","stressTesting":false},"async":false,"carId":"100600413","imei":"862551059863539","payload":null,"defend":1,"idx":7,"volume":null,"isTBeacon":null,"isSlopeStake":null,"isRFID":null,"isKickstand":null,"isCamera":null,"tbeaconRange":null}`,
	"/device/trajectory/saveDb":   `{"commandContext":{"tenantId":"1003","traceId":"178366875371047740837608","pin":"a_531546c01c80ba0","ip":"172.16.1.12","platform":"wechat","deviceId":"deviceId_178366829246366168525239","source":"ebike-rent","name":"WeChat User","stressTesting":false},"orderId":374173365401819996,"imei":"862551059858638","startTime":1783668451000,"endTime":1783668755000}`,
	"/device/paas/setInnerParam":  `{"commandContext":{"tenantId":"2","traceId":"b8778d9e-5dd7-4901-bb6a-373fff8604cd","pin":"b_52613a003b00c5d","ip":"172.16.1.159","platform":"android","deviceId":"7fc4d3f2-5edc-38bb-b0b1-f9dfd31dc083","source":"ebike-operation","name":"桐乡公共自行车测试2","stressTesting":false},"async":true,"carId":null,"imei":"864423069955192","payload":null,"isFenceEnable":0}`,
	"/device/paas/voice":          `{"async":false,"carId":null,"commandContext":{"tenantId":"2","traceId":"178366928554964891358764","pin":"a_527eb0081c818f4","ip":"172.16.3.22","platform":"wechat","deviceId":"deviceId_178243467693520440639609","source":"ebike-service-client","name":"WeChat User","stressTesting":false},"idx":9,"imei":"864423069955192","payload":null,"volume":null}`,
}

// Additional production RECORD samples from prod log 668d758bf5-2r58d, covering
// shapes not exercised above: acc=0 (lock-off) with idx set, setInnerParam with
// freqMove/freqNorm keys explicitly present-but-null, and a bluetooth lookup
// whose device-side result is null (no matching beacon).
var recordSamples2 = map[string]string{
	"/device/paas/lock-acc0": `{"commandContext":{"tenantId":"1000","traceId":"178390983442057321600308","pin":"a_44bcac401080ad5","ip":"172.16.1.12","platform":"wechat","deviceId":"deviceId_178390917903141365544202","source":"ebike-rent","name":"辜丽春","stressTesting":false},"async":true,"carId":"100600916","imei":"866940070059332","payload":null,"acc":0,"izRiskControl":true,"dt":null,"idx":1,"volume":null,"isIgnoreFence":null,"isTBeacon":null,"isKickstand":null,"singleSpeedLimit":null}`,
	"/device/paas/setInnerParam-nullfreq": `{"commandContext":{"tenantId":"1000","traceId":"195d8821-c067-469d-9346-a8c1ed422534","pin":"b_3ef8abe804804c8","ip":"172.16.1.159","platform":"android","deviceId":"499c269d-9534-3cd6-b794-9a37fd21bd3b","source":"ebike-operation","name":"张人福","stressTesting":false},"async":true,"carId":null,"imei":"866940070129002","payload":null,"mode":null,"freqNorm":null,"freqMove":null,"isFenceEnable":0}`,
	"/device/paas/bluetooth-nullresult": `{"commandContext":{"tenantId":"1003","traceId":"50ffe084-b81e-4d75-abd3-6ced7df5e126","pin":"ebike-device-consume","ip":null,"platform":null,"deviceId":null,"source":null,"name":null,"stressTesting":false},"async":false,"carId":null,"imei":"862551059877364","payload":null,"token":736415771,"name":null}`,
}

// TestRecordPayload2_AllPathsBind covers the prod 668d758bf5-2r58d samples: bind
// must succeed and produce the expected forwarded shape, with no panics on the
// explicit-null / off-command edge cases these samples exercise.
func TestRecordPayload2_AllPathsBind(t *testing.T) {
	t.Run("lock_acc0_lockOff", func(t *testing.T) {
		// acc=0 (lock-off): izRiskControl no-risk-control side effect must NOT
		// fire (Java/Go both gate it on acc==1), and idx must forward as-is.
		raw, cc := bindCommandJSON(t, recordSamples2["/device/paas/lock-acc0"])
		gw := prepCommandBody(raw, cc)
		if gw["acc"].(float64) != 0 {
			t.Fatalf("acc=%v, want 0", gw["acc"])
		}
		if gw["idx"].(float64) != 1 {
			t.Fatalf("idx=%v, want 1", gw["idx"])
		}
		if acc := bodyInt(raw, "acc"); acc == 1 {
			t.Fatal("acc must not be treated as 1 (lock-off must skip no_risk_control_ex)")
		}
	})

	t.Run("setInnerParam_explicitNullFreq", func(t *testing.T) {
		// freqMove/freqNorm are present with an explicit JSON null (not
		// omitted): innerParamBody must still inject freq_move/freq_norm keys
		// (as nil), matching Java's null-safe field copy, and must not panic.
		raw, cc := bindCommandJSON(t, recordSamples2["/device/paas/setInnerParam-nullfreq"])
		gw := innerParamBody(raw, cc)
		if v, ok := gw["freq_move"]; !ok || v != nil {
			t.Fatalf("freq_move=%v (ok=%v), want present+nil", v, ok)
		}
		if v, ok := gw["freq_norm"]; !ok || v != nil {
			t.Fatalf("freq_norm=%v (ok=%v), want present+nil", v, ok)
		}
		if gw["isFenceEnable"].(float64) != 0 {
			t.Fatalf("isFenceEnable=%v", gw["isFenceEnable"])
		}
	})

	t.Run("bluetooth_nullDeviceResult", func(t *testing.T) {
		// Device-side lookup found no beacon (Java result:null): the forwarded
		// gateway body must still bind/transform cleanly regardless of the
		// eventual (out-of-scope-here) device response.
		raw, cc := bindCommandJSON(t, recordSamples2["/device/paas/bluetooth-nullresult"])
		gw := prepCommandBody(raw, cc)
		if gw["token"].(float64) != 736415771 {
			t.Fatalf("token=%v", gw["token"])
		}
		if cc.TenantID != "1003" {
			t.Fatalf("tenant=%q", cc.TenantID)
		}
	})
}

func bindCommandJSON(t *testing.T, body string) (map[string]interface{}, *dto.CommandContext) {
	t.Helper()
	raw := map[string]interface{}{}
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		t.Fatal(err)
	}
	var ctx struct {
		CommandContext *dto.CommandContext `json:"commandContext"`
	}
	if err := json.Unmarshal([]byte(body), &ctx); err != nil {
		t.Fatal(err)
	}
	return raw, ctx.CommandContext
}

func helmetBody(raw map[string]interface{}, cc *dto.CommandContext) map[string]interface{} {
	body := prepCommandBody(raw, cc)
	if sw, ok := raw["sw"]; ok {
		if n, _ := asInt(sw); n == 0 {
			body["idx"] = 22
		}
	}
	return body
}

func TestRecordPayload_AllPathsBind(t *testing.T) {
	t.Run("helmetLock", func(t *testing.T) {
		raw, cc := bindCommandJSON(t, recordSamples["/device/paas/helmetLock"])
		gw := helmetBody(raw, cc)
		if gw["imei"] != "862551059864305" || gw["idx"] != 22 {
			t.Fatalf("gw=%v", gw)
		}
		if cc.TenantID != "1003" {
			t.Fatalf("tenant=%q", cc.TenantID)
		}
	})

	t.Run("deviceInfo", func(t *testing.T) {
		var q dto.DeviceInfoQry
		if err := json.Unmarshal([]byte(recordSamples["/device/paas/deviceInfo"]), &q); err != nil {
			t.Fatal(err)
		}
		if q.Imei != "862551059864305" || q.CommandContext.TenantID != "1003" {
			t.Fatalf("q=%+v", q)
		}
		if q.CommandContext.IP == nil || q.CommandContext.Source == nil {
			t.Fatalf("audit fields missing: %+v", q.CommandContext)
		}
	})

	t.Run("bluetooth", func(t *testing.T) {
		raw, cc := bindCommandJSON(t, recordSamples["/device/paas/bluetooth"])
		gw := prepCommandBody(raw, cc)
		if gw["token"].(float64) != 144725592 {
			t.Fatalf("token=%v", gw["token"])
		}
		if cc.TraceID != "266c257d-50a1-41d3-b84f-51fd42f1b288" {
			t.Fatalf("trace=%q", cc.TraceID)
		}
	})

	t.Run("scanLocation", func(t *testing.T) {
		var cmd dto.ScanLocationChangeCmd
		if err := json.Unmarshal([]byte(recordSamples["/device/scanLocation/change"]), &cmd); err != nil {
			t.Fatal(err)
		}
		if cmd.CarID != "100600167" || cmd.Lng == nil || cmd.Lat == nil {
			t.Fatalf("cmd=%+v", cmd)
		}
	})

	t.Run("lock", func(t *testing.T) {
		raw, cc := bindCommandJSON(t, recordSamples["/device/paas/lock"])
		gw := prepCommandBody(raw, cc)
		if gw["acc"].(float64) != 1 || !gw["izRiskControl"].(bool) || !gw["async"].(bool) {
			t.Fatalf("gw=%v", gw)
		}
		if cc.TraceID != gw["traceId"] {
			t.Fatalf("traceId mismatch")
		}
	})

	t.Run("defend", func(t *testing.T) {
		raw, cc := bindCommandJSON(t, recordSamples["/device/paas/defend"])
		gw := prepCommandBody(raw, cc)
		if gw["defend"].(float64) != 1 || gw["idx"].(float64) != 7 {
			t.Fatalf("gw=%v", gw)
		}
		if cc.Source == nil || *cc.Source != "ebike-rent" {
			t.Fatalf("source=%v", cc.Source)
		}
	})

	t.Run("trajectorySaveDb", func(t *testing.T) {
		var cmd dto.TrajectoryDbCmd
		if err := json.Unmarshal([]byte(recordSamples["/device/trajectory/saveDb"]), &cmd); err != nil {
			t.Fatal(err)
		}
		if cmd.OrderId.Int64() != 374173365401819996 || cmd.Imei != "862551059858638" {
			t.Fatalf("cmd=%+v", cmd)
		}
		assertTenantRecord(t, cmd.CommandContext)
	})

	t.Run("setInnerParam", func(t *testing.T) {
		raw, cc := bindCommandJSON(t, recordSamples["/device/paas/setInnerParam"])
		gw := prepCommandBody(raw, cc)
		if gw["isFenceEnable"].(float64) != 0 || !gw["async"].(bool) {
			t.Fatalf("gw=%v", gw)
		}
		if cc.Name == nil {
			t.Fatalf("name missing")
		}
	})

	t.Run("voice", func(t *testing.T) {
		raw, cc := bindCommandJSON(t, recordSamples["/device/paas/voice"])
		gw := prepCommandBody(raw, cc)
		if gw["idx"].(float64) != 9 {
			t.Fatalf("idx=%v", gw["idx"])
		}
		if cc.Platform != "wechat" {
			t.Fatalf("platform=%q", cc.Platform)
		}
	})
}

func assertTenantRecord(t *testing.T, cc *dto.CommandContext) {
	t.Helper()
	if cc == nil || cc.TenantID == "" {
		t.Fatal("commandContext.tenantId missing")
	}
}

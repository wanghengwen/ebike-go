package order

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCloseOrderDownstreamBody(t *testing.T) {
	orderID := int64(999)
	payType := 1
	body := closeOrderCmdBody{OrderId: &orderID, PayType: &payType}

	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, `"orderId":999`) || !strings.Contains(s, `"payType":1`) {
		t.Fatalf("close order body shape mismatch: %s", s)
	}
}

func TestCarSearchVoiceDownstreamBody(t *testing.T) {
	body := voiceCmdBody{Async: false, Imei: "860123456789012", Idx: 9}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	for _, key := range []string{
		`"async":false`,
		`"imei":"860123456789012"`,
		`"idx":9`,
	} {
		if !strings.Contains(s, key) {
			t.Errorf("voice cmd missing %s in %s", key, s)
		}
	}
}

func TestConvertCInvoiceCOParsesStringOrderIds(t *testing.T) {
	raw := []byte(`{
		"serviceId":"274684702135689202",
		"id":"368394944319066945",
		"createdAt":"2026-06-09 12:01:04",
		"phone":"+86-13632021003",
		"title":"广州市苏尔派视听科技有限公司",
		"companyEin":"91440101MA9XMKBR98",
		"type":1,
		"amount":1200,
		"email":"1452875964@qq.com",
		"content":"运输服务",
		"state":1,
		"bank":"",
		"companyAddress":"",
		"bankAccount":"",
		"companyPhone":"",
		"opResult":null,
		"mark":"",
		"opManPhone":"+86-13787796908",
		"userPin":"a_4970de48108238c",
		"opManPin":"b_3731da38048118c",
		"dealTime":"2026-06-12 13:51:30",
		"orderIds":"[\"331089826894843934\",\"331098350257444022\"]"
	}`)
	co, ok := convertCInvoiceCO(raw)
	if !ok {
		t.Fatal("convertCInvoiceCO failed")
	}
	if len(co.OrderIds) != 2 || co.OrderIds[0] != "331089826894843934" {
		t.Fatalf("orderIds = %#v", co.OrderIds)
	}
	out, err := json.Marshal(co)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, `"orderIds":["331089826894843934","331098350257444022"]`) {
		t.Fatalf("expected array orderIds in output, got %s", s)
	}
	if !strings.Contains(s, `"commandContext":null`) {
		t.Fatalf("expected commandContext:null, got %s", s)
	}
}

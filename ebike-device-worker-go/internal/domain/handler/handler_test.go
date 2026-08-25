package handler_test

import (
	"encoding/json"
	"testing"

	"ebike-device-worker-go/internal/domain/constants"
	"ebike-device-worker-go/internal/domain/handler"
	"ebike-device-worker-go/internal/domain/message"
)

func TestGetPersistMessages(t *testing.T) {
	raw := `{"imei":"860123456789012","bussinessType":"ebike","data":{"cmd":3,"wgs84Lng":114.1,"wgs84Lat":22.2}}`
	var msg message.DeviceMessageDTO
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		// build manually when data is nested object in outer json
	}
	msg = message.DeviceMessageDTO{
		Imei:          "860123456789012",
		BussinessType: "ebike",
		MsgType:       "data",
		Data:          json.RawMessage(`{"cmd":3,"wgs84Lng":114.1,"wgs84Lat":22.2}`),
	}
	jsonList, structList := handler.ExportGetPersistMessages([]message.DeviceMessageDTO{msg})
	if len(jsonList) != 1 || len(structList) != 1 {
		t.Fatalf("expected 1 message, got json=%d struct=%d", len(jsonList), len(structList))
	}
	if structList[0].Cmd != constants.CmdGPS1 {
		t.Fatalf("unexpected cmd %d", structList[0].Cmd)
	}
	if jsonList[0]["isOnline"] != float64(1) && jsonList[0]["isOnline"] != 1 {
		t.Fatalf("expected isOnline=1 got %v", jsonList[0]["isOnline"])
	}
	if _, ok := structList[0].Data["isOnline"]; ok {
		t.Fatalf("struct data must use original payload without isOnline")
	}
	if _, ok := structList[0].Data["lastOnlineTime"]; ok {
		t.Fatalf("struct data must use original payload without lastOnlineTime")
	}
	if _, ok := structList[0].Data["deviceType"]; ok {
		t.Fatalf("struct data must use original payload without deviceType")
	}
}

func TestGetPersistMessagesUsesReceiveDataTimeMillis(t *testing.T) {
	raw := `{"imei":"860123456789012","bussinessType":"ebike","msgType":"data","receiveDataTime":1700000123456,"data":{"cmd":2}}`
	var msg message.DeviceMessageDTO
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	_, structList := handler.ExportGetPersistMessages([]message.DeviceMessageDTO{msg})
	if len(structList) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(structList))
	}
	if structList[0].ReceiveDataTime.Unix() != 1700000123 {
		t.Fatalf("receiveDataTime=%v", structList[0].ReceiveDataTime)
	}
}

func TestGetPersistMessagesDataAsJSONString(t *testing.T) {
	msg := message.DeviceMessageDTO{
		Imei:          "860123456789012",
		BussinessType: "ebike",
		MsgType:       "data",
		Data:          json.RawMessage(`"{\"cmd\":3,\"wgs84Lng\":114.1,\"wgs84Lat\":22.2}"`),
	}
	_, structList := handler.ExportGetPersistMessages([]message.DeviceMessageDTO{msg})
	if len(structList) != 1 || structList[0].Cmd != constants.CmdGPS1 {
		t.Fatalf("unexpected structList=%+v", structList)
	}
}

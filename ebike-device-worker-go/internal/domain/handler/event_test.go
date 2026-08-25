package handler_test

import (
	"encoding/json"
	"testing"

	"ebike-device-worker-go/internal/domain/constants"
	"ebike-device-worker-go/internal/domain/handler"
	"ebike-device-worker-go/internal/domain/message"
)

func TestEventLogoutSkipsRedis(t *testing.T) {
	msgs := []message.DeviceMessageDTO{
		{Imei: "logout-imei", BussinessType: "ebike", Data: json.RawMessage(`{"cmd":1001}`)},
		{Imei: "after-imei", BussinessType: "ebike", Data: json.RawMessage(`{"cmd":2}`)},
	}
	keys, vals, structList := handler.ExportEventRedisBatch(msgs)
	if len(keys) != 0 || len(vals) != 0 {
		t.Fatalf("logout should skip redis batch, keys=%v vals=%v", keys, vals)
	}
	if len(structList) != 2 {
		t.Fatalf("expected 2 persist structs, got %d", len(structList))
	}
}

func TestEventLogoutSetsOfflineOnBreak(t *testing.T) {
	msgs := []message.DeviceMessageDTO{
		{Imei: "logout-imei", BussinessType: "ebike", Data: json.RawMessage(`{"cmd":1001}`)},
	}
	jsonList, structList := handler.ExportGetPersistMessages(msgs)
	handler.ExportCollectEventRedisBatch(jsonList, structList)
	if jsonList[0]["isOnline"] != float64(0) && jsonList[0]["isOnline"] != 0 {
		t.Fatalf("logout should set isOnline=0, got %v", jsonList[0]["isOnline"])
	}
	if _, ok := structList[0].Data["isOnline"]; ok {
		t.Fatalf("logout PG payload must stay original without isOnline")
	}
}

func TestEventPingBeforeLogoutWritesOnlyPing(t *testing.T) {
	msgs := []message.DeviceMessageDTO{
		{Imei: "ping-imei", BussinessType: "ebike", Data: json.RawMessage(`{"cmd":2}`)},
		{Imei: "logout-imei", BussinessType: "ebike", Data: json.RawMessage(`{"cmd":1001}`)},
	}
	keys, vals, structList := handler.ExportEventRedisBatch(msgs)
	if len(keys) != 1 || len(vals) != 1 {
		t.Fatalf("expected one redis entry, keys=%v vals=%v", keys, vals)
	}
	wantKey := constants.RedisKeyDeviceEbike + "ping-imei"
	if keys[0] != wantKey {
		t.Fatalf("key=%s want=%s", keys[0], wantKey)
	}
	if structList[0].DeviceID != "ping-imei" || structList[1].DeviceID != "logout-imei" {
		t.Fatalf("struct order mismatch: %+v", structList)
	}
}

func TestEventLogoutCmdConstant(t *testing.T) {
	if constants.CmdLogout != 1001 {
		t.Fatalf("cmd logout=%d", constants.CmdLogout)
	}
}

package push

import (
	"encoding/json"
	"testing"

	"ebike-device-worker-go/internal/domain/constants"
	"ebike-device-worker-go/internal/domain/message"
)

func TestBuildSendDataPreservesVoltageFromStringEncodedData(t *testing.T) {
	kafkaMsg := `{"imei":"860123456789012","msgType":"data","bussinessType":"ebike","data":"{\"cmd\":3,\"voltage\":48000,\"gsm\":25}"}`
	var dto message.DeviceMessageDTO
	if err := json.Unmarshal([]byte(kafkaMsg), &dto); err != nil {
		t.Fatal(err)
	}

	s := &Service{}
	out := s.buildSendData(dto, map[string]interface{}{
		"appId":          1,
		"appName":        "tenant",
		"deviceDataType": "gps",
	}, "data")

	if out["deviceDataType"] != "gps" {
		t.Fatalf("deviceDataType=%v", out["deviceDataType"])
	}
	if out["voltage"] != float64(48000) {
		t.Fatalf("voltage=%v want 48000", out["voltage"])
	}
	if out["imei"] != "860123456789012" {
		t.Fatalf("imei=%v", out["imei"])
	}
}

func TestDeviceDataTypeByCmdShuakaYuhuan(t *testing.T) {
	if got := deviceDataTypeByCmd(constants.CmdShuaka); got != "shuaka" {
		t.Fatalf("deviceDataTypeByCmd(%d)=%q want shuaka", constants.CmdShuaka, got)
	}
}

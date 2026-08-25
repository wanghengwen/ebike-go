package message_test

import (
	"encoding/json"
	"testing"

	"ebike-device-worker-go/internal/domain/message"
)

func TestParseDataJSONObject(t *testing.T) {
	data, err := message.ParseData(json.RawMessage(`{"cmd":3,"voltage":48000}`))
	if err != nil {
		t.Fatal(err)
	}
	if data["voltage"] != float64(48000) {
		t.Fatalf("voltage=%v", data["voltage"])
	}
}

func TestParseDataJSONString(t *testing.T) {
	inner := `"{\"cmd\":3,\"voltage\":48000,\"gsm\":25}"`
	data, err := message.ParseData(json.RawMessage(inner))
	if err != nil {
		t.Fatal(err)
	}
	if data["voltage"] != float64(48000) {
		t.Fatalf("voltage=%v", data["voltage"])
	}
}

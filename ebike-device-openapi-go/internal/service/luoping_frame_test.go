package service

import (
	"encoding/binary"
	"testing"

	"ebike-device-openapi-go/internal/decode/xiaoan"
	lpcrypto "ebike-device-openapi-go/internal/pkg/crypto"
)

func TestDecodeXiaoanFramePing(t *testing.T) {
	frame := xiaoan.EncodeFrame(2, 1, []byte{0x1F, 0x30})
	got, err := decodeXiaoanFrame(frame)
	if err != nil {
		t.Fatal(err)
	}
	if got.Cmd != 2 || got.Sequence != 1 || got.MsgType == "" {
		t.Fatalf("cmd=%d seq=%d msgType=%s", got.Cmd, got.Sequence, got.MsgType)
	}
}

func TestDecodeXiaoanFrameWildReply(t *testing.T) {
	frame, err := xiaoan.EncodeWildCmdFrame(3, 33, map[string]interface{}{"acc": 1}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := decodeXiaoanFrame(frame)
	if err != nil {
		t.Fatal(err)
	}
	if got.Cmd != 0 || got.Sequence != 3 || got.MsgType != "event" {
		t.Fatalf("cmd=%d seq=%d msgType=%s", got.Cmd, got.Sequence, got.MsgType)
	}
	if got.Map["json"] == "" {
		t.Fatalf("missing json body: %v", got.Map)
	}
}

func TestDecodeXiaoanFrameBin35(t *testing.T) {
	body := make([]byte, 35)
	binary.BigEndian.PutUint32(body[0:4], 0x00140401) // 20.4.1
	body[4] = 100
	copy(body[5:20], []byte("864423069952637"))
	copy(body[20:35], []byte("460081944102200"))
	frame := xiaoan.EncodeFrame(35, 9, body)

	got, err := decodeXiaoanFrame(frame)
	if err != nil {
		t.Fatal(err)
	}
	msg, ok := got.Typed.(*xiaoan.Bin35LoginMessage)
	if !ok || msg == nil {
		t.Fatalf("typed=%T", got.Typed)
	}
	if msg.Version != 0x00140401 || msg.DeviceType != 100 {
		t.Fatalf("version=%d deviceType=%d", msg.Version, msg.DeviceType)
	}
	if msg.Imei != "864423069952637" || msg.Imsi != "460081944102200" {
		t.Fatalf("imei=%s imsi=%s", msg.Imei, msg.Imsi)
	}
}

// Uplink data must not carry an inner imei: the 小安 TCP path omits it and
// device-worker overwrites it with the envelope imei anyway.
func TestDecodeXiaoanFrameOmitsInnerImei(t *testing.T) {
	frame := xiaoan.EncodeFrame(2, 1, []byte{0x1F, 0x30})
	got, err := decodeXiaoanFrame(frame)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.Map["imei"]; ok {
		t.Fatalf("decoded map should not contain imei: %v", got.Map)
	}
}

func TestEncryptDecryptBrptPayload(t *testing.T) {
	deviceID := "2604000274"
	imei := "868480084776843"
	frame := xiaoan.EncodeFrame(2, 7, []byte{0x1F, 0x30})
	enc, err := lpcrypto.EncryptLuopingMQTTWithKeyIndex(frame, deviceID, imei, 5)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := lpcrypto.DecryptLuopingMQTT(enc, deviceID, imei)
	if err != nil {
		t.Fatal(err)
	}
	got, err := decodeXiaoanFrame(plain)
	if err != nil {
		t.Fatal(err)
	}
	if got.Cmd != 2 || got.Sequence != 7 {
		t.Fatalf("cmd=%d seq=%d", got.Cmd, got.Sequence)
	}
}

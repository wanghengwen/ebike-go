package xiaoan

import (
	"bytes"
	"encoding/json"
	"testing"
)

// buildFrame assembles a 小安 frame via EncodeFrame.
func buildFrame(cmd, seq byte, body []byte) []byte {
	return EncodeFrame(cmd, seq, body)
}

func TestParseFrameBin2(t *testing.T) {
	frame := buildFrame(2, 1, []byte{0x1F, 0x30}) // gsm=31, voltage=48
	header, buf, err := ParseFrame(frame)
	if err != nil {
		t.Fatalf("parse frame: %v", err)
	}
	if header.Cmd != 2 || header.Sequence != 1 || header.Length != 2 || header.Magic != XiaoanMagic {
		t.Fatalf("unexpected header: %+v", header)
	}
	msg, err := DecodeHex(header, buf)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	ping, ok := msg.(*Bin2PingMessage)
	if !ok {
		t.Fatalf("expected *Bin2PingMessage, got %T", msg)
	}
	if ping.GsmSignal != 31 || ping.Voltage != 48 {
		t.Fatalf("unexpected fields: gsm=%d voltage=%d", ping.GsmSignal, ping.Voltage)
	}
}

func TestParseFrameIgnoresTrailingBytes(t *testing.T) {
	// body length=2, but 2 extra trailing bytes (e.g. checksum) must be ignored.
	frame := append(buildFrame(2, 7, []byte{0x1F, 0x30}), 0xDE, 0xAD)
	header, buf, err := ParseFrame(frame)
	if err != nil {
		t.Fatalf("parse frame: %v", err)
	}
	if header.Length != 2 {
		t.Fatalf("length=%d", header.Length)
	}
	if _, err := DecodeHex(header, buf); err != nil {
		t.Fatalf("decode with trailing bytes: %v", err)
	}
}

func TestParseFrameBadMagic(t *testing.T) {
	frame := []byte{0x12, 0x34, 0x02, 0x01, 0x00, 0x00}
	if _, _, err := ParseFrame(frame); err == nil {
		t.Fatal("expected magic error")
	}
}

func TestParseFrameTruncatedBody(t *testing.T) {
	// length declares 5 but only 2 body bytes present.
	frame := []byte{0xAA, 0x55, 0x02, 0x01, 0x00, 0x05, 0x01, 0x02}
	if _, _, err := ParseFrame(frame); err == nil {
		t.Fatal("expected truncated body error")
	}
}

func TestParseFrameTooShort(t *testing.T) {
	if _, _, err := ParseFrame([]byte{0xAA, 0x55}); err == nil {
		t.Fatal("expected too short error")
	}
}

func TestEncodeWildCmdFrame(t *testing.T) {
	frame, err := EncodeWildCmdFrame(9, 33, map[string]interface{}{"acc": 1}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	header, body, err := ParseFrame(frame)
	if err != nil {
		t.Fatal(err)
	}
	if header.Cmd != WildCmdHeader || header.Sequence != 9 {
		t.Fatalf("header=%+v", header)
	}
	rawBody := body.ReadBytes(body.ReadableBytes())
	var m map[string]interface{}
	if err := json.Unmarshal(rawBody, &m); err != nil {
		t.Fatal(err)
	}
	if int(m["c"].(float64)) != 33 {
		t.Fatalf("c=%v", m["c"])
	}
	param := m["param"].(map[string]interface{})
	if int(param["acc"].(float64)) != 1 {
		t.Fatalf("param=%v", param)
	}
	// round-trip EncodeFrame
	again := EncodeFrame(2, 1, []byte{1, 2})
	if !bytes.Equal(again, []byte{0xAA, 0x55, 2, 1, 0, 2, 1, 2}) {
		t.Fatalf("EncodeFrame bytes=%x", again)
	}
}

func TestEncodeLoginAck(t *testing.T) {
	frame := EncodeLoginAck(5, 1700000000)
	header, body, err := ParseFrame(frame)
	if err != nil {
		t.Fatal(err)
	}
	if header.Cmd != 35 || header.Sequence != 5 || header.Length != 4 {
		t.Fatalf("header=%+v", header)
	}
	ts := body.ReadUnsignedInt()
	if ts != 1700000000 {
		t.Fatalf("ts=%d", ts)
	}
}

func TestEncodeAutoReplayCmd2(t *testing.T) {
	frame := EncodeAutoReplay(2, 7)
	header, body, err := ParseFrame(frame)
	if err != nil {
		t.Fatal(err)
	}
	if header.Cmd != 2 || header.Sequence != 7 || header.Length != 0 {
		t.Fatalf("header=%+v", header)
	}
	if body.ReadableBytes() != 0 {
		t.Fatalf("body should be empty, remain=%d", body.ReadableBytes())
	}
	if !bytes.Equal(frame, []byte{0xAA, 0x55, 0x02, 0x07, 0x00, 0x00}) {
		t.Fatalf("frame=%x", frame)
	}
}

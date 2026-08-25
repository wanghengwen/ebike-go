package xiaoan

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"ebike-device-openapi-go/internal/api/dto"
)

// XiaoanMagic is the 2-byte frame magic of the 小安 protocol (0xAA55).
const XiaoanMagic = 0xAA55

// WildCmdHeader is the header cmd used by gateway /ecu/wild (Bin0Encode).
// Body is ASCII JSON {"c":<bizCmd>,"param":{...}}.
const WildCmdHeader int16 = 0

// frameHeaderLen is magic(2)+cmd(1)+sequence(1)+length(2).
const frameHeaderLen = 6

// EncodeFrame builds a complete 小安 frame:
//
//	magic(2, 0xAA55) | cmd(1) | sequence(1) | length(2 BE) | body
func EncodeFrame(cmd, sequence byte, body []byte) []byte {
	if body == nil {
		body = []byte{}
	}
	out := make([]byte, frameHeaderLen+len(body))
	binary.BigEndian.PutUint16(out[0:2], XiaoanMagic)
	out[2] = cmd
	out[3] = sequence
	binary.BigEndian.PutUint16(out[4:6], uint16(len(body)))
	copy(out[6:], body)
	return out
}

// EncodeWildCmdFrame builds the same downlink as gateway /ecu/wild:
// header cmd=0, body = US-ASCII JSON {"c":bizCmd,"param":..., optional dt/tm}.
func EncodeWildCmdFrame(sequence byte, bizCmd int16, params map[string]interface{}, dt, tm *int) ([]byte, error) {
	payload := map[string]interface{}{"c": bizCmd}
	if dt != nil && tm != nil {
		payload["dt"] = *dt
		payload["tm"] = *tm
	}
	if params != nil {
		payload["param"] = params
	} else {
		payload["param"] = map[string]interface{}{}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return EncodeFrame(byte(WildCmdHeader), sequence, body), nil
}

// EncodeLoginAck builds the gateway LoginHandler.replayLogin response:
//
//	AA55 | cmd=35 | sequence | len=4 | unix_seconds uint32 BE
func EncodeLoginAck(sequence byte, unixSec int64) []byte {
	if unixSec <= 0 {
		unixSec = time.Now().Unix()
	}
	body := make([]byte, 4)
	binary.BigEndian.PutUint32(body, uint32(unixSec))
	return EncodeFrame(35, sequence, body)
}

// EncodeAutoReplay builds gateway ReplayMessage for auto-ack cmds (e.g. ping cmd=2):
//
//	AA55 | cmd | sequence | len=0
func EncodeAutoReplay(cmd, sequence byte) []byte {
	return EncodeFrame(cmd, sequence, nil)
}

// ParseFrame parses a complete 小安 protocol frame:
//
//	magic(2, 0xAA55) | cmd(1) | sequence(1) | length(2) | body(length) [| trailing...]
//
// It validates the magic, reads the header, and returns a ByteBuf positioned at
// the body (exactly `length` bytes) so callers can feed it straight into
// DecodeHex. Any bytes after body (e.g. a trailing checksum) are ignored, since
// the body length is authoritative — this mirrors the gateway's Message.decode
// (ebike-device-gateway) frame layout.
func ParseFrame(raw []byte) (*dto.MessageHeader, *ByteBuf, error) {
	if len(raw) < frameHeaderLen {
		return nil, nil, fmt.Errorf("xiaoan frame too short: %d bytes (need >=%d header)", len(raw), frameHeaderLen)
	}
	buf := NewByteBuf(raw)
	magic := int(buf.ReadUnsignedShort())
	if magic != XiaoanMagic {
		return nil, nil, fmt.Errorf("invalid xiaoan magic: 0x%04X (want 0x%04X)", magic, XiaoanMagic)
	}
	cmd := int16(buf.ReadUnsignedByte())
	sequence := int16(buf.ReadUnsignedByte())
	length := int(buf.ReadUnsignedShort())
	if buf.Err() != nil {
		return nil, nil, buf.Err()
	}
	body := buf.ReadBytes(length)
	if buf.Err() != nil {
		return nil, nil, fmt.Errorf("xiaoan frame body: %w", buf.Err())
	}
	header := &dto.MessageHeader{
		Magic:    magic,
		Cmd:      cmd,
		Sequence: sequence,
		Length:   length,
	}
	return header, NewByteBuf(body), nil
}

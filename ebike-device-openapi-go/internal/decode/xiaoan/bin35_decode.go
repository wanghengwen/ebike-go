package xiaoan

import (
	"fmt"
	"strings"
	"time"

	"ebike-device-openapi-go/internal/api/dto"
)

// Bin35LoginMessage mirrors Java gateway LoginMessage / Bin35Decode body.
// Binary layout (35 bytes):
//
//	version uint32 | equipmentType uint8 | imei ASCII[15] | imsi ASCII[15]
type Bin35LoginMessage struct {
	Imei          string `json:"imei,omitempty"`
	Imsi          string `json:"imsi,omitempty"`
	Version       int64  `json:"version"`
	DeviceType    int    `json:"deviceType"` // Java LoginCmd.deviceType ← equipmentType
	MsgType       string `json:"msgType"`
	Cmd           int16  `json:"cmd"`
	BussinessType string `json:"bussinessType"`
	Timestamp     int64  `json:"timestamp"`
}

// Bin35Decode parses 小安 cmd=35 登录包 body（对齐 gateway Bin35Decode.java）。
type Bin35Decode struct{}

func (d *Bin35Decode) Decode(header *dto.MessageHeader, data *ByteBuf) (interface{}, error) {
	// Java: if (data.capacity() < 35) return null;
	if data.ReadableBytes() < 35 {
		return nil, fmt.Errorf("bin35 body too short: need 35, remain %d", data.ReadableBytes())
	}

	version := int64(data.ReadUnsignedInt())
	equipmentType := int(data.ReadUnsignedByte())
	// ASCII fields are often NUL-padded; strip NULs then whitespace (Java String.trim).
	imei := strings.TrimSpace(strings.ReplaceAll(data.ReadCharSequence(15), "\x00", ""))
	imsi := strings.TrimSpace(strings.ReplaceAll(data.ReadCharSequence(15), "\x00", ""))
	if data.Err() != nil {
		return nil, data.Err()
	}

	msg := &Bin35LoginMessage{
		Imei:          imei,
		Imsi:          imsi,
		Version:       version,
		DeviceType:    equipmentType,
		MsgType:       "event",
		Cmd:           35,
		BussinessType: "ebike",
		Timestamp:     time.Now().Unix(),
	}
	return msg, nil
}

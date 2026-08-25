package xiaoan

import "ebike-device-openapi-go/internal/api/dto"

type Bin2PingMessage struct {
	Imei          *string `json:"imei,omitempty"`
	MsgType       string  `json:"msgType"`
	Cmd           int16   `json:"cmd"`
	BussinessType string  `json:"bussinessType"`
	Timestamp     int64   `json:"timestamp"`
	GsmSignal     int8    `json:"gsmSignal"`
	Voltage       uint16  `json:"voltage"`
}

type Bin2Decode struct{}

func (d *Bin2Decode) Decode(header *dto.MessageHeader, data *ByteBuf) (interface{}, error) {
	msg := &Bin2PingMessage{
		MsgType:       "data",
		Cmd:           2, // CMD_PING
		BussinessType: "ebike",
	}

	gsmSignal := data.ReadUnsignedByte()
	voltage := data.ReadUnsignedByte()

	msg.GsmSignal = int8(gsmSignal)
	msg.Voltage = uint16(voltage)

	return msg, nil
}

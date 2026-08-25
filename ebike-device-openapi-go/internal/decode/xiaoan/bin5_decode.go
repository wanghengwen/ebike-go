package xiaoan

import "ebike-device-openapi-go/internal/api/dto"

type Bin5AlarmMessage struct {
	Imei          *string `json:"imei,omitempty"`
	MsgType       string  `json:"msgType"`
	Cmd           int16   `json:"cmd"`
	BussinessType string  `json:"bussinessType"`
	Timestamp     int64   `json:"timestamp"`
	Type          uint8   `json:"type"`
}

type Bin5Decode struct{}

func (d *Bin5Decode) Decode(header *dto.MessageHeader, data *ByteBuf) (interface{}, error) {
	msg := &Bin5AlarmMessage{
		MsgType:       "alarm",
		Cmd:           5, // CMD_ALARM
		BussinessType: "ebike",
	}

	alarmType := data.ReadUnsignedByte()
	msg.Type = alarmType

	return msg, nil
}

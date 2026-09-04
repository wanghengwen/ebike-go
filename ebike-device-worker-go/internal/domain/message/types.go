package message

import (
	"encoding/json"
	"time"
)

// DeviceMessageDTO mirrors Java infrastructure.dto.DeviceMessageDTO.
type DeviceMessageDTO struct {
	Imei            string             `json:"imei"`
	MsgType         string             `json:"msgType"`
	Identifier      string             `json:"identifier"`
	BussinessType   string             `json:"bussinessType"`
	MsgID           string             `json:"msgId"`
	Data            json.RawMessage    `json:"data"`
	Payload         json.RawMessage    `json:"payload"`
	ReceiveDataTime *ReceiveMillisTime `json:"receiveDataTime"`
}

// MessageInfoDTO mirrors Java infrastructure.dto.MessageInfoDTO.
type MessageInfoDTO struct {
	DeviceID        string
	MsgType         string
	BussinessType   string
	Cmd             int
	Data            map[string]interface{}
	ReceiveDataTime time.Time
}

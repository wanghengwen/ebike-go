package xiaoan

import (
	"ebike-device-openapi-go/internal/api/dto"
	"fmt"
)

type Decoder interface {
	Decode(header *dto.MessageHeader, data *ByteBuf) (interface{}, error)
}

var decoders = map[int16]Decoder{
	2:   &Bin2Decode{},
	5:   &Bin5Decode{},
	35:  &Bin35Decode{}, // CMD_LOGIN 登录（gateway Bin35Decode → imei/imsi/version/deviceType）
	41:  &Bin41Decode{},
	66:  &Bin66Decode{},
	68:  &Bin68Decode{},
	70:  &Bin70Decode{},
	72:  &Bin72Decode{},
	201: &Bin201Decode{}, // CMD_SHUAKA 刷卡中控，feature/Bin201Decode-20240531
}

func DecodeHex(header *dto.MessageHeader, data *ByteBuf) (interface{}, error) {
	decoder, ok := decoders[header.Cmd]
	if !ok {
		return nil, fmt.Errorf("no decoder found for cmd: %d", header.Cmd)
	}
	msg, err := decoder.Decode(header, data)
	if err != nil {
		return nil, err
	}
	if data.Err() != nil {
		return nil, data.Err()
	}
	return msg, nil
}

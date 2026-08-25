package xiaoan

import (
	"ebike-device-openapi-go/internal/api/dto"
	"ebike-device-openapi-go/internal/pkg/utils"
)

type Bin41GpsMessage struct {
	Imei          *string `json:"imei,omitempty"`
	MsgType       string  `json:"msgType"`
	Cmd           int16   `json:"cmd"`
	BussinessType string  `json:"bussinessType"`
	Sw            uint32  `json:"sw"`
	Gsm           uint16  `json:"gsm"`
	Voltage       uint64  `json:"voltage"`
	Timestamp     uint64  `json:"timestamp"`
	Lng           float64 `json:"lng"`
	Lat           float64 `json:"lat"`
	Speed         uint16  `json:"speed"`
	Course        uint16  `json:"course"`
	Hdop          float32 `json:"hdop"`
	Satellite     uint16  `json:"satellite"`

	Defend             uint8 `json:"defend"`
	Acc                uint8 `json:"acc"`
	BackWheelLock      uint8 `json:"backWheelLock"`
	BatteryLock        uint8 `json:"batteryLock"`
	BatteryConnect     uint8 `json:"batteryConnect"`
	IsGPSFastMode      uint8 `json:"isGPSFastMode"`
	IsMoving           uint8 `json:"isMoving"`
	IsWheelSpan        uint8 `json:"isWheelSpan"`
	IsHelmetUnlock     uint8 `json:"isHelmetUnlock"`
	IsSleepMode        uint8 `json:"isSleepMode"`
	IsMoveAlarmOn      uint8 `json:"isMoveAlarmOn"`
	IsOverSpeedOn      uint8 `json:"isOverSpeedOn"`
	IsGPSUnfixed       uint8 `json:"isGPSUnfixed"`
	IsGyroFixed        uint8 `json:"isGyroFixed"`
	IsFenceEnable      uint8 `json:"isFenceEnable"`
	IsOutofServAera    uint8 `json:"isOutofServAera"`
	Wgs84Lng           float32 `json:"wgs84Lng"`
	Wgs84Lat           float32 `json:"wgs84Lat"`
	FenceVersion       uint64  `json:"fenceVersion"`
}

type Bin41Decode struct{}

func (d *Bin41Decode) Decode(header *dto.MessageHeader, data *ByteBuf) (interface{}, error) {
	msg := &Bin41GpsMessage{
		MsgType:       "data",
		Cmd:           3, // CmdConstant.CMD_GPS_1 = 3
		BussinessType: "ebike",
	}

	sw := uint32(data.ReadUnsignedShort())
	msg.Gsm = uint16(data.ReadUnsignedByte())
	msg.Voltage = uint64(data.ReadUnsignedInt())
	msg.Timestamp = uint64(data.ReadUnsignedInt())
	wgs84Lng := data.ReadFloatLE()
	wgs84Lat := data.ReadFloatLE()
	msg.Speed = uint16(data.ReadUnsignedByte())
	msg.Course = data.ReadUnsignedShort()
	msg.Hdop = data.ReadFloatLE()
	msg.Satellite = uint16(data.ReadUnsignedByte())

	msg.Sw = sw
	msg.Defend = uint8(sw & 1)
	msg.Acc = uint8(sw >> 1 & 1)
	msg.BackWheelLock = uint8(sw >> 2 & 1)
	msg.BatteryLock = uint8(sw >> 3 & 1)
	msg.BatteryConnect = uint8(sw >> 4 & 1)
	msg.IsGPSFastMode = uint8(sw >> 5 & 1)
	msg.IsMoving = uint8(sw >> 6 & 1)
	msg.IsWheelSpan = uint8(sw >> 7 & 1)
	msg.IsHelmetUnlock = uint8(sw >> 8 & 1)
	msg.IsSleepMode = uint8(sw >> 9 & 1)
	msg.IsMoveAlarmOn = uint8(sw >> 10 & 1)
	msg.IsOverSpeedOn = uint8(sw >> 11 & 1)
	msg.IsGPSUnfixed = uint8(sw >> 12 & 1)
	msg.IsGyroFixed = uint8(sw >> 13 & 1)
	msg.IsFenceEnable = uint8(sw >> 14 & 1)
	msg.IsOutofServAera = uint8(sw >> 15 & 1)

	msg.Wgs84Lng = wgs84Lng
	msg.Wgs84Lat = wgs84Lat
	
	lng, lat := utils.TransformWGS84ToGCJ02(float64(wgs84Lng), float64(wgs84Lat))
	msg.Lng = lng
	msg.Lat = lat

	return msg, nil
}

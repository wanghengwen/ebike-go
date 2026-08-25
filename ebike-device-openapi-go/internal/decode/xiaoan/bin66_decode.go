package xiaoan

import (
	"strings"

	"ebike-device-openapi-go/internal/api/dto"
)

type Bin66BmsInfoMessage struct {
	Imei          *string `json:"imei,omitempty"`
	MsgType       string  `json:"msgType"`
	Cmd           int16   `json:"cmd"`
	BussinessType string  `json:"bussinessType"`

	Sn               string `json:"sn"`
	HardVersion      uint16 `json:"hardVersion"`
	SoftVersion      uint16 `json:"softVersion"`
	MosTemperature   int16  `json:"mosTemperature"`
	MosState         uint8  `json:"mosState"`
	MaxVoltage       uint8  `json:"maxVoltage"`
	MinVoltage       uint8  `json:"minVoltage"`
	Soh              uint8  `json:"soh"`
	Capacity         uint16 `json:"capacity"`
	RemainCapacity   uint16 `json:"remainCapacity"`
	Soc              uint8  `json:"soc"`
	CycleLifeCounter uint16 `json:"cycleLifeCounter"`
	Voltage          uint16 `json:"voltage"`
	Current          int16  `json:"current"`
	Timestamp        uint32 `json:"timestamp"`

	// 原始故障位图。下面的派生布尔位无法还原它：IsDischargeOverCurrent 被 bit3 覆盖、
	// IsChargeOverCurrent 恒为 0，需要按位透传故障状态时只能用这个字段。
	Fault uint16 `json:"fault"`

	IsChargeTemperatureHigh    uint8 `json:"isChargeTemperatureHigh"`
	IsChargeShortCircuit       uint8 `json:"isChargeShortCircuit"`
	IsDischargeOverCurrent     uint8 `json:"isDischargeOverCurrent"`
	IsChargeOverCurrent        uint8 `json:"isChargeOverCurrent"`
	IsUnderVoltage             uint8 `json:"isUnderVoltage"`
	IsOverVoltage              uint8 `json:"isOverVoltage"`
	IsNodeUnderVoltage         uint8 `json:"isNodeUnderVoltage"`
	IsNodeOverVoltage          uint8 `json:"isNodeOverVoltage"`
	IsErrorCollect             uint8 `json:"isErrorCollect"`
	IsDischargeLowTemperature  uint8 `json:"isDischargeLowTemperature"`
	IsChargeLowTemperature     uint8 `json:"isChargeLowTemperature"`
	IsDischargeHighTemperature uint8 `json:"isDischargeHighTemperature"`
}

type Bin66Decode struct{}

func (d *Bin66Decode) Decode(header *dto.MessageHeader, data *ByteBuf) (interface{}, error) {
	msg := &Bin66BmsInfoMessage{
		MsgType:       "data",
		Cmd:           66, // CMD_BMS_INFO
		BussinessType: "ebike",
	}

	snRaw := data.ReadCharSequence(20)
	// Clean up \u0000 bytes and trim spaces (matching Java: .trim().replaceAll("\u0000", ""))
	cleanSn := make([]byte, 0, len(snRaw))
	for i := 0; i < len(snRaw); i++ {
		if snRaw[i] != 0 {
			cleanSn = append(cleanSn, snRaw[i])
		}
	}
	msg.Sn = strings.TrimSpace(string(cleanSn))

	msg.HardVersion = data.ReadUnsignedShort()
	msg.SoftVersion = data.ReadUnsignedShort()
	msg.MosTemperature = data.ReadShort()
	msg.MosState = data.ReadUnsignedByte()
	msg.MaxVoltage = data.ReadUnsignedByte()
	msg.MinVoltage = data.ReadUnsignedByte()
	msg.Soh = data.ReadUnsignedByte()

	bmsFault := data.ReadUnsignedShort()
	msg.Capacity = data.ReadUnsignedShort()
	msg.RemainCapacity = data.ReadUnsignedShort()
	msg.Soc = data.ReadUnsignedByte()
	msg.CycleLifeCounter = data.ReadUnsignedShort()
	msg.Voltage = data.ReadUnsignedShort()
	msg.Current = data.ReadShort()
	msg.Timestamp = data.ReadUnsignedInt()

	msg.Fault = bmsFault
	msg.IsChargeTemperatureHigh = uint8(bmsFault & 1)
	msg.IsChargeShortCircuit = uint8((bmsFault >> 1) & 1)
	// NOTE(Java-Bug-Replication): These two lines intentionally replicate a bug from the original Java code.
	// In the original Java code (Bin66Decode.java L91-92):
	// bmsMessage.setIsDischargeOverCurrent(isDischargeOverCurrent);
	// bmsMessage.setIsDischargeOverCurrent(isChargeOverCurrent); // BUG: overrides the discharge flag with the charge flag
	msg.IsDischargeOverCurrent = uint8((bmsFault >> 3) & 1)
	// As a result of the bug above, the original Java code never sets isChargeOverCurrent, so it remains its default value (0).
	msg.IsChargeOverCurrent = 0
	msg.IsUnderVoltage = uint8((bmsFault >> 4) & 1)
	msg.IsOverVoltage = uint8((bmsFault >> 5) & 1)
	msg.IsNodeUnderVoltage = uint8((bmsFault >> 6) & 1)
	msg.IsNodeOverVoltage = uint8((bmsFault >> 7) & 1)
	msg.IsErrorCollect = uint8((bmsFault >> 8) & 1)
	msg.IsDischargeLowTemperature = uint8((bmsFault >> 9) & 1)
	msg.IsChargeLowTemperature = uint8((bmsFault >> 10) & 1)
	msg.IsDischargeHighTemperature = uint8((bmsFault >> 11) & 1)

	return msg, nil
}

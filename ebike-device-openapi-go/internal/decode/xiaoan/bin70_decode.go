package xiaoan

import "ebike-device-openapi-go/internal/api/dto"

type Bin70NotifyFaultMessage struct {
	Imei          *string `json:"imei,omitempty"`
	MsgType       string  `json:"msgType"`
	Cmd           int16   `json:"cmd"`
	BussinessType string  `json:"bussinessType"`
	Timestamp     int64   `json:"timestamp"`

	EtcFault int32 `json:"etcFault"`
	BmsFault int32 `json:"bmsFault"`
	EcuFault int32 `json:"ecuFault"`

	LockedRotor                 uint8 `json:"lockedRotor"`
	ThrottleFault               uint8 `json:"throttleFault"`
	IsUnderVoltage              uint8 `json:"isUnderVoltage"`
	IsOverVoltage               uint8 `json:"isOverVoltage"`
	BrakeFault                  uint8 `json:"brakeFault"`
	HallFault                   uint8 `json:"hallFault"`
	VehicleSpeeding             uint8 `json:"vehicleSpeeding"`
	Upgrading                   uint8 `json:"upgrading"`
	ControlFailure              uint8 `json:"controlFailure"`
	ControlCommunicationFailure uint8 `json:"controlCommunicationFailure"`

	Overheating                uint8 `json:"overheating"`
	ShortCircuitFault          uint8 `json:"shortCircuitFault"`
	DischargeOverCurrent       uint8 `json:"dischargeOverCurrent"`
	ChargeOverCurrent          uint8 `json:"chargeOverCurrent"`
	UnderVoltageFault          uint8 `json:"underVoltageFault"`
	OverVoltageFault           uint8 `json:"overVoltageFault"`
	OverloadFailure            uint8 `json:"overloadFailure"`
	SingleUnderVoltageFault    uint8 `json:"singleUnderVoltageFault"`
	SingleOverVoltageFault     uint8 `json:"singleOverVoltageFault"`
	FrontAcquisitionError      uint8 `json:"frontAcquisitionError"`
	IsDischargeLowTemperature  uint8 `json:"isDischargeLowTemperature"`
	IsChargeLowTemperature     uint8 `json:"isChargeLowTemperature"`
	IsDischargeHighTemperature uint8 `json:"isDischargeHighTemperature"`
	BmsCommunicationFailure    uint8 `json:"bmsCommunicationFailure"`

	SensorFailure                  uint8 `json:"sensorFailure"`
	ElectricDoorLockFailure        uint8 `json:"electricDoorLockFailure"`
	RfidFailure                    uint8 `json:"rfidFailure"`
	ReturnButtonFailure            uint8 `json:"returnButtonFailure"`
	InstrumentCommunicationFailure uint8 `json:"instrumentCommunicationFailure"`
}

type Bin70Decode struct{}

func (d *Bin70Decode) Decode(header *dto.MessageHeader, data *ByteBuf) (interface{}, error) {
	msg := &Bin70NotifyFaultMessage{
		MsgType:       "alarm",
		Cmd:           70, // CMD_FAULT
		BussinessType: "ebike",
	}

	etcFault := data.ReadUnsignedShort()
	bmsFault := data.ReadUnsignedShort()
	ecuFault := data.ReadUnsignedShort()

	msg.EtcFault = int32(etcFault)
	msg.BmsFault = int32(bmsFault)
	msg.EcuFault = int32(ecuFault)

	msg.LockedRotor = uint8(etcFault & 1)
	msg.ThrottleFault = uint8((etcFault >> 1) & 1)
	msg.IsUnderVoltage = uint8((etcFault >> 2) & 1)
	msg.IsOverVoltage = uint8((etcFault >> 3) & 1)
	msg.BrakeFault = uint8((etcFault >> 4) & 1)
	msg.HallFault = uint8((etcFault >> 5) & 1)
	msg.VehicleSpeeding = uint8((etcFault >> 6) & 1)
	msg.Upgrading = uint8((etcFault >> 7) & 1)
	msg.ControlFailure = uint8((etcFault >> 8) & 1)
	msg.ControlCommunicationFailure = uint8((etcFault >> 15) & 1)

	msg.Overheating = uint8(bmsFault & 1)
	msg.ShortCircuitFault = uint8((bmsFault >> 1) & 1)
	msg.DischargeOverCurrent = uint8((bmsFault >> 2) & 1)
	msg.ChargeOverCurrent = uint8((bmsFault >> 3) & 1)
	msg.UnderVoltageFault = uint8((bmsFault >> 4) & 1)
	// Java BUG REPLICATION: In the original Java code, Bin70Decode forgot to set the overVoltageFault property
	// (missing `msg.setOverVoltageFault(...)`), so it always outputs as 0/null in JSON regardless of bit 5.
	// We force it to 0 here to maintain exact JSON output compatibility.
	msg.OverVoltageFault = 0 // originally: uint8((bmsFault >> 5) & 1)
	msg.OverloadFailure = uint8((ecuFault >> 3) & 1)
	msg.SingleUnderVoltageFault = uint8((bmsFault >> 6) & 1)
	msg.SingleOverVoltageFault = uint8((bmsFault >> 7) & 1)
	msg.FrontAcquisitionError = uint8((bmsFault >> 8) & 1)
	msg.IsDischargeLowTemperature = uint8((bmsFault >> 9) & 1)
	msg.IsChargeLowTemperature = uint8((bmsFault >> 10) & 1)
	msg.IsDischargeHighTemperature = uint8((bmsFault >> 11) & 1)
	msg.BmsCommunicationFailure = uint8((bmsFault >> 12) & 1)

	msg.SensorFailure = uint8(ecuFault & 1)
	msg.ElectricDoorLockFailure = uint8((ecuFault >> 1) & 1)
	msg.RfidFailure = uint8((ecuFault >> 2) & 1)
	msg.ReturnButtonFailure = uint8((ecuFault >> 4) & 1)
	msg.InstrumentCommunicationFailure = uint8((ecuFault >> 5) & 1)

	return msg, nil
}

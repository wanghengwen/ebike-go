package xiaoan

import "ebike-device-openapi-go/internal/api/dto"

type Bin72NotifyMessage struct {
	Imei          *string `json:"imei,omitempty"`
	MsgType       string  `json:"msgType"`
	Cmd           int16   `json:"cmd"`
	BussinessType string  `json:"bussinessType"`
	Timestamp     int64   `json:"timestamp"`

	Type     int    `json:"type"`
	Payload  int    `json:"payload"`
	Reserved uint32 `json:"reserved"`

	// Type 2 fields
	LockedRotor    *uint8 `json:"lockedRotor,omitempty"`
	ThrottleFault  *uint8 `json:"throttleFault,omitempty"`
	IsUndervoltage *uint8 `json:"isUndervoltage,omitempty"`
	IsOvervoltage  *uint8 `json:"isOvervoltage,omitempty"`
	BrakeFault     *uint8 `json:"brakeFault,omitempty"`
	HallFault      *uint8 `json:"hallFault,omitempty"`

	// Type 7 fields
	ControlUpgradeSuc  *uint8 `json:"controlUpgradeSuc,omitempty"`
	ControlUpgradeFail *uint8 `json:"controlUpgradeFail,omitempty"`
	BmsUpgradeSuc      *uint8 `json:"bmsUpgradeSuc,omitempty"`
	BmsUpgradeFail     *uint8 `json:"bmsUpgradeFail,omitempty"`
	RfidUpgradeSuc     *uint8 `json:"rfidUpgradeSuc,omitempty"`
	RfidUpgradeFail    *uint8 `json:"rfidUpgradeFail,omitempty"`
	CameraUpgradeSuc   *uint8 `json:"cameraUpgradeSuc,omitempty"`
	CameraUpgradeFail  *uint8 `json:"cameraUpgradeFail,omitempty"`

	// Type 8 fields
	ControlCommunicationDisconnected *uint8 `json:"controlCommunicationDisconnected,omitempty"`
	ControlCommunicationRecovery     *uint8 `json:"controlCommunicationRecovery,omitempty"`
	BmsCommunicationDisconnected     *uint8 `json:"bmsCommunicationDisconnected,omitempty"`
	BmsCommunicationRecovery         *uint8 `json:"bmsCommunicationRecovery,omitempty"`
	RfidCommunicationDisconnected    *uint8 `json:"rfidCommunicationDisconnected,omitempty"`
	RfidCommunicationRecovery        *uint8 `json:"rfidCommunicationRecovery,omitempty"`
	CameraCommunicationDisconnected  *uint8 `json:"cameraCommunicationDisconnected,omitempty"`
	CameraCommunicationRecovery      *uint8 `json:"cameraCommunicationRecovery,omitempty"`

	// Type 9 fields
	FrontSaddleContact *uint8 `json:"frontSaddleContact,omitempty"`
	CentSaddleContact  *uint8 `json:"centSaddleContact,omitempty"`
	BackSaddleContact  *uint8 `json:"backSaddleContact,omitempty"`

	// Type 10 fields
	ReturnConditions        *uint8 `json:"returnConditions,omitempty"`
	SwitchOn                *uint8 `json:"switchOn,omitempty"`
	OneClickReturn          *uint8 `json:"oneClickReturn,omitempty"`
	GpsNotInParkingArea     *uint8 `json:"gpsNotInParkingArea,omitempty"`
	RfidInductionFailed     *uint8 `json:"rfidInductionFailed,omitempty"`
	WrongParkingAngle       *uint8 `json:"wrongParkingAngle,omitempty"`
	RoadSpikeSignalIsWeak   *uint8 `json:"roadSpikeSignalIsWeak,omitempty"`
	CameraRecognitionFailed *uint8 `json:"cameraRecognitionFailed,omitempty"`
	HelmetNotReturned       *uint8 `json:"helmetNotReturned,omitempty"`
}

type Bin72Decode struct{}

func (d *Bin72Decode) Decode(header *dto.MessageHeader, data *ByteBuf) (interface{}, error) {
	msg := &Bin72NotifyMessage{
		MsgType:       "alarm",
		Cmd:           5, // CmdConstant.CMD_ALARM = 5
		BussinessType: "ebike",
	}

	t := data.ReadUnsignedShort()
	payload := data.ReadUnsignedShort()
	reserved := data.ReadUnsignedInt()

	msg.Payload = int(payload)
	msg.Reserved = reserved

	switch t {
	case 2:
		msg.LockedRotor = ptrUint8(uint8(payload & 1))
		msg.ThrottleFault = ptrUint8(uint8((payload >> 1) & 1))
		msg.IsUndervoltage = ptrUint8(uint8((payload >> 2) & 1))
		msg.IsOvervoltage = ptrUint8(uint8((payload >> 3) & 1))
		msg.BrakeFault = ptrUint8(uint8((payload >> 4) & 1))
		msg.HallFault = ptrUint8(uint8((payload >> 5) & 1))
	case 7:
		msg.ControlUpgradeSuc = ptrUint8(uint8((payload >> 1) & 1))
		msg.ControlUpgradeFail = ptrUint8(uint8((payload >> 2) & 1))
		msg.BmsUpgradeSuc = ptrUint8(uint8((payload >> 3) & 1))
		msg.BmsUpgradeFail = ptrUint8(uint8((payload >> 4) & 1))
		msg.RfidUpgradeSuc = ptrUint8(uint8((payload >> 5) & 1))
		msg.RfidUpgradeFail = ptrUint8(uint8((payload >> 6) & 1))
		msg.CameraUpgradeSuc = ptrUint8(uint8((payload >> 7) & 1))
		msg.CameraUpgradeFail = ptrUint8(uint8((payload >> 8) & 1))
	case 8:
		msg.ControlCommunicationDisconnected = ptrUint8(uint8((payload >> 2) & 1))
		msg.ControlCommunicationRecovery = ptrUint8(uint8((payload >> 3) & 1))
		msg.BmsCommunicationDisconnected = ptrUint8(uint8((payload >> 4) & 1))
		msg.BmsCommunicationRecovery = ptrUint8(uint8((payload >> 5) & 1))
		msg.RfidCommunicationDisconnected = ptrUint8(uint8((payload >> 6) & 1))
		msg.RfidCommunicationRecovery = ptrUint8(uint8((payload >> 7) & 1))
		msg.CameraCommunicationDisconnected = ptrUint8(uint8((payload >> 8) & 1))
		msg.CameraCommunicationRecovery = ptrUint8(uint8((payload >> 9) & 1))
	case 9:
		msg.FrontSaddleContact = ptrUint8(uint8(payload & 1))
		msg.CentSaddleContact = ptrUint8(uint8((payload >> 1) & 1))
		msg.BackSaddleContact = ptrUint8(uint8((payload >> 2) & 1))
	case 10:
		msg.ReturnConditions = ptrUint8(uint8(payload & 1))
		msg.SwitchOn = ptrUint8(uint8((payload >> 1) & 1))
		msg.OneClickReturn = ptrUint8(uint8((payload >> 2) & 1))
		msg.GpsNotInParkingArea = ptrUint8(uint8((reserved >> 1) & 1))
		msg.RfidInductionFailed = ptrUint8(uint8((reserved >> 2) & 1))
		msg.WrongParkingAngle = ptrUint8(uint8((reserved >> 3) & 1))
		msg.RoadSpikeSignalIsWeak = ptrUint8(uint8((reserved >> 4) & 1))
		msg.CameraRecognitionFailed = ptrUint8(uint8((reserved >> 5) & 1))
		msg.HelmetNotReturned = ptrUint8(uint8((reserved >> 6) & 1))
	}

	// NOTE(Java-Quirk-Replication): In Java's Bin72Decode.java, the original author explicitly wrote:
	// "// 和老预警区分 type += 1000;"
	// We preserve this magic number addition to ensure strict downstream compatibility.
	msg.Type = int(t) + 1000
	return msg, nil
}

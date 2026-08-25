package xiaoan

import (
	"ebike-device-openapi-go/internal/api/dto"
	"ebike-device-openapi-go/internal/pkg/utils"
	"encoding/hex"
	"sort"
	"strings"
)

type Bin68GpsMessage struct {
	Imei          *string `json:"imei,omitempty"`
	MsgType       string  `json:"msgType"`
	Cmd           int16   `json:"cmd"`
	BussinessType string  `json:"bussinessType"`

	Sw           uint64  `json:"sw"`
	Gsm          uint8   `json:"gsm"`
	Voltage      uint64  `json:"voltage"`
	Timestamp    uint32  `json:"timestamp"`
	Wgs84Lng     float64 `json:"wgs84Lng"`
	Wgs84Lat     float64 `json:"wgs84Lat"`
	Lng          float64 `json:"lng"`
	Lat          float64 `json:"lat"`
	Speed        uint16  `json:"speed"`
	Course       uint16  `json:"course"`
	Hdop         uint16  `json:"hdop"`
	Satellite    uint8   `json:"satellite"`
	TotalMiles   uint32  `json:"totalMiles"`
	FenceVersion uint32  `json:"fenceVersion"`

	Defend          uint8 `json:"defend"`
	Acc             uint8 `json:"acc"`
	BackWheelLock   uint8 `json:"backWheelLock"`
	BatteryLock     uint8 `json:"batteryLock"`
	BatteryConnect  uint8 `json:"batteryConnect"`
	IsGPSFastMode   uint8 `json:"isGPSFastMode"`
	IsMoving        uint8 `json:"isMoving"`
	IsWheelSpan     uint8 `json:"isWheelSpan"`
	IsHelmetUnlock  uint8 `json:"isHelmetUnlock"`
	IsSleepMode     uint8 `json:"isSleepMode"`
	IsMoveAlarmOn   uint8 `json:"isMoveAlarmOn"`
	IsOverSpeedOn   uint8 `json:"isOverSpeedOn"`
	IsGPSUnfixed    uint8 `json:"isGPSUnfixed"`
	IsGyroFixed     uint8 `json:"isGyroFixed"`
	IsFenceEnable   uint8 `json:"isFenceEnable"`
	IsOutofServAera uint8 `json:"isOutofServAera"`
	IsPowerCut      uint8 `json:"isPowerCut"`

	CourseX *int32 `json:"courseX,omitempty"`
	CourseY *int32 `json:"courseY,omitempty"`
	CourseZ *int32 `json:"courseZ,omitempty"`

	BmsSoc       *uint8   `json:"bmsSoc,omitempty"`
	BmsVoltage   int32   `json:"bmsVoltage"`
	BmsSN        *string  `json:"bmsSN,omitempty"`
	SpikeRssi    *string  `json:"spikeRssi,omitempty"`
	RfidAck      *int16   `json:"rfidAck,omitempty"`
	VoltageState *uint8   `json:"voltageState,omitempty"`

	HeadingAngle        *int16 `json:"headingAngle,omitempty"`
	ControllerSpeed     int32 `json:"controllerSpeed"`
	IsSupportHelmetLock *uint8 `json:"isSupportHelmetLock,omitempty"`
	HelmetType          *int32 `json:"helmetType,omitempty"`
	HelmetLock          int32 `json:"helmetLock"`
	HelmetReact         int32 `json:"helmetReact"`

	KickStandType          *uint8  `json:"kickStandType,omitempty"`
	KickStandRFIDState     *uint8  `json:"kickStandRFIDState,omitempty"`
	KickStandMagneticState *uint8  `json:"kickStandMagneticState,omitempty"`
	KickStandRFIDData      *string `json:"kickStandRFIDData,omitempty"`
	Mile                   *uint32 `json:"mile,omitempty"`
	BikeState              *int16  `json:"bikeState,omitempty"`
	ParkState              *int16  `json:"parkState,omitempty"`
	RemainMiles            *uint64 `json:"remainMiles,omitempty"`

	Overload               *int16   `json:"overload,omitempty"`
	OverloadThreshold      *int16   `json:"overloadThreshold,omitempty"`
	IsSupportOverload      *uint8   `json:"isSupportOverload,omitempty"`
	HelmetSOC              *int16   `json:"helmetSOC,omitempty"`
	HelmetState            *int16   `json:"helmetState,omitempty"`
	HelmetHasAngle         *int16   `json:"helmetHasAngle,omitempty"`
	HelmetHasCapacitance   *int16   `json:"helmetHasCapacitance,omitempty"`
	HelmetHasTinfrared     *int16   `json:"helmetHasTinfrared,omitempty"`
	HelmetHasPressure      *int16   `json:"helmetHasPressure,omitempty"`
	HelmetAngle            *int16   `json:"helmetAngle,omitempty"`
	// CRITICAL: Must be []byte, NOT []int. Java's original type is byte[], which Fastjson/Jackson
	// serialize as a Base64 string by default. Go's encoding/json base64-encodes []byte naturally.
	HelmetCapacitance      []byte `json:"helmetCapacitance"`
	HelmeTinfrared         *int16   `json:"helmeTinfrared,omitempty"`
	HelmePressure          *int16   `json:"helmePressure,omitempty"`
	HelmetAngleFault       *int16   `json:"helmetAngleFault,omitempty"`
	HelmetCapacitanceFault *int16   `json:"helmetCapacitanceFault,omitempty"`
	HelmetTinfraredFault   *int16   `json:"helmetTinfraredFault,omitempty"`
	HelmetPressureFault    *int16   `json:"helmetPressureFault,omitempty"`

	HelmetBind *uint8 `json:"helmetBind,omitempty"`

	ElectricEnergy *float64 `json:"electricEnergy,omitempty"`

	LastHelmetLockState *uint8  `json:"lastHelmetLockState,omitempty"`
	ReturnHelmet        *uint8  `json:"returnHelmet,omitempty"`
	HelmetFault         *int32  `json:"helmetFault,omitempty"`
	HelmetVoltage       *int32  `json:"helmetVoltage,omitempty"`
	HelmetSn            *string `json:"helmetSn,omitempty"`
	BmsManufacturer     *string `json:"bmsManufacturer,omitempty"`

	AssetID   *string `json:"assetID,omitempty"`
	AssetType *string `json:"assetType,omitempty"`
}


func ptrString(v string) *string { return &v }
func ptrUint8(v uint8) *uint8 { return &v }
func ptrUint16(v uint16) *uint16 { return &v }
func ptrUint32(v uint32) *uint32 { return &v }
func ptrUint64(v uint64) *uint64 { return &v }
func ptrInt16(v int16) *int16 { return &v }
func ptrInt32(v int32) *int32 { return &v }
func ptrInt64(v int64) *int64 { return &v }
func ptrFloat64(v float64) *float64 { return &v }

type Bin68Decode struct{}

func (d *Bin68Decode) Decode(header *dto.MessageHeader, data *ByteBuf) (interface{}, error) {
	msg := &Bin68GpsMessage{
		MsgType:       "data",
		Cmd:           3, // CmdConstant.CMD_GPS_1 = 3
		BussinessType: "ebike",
		HelmetCapacitance: []byte{0, 0, 0, 0},
		HelmetBind: ptrUint8(0),
	}

	sw := uint64(data.ReadUnsignedInt())
	msg.Sw = sw

	msg.Gsm = data.ReadUnsignedByte()
	msg.Voltage = uint64(data.ReadUnsignedInt())
	msg.Timestamp = data.ReadUnsignedInt()

	_wgs84Lng := data.ReadUnsignedInt()
	_wgs84Lat := data.ReadUnsignedInt()

	wgs84Lng := float64(_wgs84Lng) * 0.000001
	wgs84Lat := float64(_wgs84Lat) * 0.000001
	
	msg.Wgs84Lng = wgs84Lng
	msg.Wgs84Lat = wgs84Lat

	lng, lat := utils.TransformWGS84ToGCJ02(wgs84Lng, wgs84Lat)
	msg.Lng = lng
	msg.Lat = lat

	msg.Speed = data.ReadUnsignedShort()
	msg.Course = data.ReadUnsignedShort()
	msg.Hdop = data.ReadUnsignedShort()
	msg.Satellite = data.ReadUnsignedByte()
	msg.TotalMiles = data.ReadUnsignedInt()
	msg.FenceVersion = data.ReadUnsignedInt()

	msg.Defend = uint8(sw & 1)
	msg.Acc = uint8((sw >> 1) & 1)
	msg.BackWheelLock = uint8((sw >> 2) & 1)
	msg.BatteryLock = uint8((sw >> 3) & 1)
	msg.BatteryConnect = uint8((sw >> 4) & 1)
	msg.IsGPSFastMode = uint8((sw >> 5) & 1)
	msg.IsMoving = uint8((sw >> 6) & 1)
	msg.IsWheelSpan = uint8((sw >> 7) & 1)
	msg.IsHelmetUnlock = uint8((sw >> 8) & 1)
	msg.IsSleepMode = uint8((sw >> 9) & 1)
	msg.IsMoveAlarmOn = uint8((sw >> 10) & 1)
	msg.IsOverSpeedOn = uint8((sw >> 11) & 1)
	msg.IsGPSUnfixed = uint8((sw >> 12) & 1)
	msg.IsGyroFixed = uint8((sw >> 13) & 1)
	msg.IsFenceEnable = uint8((sw >> 14) & 1)
	msg.IsOutofServAera = uint8((sw >> 15) & 1)
	msg.IsPowerCut = uint8((sw >> 18) & 1)

	// Decode TLV.
	// NOTE(Java-Alignment): The original Java (Bin68Decode + TlvUtil.buildTlvData + TLVDecode) reads all
	// TLVs into a TreeMap<Integer, ByteBuf>, which de-duplicates by tag (keeping the LAST occurrence) and
	// iterates tags in ASCENDING order. We replicate both semantics here.
	// Each TLV value is a sub-buffer of exactly `length` bytes; in Java any handler that reads past
	// `length` throws IndexOutOfBoundsException, and since the decode chain has no try/catch the WHOLE
	// message is dropped. We mirror that by returning an error when the main buffer underruns while
	// slicing TLVs, or when a value sub-buffer is later read out of bounds.
	type tlvEntry struct {
		length uint8
		value  *ByteBuf
	}
	tlvMap := make(map[uint8]tlvEntry)
	for data.ReadableBytes() > 0 {
		tag := data.ReadUnsignedByte()
		length := data.ReadUnsignedByte()
		value := data.ReadBytes(int(length))
		if data.Err() != nil {
			// Mirrors Netty data.readBytes(len) throwing IndexOutOfBoundsException during buildTlvData.
			return nil, data.Err()
		}
		// TreeMap.put semantics: a later occurrence of the same tag overwrites the earlier one.
		tlvMap[tag] = tlvEntry{length: length, value: NewByteBuf(value)}
	}

	tags := make([]int, 0, len(tlvMap))
	for tag := range tlvMap {
		tags = append(tags, int(tag))
	}
	sort.Ints(tags)

	for _, t := range tags {
		tag := uint8(t)
		entry := tlvMap[tag]
		length := entry.length
		tagValue := entry.value

		switch tag {
		case 0x00: // angle
			// NOTE(Java-Bug-Replication): This block intentionally replicates a bug from the original Java code.
			// In the original Java code (TLVDecode.java L28-30):
			// case 0x00: //angle
			//     message.setCourseX(tagValue.readUnsignedShort());
			//     message.setCourseX(tagValue.readUnsignedShort()); // BUG: Overwrites CourseX with the second short, skipping CourseY.
			_ = tagValue.ReadUnsignedShort()
			msg.CourseX = ptrInt32(int32(tagValue.ReadUnsignedShort()))
			// As a result of the bug above, the original Java code never sets CourseY, so it remains its default value (null).
			msg.CourseY = nil
			msg.CourseZ = ptrInt32(int32(tagValue.ReadUnsignedShort()))
		case 0x01: // bmsSoc
			msg.BmsSoc = ptrUint8(tagValue.ReadUnsignedByte())
		case 0x02: // bmsVoltage
			msg.BmsVoltage = int32(tagValue.ReadUnsignedShort())
		case 0x03: // etcSpeed
			msg.ControllerSpeed = int32(tagValue.ReadUnsignedShort())
		case 0x04: // bmsSN
			msg.BmsSN = ptrString(strings.TrimSpace(tagValue.ReadCharSequence(20)))
		case 0x05: // RFID
			if length == 16 {
				msg.RfidAck = ptrInt16(0)
				msg.SpikeRssi = ptrString(strings.TrimSpace(tagValue.ReadCharSequence(16)))
			} else {
				msg.RfidAck = ptrInt16(tagValue.ReadShort())
			}
		case 0x09: // headingAngle
			msg.HeadingAngle = ptrInt16(tagValue.ReadShort())
		case 0x0A: // voltageState
			msg.VoltageState = ptrUint8(tagValue.ReadUnsignedByte())
		case 0x0B: // helmetLockState
			msg.IsSupportHelmetLock = ptrUint8(1)
			helmetLockState := tagValue.ReadUnsignedByte()
			helmetType := int32(helmetLockState & 1)
			helmetReact := int32((helmetLockState >> 2) & 1)
			helmetLock := int32((helmetLockState >> 3) & 1)
			if helmetType == 0 { // 4芯
				helmetLock = helmetReact
			}
			msg.HelmetType = ptrInt32(helmetType)
			msg.HelmetReact = helmetReact
			msg.HelmetLock = helmetLock
		case 0x0C: // kickStandState
			kickStandType := tagValue.ReadUnsignedByte()
			msg.KickStandType = ptrUint8(kickStandType)
			msg.KickStandMagneticState = ptrUint8(tagValue.ReadUnsignedByte())
			if kickStandType == 1 {
				msg.KickStandRFIDState = ptrUint8(tagValue.ReadUnsignedByte())
				dataLen := int(length) - 3
				if dataLen > 0 {
					msg.KickStandRFIDData = ptrString(strings.TrimSpace(tagValue.ReadCharSequence(dataLen)))
				}
			}
		case 0x0D: // tBeaconRealRssi
			// ignored in java
		case 0x0E: // mile
			msg.Mile = ptrUint32(uint32(tagValue.ReadUnsignedShort()))
		case 0x0F: // bikeState
			msg.BikeState = ptrInt16(int16(tagValue.ReadUnsignedByte()))
		case 0x10: // parkState
			msg.ParkState = ptrInt16(int16(tagValue.ReadUnsignedByte()))
		case 0x11: // remainMiles
			msg.RemainMiles = ptrUint64(uint64(tagValue.ReadUnsignedInt()))
		case 0x12: // SHelmet
			msg.HelmetBind = ptrUint8(1)
			msg.HelmetSOC = ptrInt16(int16(tagValue.ReadUnsignedByte()))
			helmetState := tagValue.ReadUnsignedByte()
			msg.HelmetState = ptrInt16(int16(helmetState & 1))

			helmetFuncType := tagValue.ReadUnsignedByte()
			msg.HelmetHasAngle = ptrInt16(int16(helmetFuncType & 1))
			msg.HelmetHasCapacitance = ptrInt16(int16((helmetFuncType >> 1) & 1))
			msg.HelmetHasTinfrared = ptrInt16(int16((helmetFuncType >> 2) & 1))
			msg.HelmetHasPressure = ptrInt16(int16((helmetFuncType >> 3) & 1))

			msg.HelmetAngle = ptrInt16(int16(tagValue.ReadUnsignedByte()))

			helmetCapacitanceNum := tagValue.ReadUnsignedByte()
			msg.HelmetCapacitance = []byte{
				byte(helmetCapacitanceNum & 1),
				byte((helmetCapacitanceNum >> 1) & 1),
				byte((helmetCapacitanceNum >> 2) & 1),
				byte((helmetCapacitanceNum >> 3) & 1),
			}

			msg.HelmeTinfrared = ptrInt16(int16(tagValue.ReadUnsignedByte()))
			msg.HelmePressure = ptrInt16(int16(tagValue.ReadUnsignedByte()))
			helmetFault := int32(tagValue.ReadUnsignedByte())
			msg.HelmetFault = ptrInt32(helmetFault)

			msg.HelmetAngleFault = ptrInt16(int16(helmetFault & 1))
			msg.HelmetCapacitanceFault = ptrInt16(int16((helmetFault >> 1) & 1))
			msg.HelmetTinfraredFault = ptrInt16(int16((helmetFault >> 2) & 1))
			msg.HelmetPressureFault = ptrInt16(int16((helmetFault >> 3) & 1))

		case 0x13: // overload
			msg.Overload = ptrInt16(int16(tagValue.ReadUnsignedByte()))
			msg.OverloadThreshold = ptrInt16(int16(tagValue.ReadUnsignedByte()))
			msg.IsSupportOverload = ptrUint8(1)
		case 0x16:
			// ignored in java
		case 0x17:
			msg.AssetID = ptrString(tagValue.ReadCharSequence(int(length)))
		case 0x18:
			msg.AssetType = ptrString(tagValue.ReadCharSequence(int(length)))
		case 0x19: // helmetLock 11 bytes
			msg.HelmetBind = ptrUint8(1)
			msg.IsSupportHelmetLock = ptrUint8(1)
			msg.HelmetType = ptrInt32(1)

			helmetLockYKT := tagValue.ReadUnsignedByte()
			msg.LastHelmetLockState = ptrUint8(helmetLockYKT & 1)
			msg.HelmetReact = int32((helmetLockYKT >> 1) & 1)
			msg.HelmetLock = int32((helmetLockYKT >> 2) & 1)
			msg.HelmetState = ptrInt16(int16((helmetLockYKT >> 5) & 1))
			msg.ReturnHelmet = ptrUint8((helmetLockYKT >> 6) & 1)

			msg.HelmetFault = ptrInt32(int32(tagValue.ReadUnsignedShort()))
			msg.HelmetVoltage = ptrInt32(int32(tagValue.ReadUnsignedShort() * 10))

			if tagValue.ReadableBytes() > 0 {
				bs := tagValue.ReadBytes(6)
				helmetSn := hex.EncodeToString(bs)
				for len(helmetSn) < 12 {
					helmetSn = "0" + helmetSn
				}
				msg.HelmetSn = ptrString(helmetSn)
			}
		case 0x20: // electricEnergy
			msg.ElectricEnergy = ptrFloat64(float64(tagValue.ReadUnsignedShort()) * 0.1)
		case 0x21:
			msg.BmsManufacturer = ptrString(strings.TrimSpace(tagValue.ReadCharSequence(16)))
		default:
			// Ignore other TLV
		}

		// Java-Alignment: a handler reading past the TLV value length throws IndexOutOfBoundsException in
		// Netty, which (absent any try/catch) drops the whole message. Mirror it by failing the decode.
		if tagValue.Err() != nil {
			return nil, tagValue.Err()
		}
	}

	return msg, nil
}

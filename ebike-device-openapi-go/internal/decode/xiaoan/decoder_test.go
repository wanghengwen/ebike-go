package xiaoan

import (
	"encoding/binary"
	"math"
	"testing"

	"ebike-device-openapi-go/internal/api/dto"
)

func headerFor(cmd int16) *dto.MessageHeader {
	return &dto.MessageHeader{Cmd: cmd}
}

func TestDecodeBin2(t *testing.T) {
	body := []byte{0x1F, 0x30} // gsm=31, voltage=48
	msg, err := DecodeHex(headerFor(2), NewByteBuf(body))
	if err != nil {
		t.Fatalf("decode bin2: %v", err)
	}
	ping, ok := msg.(*Bin2PingMessage)
	if !ok {
		t.Fatalf("expected *Bin2PingMessage, got %T", msg)
	}
	if ping.MsgType != "data" || ping.Cmd != 2 || ping.BussinessType != "ebike" {
		t.Fatalf("unexpected meta: %+v", ping)
	}
	if ping.GsmSignal != 31 || ping.Voltage != 48 {
		t.Fatalf("unexpected fields: gsm=%d voltage=%d", ping.GsmSignal, ping.Voltage)
	}
}

func TestDecodeBin5(t *testing.T) {
	body := []byte{0x03}
	msg, err := DecodeHex(headerFor(5), NewByteBuf(body))
	if err != nil {
		t.Fatalf("decode bin5: %v", err)
	}
	alarm, ok := msg.(*Bin5AlarmMessage)
	if !ok {
		t.Fatalf("expected *Bin5AlarmMessage, got %T", msg)
	}
	if alarm.MsgType != "alarm" || alarm.Cmd != 5 || alarm.Type != 3 {
		t.Fatalf("unexpected alarm: %+v", alarm)
	}
}

func TestDecodeBin41OutputsCmd3(t *testing.T) {
	body := make([]byte, 27)
	binary.BigEndian.PutUint16(body[0:2], 0)
	body[2] = 31
	binary.BigEndian.PutUint32(body[3:7], 48000)
	binary.BigEndian.PutUint32(body[7:11], 1700000000)
	binary.LittleEndian.PutUint32(body[11:15], math.Float32bits(118.378037))
	binary.LittleEndian.PutUint32(body[15:19], math.Float32bits(33.780653))
	body[19] = 0
	binary.BigEndian.PutUint16(body[20:22], 228)
	binary.LittleEndian.PutUint32(body[22:26], math.Float32bits(0.6))
	body[26] = 24

	msg, err := DecodeHex(headerFor(41), NewByteBuf(body))
	if err != nil {
		t.Fatalf("decode bin41: %v", err)
	}
	gps, ok := msg.(*Bin41GpsMessage)
	if !ok {
		t.Fatalf("expected *Bin41GpsMessage, got %T", msg)
	}
	if gps.Cmd != 3 || gps.MsgType != "data" {
		t.Fatalf("bin41 should output cmd=3 data, got cmd=%d msgType=%s", gps.Cmd, gps.MsgType)
	}
	if gps.Gsm != 31 || gps.Satellite != 24 {
		t.Fatalf("unexpected gps fields: %+v", gps)
	}
}

func TestDecodeBin66(t *testing.T) {
	body := make([]byte, 47)
	copy(body[0:20], []byte("TEST-SN-00000000001"))
	binary.BigEndian.PutUint16(body[20:22], 1)
	binary.BigEndian.PutUint16(body[22:24], 2)
	binary.BigEndian.PutUint16(body[24:26], 0xFF00) // mos temp
	body[26] = 1
	body[29] = 80 // soh
	binary.BigEndian.PutUint16(body[30:32], 0x0008) // bmsFault bit3 set
	body[37] = 65 // soc
	binary.BigEndian.PutUint32(body[43:47], 1700000000)

	msg, err := DecodeHex(headerFor(66), NewByteBuf(body))
	if err != nil {
		t.Fatalf("decode bin66: %v", err)
	}
	bms, ok := msg.(*Bin66BmsInfoMessage)
	if !ok {
		t.Fatalf("expected *Bin66BmsInfoMessage, got %T", msg)
	}
	if bms.Cmd != 66 || bms.MsgType != "data" || bms.Sn != "TEST-SN-00000000001" {
		t.Fatalf("unexpected bms meta: %+v", bms)
	}
	if bms.IsDischargeOverCurrent != 1 || bms.IsChargeOverCurrent != 0 {
		t.Fatalf("bin66 java bug replication failed: discharge=%d charge=%d",
			bms.IsDischargeOverCurrent, bms.IsChargeOverCurrent)
	}
}

func TestDecodeBin68OutputsCmd3(t *testing.T) {
	body := make([]byte, 36)
	binary.BigEndian.PutUint32(body[0:4], 1)
	body[4] = 31
	binary.BigEndian.PutUint32(body[5:9], 48000)
	binary.BigEndian.PutUint32(body[9:13], 1700000000)
	binary.BigEndian.PutUint32(body[13:17], 118378037)
	binary.BigEndian.PutUint32(body[17:21], 33780653)
	binary.BigEndian.PutUint16(body[21:23], 0)
	binary.BigEndian.PutUint16(body[23:25], 228)
	binary.BigEndian.PutUint16(body[25:27], 6)
	body[27] = 24
	binary.BigEndian.PutUint32(body[28:32], 1000)
	binary.BigEndian.PutUint32(body[32:36], 2)

	msg, err := DecodeHex(headerFor(68), NewByteBuf(body))
	if err != nil {
		t.Fatalf("decode bin68: %v", err)
	}
	gps, ok := msg.(*Bin68GpsMessage)
	if !ok {
		t.Fatalf("expected *Bin68GpsMessage, got %T", msg)
	}
	if gps.Cmd != 3 || gps.MsgType != "data" {
		t.Fatalf("bin68 should output cmd=3, got cmd=%d", gps.Cmd)
	}
	if gps.Gsm != 31 || gps.Satellite != 24 {
		t.Fatalf("unexpected bin68 fields: %+v", gps)
	}
}

func TestDecodeBin70OverVoltageFaultAlwaysZero(t *testing.T) {
	body := make([]byte, 6)
	binary.BigEndian.PutUint16(body[0:2], 0)
	binary.BigEndian.PutUint16(body[2:4], 0x0020) // bit5 overVoltage set
	binary.BigEndian.PutUint16(body[4:6], 0)

	msg, err := DecodeHex(headerFor(70), NewByteBuf(body))
	if err != nil {
		t.Fatalf("decode bin70: %v", err)
	}
	fault, ok := msg.(*Bin70NotifyFaultMessage)
	if !ok {
		t.Fatalf("expected *Bin70NotifyFaultMessage, got %T", msg)
	}
	if fault.Cmd != 70 || fault.MsgType != "alarm" {
		t.Fatalf("unexpected meta: %+v", fault)
	}
	if fault.OverVoltageFault != 0 {
		t.Fatalf("bin70 overVoltageFault must stay 0 for java compatibility, got %d", fault.OverVoltageFault)
	}
}

func TestDecodeBin72TypePlus1000(t *testing.T) {
	body := make([]byte, 8)
	binary.BigEndian.PutUint16(body[0:2], 2)
	binary.BigEndian.PutUint16(body[2:4], 0x003F)
	binary.BigEndian.PutUint32(body[4:8], 0)

	msg, err := DecodeHex(headerFor(72), NewByteBuf(body))
	if err != nil {
		t.Fatalf("decode bin72: %v", err)
	}
	notify, ok := msg.(*Bin72NotifyMessage)
	if !ok {
		t.Fatalf("expected *Bin72NotifyMessage, got %T", msg)
	}
	if notify.Cmd != 5 || notify.MsgType != "alarm" || notify.Type != 1002 {
		t.Fatalf("unexpected bin72 output: %+v", notify)
	}
	if notify.LockedRotor == nil || *notify.LockedRotor != 1 {
		t.Fatalf("expected lockedRotor=1, got %+v", notify.LockedRotor)
	}
}

// bin68FixedHeader builds the 36-byte fixed portion of a Bin68 GPS message.
func bin68FixedHeader() []byte {
	body := make([]byte, 36)
	binary.BigEndian.PutUint32(body[0:4], 1)
	body[4] = 31
	binary.BigEndian.PutUint32(body[5:9], 48000)
	binary.BigEndian.PutUint32(body[9:13], 1700000000)
	binary.BigEndian.PutUint32(body[13:17], 118378037)
	binary.BigEndian.PutUint32(body[17:21], 33780653)
	binary.BigEndian.PutUint16(body[21:23], 0)
	binary.BigEndian.PutUint16(body[23:25], 228)
	binary.BigEndian.PutUint16(body[25:27], 6)
	body[27] = 24
	binary.BigEndian.PutUint32(body[28:32], 1000)
	binary.BigEndian.PutUint32(body[32:36], 2)
	return body
}

// TestDecodeBin68TLVAscendingAndDedup verifies Java TreeMap alignment: duplicate tags keep the LAST
// occurrence, and tags are processed in ASCENDING order (so a higher tag wins when two tags touch the
// same field, regardless of wire order).
func TestDecodeBin68TLVAscendingAndDedup(t *testing.T) {
	body := bin68FixedHeader()
	tlv := []byte{
		0x01, 0x01, 0x0A, // bmsSoc=10
		0x01, 0x01, 0x14, // bmsSoc=20 -> dedup keeps last (20)
		// 0x19 (helmetLock, 5 bytes): sets HelmetType=1
		0x19, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00,
		// 0x0B (helmetLockState, 1 byte, value 0): would set HelmetType=0
		0x0B, 0x01, 0x00,
	}
	full := append(body, tlv...)

	msg, err := DecodeHex(headerFor(68), NewByteBuf(full))
	if err != nil {
		t.Fatalf("decode bin68: %v", err)
	}
	gps := msg.(*Bin68GpsMessage)

	if gps.BmsSoc == nil || *gps.BmsSoc != 20 {
		t.Fatalf("dedup failed: expected bmsSoc=20 (last occurrence), got %v", gps.BmsSoc)
	}
	// Wire order is [0x19, 0x0B]; if processed in wire order, 0x0B (last) would set HelmetType=0.
	// Ascending order processes 0x0B then 0x19, so 0x19 wins -> HelmetType=1.
	if gps.HelmetType == nil || *gps.HelmetType != 1 {
		t.Fatalf("ascending order failed: expected helmetType=1 (tag 0x19 wins), got %v", gps.HelmetType)
	}
}

// TestDecodeBin68TruncatedTLVDropsMessage verifies Java alignment: a TLV whose declared length is
// shorter than what the handler reads (here 0x04 bmsSN always reads 20 bytes) makes the whole message
// decode fail, mirroring Netty's IndexOutOfBoundsException dropping the message in the original Java.
func TestDecodeBin68TruncatedTLVDropsMessage(t *testing.T) {
	body := bin68FixedHeader()
	// tag 0x04 (bmsSN) declares length 2 but the handler reads 20 bytes -> out of bounds.
	tlv := []byte{0x04, 0x02, 0xAA, 0xBB}
	full := append(body, tlv...)

	if _, err := DecodeHex(headerFor(68), NewByteBuf(full)); err == nil {
		t.Fatal("expected decode error for truncated TLV value, got nil")
	}
}

func TestDecodeBin201ShuaKa(t *testing.T) {
	body := make([]byte, 42)
	copy(body[0:20], []byte("CARD1234567890123456"))
	copy(body[20:38], []byte("110101199001011234"))
	binary.BigEndian.PutUint32(body[38:42], 1700000000)

	msg, err := DecodeHex(headerFor(201), NewByteBuf(body))
	if err != nil {
		t.Fatalf("decode bin201: %v", err)
	}
	shuaka, ok := msg.(*Bin201ShuaKaMessage)
	if !ok {
		t.Fatalf("expected *Bin201ShuaKaMessage, got %T", msg)
	}
	if shuaka.MsgType != "data" || shuaka.Cmd != 201 || shuaka.BussinessType != "ebike" {
		t.Fatalf("unexpected meta: %+v", shuaka)
	}
	if shuaka.CardNo != "CARD1234567890123456" || shuaka.AuthNo != "110101199001011234" {
		t.Fatalf("unexpected card fields: cardNo=%q authNo=%q", shuaka.CardNo, shuaka.AuthNo)
	}
	if shuaka.Timestamp != 1700000000 {
		t.Fatalf("unexpected timestamp: %d", shuaka.Timestamp)
	}
}

func TestDecodeBin201TrimsPadding(t *testing.T) {
	body := make([]byte, 42)
	// Java String.trim() 移除首尾所有码点 <= 0x20 的字符。
	// cardNo：ASCII 内容后用 \0 填充（strings.TrimSpace 无法去除 \0，javaTrim 可以）。
	copy(body[0:20], []byte("CARD123")) // 剩余 13 字节保持 0x00
	// authNo：首尾用空格填充，共 18 字节。
	copy(body[20:38], []byte(" 110101199001011234")[:18])
	binary.BigEndian.PutUint32(body[38:42], 1)

	msg, err := DecodeHex(headerFor(201), NewByteBuf(body))
	if err != nil {
		t.Fatalf("decode bin201: %v", err)
	}
	shuaka := msg.(*Bin201ShuaKaMessage)
	if shuaka.CardNo != "CARD123" {
		t.Fatalf("expected \\0-padded cardNo trimmed, got %q", shuaka.CardNo)
	}
	if shuaka.AuthNo != "11010119900101123" {
		t.Fatalf("expected space-trimmed authNo, got %q", shuaka.AuthNo)
	}
}

func TestDecodeBin201TruncatedBuffer(t *testing.T) {
	_, err := DecodeHex(headerFor(201), NewByteBuf(make([]byte, 41)))
	if err == nil {
		t.Fatal("expected out-of-bounds error for truncated bin201 body")
	}
}

func TestDecodeBin35Login(t *testing.T) {
	// version=0x0003000B (3.0.11), equipmentType=100, imei/imsi ASCII[15]
	body := make([]byte, 35)
	binary.BigEndian.PutUint32(body[0:4], 0x0003000B)
	body[4] = 100
	copy(body[5:20], []byte("864908050917961"))
	copy(body[20:35], []byte("460081944102200"))

	msg, err := DecodeHex(headerFor(35), NewByteBuf(body))
	if err != nil {
		t.Fatalf("decode bin35: %v", err)
	}
	login, ok := msg.(*Bin35LoginMessage)
	if !ok {
		t.Fatalf("expected *Bin35LoginMessage, got %T", msg)
	}
	if login.Cmd != 35 || login.MsgType != "event" {
		t.Fatalf("meta: %+v", login)
	}
	if login.Version != 0x0003000B || login.DeviceType != 100 {
		t.Fatalf("version/deviceType: %+v", login)
	}
	if login.Imei != "864908050917961" || login.Imsi != "460081944102200" {
		t.Fatalf("imei/imsi: %+v", login)
	}
}

func TestDecodeBin35ShortBody(t *testing.T) {
	if _, err := DecodeHex(headerFor(35), NewByteBuf(make([]byte, 10))); err == nil {
		t.Fatal("expected short-body error")
	}
}

func TestDecodeUnknownCmd(t *testing.T) {
	_, err := DecodeHex(headerFor(99), NewByteBuf([]byte{0x01}))
	if err == nil {
		t.Fatal("expected error for unknown cmd")
	}
}

func TestDecodeBin2TruncatedBuffer(t *testing.T) {
	_, err := DecodeHex(headerFor(2), NewByteBuf([]byte{0x01}))
	if err == nil {
		t.Fatal("expected out-of-bounds error for truncated bin2 body")
	}
}

package com.luopingtech.ebike.rider.domain.ble

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

/**
 * ACK / notify parse. Multi-byte `resolveAck` must stay on `code == 0`
 * (UniApp compared CRC without `& 0xff`).
 *
 * Device-info / beacon / RFID layouts below are structurally valid CRC frames
 * or documented mocks — **real-vehicle captures still needed** to lock field
 * offsets against firmware.
 */
class BleFrameParserTest {
    @Test
    fun resolveAck_oneByteSuccess() {
        // length=1, retCode=0, crc must equal cmd+1+0
        val hex = "2c01002d"
        val ack = BleFrameParser.resolveAck(hex)
        assertEquals(0x2c, ack.cmd)
        assertEquals(0, ack.code)
        assertTrue(BleFrameParser.crcValid(hex))
    }

    @Test
    fun resolveAck_oneByteNack() {
        // retCode=1, crc = cmd+1+1 = 0x2e
        val hex = "2c01012e"
        val ack = BleFrameParser.resolveAck(hex)
        assertEquals(1, ack.code)
    }

    @Test
    fun resolveAck_lenGt1_locksCodeZero_productionSemantics() {
        // Constructed GET_DEVICE_INFO command frame: length=4, retCode is the
        // 4-byte token (0x0A0A0505). CRC byte never equals cmd+len+retCode
        // without a mask → code falls through to 0. Upper layer treats that
        // as success. Do not "fix".
        val hex = BleFrame.build(BleCommands.GET_DEVICE_INFO)
        assertEquals("41040a0a050563", hex)
        val ack = BleFrameParser.resolveAck(hex)
        assertEquals(0x41, ack.cmd)
        assertEquals(0, ack.code)
    }

    @Test
    fun hexToFloat32_42_littleEndian() {
        // 42.0f = 0x42280000; LE bytes 00 00 28 42
        assertEquals(42.0, BleConvert.hexToFloat32("00002842", bigEndian = false), 1e-6)
        assertEquals(42.0, BleConvert.hexToFloat32("42280000", bigEndian = true), 1e-6)
    }

    @Test
    fun parseNotify_gps_synthetic() {
        // cmd + len + ts(4) + lon_le(4) + lat_le(4) + speed(1) + course(2)
        val data = "00000001" + "00002842" + "00002842" + "0a" + "005a"
        val hex = notifyFrame(0x32, data)
        assertTrue(BleFrameParser.crcValid(hex))
        val decoded = BleFrameParser.parseNotify(hex)
        assertEquals(0x32, decoded.cmd)
        assertEquals("1", decoded.fields["timestamp"])
        assertEquals(42.0, decoded.fields["longitude"]?.toDoubleOrNull() ?: 0.0, 1e-4)
        assertEquals(42.0, decoded.fields["latitude"]?.toDoubleOrNull() ?: 0.0, 1e-4)
        assertEquals("10", decoded.fields["speed"])
        assertEquals("90", decoded.fields["course"])
    }

    @Test
    fun parseNotify_crcFail_emptyFields() {
        // Needs a real-vehicle 0x41 capture to lock offsets. This only
        // asserts the crc-fail path (firmware samples TBD).
        val hex = "41040000000000"
        val decoded = BleFrameParser.parseNotify(hex)
        assertEquals(0x41, decoded.cmd)
        assertTrue(decoded.fields.isEmpty())
    }

    @Test
    fun parseNotify_unknownCmd_keepsRaw() {
        val hex = BleFrame.build(BleCommands.UNLOCK, BleFrame.MUTE_PAYLOAD)
        val decoded = BleFrameParser.parseNotify(hex)
        assertEquals(0x2c, decoded.cmd)
        assertEquals(hex, decoded.fields["raw"])
    }

    @Test
    fun numberToMacAddress_reversesPairs() {
        assertEquals("ff:ee:dd:cc:bb:aa", BleConvert.numberToMacAddress("aabbccddeeff"))
    }

    private fun notifyFrame(cmd: Int, dataHex: String): String {
        val data = BleConvert.hexToBytes(dataHex)
        val datalen = data.size
        var sum = 0
        for (el in data) sum += el
        val crc = (cmd + sum + datalen) and 0xff
        return BleConvert.toPad2HexStr(cmd) +
            BleConvert.toPad2HexStr(datalen) +
            dataHex +
            BleConvert.toPad2HexStr(crc)
    }
}

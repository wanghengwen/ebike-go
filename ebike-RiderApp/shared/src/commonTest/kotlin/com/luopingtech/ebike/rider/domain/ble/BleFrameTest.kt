package com.luopingtech.ebike.rider.domain.ble

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

/**
 * Golden command frames. CRC is `(cmd + Σdata + datalen) & 0xff` — hand-calculated
 * against UniApp `buildCmd.ts`. Default token `0A 0A 05 05`.
 *
 * Planning note: LOCK+voice7 was listed as `…645a`; the TS formula yields `…64ba`.
 */
class BleFrameTest {
    @Test
    fun unlockMute_defaultToken() {
        val hex = BleFrame.build(BleCommands.UNLOCK, BleFrame.MUTE_PAYLOAD)
        assertEquals("2c060a0a0505000050", hex)
        assertTrue(BleFrameParser.crcValid(hex))
    }

    @Test
    fun lockMute_defaultToken() {
        val hex = BleFrame.build(BleCommands.LOCK, BleFrame.MUTE_PAYLOAD)
        assertEquals("2b060a0a050500004f", hex)
        assertTrue(BleFrameParser.crcValid(hex))
    }

    @Test
    fun lockTempVoice7Vol100_defaultToken() {
        val hex = BleFrame.build(
            BleCommands.LOCK,
            intArrayOf(BleVoice.TEMP_LOCK, BleVoice.DEFAULT_VOLUME),
        )
        assertEquals("2b060a0a05050764ba", hex)
        assertTrue(BleFrameParser.crcValid(hex))
    }

    @Test
    fun playVoiceStart_defaultToken() {
        val hex = BleFrame.build(
            BleCommands.PLAY_VOICE,
            intArrayOf(BleVoice.START, BleVoice.DEFAULT_VOLUME),
        )
        assertEquals("28060a0a05050364b3", hex)
        assertTrue(BleFrameParser.crcValid(hex))
    }

    @Test
    fun getDeviceInfo_defaultToken() {
        val hex = BleFrame.build(BleCommands.GET_DEVICE_INFO)
        assertEquals("41040a0a050563", hex)
        assertTrue(BleFrameParser.crcValid(hex))
    }

    @Test
    fun getGps_defaultToken() {
        val hex = BleFrame.build(BleCommands.GET_GPS)
        assertEquals("32040a0a050554", hex)
        assertTrue(BleFrameParser.crcValid(hex))
    }

    @Test
    fun getLastBeacon_defaultToken() {
        val hex = BleFrame.build(BleCommands.GET_LAST_BEACON)
        assertEquals("42040a0a050564", hex)
        assertTrue(BleFrameParser.crcValid(hex))
    }

    @Test
    fun getRfid_defaultToken() {
        val hex = BleFrame.build(BleCommands.GET_RFID)
        assertEquals("54040a0a050576", hex)
        assertTrue(BleFrameParser.crcValid(hex))
    }

    @Test
    fun tokenBytes_defaultDecimal() {
        val bytes = BleFrame.tokenBytes(BleFrame.DEFAULT_TOKEN_DECIMAL)
        assertTrue(bytes.contentEquals(BleFrame.DEFAULT_TOKEN))
        assertEquals("0a0a0505", BleConvert.toHexString(bytes))
    }

    @Test
    fun tokenBytes_planningTypoDecimal_isNotDefault() {
        // 168496389 == 0x0A0B0D05, not 0x0A0A0505
        assertEquals("0a0b0d05", BleConvert.toHexString(BleFrame.tokenBytes("168496389")))
    }

    @Test
    fun tokenBytes_longDecimalTakesLow4Bytes_noThrow() {
        // 0x1A0A0505FF = 111837251071 → low 32 bits 0x0A0505FF
        val bytes = BleFrame.tokenBytes("111837251071")
        assertEquals(4, bytes.size)
        assertEquals(listOf(0x0a, 0x05, 0x05, 0xff), bytes.map { it.toInt() and 0xff })
        val hex = BleFrame.build(BleCommands.UNLOCK, BleFrame.MUTE_PAYLOAD, bytes)
        assertEquals(18, hex.length)
        assertTrue(BleFrameParser.crcValid(hex))
    }

    @Test
    fun crcValid_everyConstructedFrame() {
        val frames = listOf(
            BleFrame.build(BleCommands.UNLOCK, BleFrame.MUTE_PAYLOAD),
            BleFrame.build(BleCommands.LOCK, BleFrame.MUTE_PAYLOAD),
            BleFrame.build(BleCommands.LOCK, intArrayOf(BleVoice.TEMP_LOCK, 100)),
            BleFrame.build(BleCommands.PLAY_VOICE, intArrayOf(BleVoice.START, 100)),
            BleFrame.build(BleCommands.GET_DEVICE_INFO),
            BleFrame.build(BleCommands.GET_GPS),
            BleFrame.build(BleCommands.GET_LAST_BEACON),
            BleFrame.build(BleCommands.GET_RFID),
        )
        frames.forEach { assertTrue(BleFrameParser.crcValid(it), it) }
    }

    @Test
    fun demoImeiForCarId_matchesAdvertRule() {
        val imei = BleFrame.demoImeiForCarId("D1001-001")
        assertEquals(15, imei.length)
        assertTrue(BleFrame.matchImeiAdvert(imei, imei.substring(3, 15)))
    }

    @Test
    fun matchImeiAdvert_mid12() {
        val imei = "867567046128534"
        assertTrue(BleFrame.matchImeiAdvert(imei, "567046128534"))
        assertFalse(BleFrame.matchImeiAdvert(imei, "56704612853"))
        assertFalse(BleFrame.matchImeiAdvert(imei, "675670461285"))
        assertFalse(BleFrame.matchImeiAdvert("", "567046128534"))
    }
}

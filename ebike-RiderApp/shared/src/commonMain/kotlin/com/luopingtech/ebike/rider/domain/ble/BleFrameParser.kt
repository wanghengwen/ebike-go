package com.luopingtech.ebike.rider.domain.ble

/**
 * Notify / ACK parsers. Ported from UniApp `buildCmd.resolveResult` + `resolve.ts`.
 *
 * **Do not “fix” these.** Production C-end treats `code === 0` as success, and
 * [resolveAck] compares CRC **without** `& 0xff`. Multi-byte replies therefore
 * almost always fall through to `code: 0`.
 */
data class BleAckFrame(
    val cmd: Int,
    val code: Int,
    val data: String,
)

data class BleNotifyDecoded(
    val cmd: Int,
    val raw: String = "",
    val fields: Map<String, String> = emptyMap(),
)

object BleFrameParser {
    private const val BLE_PROTOCOL_MIN = 6

    /**
     * UniApp `resolveResult`.
     *
     * `retCode = parseInt(hex.slice(4, 4 + length*2), 16)` — for `length > 1`
     * this is a multi-byte integer. The check is
     * `crc === cmd + length + retCode` (no mask), so it fails and returns
     * `{ code: 0 }`.
     */
    fun resolveAck(result: String): BleAckFrame {
        if (result.length < 4) return BleAckFrame(cmd = 0, code = 0, data = result)
        val cmd = result.substring(0, 2).toIntOrNull(16) ?: 0
        val length = result.substring(2, 4).toIntOrNull(16) ?: 0
        val dataEnd = 4 + length * 2
        val crcEnd = dataEnd + 2
        val retHex = if (dataEnd <= result.length) {
            result.substring(4, dataEnd.coerceAtMost(result.length))
        } else {
            ""
        }
        val retCode = retHex.toLongOrNull(16)
        val crc = if (crcEnd <= result.length) {
            result.substring(dataEnd, crcEnd).toIntOrNull(16)
        } else {
            null
        }
        // JS: failed parseInt → NaN, `NaN === n` is false → code 0.
        if (retCode == null || crc == null) {
            return BleAckFrame(cmd = cmd, code = 0, data = result)
        }
        return if (crc.toLong() == cmd.toLong() + length.toLong() + retCode) {
            BleAckFrame(cmd = cmd, code = retCode.toInt(), data = result)
        } else {
            BleAckFrame(cmd = cmd, code = 0, data = result)
        }
    }

    /**
     * UniApp `crcCheckValid`. Loop is `i < len - 1; i += 2` (includes the CRC
     * byte) then `(sum & 0xff) === ((crc + crc) & 0xff)`.
     */
    fun crcValid(bleRsp: String): Boolean {
        val len = bleRsp.length
        if (len < BLE_PROTOCOL_MIN || len % 2 != 0) return false
        val crc = bleRsp.substring(len - 2).toIntOrNull(16) ?: return false
        var sum = 0
        var i = 0
        while (i < len - 1) {
            val pair = bleRsp.substring(i, (i + 2).coerceAtMost(len))
            sum += pair.toIntOrNull(16) ?: return false
            i += 2
        }
        return (sum and 0xff) == ((crc + crc) and 0xff)
    }

    fun resolveStatusResult(statusRsp: String): Map<String, String> {
        if (!crcValid(statusRsp)) return emptyMap()
        val mode = statusRsp.substrInt(4, 2)
        val gsm = statusRsp.substrInt(6, 2)
        val sw = statusRsp.substrInt(8, 2)
        val voltageMv = statusRsp.substrInt(10, 8)
        return mapOf(
            "mode" to mode.toString(),
            "gsm" to gsm.toString(),
            "acc" to if ((sw and 0b10) > 1) "1" else "0",
            "defend" to (sw and 0b01).toString(),
            "backWheel" to if ((sw and 0b0100) > 2) "1" else "0",
            "backSeat" to if ((sw and 0b1000) > 3) "1" else "0",
            "voltageMv" to voltageMv.toString(),
        )
    }

    fun resolveGpsInfo(data: String): Map<String, String> {
        if (data.isEmpty() || !crcValid(data)) return emptyMap()
        return mapOf(
            "timestamp" to data.substrInt(4, 8).toString(),
            "longitude" to BleConvert.hexToFloat32(data.safeSubstr(12, 8), bigEndian = false).toString(),
            "latitude" to BleConvert.hexToFloat32(data.safeSubstr(20, 8), bigEndian = false).toString(),
            "speed" to data.substrInt(28, 2).toString(),
            "course" to data.substrInt(30, 4).toString(),
        )
    }

    fun resolveDeviceStatusInfo(data: String): Map<String, String> {
        if (data.isEmpty() || !crcValid(data)) return emptyMap()
        val sw = data.substrInt(6, 8)
        return mapOf(
            "gsm" to data.substrInt(14, 2).toString(),
            "voltage" to data.substrInt(16, 4).toString(),
            "GPSMajorVsn" to data.substrInt(20, 2).toString(),
            "GPSManorVsn" to data.substrInt(22, 2).toString(),
            "GPSMicroVsn" to data.substrInt(24, 2).toString(),
            "BLEMajorVsn" to data.substrInt(26, 2).toString(),
            "BLEManorVsn" to data.substrInt(28, 2).toString(),
            "BLEMicroVsn" to data.substrInt(30, 2).toString(),
            "timestamp" to data.substrInt(32, 8).toString(),
            "longitude" to (data.substrInt(40, 8) / 1_000_000.0).toString(),
            "latitude" to (data.substrInt(48, 8) / 1_000_000.0).toString(),
            "speed" to data.substrInt(56, 2).toString(),
            "course" to data.substrInt(58, 4).toString(),
            "hdop" to data.substrInt(62, 4).toString(),
            "satellite" to data.substrInt(66, 2).toString(),
            "totalMiles" to data.substrInt(68, 8).toString(),
            "isDefendOn" to if ((sw and 0b01) > 0) "1" else "0",
            "isAccOn" to if ((sw and 0b10) > 0) "1" else "0",
            "isWheelLocked" to if ((sw and 0b100) > 1) "1" else "0",
            "isSeatLocked" to if ((sw and 0b1000) > 1) "1" else "0",
            "isPowerExist" to if ((sw and 0b10000) > 1) "1" else "0",
            "isMoving" to if ((sw and 0b1000000) > 1) "1" else "0",
            "isHelmetExist" to if ((sw and 0b10000000000000000) > 1) "1" else "0",
        )
    }

    fun resolveLastBeaconInfo(data: String): Map<String, String> {
        if (data.isEmpty() || !crcValid(data)) return emptyMap()
        return mapOf(
            "event" to data.substrInt(4, 2).toString(),
            "tBeaconAddr" to BleConvert.numberToMacAddress(data.safeSubstr(6, 12)),
            "tBeaconId" to BleConvert.hexToString(data.safeSubstr(18, 24)),
            "tBeaconSOC" to data.substrInt(42, 2).toString(),
            "tBeaconVsn" to BleConvert.hexToString(data.safeSubstr(44, 4)),
            "lat" to (data.substrInt(48, 8) / 1_000_000.0).toString(),
            "lon" to (data.substrInt(56, 8) / 1_000_000.0).toString(),
            "timestamp" to data.substrInt(64, 8).toString(),
        )
    }

    fun resolveRFIDInfo(data: String): Map<String, String> {
        if (data.isEmpty() || !crcValid(data)) return emptyMap()
        return mapOf(
            "result" to data.substrInt(4, 2).toString(),
            "version" to data.substrInt(6, 48).toString(),
            "cardID" to BleConvert.hexToString(data.safeSubstr(54, 32)),
        )
    }

    fun parseNotify(hexOrEmpty: String): BleNotifyDecoded {
        val hex = hexOrEmpty.lowercase()
        if (hex.isEmpty()) return BleNotifyDecoded(cmd = 0)
        val cmd = hex.substring(0, 2.coerceAtMost(hex.length)).toIntOrNull(16) ?: 0
        return when (cmd) {
            0x41 -> BleNotifyDecoded(cmd, hex, resolveDeviceStatusInfo(hex))
            0x32 -> BleNotifyDecoded(cmd, hex, resolveGpsInfo(hex))
            0x42 -> BleNotifyDecoded(cmd, hex, resolveLastBeaconInfo(hex))
            0x54 -> BleNotifyDecoded(cmd, hex, resolveRFIDInfo(hex))
            else -> BleNotifyDecoded(cmd, hex, mapOf("raw" to hex))
        }
    }

    fun parseNotify(bytes: ByteArray): BleNotifyDecoded =
        parseNotify(BleConvert.toHexString(bytes))

    private fun String.safeSubstr(start: Int, length: Int): String {
        if (start >= this.length) return ""
        val end = (start + length).coerceAtMost(this.length)
        return substring(start, end)
    }

    private fun String.substrInt(start: Int, length: Int): Long {
        val slice = safeSubstr(start, length)
        if (slice.isEmpty()) return 0
        return slice.toLongOrNull(16) ?: 0
    }
}

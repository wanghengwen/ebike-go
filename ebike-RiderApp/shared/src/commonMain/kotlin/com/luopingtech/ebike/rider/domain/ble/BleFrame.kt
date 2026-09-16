package com.luopingtech.ebike.rider.domain.ble

/**
 * Command-frame builder. Ported from UniApp `buildCmd.ts`.
 *
 * Wire: `cmd(1) | datalen(1) | token(4)[+payload] | crc(1)`
 * Construction CRC: `(cmd + Σdata + datalen) & 0xff`
 *
 * Token: UniApp `parseInt(String(token), 10).toString(16)` padded to 8 hex
 * digits. When the decimal value needs more than 8 hex digits, TS concatenates
 * the unpadded hex and emits extra bytes. Kotlin **takes the low 4 bytes**
 * instead of throwing or growing the token field.
 */
object BleFrame {
    val DEFAULT_TOKEN: ByteArray = byteArrayOf(0x0a, 0x0a, 0x05, 0x05)

    /** Legacy `openBike` mute payload — voice is played after ride report. */
    val MUTE_PAYLOAD: IntArray = intArrayOf(0, 0)

    /**
     * Decimal of `0x0A0A0505`. UniApp only stores the byte array.
     * `168496389` is `0x0A0B0D05` — do not use that planning typo as default.
     */
    const val DEFAULT_TOKEN_DECIMAL: String = "168428805"

    fun interface TokenProvider {
        suspend fun getToken(imei: String): String?
    }

    /**
     * Decimal token → 4 bytes, low 32 bits only.
     * Empty / non-decimal → [DEFAULT_TOKEN] (API miss falls back the same way).
     */
    fun tokenBytes(token: String): ByteArray {
        val trimmed = token.trim()
        if (trimmed.isEmpty()) return DEFAULT_TOKEN.copyOf()
        var n = 0L
        var seen = false
        for (ch in trimmed) {
            val digit = ch.digitToIntOrNull() ?: break
            seen = true
            n = n * 10L + digit.toLong()
        }
        return if (seen) low4Bytes(n) else DEFAULT_TOKEN.copyOf()
    }

    fun tokenBytes(token: Long): ByteArray = low4Bytes(token)

    fun tokenBytes(token: UInt): ByteArray = low4Bytes(token.toLong() and 0xFFFF_FFFFL)

    /**
     * UniApp `imei.indexOf(advertHex) === 3 && advertHex.length === 12`.
     */
    fun matchImeiAdvert(imei: String, advertHex: String): Boolean {
        return imei.isNotEmpty() &&
            advertHex.length == 12 &&
            imei.indexOf(advertHex) == 3
    }

    /** Demo / map pin → 15-digit IMEI whose mid-12 matches [matchImeiAdvert]. */
    fun demoImeiForCarId(carId: String): String {
        val digits = carId.filter { it.isDigit() }.padEnd(12, '0').take(12)
        return "867$digits"
    }

    fun build(
        cmd: Int,
        payload: IntArray = intArrayOf(),
        token: ByteArray = DEFAULT_TOKEN,
    ): String {
        val data = IntArray(token.size + payload.size)
        for (i in token.indices) data[i] = token[i].toInt() and 0xff
        for (i in payload.indices) data[token.size + i] = payload[i] and 0xff
        val datalen = data.size
        var datasum = 0
        for (el in data) datasum += el
        val crc = (cmd + datasum + datalen) and 0xff
        return BleConvert.toPad2HexStr(cmd) +
            BleConvert.toPad2HexStr(datalen) +
            BleConvert.toHexString(data) +
            BleConvert.toPad2HexStr(crc)
    }

    suspend fun build(
        imei: String,
        cmd: Int,
        payload: IntArray = intArrayOf(),
        getToken: TokenProvider? = null,
    ): String {
        var token = DEFAULT_TOKEN
        if (getToken != null) {
            try {
                val raw = getToken.getToken(imei)
                if (raw != null && raw.isNotEmpty()) {
                    token = tokenBytes(raw)
                }
            } catch (_: Throwable) {
                token = DEFAULT_TOKEN
            }
        }
        return build(cmd, payload, token)
    }

    fun hexFrameToBytes(hex: String): ByteArray = BleConvert.hexToByteArray(hex)

    /** Low 32 bits, big-endian. Matches TS pad-to-8-hex when the value fits. */
    private fun low4Bytes(n: Long): ByteArray {
        var x = n
        val out = ByteArray(4)
        for (i in 3 downTo 0) {
            out[i] = (x and 0xFFL).toByte()
            x = x ushr 8
        }
        return out
    }
}

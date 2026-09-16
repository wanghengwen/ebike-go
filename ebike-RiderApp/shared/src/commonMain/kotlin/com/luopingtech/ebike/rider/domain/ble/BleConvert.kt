package com.luopingtech.ebike.rider.domain.ble

import kotlin.math.pow

/**
 * BLE hex / float helpers, ported from UniApp `convert.ts`.
 *
 * Keep the bit-twiddling identical — GPS notify uses the little-endian
 * [hexToFloat32] path, and [swap32] matches JS 32-bit signed shifts.
 */
object BleConvert {
    fun toPad2HexStr(num: Int): String {
        val hex = (num and 0xff).toString(16)
        return ("0" + hex).takeLast(2)
    }

    fun toHexString(byteArray: IntArray): String =
        byteArray.joinToString("") { toPad2HexStr(it) }

    fun toHexString(bytes: ByteArray): String =
        bytes.joinToString("") { toPad2HexStr(it.toInt() and 0xff) }

    fun hexToBytes(hex: String): IntArray {
        val bytes = ArrayList<Int>(hex.length / 2)
        var c = 0
        while (c < hex.length) {
            val end = (c + 2).coerceAtMost(hex.length)
            bytes.add(hex.substring(c, end).toIntOrNull(16) ?: 0)
            c += 2
        }
        return bytes.toIntArray()
    }

    fun hexToByteArray(hex: String): ByteArray {
        val ints = hexToBytes(hex)
        return ByteArray(ints.size) { i -> ints[i].toByte() }
    }

    fun bytesToHex(bytes: IntArray): String {
        val hex = ArrayList<String>(bytes.size * 2)
        for (i in bytes.indices) {
            val current = if (bytes[i] < 0) bytes[i] + 256 else bytes[i]
            hex.add((current ushr 4).toString(16))
            hex.add((current and 0xf).toString(16))
        }
        return hex.joinToString("")
    }

    /**
     * JS `swap32` — 32-bit signed shifts (`>>` / `<<` / `|`).
     */
    fun swap32(value: Int): Int {
        return ((value and 0xff) shl 24) or
            ((value and 0xff00) shl 8) or
            ((value shr 8) and 0xff00) or
            ((value shr 24) and 0xff)
    }

    /**
     * IEEE-754 binary32 from 8 hex chars. [bigEndian] false → [swap32] first
     * (UniApp GPS notify).
     */
    fun hexToFloat32(str: String, bigEndian: Boolean = true): Double {
        val parsed = str.toLongOrNull(16) ?: return 0.0
        var intBits = parsed.toInt()
        if (!bigEndian) intBits = swap32(intBits)
        if (intBits == 0) return 0.0
        val sign = if (intBits ushr 31 != 0) -1.0 else 1.0
        var exp = ((intBits ushr 23) and 0xff) - 127
        val mantissa = ((intBits and 0x7fffff) + 0x800000).toString(2)
        var float32 = 0.0
        for (ch in mantissa) {
            if (ch == '1') float32 += 2.0.pow(exp.toDouble())
            exp--
        }
        return float32 * sign
    }

    fun hexToString(hex: String): String {
        val out = StringBuilder()
        var i = 0
        while (i < hex.length) {
            val end = (i + 2).coerceAtMost(hex.length)
            val code = hex.substring(i, end).toIntOrNull(16) ?: 0
            out.append(Char(code and 0xff))
            i += 2
        }
        return out.toString()
    }

    /** Reverse byte-pairs and join with `:`. */
    fun numberToMacAddress(number: String): String {
        val pairs = ArrayList<String>()
        var i = 0
        while (i < number.length) {
            val end = (i + 2).coerceAtMost(number.length)
            pairs.add(number.substring(i, end))
            i += 2
        }
        return pairs.asReversed().joinToString(":")
    }
}

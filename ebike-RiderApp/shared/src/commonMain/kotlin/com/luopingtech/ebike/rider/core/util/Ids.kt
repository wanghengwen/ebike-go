package com.luopingtech.ebike.rider.core.util

import com.luopingtech.ebike.rider.core.time.nowEpochMillis
import kotlin.math.round
import kotlin.random.Random

object Ids {
    /**
     * UniApp `ensureDeviceId`：`` `${Date.now()}${(Math.random()*1e11).toFixed(0)}` ``。
     * 只给新设备用；已经落过盘的值（含旧 uuid）原样保留。
     */
    fun uniAppDeviceId(
        nowMillis: Long = nowEpochMillis(),
        randomFraction: Double = Random.nextDouble(),
    ): String {
        val suffix = round(randomFraction.coerceIn(0.0, 1.0) * 1e11).toLong().coerceAtLeast(0L)
        return "$nowMillis$suffix"
    }

    fun looksLikeUniAppDeviceId(value: String): Boolean =
        value.isNotEmpty() && value.all { it.isDigit() } && value.length in 14..25

    fun uuidV4(): String {
        val bytes = ByteArray(16)
        Random.Default.nextBytes(bytes)
        bytes[6] = ((bytes[6].toInt() and 0x0F) or 0x40).toByte()
        bytes[8] = ((bytes[8].toInt() and 0x3F) or 0x80).toByte()
        val hex = bytes.joinToString("") { b ->
            val v = b.toInt() and 0xFF
            v.toString(16).padStart(2, '0')
        }
        return buildString(36) {
            append(hex, 0, 8); append('-')
            append(hex, 8, 12); append('-')
            append(hex, 12, 16); append('-')
            append(hex, 16, 20); append('-')
            append(hex, 20, 32)
        }
    }
}

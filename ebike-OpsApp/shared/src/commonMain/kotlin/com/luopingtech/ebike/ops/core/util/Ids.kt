package com.luopingtech.ebike.ops.core.util

import kotlin.random.Random

object Ids {
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

package com.luopingtech.ebike.ops.core.crypto

internal fun ByteArray.toLowerHex(): String {
    val hexChars = CharArray(size * 2)
    val digits = "0123456789abcdef"
    for (i in indices) {
        val v = this[i].toInt() and 0xFF
        hexChars[i * 2] = digits[v ushr 4]
        hexChars[i * 2 + 1] = digits[v and 0x0F]
    }
    return hexChars.concatToString()
}

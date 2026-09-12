package com.luopingtech.ebike.ops.data.auth

/**
 * Normalize ops login phone to the legacy `+{area}-{local}` form.
 */
object PhoneNormalizer {
    const val DEFAULT_AREA_CODE: String = "+86"

    /** Ensure a single leading `+`, drop spaces; empty → [DEFAULT_AREA_CODE]. */
    fun normalizeAreaCode(raw: String): String {
        var value = raw.trim().replace(" ", "")
        if (value.isEmpty()) return DEFAULT_AREA_CODE
        while (value.startsWith("++")) value = value.drop(1)
        if (!value.startsWith("+")) value = "+$value"
        return value
    }

    fun toLoginPhone(raw: String, areaCode: String = DEFAULT_AREA_CODE): String {
        val trimmed = raw.trim()
        if (trimmed.isEmpty()) return trimmed
        if (trimmed.startsWith("+")) return trimmed
        val digits = trimmed.filter { it.isDigit() }
        return "${normalizeAreaCode(areaCode)}-$digits"
    }

    /** Local mobile digits without area prefix (forget-password body). */
    fun toLocalMobile(raw: String): String {
        val trimmed = raw.trim()
        if (trimmed.isEmpty()) return trimmed
        val digits = trimmed.filter { it.isDigit() }
        return if (digits.length > 11 && digits.startsWith("86")) {
            digits.takeLast(11)
        } else {
            digits
        }
    }
}

object SmsScene {
    const val LOGIN = 1
    const val FORGET_PASSWORD = 4
}

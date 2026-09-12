package com.luopingtech.ebike.ops.domain.vehicle

/**
 * Legacy ReplaceBatteryActivity1 QR rules: strip `battery_` prefix; accept 11 or 14 digits.
 */
object BatterySnNormalizer {
    fun normalize(raw: String): String {
        var value = raw.trim()
        if (value.contains("battery_", ignoreCase = true)) {
            value = value.replace("battery_", "", ignoreCase = true).trim()
        }
        return value
    }

    fun isValid(raw: String): Boolean {
        val sn = normalize(raw)
        return sn.length == 11 || sn.length == 14
    }
}

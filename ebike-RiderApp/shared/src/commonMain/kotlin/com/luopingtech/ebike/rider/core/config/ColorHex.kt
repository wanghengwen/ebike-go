package com.luopingtech.ebike.rider.core.config

/**
 * Parse `#RGB` / `#RRGGBB` / `#AARRGGBB` (optional leading `#`) into ARGB int bits.
 * Returns null when the string is blank or invalid.
 */
object ColorHex {
    fun parseArgb(raw: String?): Long? {
        val s = raw?.trim().orEmpty()
        if (s.isEmpty()) return null
        val hex = s.removePrefix("#")
        val normalized = when (hex.length) {
            3 -> buildString {
                hex.forEach { c -> append(c); append(c) }
            }.let { "FF$it" }
            6 -> "FF$hex"
            8 -> hex
            else -> return null
        }
        return normalized.toLongOrNull(16)
    }

    fun parseArgbOr(raw: String?, fallback: Long): Long = parseArgb(raw) ?: fallback
}

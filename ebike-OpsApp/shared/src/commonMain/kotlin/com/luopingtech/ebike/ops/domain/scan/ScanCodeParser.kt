package com.luopingtech.ebike.ops.domain.scan

/**
 * Normalize QR / barcode / manual input into a lookup key.
 *
 * Supports plain carId, `IMEI:` prefix, URL query carId/imei, and URL path tails.
 * http(s) URLs always require a non-empty [allowedHosts] match (fail closed when
 * selectAppConfig did not yield qrHosts — legacy decodeCarId rejects non-matching hosts).
 * Plain carId / IMEI: prefix never go through the host check.
 */
sealed class ScanTarget {
    data class CarId(val value: String) : ScanTarget()
    data class Imei(val value: String) : ScanTarget()
}

object ScanCodeParser {
    private val carIdInQuery = Regex("""[?&]carId=([^&#]+)""", RegexOption.IGNORE_CASE)
    private val imeiInQuery = Regex("""[?&]imei=([^&#]+)""", RegexOption.IGNORE_CASE)
    private val imeiPrefix = Regex("""^IMEI[:：]\s*(.+)$""", RegexOption.IGNORE_CASE)

    fun parse(raw: String, allowedHosts: List<String> = emptyList()): ScanTarget? {
        val trimmed = raw.trim()
        if (trimmed.isEmpty()) return null

        val isUrl = trimmed.startsWith("http://", ignoreCase = true) ||
            trimmed.startsWith("https://", ignoreCase = true)
        if (isUrl) {
            val hosts = allowedHosts.map { it.trim() }.filter { it.isNotEmpty() }
            // Fail closed: empty allowlist must not accept arbitrary QR URLs.
            if (hosts.isEmpty() || !hostAllowed(trimmed, hosts)) {
                return null
            }
        }

        carIdInQuery.find(trimmed)?.groupValues?.getOrNull(1)?.let { value ->
            val decoded = decode(value)
            if (decoded.isNotBlank()) return ScanTarget.CarId(decoded)
        }
        imeiInQuery.find(trimmed)?.groupValues?.getOrNull(1)?.let { value ->
            val decoded = decode(value)
            if (decoded.isNotBlank()) return ScanTarget.Imei(decoded)
        }

        imeiPrefix.find(trimmed)?.groupValues?.getOrNull(1)?.let { value ->
            val cleaned = value.trim()
            if (cleaned.isNotBlank()) return ScanTarget.Imei(cleaned)
        }

        if (isUrl) {
            val pathTail = trimmed.substringAfterLast('/').substringBefore('?').substringBefore('#')
            if (pathTail.isNotBlank()) return ScanTarget.CarId(pathTail.trim())
        }

        // Legacy main scan treats bare tokens as carId (including 15-digit numbers).
        return ScanTarget.CarId(trimmed)
    }

    private fun hostAllowed(url: String, allowedHosts: List<String>): Boolean {
        val lower = url.lowercase()
        return allowedHosts.any { host ->
            val fragment = host.trim()
            fragment.isNotEmpty() && lower.contains(fragment.lowercase())
        }
    }

    private fun decode(value: String): String =
        value.replace("%2B", "+", ignoreCase = true)
            .replace("%3A", ":", ignoreCase = true)
            .replace("%2F", "/", ignoreCase = true)
            .replace("%20", " ")
}

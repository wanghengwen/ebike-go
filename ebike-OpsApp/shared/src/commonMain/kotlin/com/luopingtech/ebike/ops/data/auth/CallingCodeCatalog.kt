package com.luopingtech.ebike.ops.data.auth

import kotlinx.serialization.Serializable
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.Json

@Serializable
data class CallingCode(
    val regionCode: String,
    val regionName: String,
    val callingCode: String,
    val regionFlag: String = "",
) {
    val dialCode: String get() = PhoneNormalizer.normalizeAreaCode(callingCode)

    val displayLabel: String
        get() = buildString {
            if (regionFlag.isNotBlank()) {
                append(regionFlag)
                append(' ')
            }
            append(dialCode)
            append(' ')
            append(regionName)
        }
}

object CallingCodeCatalog {
    val DEFAULT: CallingCode = CallingCode(
        regionCode = "CN",
        regionName = "China mainland",
        callingCode = "+86",
        regionFlag = "🇨🇳",
    )

    private val json = Json { ignoreUnknownKeys = true; isLenient = true }

    private val all: List<CallingCode> by lazy {
        runCatching {
            json.decodeFromString(
                ListSerializer(CallingCode.serializer()),
                CallingCodesJson.EN.trim(),
            )
        }.getOrElse { emptyList() }
            .ifEmpty { listOf(DEFAULT) }
            .let { list ->
                // Legacy iOS: pin CN then PL at top.
                val cn = list.firstOrNull { it.regionCode == "CN" } ?: DEFAULT
                val pl = list.firstOrNull { it.regionCode == "PL" }
                val rest = list.filterNot { it.regionCode == "CN" || it.regionCode == "PL" }
                listOfNotNull(cn, pl) + rest
            }
    }

    fun all(): List<CallingCode> = all

    fun find(regionCode: String?, callingCode: String?): CallingCode {
        val byRegion = regionCode?.trim()?.takeIf { it.isNotEmpty() }?.let { code ->
            all.firstOrNull { it.regionCode.equals(code, ignoreCase = true) }
        }
        if (byRegion != null) return byRegion
        val dial = callingCode?.let { PhoneNormalizer.normalizeAreaCode(it) }
        if (!dial.isNullOrBlank()) {
            all.firstOrNull { it.dialCode == dial }?.let { return it }
        }
        return DEFAULT
    }

    fun filter(query: String): List<CallingCode> {
        val q = query.trim()
        if (q.isEmpty()) return all()
        return all().filter {
            it.regionName.contains(q, ignoreCase = true) ||
                it.regionCode.contains(q, ignoreCase = true) ||
                it.callingCode.contains(q, ignoreCase = true) ||
                it.dialCode.contains(q, ignoreCase = true)
        }
    }
}

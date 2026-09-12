package com.luopingtech.ebike.ops.core.i18n

/**
 * App languages. Tags match legacy [AppLanguageHelper] / Accept-Language.
 */
enum class OpsLanguage(
    val tag: String,
    val acceptLanguage: String,
) {
    ZH_CN("zh-CN", "zh-CN"),
    EN("en", "en-US"),
    ;

    companion object {
        fun fromTag(raw: String?): OpsLanguage {
            val t = raw?.trim().orEmpty()
            if (t.isEmpty()) return ZH_CN
            val lower = t.lowercase()
            return when {
                lower.startsWith("en") -> EN
                lower.startsWith("zh") -> ZH_CN
                else -> entries.firstOrNull { it.tag.equals(t, ignoreCase = true) } ?: ZH_CN
            }
        }

        fun fromSystemLanguage(language: String?): OpsLanguage =
            when (language?.lowercase()?.take(2)) {
                "en" -> EN
                else -> ZH_CN
            }
    }
}

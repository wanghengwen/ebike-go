package com.luopingtech.ebike.rider.domain.riding

/**
 * 骑行相关的数值格式化。
 *
 * 放在 shared 而不是界面层，有两个原因：Kotlin/Native 没有 `String.format`，各屏各自手拼
 * 就会出现「1.5 元」和「1.50 元」同屏；而且围栏横幅的金额（[RidingFenceTips]）与骑行页、
 * 结费屏必须是同一套规则。
 */
object RideFormat {

    /** 分 → 两位小数元。 */
    fun yuan(fen: Int): String {
        val abs = if (fen < 0) -fen else fen
        val sign = if (fen < 0) "-" else ""
        val cents = abs % 100
        return "$sign${abs / 100}.${if (cents < 10) "0$cents" else "$cents"}"
    }

    /**
     * 分 → 元，对齐 UniApp `formatMoney`：整元不带小数，有角分才带。
     * 截图「待支付 1 元」。
     */
    fun yuanPlain(fen: Int): String {
        val abs = if (fen < 0) -fen else fen
        val sign = if (fen < 0) "-" else ""
        val yuan = abs / 100
        val cents = abs % 100
        return if (cents == 0) "$sign$yuan" else "$sign$yuan.${if (cents < 10) "0$cents" else "$cents"}"
    }

    /** 秒 → `mm:ss` / `h:mm:ss`。骑行时长会超过一小时，别只留两段。 */
    fun duration(totalSeconds: Long): String {
        val safe = if (totalSeconds < 0L) 0L else totalSeconds
        val hours = safe / 3600L
        val minutes = (safe % 3600L) / 60L
        val seconds = safe % 60L
        return if (hours > 0L) {
            "$hours:${pad(minutes)}:${pad(seconds)}"
        } else {
            "${pad(minutes)}:${pad(seconds)}"
        }
    }

    /** 秒 →「X小时Y分钟Z秒」，结费页文案。 */
    fun durationChinese(totalSeconds: Long): String {
        val safe = if (totalSeconds < 0L) 0L else totalSeconds
        val hours = safe / 3600L
        val minutes = (safe % 3600L) / 60L
        val seconds = safe % 60L
        return buildString {
            if (hours > 0L) append("${hours}小时")
            if (minutes > 0L || hours > 0L) append("${minutes}分钟")
            append("${seconds}秒")
        }
    }

    /** 米 → 1 km 以下按米、以上按一位小数千米。 */
    fun distance(meters: Int): String {
        if (meters < 1_000) return "${if (meters < 0) 0 else meters} m"
        val tenths = (meters + 50) / 100
        return "${tenths / 10}.${tenths % 10} km"
    }

    /** 米 →「x.xxx公里」，结费页。 */
    fun distanceKmChinese(meters: Int): String {
        val safe = if (meters < 0) 0 else meters
        val whole = safe / 1000
        val frac = safe % 1000
        return "$whole.${frac.toString().padStart(3, '0')}公里"
    }

    /** 剩余毫秒 → `HH:MM:SS`。 */
    fun countdownHms(remainMs: Long): String {
        val total = (remainMs / 1000L).coerceAtLeast(0L)
        val h = total / 3600L
        val m = (total % 3600L) / 60L
        val s = total % 60L
        return "${pad(h)}:${pad(m)}:${pad(s)}"
    }

    private fun pad(value: Long): String = if (value < 10L) "0$value" else "$value"
}

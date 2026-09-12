package com.luopingtech.ebike.ops.domain.task

import com.luopingtech.ebike.ops.core.time.nowEpochMillis

/**
 * 对齐 Flutter `TaskUtil.getTheWaitingTimeWith`：
 * N天N时 / N时N分 / N分N秒 / N秒；起点为空返回 `--`。
 */
object TaskWaitingTime {
    fun format(startText: String?, endText: String? = null, nowMs: Long = nowEpochMillis()): String {
        val startMs = parseDateTime(startText) ?: return "--"
        val endMs = parseDateTime(endText) ?: nowMs
        val seconds = ((endMs - startMs) / 1000L).coerceAtLeast(0L)
        val d = seconds / 86_400
        val h = (seconds % 86_400) / 3_600
        val m = (seconds % 3_600) / 60
        val s = seconds % 60
        return when {
            d != 0L -> "${d}天${h}时"
            h != 0L -> "${h}时${m}分"
            m != 0L -> "${m}分${s}秒"
            s != 0L -> "${s}秒"
            else -> "0秒"
        }
    }

    fun formatSeconds(totalSeconds: Long): String {
        val seconds = totalSeconds.coerceAtLeast(0L)
        val d = seconds / 86_400
        val h = (seconds % 86_400) / 3_600
        val m = (seconds % 3_600) / 60
        val s = seconds % 60
        return when {
            d != 0L -> "${d}天${h}时"
            h != 0L -> "${h}时${m}分"
            m != 0L -> "${m}分${s}秒"
            else -> "${s}秒"
        }
    }

    /** 解析 `yyyy-MM-dd HH:mm:ss` / ISO；失败返回 null。 */
    fun parseDateTime(text: String?): Long? {
        val raw = text?.trim().orEmpty()
        if (raw.isBlank()) return null
        val normalized = raw.replace('T', ' ').removeSuffix("Z")
        val parts = normalized.split(' ')
        if (parts.size < 2) return null
        val ymd = parts[0].split('-').mapNotNull { it.toIntOrNull() }
        val hms = parts[1].substringBefore('.').split(':').mapNotNull { it.toIntOrNull() }
        if (ymd.size < 3 || hms.size < 2) return null
        val year = ymd[0]
        val month = ymd[1]
        val day = ymd[2]
        val hour = hms[0]
        val minute = hms[1]
        val second = hms.getOrElse(2) { 0 }
        return dateTimeToEpochMs(year, month, day, hour, minute, second)
    }

    private fun dateTimeToEpochMs(
        year: Int,
        month: Int,
        day: Int,
        hour: Int,
        minute: Int,
        second: Int,
    ): Long {
        val zoneOffsetMs = 8 * 3_600_000L
        var days = 0L
        for (y in 1970 until year) {
            days += if (isLeap(y)) 366 else 365
        }
        val mdays = intArrayOf(31, if (isLeap(year)) 29 else 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31)
        for (m in 1 until month) {
            days += mdays[m - 1]
        }
        days += (day - 1)
        val tod = hour * 3_600_000L + minute * 60_000L + second * 1_000L
        return days * 86_400_000L + tod - zoneOffsetMs
    }

    private fun isLeap(year: Int): Boolean =
        (year % 4 == 0 && year % 100 != 0) || year % 400 == 0
}

package com.luopingtech.ebike.ops.domain.analysis

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.time.nowEpochMillis

/**
 * 线下运维时间筛（对齐 Flutter TimeUtil + MyTask）。
 * 区间一律按东八区日历日切分，格式化为 `yyyy-MM-dd HH:mm:ss`。
 */
enum class OfflineOpsPeriod {
    Today,
    Yesterday,
    ThisWeek,
    LastWeek,
    ThisMonth,
    LastMonth,
}

enum class OfflineOpsTrendGrain {
    Daily,
    Weekly,
    Monthly,
}

data class TimeRange(
    val startMs: Long,
    val endMs: Long,
) {
    val startText: String get() = OfflineOpsTimeRanges.formatDateTime(startMs)
    val endText: String get() = OfflineOpsTimeRanges.formatDateTime(endMs)
}

object OfflineOpsTimeRanges {
    private const val DAY_MS = 86_400_000L
    private const val ZONE_OFFSET_MS = 8 * 3_600_000L // Asia/Shanghai

    fun lastDays(days: Int, nowMs: Long = nowEpoch()): TimeRange {
        val n = days.coerceAtLeast(1)
        val todayStart = dayStart(nowMs)
        val start = todayStart - (n - 1) * DAY_MS
        return TimeRange(start, todayStart + DAY_MS - 1_000L)
    }

    /**
     * 原版异议工单默认：`(now - 7天)` 的日历日 00:00:00 ~ 今天 23:59:59
     *（跨度为 8 个日历日，与 [lastDays] 的「含今天共 N 日」不同）。
     */
    fun objectionDefaultRange(nowMs: Long = nowEpoch()): TimeRange {
        val todayStart = dayStart(nowMs)
        return TimeRange(todayStart - 7 * DAY_MS, todayStart + DAY_MS - 1_000L)
    }

    /** `yyyy-MM-dd`（东八区日历日）。 */
    fun formatDate(epochMs: Long): String {
        val p = ymd(dayStart(epochMs))
        return pad4(p[0]) + "-" + pad2(p[1]) + "-" + pad2(p[2])
    }

    /** 解析 `yyyy-MM-dd` 或 `yyyy-MM-dd HH:mm:ss`，返回该日 00:00:00（东八区）毫秒；失败返回 null。 */
    fun parseDateStart(text: String): Long? {
        val datePart = text.trim().take(10)
        if (datePart.length != 10) return null
        val parts = datePart.split('-')
        if (parts.size != 3) return null
        val y = parts[0].toIntOrNull() ?: return null
        val m = parts[1].toIntOrNull() ?: return null
        val d = parts[2].toIntOrNull() ?: return null
        if (m !in 1..12 || d !in 1..daysInMonth(y, m)) return null
        return dateStart(y, m, d)
    }

    fun dayEndOf(dayStartMs: Long): Long = dayStart(dayStartMs) + DAY_MS - 1_000L

    fun yesterdayDateText(nowMs: Long = nowEpoch()): String = formatDate(dayStart(nowMs) - DAY_MS)

    fun todayDateText(nowMs: Long = nowEpoch()): String = formatDate(dayStart(nowMs))

    /** 选中日 00:00:00 的秒级时间戳。 */
    fun dateStartEpochSec(dateText: String): Long? = parseDateStart(dateText)?.div(1_000L)

    /** 选中日 23:59:59 的秒级时间戳（对齐原版 `+86400-1`）。 */
    fun dateEndEpochSec(dateText: String): Long? =
        parseDateStart(dateText)?.let { it / 1_000L + 86_400L - 1L }

    fun ymdParts(epochMs: Long): IntArray = ymd(dayStart(epochMs))

    fun dateStartOf(year: Int, month: Int, day: Int): Long = dateStart(year, month, day)

    fun daysInMonth(year: Int, month: Int): Int {
        val mdays = intArrayOf(31, if (isLeap(year)) 29 else 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31)
        return mdays[(month - 1).coerceIn(0, 11)]
    }

    /** 周一=0 … 周日=6（按东八区日历日）。 */
    fun weekdayMondayIndex(dayStartMs: Long): Int {
        val localDay = (dayStart(dayStartMs) + ZONE_OFFSET_MS) / DAY_MS
        return ((localDay + 3) % 7).toInt()
    }

    fun period(period: OfflineOpsPeriod, nowMs: Long = nowEpoch()): TimeRange {
        val todayStart = dayStart(nowMs)
        return when (period) {
            OfflineOpsPeriod.Today -> TimeRange(todayStart, todayStart + DAY_MS - 1_000L)
            OfflineOpsPeriod.Yesterday -> {
                val y = todayStart - DAY_MS
                TimeRange(y, todayStart - 1_000L)
            }
            OfflineOpsPeriod.ThisWeek -> {
                val monday = weekMondayStart(todayStart)
                TimeRange(monday, monday + 7 * DAY_MS)
            }
            OfflineOpsPeriod.LastWeek -> {
                val monday = weekMondayStart(todayStart)
                TimeRange(monday - 7 * DAY_MS, monday)
            }
            OfflineOpsPeriod.ThisMonth -> {
                val monthStart = monthStart(todayStart)
                TimeRange(monthStart, todayStart + DAY_MS - 1_000L)
            }
            OfflineOpsPeriod.LastMonth -> {
                val thisMonth = monthStart(todayStart)
                val lastMonth = previousMonthStart(thisMonth)
                TimeRange(lastMonth, thisMonth)
            }
        }
    }

    fun trend(grain: OfflineOpsTrendGrain, nowMs: Long = nowEpoch()): TimeRange {
        val todayStart = dayStart(nowMs)
        return when (grain) {
            OfflineOpsTrendGrain.Daily -> {
                val start = todayStart - 14 * DAY_MS
                TimeRange(start, todayStart + DAY_MS - 1_000L)
            }
            OfflineOpsTrendGrain.Weekly -> {
                val start = todayStart - 55 * DAY_MS
                TimeRange(start, todayStart + DAY_MS - 1_000L)
            }
            OfflineOpsTrendGrain.Monthly -> {
                val thisMonth = monthStart(todayStart)
                val start = previousMonthStart(previousMonthStart(thisMonth))
                val nextMonth = nextMonthStart(thisMonth)
                TimeRange(start, nextMonth - 1_000L)
            }
        }
    }

    fun dayLabelsForDaily(nowMs: Long = nowEpoch()): List<String> {
        val todayStart = dayStart(nowMs)
        return (14 downTo 0).map { offset ->
            formatMonthDay(todayStart - offset * DAY_MS)
        }
    }

    fun formatUpdateTime(ms: Long = nowEpoch()): String = formatDateTime(ms)

    fun formatDateTime(epochMs: Long): String {
        val p = ymd(dayStart(epochMs))
        val local = epochMs + ZONE_OFFSET_MS
        val tod = ((local % DAY_MS) + DAY_MS) % DAY_MS
        val h = (tod / 3_600_000L).toInt()
        val mi = ((tod % 3_600_000L) / 60_000L).toInt()
        val s = ((tod % 60_000L) / 1_000L).toInt()
        return pad4(p[0]) + "-" + pad2(p[1]) + "-" + pad2(p[2]) + " " +
            pad2(h) + ":" + pad2(mi) + ":" + pad2(s)
    }

    fun formatDurationSeconds(seconds: Double): String {
        val total = seconds.toLong().coerceAtLeast(0)
        val h = total / 3600
        val m = (total % 3600) / 60
        val s = total % 60
        return when {
            h > 0 -> Strings.t(Str.DurationHms, h, m, s)
            m > 0 -> Strings.t(Str.DurationMs, m, s)
            else -> Strings.t(Str.DurationZeroMs, s)
        }
    }

    private fun nowEpoch(): Long = nowEpochMillis()

    private fun dayStart(epochMs: Long): Long {
        val local = epochMs + ZONE_OFFSET_MS
        return local - (local % DAY_MS) - ZONE_OFFSET_MS
    }

    private fun weekMondayStart(dayStartMs: Long): Long {
        val localDay = (dayStartMs + ZONE_OFFSET_MS) / DAY_MS
        val dowFromMonday = ((localDay + 3) % 7).toInt()
        return dayStartMs - dowFromMonday * DAY_MS
    }

    private fun monthStart(dayStartMs: Long): Long {
        val parts = ymd(dayStartMs)
        return dateStart(parts[0], parts[1], 1)
    }

    private fun previousMonthStart(monthStartMs: Long): Long {
        val parts = ymd(monthStartMs)
        var y = parts[0]
        var m = parts[1] - 1
        if (m < 1) {
            m = 12
            y -= 1
        }
        return dateStart(y, m, 1)
    }

    private fun nextMonthStart(monthStartMs: Long): Long {
        val parts = ymd(monthStartMs)
        var y = parts[0]
        var m = parts[1] + 1
        if (m > 12) {
            m = 1
            y += 1
        }
        return dateStart(y, m, 1)
    }

    private fun dateStart(year: Int, month: Int, day: Int): Long {
        var days = 0L
        for (y in 1970 until year) {
            days += if (isLeap(y)) 366 else 365
        }
        val mdays = intArrayOf(31, if (isLeap(year)) 29 else 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31)
        for (m in 1 until month) {
            days += mdays[m - 1]
        }
        days += (day - 1)
        return days * DAY_MS - ZONE_OFFSET_MS
    }

    private fun ymd(dayStartMs: Long): IntArray {
        var days = ((dayStartMs + ZONE_OFFSET_MS) / DAY_MS).toInt()
        var y = 1970
        while (true) {
            val diy = if (isLeap(y)) 366 else 365
            if (days < diy) break
            days -= diy
            y++
        }
        val mdays = intArrayOf(31, if (isLeap(y)) 29 else 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31)
        var m = 1
        while (m <= 12 && days >= mdays[m - 1]) {
            days -= mdays[m - 1]
            m++
        }
        return intArrayOf(y, m, days + 1)
    }

    private fun isLeap(year: Int): Boolean =
        (year % 4 == 0 && year % 100 != 0) || year % 400 == 0

    private fun formatMonthDay(dayStartMs: Long): String {
        val p = ymd(dayStartMs)
        return "${p[1]}.${p[2]}"
    }

    private fun pad2(v: Int): String = v.toString().padStart(2, '0')
    private fun pad4(v: Int): String = v.toString().padStart(4, '0')
}

package com.luopingtech.ebike.ops.domain.order

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings

/**
 * 订单查询展示格式化（金额分→元、时长、里程等）。
 */
object OrderFormat {
    private const val Dash = "--"

    fun yuan(fen: Long?): String {
        if (fen == null) return Dash
        val abs = if (fen < 0) -fen else fen
        val sign = if (fen < 0) "-" else ""
        val whole = abs / 100
        val frac = (abs % 100).toString().padStart(2, '0')
        return "$sign$whole.$frac"
    }

    fun yuanWithUnit(fen: Long?): String {
        if (fen == null) return Dash
        return Strings.t(Str.OrderAmountYuan, yuan(fen))
    }

    /**
     * 骑行时长。接口多是毫秒字符串；位数 > 10 当毫秒先 /1000，否则当秒。
     */
    fun duration(raw: String?): String {
        if (raw.isNullOrBlank()) return Dash
        val value = raw.toLongOrNull() ?: return Dash
        if (value < 0) return Dash
        val totalSeconds = if (raw.trimStart('-').length > 10) value / 1000 else value
        val days = totalSeconds / 86400
        val hours = (totalSeconds % 86400) / 3600
        val minutes = (totalSeconds % 3600) / 60
        val seconds = totalSeconds % 60
        val parts = mutableListOf<String>()
        if (days > 0) parts += Strings.t(Str.DurationDays, days)
        if (hours > 0) parts += Strings.t(Str.DurationHours, hours)
        if (minutes > 0) parts += Strings.t(Str.DurationMinutes, minutes)
        if (seconds > 0 || parts.isEmpty()) parts += Strings.t(Str.DurationSeconds, seconds)
        return parts.joinToString("")
    }

    fun distance(meters: Long?): String {
        if (meters == null) return Dash
        if (meters < 1000) return Strings.t(Str.DistanceMeters, meters)
        val whole = meters / 1000
        val frac = ((meters % 1000) / 10).toString().padStart(2, '0')
        return Strings.t(Str.DistanceKilometers, "$whole.$frac")
    }

    fun payStateLabel(izPaid: Int?): String = when (izPaid) {
        OrderPayStates.Riding -> Strings.t(Str.OrderPayRiding)
        OrderPayStates.Frozen -> Strings.t(Str.OrderPayFrozen)
        OrderPayStates.ToPay -> Strings.t(Str.OrderPayToPay)
        OrderPayStates.Paid -> Strings.t(Str.OrderPayPaid)
        else -> Dash
    }

    fun isSettled(izPaid: Int?): Boolean = izPaid == OrderPayStates.Paid

    fun ridingStateLabel(state: Int?): String = when (state) {
        1 -> Strings.t(Str.RidingStateRegistered)
        2 -> Strings.t(Str.RidingStateVerified)
        3 -> Strings.t(Str.RidingStateAvailable)
        4 -> Strings.t(Str.RidingStateBooking)
        5 -> Strings.t(Str.RidingStateTempPark)
        6 -> Strings.t(Str.RidingStateRiding)
        7 -> Strings.t(Str.RidingStateToPay)
        else -> Dash
    }

    fun orDash(value: String?): String {
        val text = value?.trim().orEmpty()
        return if (text.isEmpty() || text == "-") Dash else text
    }

    fun isoDateTime(value: String?): String {
        val text = value?.trim().orEmpty()
        if (text.isEmpty()) return Dash
        return text.replace('T', ' ')
    }
}

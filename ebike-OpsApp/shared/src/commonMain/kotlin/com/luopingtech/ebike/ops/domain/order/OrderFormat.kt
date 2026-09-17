package com.luopingtech.ebike.ops.domain.order

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.core.time.nowEpochMillis

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
     * 骑行时长。对齐遗留 TimeStampUtils.timestampFormat：
     * `ridingTime` 是毫秒；位数 > 10 时先 /1000 再按毫秒拆天/时/分/秒。
     * （20 分钟 = 1_200_000，若当秒会误显示成约 14 天。）
     */
    fun duration(raw: String?): String {
        if (raw.isNullOrBlank()) return Dash
        var millis = raw.toLongOrNull() ?: return Dash
        if (millis < 0) return Dash
        if (raw.trimStart('-').length > 10) millis /= 1000
        val days = millis / 86_400_000
        millis %= 86_400_000
        val hours = millis / 3_600_000
        millis %= 3_600_000
        val minutes = millis / 60_000
        millis %= 60_000
        val seconds = millis / 1000
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

    /**
     * 对齐遗留 OrderItemModel.getPayStr：
     * carState 2/3 → 骑行中/临停中；否则「金额+已支付/未支付」。
     */
    fun legacyPayBanner(order: OrderRecord): Pair<String, PayTone> {
        when (order.carState) {
            2 -> return Strings.t(Str.OrderPayRiding) to PayTone.Blue
            3 -> return Strings.t(Str.OrderTempParking) to PayTone.Blue
        }
        val amountRaw = yuan(order.payCost)
        val amount = if (amountRaw == Dash) {
            Dash
        } else {
            // 对齐遗留 getPayStr：`1.5元已支付`（无空格）
            Strings.t(Str.OrderAmountYuan, amountRaw).replace(" ", "")
        }
        val state = when (order.izPaid) {
            OrderPayStates.Paid -> Strings.t(Str.OrderPayPaid)
            null -> Dash
            else -> Strings.t(Str.OrderPayUnpaid)
        }
        val text = when {
            amount == Dash && state == Dash -> Dash
            amount == Dash -> state
            state == Dash -> amount
            else -> "$amount$state"
        }
        val tone = if (order.izPaid == OrderPayStates.Paid) PayTone.Green else PayTone.Orange
        return text to tone
    }

    /** 对齐 item_order_history：无有效 id 时显示「骑行中」。 */
    fun displayOrderId(order: OrderRecord): String {
        val id = order.id.trim()
        if (id.isEmpty() || id == "0" || id == "null") {
            return Strings.t(Str.OrderPayRiding)
        }
        return id
    }

    enum class PayTone { Blue, Green, Orange }

    fun isRidingLike(carState: Int?): Boolean = carState == 2 || carState == 3

    /** 用户 ridingState 5临停 / 6骑行中。 */
    fun isUserRiding(ridingState: Int?): Boolean = ridingState == 5 || ridingState == 6

    /** 是否展示进行中操作栏：用户骑行中或末单 carState 2/3。 */
    fun showInProgressActions(userRidingState: Int?, carState: Int?): Boolean =
        isUserRiding(userRidingState) || isRidingLike(carState)

    /** 临停：ridingState==5 或 carState==3（用于显示「启动」）。 */
    fun isTempParking(userRidingState: Int?, carState: Int?): Boolean =
        userRidingState == 5 || carState == 3

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

    /**
     * 对齐 UserDetailInfoViewModel.handleDepositSate：
     * 已缴诚信金 →「诚信金/xx元」；押金卡 →「会员卡/xx元/n天」；否则「0元」。
     */
    fun depositStatus(user: OrderUserDetail?, assets: OrderUserAssets?): String {
        if (user?.izDeposited == true) {
            val yuanText = yuan(assets?.depositedMountFen ?: 0L)
            return Strings.t(Str.OrderQueryDepositPaidFmt, yuanText)
        }
        if (user?.izDepositCard == true) {
            val yuanText = yuan(assets?.depositCardDiscountFen ?: 0L)
            val days = orDash(user.depositedCardDays)
            return Strings.t(Str.OrderQueryDepositCardFmt, yuanText, days)
        }
        return Strings.t(Str.OrderQueryDepositZero)
    }

    /** 对齐 UserAssetsModel.getDepositCardText（实际展示骑行卡名）。 */
    fun ridingCardText(assets: OrderUserAssets?): String {
        val name = assets?.ridingCardName?.trim().orEmpty()
        return if (name.isEmpty()) Strings.t(Str.OrderQueryRideCardNone) else name
    }

    /** 对齐 UserAssetsModel.getWalletBalanceText；无资产时回落到用户详情 balance。 */
    fun walletBalanceText(assets: OrderUserAssets?, userBalanceFen: Long?): String {
        val fen = assets?.walletBalanceFen ?: userBalanceFen
        return if (fen == null) Dash else yuanWithUnit(fen)
    }

    fun paidDeposit(user: OrderUserDetail?): Boolean =
        user?.izDeposited == true || user?.izDepositCard == true

    /** 对齐 UserAssetsModel.getIntegrityAmount */
    fun integrityAmount(user: OrderUserDetail?, assets: OrderUserAssets?): String {
        if (!paidDeposit(user)) return Strings.t(Str.OrderQueryDepositZero)
        if (assets?.depositedStats == 1) return yuanWithUnit(assets.depositedMountFen ?: 0L)
        if (!assets?.depositCardExpiredDate.isNullOrBlank() || (assets?.depositCardDiscountFen ?: 0) > 0) {
            return yuanWithUnit(assets?.depositCardDiscountFen ?: 0L)
        }
        return Dash
    }

    /** 对齐 UserAssetsModel.getContinuedTime */
    fun depositExpireText(user: OrderUserDetail?, assets: OrderUserAssets?): String {
        if (!paidDeposit(user)) return Dash
        if (assets?.depositedStats == 1 || user?.izDeposited == true) {
            return Strings.t(Str.OrderQueryDepositForever)
        }
        return orDash(assets?.depositCardExpiredDate)
    }

    fun rideRemainTimes(card: OrderRideCard): String =
        Strings.t(Str.OrderQueryTimesFmt, card.remainTimes.ifBlank { "0" })

    fun rideRemainingTime(card: OrderRideCard): String {
        val ms = card.expiredAtMs ?: return Dash
        val diff = ms - nowEpochMillis()
        if (diff <= 0) return "0"
        val days = diff / 86_400_000L
        return Strings.t(Str.DurationDays, days)
    }

    fun durationDaysOrDash(raw: String): String {
        val text = raw.trim()
        if (text.isEmpty() || text == Dash) return Dash
        return Strings.t(Str.DurationDays, text)
    }
}

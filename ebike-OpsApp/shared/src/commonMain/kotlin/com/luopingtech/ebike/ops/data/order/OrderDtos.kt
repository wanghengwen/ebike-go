package com.luopingtech.ebike.ops.data.order

import com.luopingtech.ebike.ops.data.trajectory.TrackPointDto
import com.luopingtech.ebike.ops.domain.model.LastOrder
import com.luopingtech.ebike.ops.domain.order.OrderRecord
import com.luopingtech.ebike.ops.domain.order.OrderDepositRecord
import com.luopingtech.ebike.ops.domain.order.OrderRideCard
import com.luopingtech.ebike.ops.domain.order.OrderRideCardRecord
import com.luopingtech.ebike.ops.domain.order.OrderUserAssets
import com.luopingtech.ebike.ops.domain.order.OrderUserDetail
import com.luopingtech.ebike.ops.domain.order.OrderUserPageItem
import com.luopingtech.ebike.ops.domain.order.OrderWalletInfo
import com.luopingtech.ebike.ops.domain.order.OrderWalletRecord
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.longOrNull

@Serializable
data class LastOrderDto(
    val carId: String? = null,
    val startLat: Double? = null,
    val startLng: Double? = null,
    val endLat: Double? = null,
    val endLng: Double? = null,
    val deviceTrajectory: List<TrackPointDto>? = null,
    val id: JsonElement? = null,
    val userPin: String? = null,
    val phone: String? = null,
    val userPhone: String? = null,
    val startTime: String? = null,
    val endTime: String? = null,
    val userName: String? = null,
    val originCost: Long? = null,
    val payCost: Long? = null,
    val mile: Long? = null,
    val ridingTime: JsonElement? = null,
    val izPaid: Int? = null,
    val payTime: String? = null,
    val hasPaid: Long? = null,
    val dispatchCost: Long? = null,
    val helmetPenalty: Long? = null,
    val carState: Int? = null,
) {
    fun toDomain(): LastOrder {
        val track = deviceTrajectory.orEmpty().mapNotNull { it.toDomain() }
        return LastOrder(
            carId = carId.orEmpty(),
            startLat = startLat ?: 0.0,
            startLng = startLng ?: 0.0,
            endLat = endLat,
            endLng = endLng,
            trajectory = track,
            id = id.toFlexibleString(),
            userPin = userPin.orEmpty(),
            userPhone = phone?.takeIf { it.isNotBlank() } ?: userPhone.orEmpty(),
            startTime = startTime.orEmpty(),
            endTime = endTime.orEmpty(),
        )
    }

    fun toOrderRecord(): OrderRecord = OrderRecord(
        id = id.toFlexibleString(),
        carId = carId.orEmpty(),
        phone = phone?.takeIf { it.isNotBlank() } ?: userPhone.orEmpty(),
        userPin = userPin.orEmpty(),
        userName = userName.orEmpty(),
        originCost = originCost,
        payCost = payCost,
        mile = mile,
        ridingTimeRaw = ridingTime.toFlexibleString(),
        izPaid = izPaid,
        startLat = startLat,
        startLng = startLng,
        endLat = endLat,
        endLng = endLng,
        startTime = startTime.orEmpty(),
        endTime = endTime.orEmpty(),
        payTime = payTime.orEmpty(),
        hasPaid = hasPaid,
        dispatchCost = dispatchCost,
        helmetPenalty = helmetPenalty,
        carState = carState,
        trajectory = deviceTrajectory.orEmpty().mapNotNull { it.toDomain() },
    )
}

@Serializable
data class OrderPageDto(
    val list: List<LastOrderDto>? = null,
    val count: String? = null,
    val pageNum: Int = 1,
    val pageSize: Int = 10,
)

@Serializable
data class OrderUserDetailDto(
    val pin: String? = null,
    val authName: String? = null,
    val phone: String? = null,
    val authNo: String? = null,
    val createdAt: String? = null,
    val izAuth: Boolean? = null,
    val ridingState: Int? = null,
    val balance: Long? = null,
    val serviceName: String? = null,
    val izDeposited: Boolean? = null,
    val izDepositCard: Boolean? = null,
    val depositedCardDays: JsonElement? = null,
    val izRidingType: Int? = null,
) {
    fun toDomain(): OrderUserDetail = OrderUserDetail(
        pin = pin.orEmpty(),
        authName = authName.orEmpty(),
        phone = phone.orEmpty(),
        authNo = authNo.orEmpty(),
        createdAt = createdAt.orEmpty(),
        izAuth = izAuth,
        ridingState = ridingState,
        balance = balance,
        serviceName = serviceName.orEmpty(),
        izDeposited = izDeposited,
        izDepositCard = izDepositCard,
        depositedCardDays = depositedCardDays.toFlexibleString(),
        izRidingType = izRidingType,
    )
}

@Serializable
data class OrderUserAssetsDto(
    val userWallet: OrderUserWalletDto? = null,
    val userRidingCard: OrderUserRidingCardDto? = null,
    val userDepositCard: OrderUserDepositCardDto? = null,
) {
    fun toDomain(): OrderUserAssets = OrderUserAssets(
        walletBalanceFen = userWallet?.balance.toFen(),
        depositedMountFen = userWallet?.depositedMount.toFen(),
        depositedStats = userWallet?.depositedStats,
        ridingCardName = userRidingCard?.name.orEmpty(),
        depositCardDiscountFen = userDepositCard?.content?.discountMoney.toFen(),
        depositCardExpiredDate = userDepositCard?.expiredDate.orEmpty(),
        walletPresentFen = userWallet?.present.toFen(),
        walletRechargeFen = userWallet?.recharge.toFen(),
    )
}

@Serializable
data class OrderUserWalletDto(
    val balance: JsonElement? = null,
    val depositedMount: JsonElement? = null,
    val depositedStats: Int? = null,
    val pin: String? = null,
    val present: JsonElement? = null,
    val recharge: JsonElement? = null,
)

@Serializable
data class OrderUserRidingCardDto(
    val name: String? = null,
    val cardId: String? = null,
)

@Serializable
data class OrderUserDepositCardDto(
    val id: String? = null,
    val pin: String? = null,
    val expiredDate: String? = null,
    val content: OrderUserDepositCardContentDto? = null,
)

@Serializable
data class OrderUserDepositCardContentDto(
    val type: Int? = null,
    @SerialName("service_id") val serviceId: String? = null,
    val name: String? = null,
    @SerialName("card_duration_day") val cardDurationDay: Int? = null,
    val money: JsonElement? = null,
    @SerialName("discount_money") val discountMoney: JsonElement? = null,
)

@Serializable
data class OrderUserPageItemDto(
    val pin: String? = null,
    val authName: String? = null,
    val phone: String? = null,
    val izAuth: Boolean? = null,
    val ridingState: Int? = null,
    val createdAt: String? = null,
    val balance: Long? = null,
) {
    fun toDomain(): OrderUserPageItem = OrderUserPageItem(
        pin = pin.orEmpty(),
        authName = authName.orEmpty(),
        phone = phone.orEmpty(),
        izAuth = izAuth,
        ridingState = ridingState,
        createdAt = createdAt.orEmpty(),
        balance = balance,
    )
}

@Serializable
data class OrderUserPageDto(
    val list: List<OrderUserPageItemDto>? = null,
    val count: Int = 0,
    val pageNum: Int = 1,
    val pageSize: Int = 15,
)

@Serializable
data class OrderWalletInfoDto(
    val balance: JsonElement? = null,
    val depositedMount: JsonElement? = null,
    val depositedStats: Int? = null,
    val pin: String? = null,
    val present: JsonElement? = null,
    val recharge: JsonElement? = null,
) {
    fun toDomain(): OrderWalletInfo = OrderWalletInfo(
        balanceFen = balance.toFen(),
        rechargeFen = recharge.toFen(),
        presentFen = present.toFen(),
        depositedMountFen = depositedMount.toFen(),
        depositedStats = depositedStats,
    )
}

@Serializable
data class OrderWalletRecordDto(
    val amount: JsonElement? = null,
    @SerialName("change_type") val changeType: String? = null,
    val channel: String? = null,
    @SerialName("iz_refund") val izRefund: Int? = null,
    @SerialName("merchant_trade_no") val merchantTradeNo: String? = null,
    @SerialName("paid_at") val paidAt: String? = null,
    @SerialName("pin_name") val pinName: String? = null,
    @SerialName("present_amount") val presentAmount: JsonElement? = null,
    @SerialName("recharge_amount") val rechargeAmount: JsonElement? = null,
    val type: String? = null,
) {
    fun toDomain(): OrderWalletRecord = OrderWalletRecord(
        amountFen = amount.toFen(),
        changeType = changeType.orEmpty(),
        channel = channel.orEmpty(),
        izRefund = izRefund ?: 0,
        merchantTradeNo = merchantTradeNo.orEmpty(),
        paidAt = paidAt.orEmpty(),
        pinName = pinName.orEmpty(),
        presentAmountFen = presentAmount.toFen(),
        rechargeAmountFen = rechargeAmount.toFen(),
        type = type.orEmpty(),
    )
}

@Serializable
data class OrderDepositRecordDto(
    @SerialName("paid_at") val paidAt: String? = null,
    @SerialName("deposit_type") val depositType: String? = null,
    val amount: JsonElement? = null,
    val duration: String? = null,
    val state: String? = null,
    val type: String? = null,
    val channel: String? = null,
    @SerialName("merchant_trade_no") val merchantTradeNo: String? = null,
    @SerialName("iz_refund") val izRefund: Int? = null,
) {
    fun toDomain(): OrderDepositRecord = OrderDepositRecord(
        paidAt = paidAt.orEmpty(),
        depositType = depositType.orEmpty(),
        amountFen = amount.toFen(),
        duration = duration.orEmpty(),
        state = state.orEmpty(),
        type = type.orEmpty(),
        channel = channel.orEmpty(),
        merchantTradeNo = merchantTradeNo.orEmpty(),
        izRefund = izRefund ?: 0,
    )
}

@Serializable
data class OrderRideCardListDto(
    val used: List<OrderRideCardDto>? = null,
    val expired: List<OrderRideCardDto>? = null,
)

@Serializable
data class OrderRideCardDto(
    val name: String? = null,
    val cardExpiredDate: String? = null,
    val remainTimes: JsonElement? = null,
) {
    fun toDomain(): OrderRideCard = OrderRideCard(
        name = name.orEmpty(),
        cardExpiredDate = cardExpiredDate.orEmpty(),
        remainTimes = remainTimes.toFlexibleString().ifBlank { "0" },
        expiredAtMs = cardExpiredDate?.toLongOrNull(),
    )
}

@Serializable
data class OrderRideCardRecordDto(
    @SerialName("paid_at") val paidAt: String? = null,
    val name: String? = null,
    val amount: JsonElement? = null,
    val duration: String? = null,
    val type: String? = null,
    val channel: String? = null,
    @SerialName("merchant_trade_no") val merchantTradeNo: String? = null,
    @SerialName("iz_refund") val izRefund: Int? = null,
) {
    fun toDomain(): OrderRideCardRecord = OrderRideCardRecord(
        paidAt = paidAt.orEmpty(),
        name = name.orEmpty(),
        amountFen = amount.toFen(),
        duration = duration.orEmpty(),
        type = type.orEmpty(),
        channel = channel.orEmpty(),
        merchantTradeNo = merchantTradeNo.orEmpty(),
        izRefund = izRefund ?: 0,
    )
}

private fun JsonElement?.toFlexibleString(): String = when (this) {
    null -> ""
    is JsonPrimitive -> contentOrNull
        ?: longOrNull?.toString()
        ?: content
    else -> toString().trim('"')
}

private fun JsonElement?.toFen(): Long? {
    if (this == null) return null
    if (this !is JsonPrimitive) return null
    longOrNull?.let { return it }
    contentOrNull?.toLongOrNull()?.let { return it }
    contentOrNull?.toDoubleOrNull()?.let { return it.toLong() }
    return null
}

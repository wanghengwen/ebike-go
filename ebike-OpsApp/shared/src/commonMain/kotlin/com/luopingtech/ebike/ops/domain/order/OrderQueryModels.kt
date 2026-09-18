package com.luopingtech.ebike.ops.domain.order

import com.luopingtech.ebike.ops.domain.model.TrackPoint

/**
 * 商户端订单列表 / 末单元素（后端 `bLastOrderDetailCO`）。
 * 金额单位：分；`ridingTime`：毫秒字符串（历史也可能是秒）；`mile`：米。
 */
data class OrderRecord(
    val id: String = "",
    val carId: String = "",
    val phone: String = "",
    val userPin: String = "",
    val userName: String = "",
    /** 分。 */
    val originCost: Long? = null,
    /** 分。 */
    val payCost: Long? = null,
    /** 米。 */
    val mile: Long? = null,
    /** 骑行时长毫秒（对齐遗留 TimeStampUtils.timestampFormat）。 */
    val ridingTimeRaw: String = "",
    val izPaid: Int? = null,
    val startLat: Double? = null,
    val startLng: Double? = null,
    val endLat: Double? = null,
    val endLng: Double? = null,
    val startTime: String = "",
    val endTime: String = "",
    val payTime: String = "",
    /** 分。仅 detailLast 可靠。 */
    val hasPaid: Long? = null,
    val dispatchCost: Long? = null,
    val helmetPenalty: Long? = null,
    val carState: Int? = null,
    /** 设备轨迹；列表可能为空，需 orderDetail 补全。 */
    val trajectory: List<TrackPoint> = emptyList(),
)

/** `izPaid`：后端 `OrderIzPayEnum`。 */
object OrderPayStates {
    const val Riding: Int = 1
    const val Frozen: Int = 2
    const val ToPay: Int = 3
    const val Paid: Int = 4
}

data class OrderListQuery(
    val pageNum: Int = 1,
    val pageSize: Int = 10,
    val userPin: String? = null,
    val carId: String? = null,
    val imei: String? = null,
    val serviceId: Long? = null,
    val izPaid: Int? = null,
    /** `[起, 止]` 毫秒；对应请求体 `startTime`。 */
    val startTimeMs: Pair<Long, Long>? = null,
    /**
     * `[起, 止]` 毫秒；对应请求体 `endTime`。
     * 原版首页近期订单用 `endTime: [now-7d, now]`。
     */
    val endTimeMs: Pair<Long, Long>? = null,
)

data class OrderUserDetail(
    val pin: String,
    val authName: String = "",
    val phone: String = "",
    /** 身份证号，对齐 UserDetailModel.authNo。 */
    val authNo: String = "",
    val createdAt: String = "",
    val izAuth: Boolean? = null,
    val ridingState: Int? = null,
    /** 分。详情页钱包以资产接口为准，此字段仅兜底。 */
    val balance: Long? = null,
    val serviceName: String = "",
    /** 是否缴纳押金（诚信金）。 */
    val izDeposited: Boolean? = null,
    /** 是否购买押金卡/会员卡。 */
    val izDepositCard: Boolean? = null,
    /** 押金卡剩余天数。 */
    val depositedCardDays: String = "",
    /** 0 一键免押 1押金 2押金卡 3职业认证 4微信支付分 */
    val izRidingType: Int? = null,
)

/** 对齐 UserAssetsModel：/business/ebike-account/user_account。 */
data class OrderUserAssets(
    /** 钱包余额，分。 */
    val walletBalanceFen: Long? = null,
    /** 已缴诚信金金额，分。 */
    val depositedMountFen: Long? = null,
    val depositedStats: Int? = null,
    val ridingCardName: String = "",
    /** 押金卡优惠金额，分。 */
    val depositCardDiscountFen: Long? = null,
    /** 押金卡到期时间。 */
    val depositCardExpiredDate: String = "",
    /** 赠送余额，分。 */
    val walletPresentFen: Long? = null,
    /** 充值余额，分。 */
    val walletRechargeFen: Long? = null,
)

data class OrderWalletInfo(
    val balanceFen: Long? = null,
    val rechargeFen: Long? = null,
    val presentFen: Long? = null,
    val depositedMountFen: Long? = null,
    val depositedStats: Int? = null,
)

data class OrderWalletRecord(
    val amountFen: Long? = null,
    val changeType: String = "",
    val channel: String = "",
    val izRefund: Int = 0,
    val merchantTradeNo: String = "",
    val paidAt: String = "",
    val pinName: String = "",
    val presentAmountFen: Long? = null,
    val rechargeAmountFen: Long? = null,
    val type: String = "",
)

data class OrderDepositRecord(
    val paidAt: String = "",
    val depositType: String = "",
    val amountFen: Long? = null,
    val duration: String = "",
    val state: String = "",
    val type: String = "",
    val channel: String = "",
    val merchantTradeNo: String = "",
    val izRefund: Int = 0,
)

data class OrderRideCard(
    val name: String = "",
    val cardExpiredDate: String = "",
    val remainTimes: String = "",
    val expiredAtMs: Long? = null,
)

data class OrderRideCardRecord(
    val paidAt: String = "",
    val name: String = "",
    val amountFen: Long? = null,
    val duration: String = "",
    val type: String = "",
    val channel: String = "",
    val merchantTradeNo: String = "",
    val izRefund: Int = 0,
)

data class OrderUserPageItem(
    val pin: String,
    val authName: String = "",
    val phone: String = "",
    val izAuth: Boolean? = null,
    val ridingState: Int? = null,
    val createdAt: String = "",
    /** 分。 */
    val balance: Long? = null,
)

enum class OrderSearchKind {
    Phone,
    Imei,
    CarId,
    Name,
}

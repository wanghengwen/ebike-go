package com.luopingtech.ebike.ops.domain.order

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
    /** 毫秒或秒的原始字符串。 */
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
    /** `[起, 止]` 毫秒。 */
    val startTimeMs: Pair<Long, Long>? = null,
)

data class OrderUserDetail(
    val pin: String,
    val authName: String = "",
    val phone: String = "",
    val createdAt: String = "",
    val izAuth: Boolean? = null,
    val ridingState: Int? = null,
    /** 分。 */
    val balance: Long? = null,
    val serviceName: String = "",
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

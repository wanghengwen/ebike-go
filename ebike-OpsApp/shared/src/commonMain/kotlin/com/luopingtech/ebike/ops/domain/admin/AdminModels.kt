package com.luopingtech.ebike.ops.domain.admin

import com.luopingtech.ebike.ops.domain.order.OrderRecord

/** 列表状态筛：null=全部，0=待处理，1=已处理。 */
object ObjectionStates {
    const val Pending = 0
    const val Processed = 1
}

data class ObjectionOrderItem(
    val id: String,
    val orderId: String,
    val userName: String,
    val phone: String,
    val carId: String,
    val state: Int,
    val userReason: String = "",
    val createdAt: String = "",
    val izPaid: Int? = null,
    val initiator: Int = 0,
    val opType: Int? = null,
    val payCost: Long? = null,
    val dispatchCost: Long? = null,
    val helmetPenalty: Long? = null,
    val refundCost: Long? = null,
    val refundDispatchCost: Long? = null,
    val refundHelmetPenalty: Long? = null,
    val refundCardTimes: Int? = null,
    val updatePayCost: Long? = null,
    val updateDispatchCost: Long? = null,
    val updateHelmetPenalty: Long? = null,
    val opReason: String = "",
    val photoUrl: String = "",
    val opManName: String = "",
    val opManPhone: String = "",
    val dealAt: String = "",
    val favorableTypes: List<Int> = emptyList(),
    val ridingTimeRaw: String = "",
    val mile: Long? = null,
    val originCost: Long? = null,
)

data class ObjectionOrderDetail(
    val ticket: ObjectionOrderItem,
    val order: OrderRecord? = null,
)

data class ObjectionPageQuery(
    val serviceId: String,
    val pageNum: Int = 1,
    val pageSize: Int = 10,
    /** null=全部 */
    val state: Int? = null,
    val keyword: String = "",
    val createdTimeStart: String = "",
    val createdTimeEnd: String = "",
)

data class ObjectionDealRequest(
    val id: String,
    val orderId: String,
    val initiator: Int,
    /**
     * 对齐原版：UI 选「费用合理」时本字段为 true，
     * 请求体发 `izAccepted = !feeReasonable`。
     */
    val feeReasonable: Boolean,
    /** 已支付 + 通过（采纳）时走 refund* 分支。 */
    val isPaidRefund: Boolean,
    val opReason: String,
    val refundCostFen: Long? = null,
    val refundDispatchCostFen: Long? = null,
    val refundHelmetPenaltyFen: Long? = null,
    val refundCardTimes: Int? = null,
    val modifyPayCostFen: Long? = null,
    val modifyDispatchCostFen: Long? = null,
    val modifyHelmetPenaltyFen: Long? = null,
    val remindWay: List<Int> = emptyList(),
)

/** 职业认证 / 实名换绑审核状态：0 待审 1 通过 2 未通过 3 已取消；换绑另有 4 人脸核验通过。 */
object CertificationAuditStates {
    const val Pending = 0
    const val Passed = 1
    const val Rejected = 2
    const val Cancelled = 3
    const val FacePassed = 4
}

data class CareerAuditItem(
    val id: String,
    val name: String,
    val phone: String,
    val auditState: Int,
    val createdAt: String = "",
    val company: String = "",
    val certNo: String = "",
    val frontCard: String = "",
    val backCard: String = "",
)

data class CareerAuditDetail(
    val id: String,
    val name: String,
    val phone: String,
    val company: String = "",
    val certNo: String = "",
    val frontCard: String = "",
    val backCard: String = "",
    val auditState: Int,
    val createdAt: String = "",
    val auditName: String = "",
    val auditPhone: String = "",
    val auditAt: String = "",
    val reason: String = "",
)

data class BlacklistItem(
    val id: String,
    val authName: String,
    val phone: String,
    val reason: String,
    val state: Int,
    val createdAt: String = "",
)

data class IdBindAuditItem(
    val id: String,
    val authName: String,
    val applyPhone: String,
    val originPhone: String,
    val auditState: Int,
    val createdAt: String = "",
    /** 1 手机号换绑 / 2 身份证换绑 */
    val applyType: Int = 0,
    val authNo: String = "",
    val frontCard: String = "",
    val backCard: String = "",
)

data class IdBindAuditDetail(
    val id: String,
    val authName: String,
    val applyPhone: String,
    val originPhone: String,
    val authNo: String = "",
    val applyType: Int = 0,
    val frontCard: String = "",
    val backCard: String = "",
    val auditState: Int,
    val createdAt: String = "",
    val dealerName: String = "",
    val dealerPhone: String = "",
    val dealTime: String = "",
    val reason: String = "",
)

/** 职业认证 / 实名换绑列表查询（对齐原版 CertificationReq / ChangeBindCertificationReq）。 */
data class CertificationPageQuery(
    val serviceId: String,
    val pageNum: Int = 1,
    val pageSize: Int = 20,
    /** null = 全部状态 */
    val auditState: Int? = null,
    val keyword: String = "",
    val startTime: String = "",
    val endTime: String = "",
)

/**
 * 审核提交。
 * [remindTypes] 为弹窗勾选值：0 系统 / 1 短信 / 2 App；
 * 换绑提交时 Repository 会 +1 对齐后端 1/2/3。
 */
data class CertificationAuditSubmit(
    val id: String,
    val pass: Boolean,
    val reason: String = "",
    val remindTypes: List<Int> = emptyList(),
)

fun CareerAuditDetail.showAuditActions(): Boolean =
    auditState == CertificationAuditStates.Pending

fun CareerAuditDetail.showAuditResult(): Boolean =
    auditState != CertificationAuditStates.Pending &&
        auditState != CertificationAuditStates.Cancelled

fun CareerAuditDetail.showRejectReason(): Boolean =
    auditState == CertificationAuditStates.Rejected

fun IdBindAuditDetail.showAuditActions(): Boolean =
    auditState == CertificationAuditStates.Pending

fun IdBindAuditDetail.showAuditResult(): Boolean =
    auditState != CertificationAuditStates.Pending &&
        auditState != CertificationAuditStates.Cancelled &&
        auditState != CertificationAuditStates.FacePassed

/** 申请类型文案：1 手机号换绑 / 2 身份证换绑。 */
fun applyTypeLabel(applyType: Int, phoneBind: String, idBind: String): String = when (applyType) {
    1 -> phoneBind
    2 -> idBind
    else -> "--"
}

data class OperationLogItem(
    val time: String = "--",
    val name: String = "--",
    val phone: String = "--",
    val carId: String = "--",
    val imei: String = "--",
    val eventName: String = "--",
    val result: String = "",
    /** 兼容旧字段；列表主文案用 eventName，详情用 result。 */
    val content: String = "",
) {
    val operatorName: String get() = name
}

enum class OperationLogKind {
    Car,
    Device,
    Operator,
    ;

    /** eventTree 接口 type：车辆1 / 操作人2 / 设备3。 */
    val eventTreeType: Int
        get() = when (this) {
            Car -> 1
            Operator -> 2
            Device -> 3
        }
}

data class OperationLogEventNode(
    val id: Long,
    val eventName: String,
    val children: List<OperationLogEventNode> = emptyList(),
)

data class OperationLogListQuery(
    val kind: OperationLogKind,
    val carId: String? = null,
    val imei: String? = null,
    val phone: String? = null,
    /** 秒级时间戳字符串 */
    val startTimeSec: String? = null,
    /** 秒级；请求字段名对齐原版拼写 `entTime` */
    val endTimeSec: String? = null,
    /** 事件树选中 id，逗号拼接 */
    val eventTypeIds: String? = null,
    val pageNum: Int = 1,
    val pageSize: Int = 10,
)

fun ObjectionOrderItem.isProcessed(): Boolean = state == ObjectionStates.Processed

fun ObjectionOrderItem.stateLabel(): String =
    if (isProcessed()) "已处理" else "待处理"

/** 支付状态：3 未支付 / 4 已支付（对齐原版）。 */
fun ObjectionOrderItem.isPaid(): Boolean = izPaid == 4

fun ObjectionOrderItem.isUnpaid(): Boolean = izPaid == 3

fun ObjectionOrderItem.opTypeLabel(feeReasonable: Boolean = false): String = when {
    feeReasonable -> "费用合理"
    izPaid == 4 -> "退回费用"
    else -> "修改订单"
}

/** 处理页用到的金额/优惠上下文（工单 + 订单详情合并）。 */
data class ObjectionProcessAmounts(
    val izPaid: Int?,
    val payCostFen: Long,
    val originCostFen: Long,
    val dispatchCostFen: Long,
    val helmetPenaltyFen: Long,
    val favorableTypes: List<Int>,
) {
    fun isPaid(): Boolean = izPaid == 4
    fun isUnpaid(): Boolean = izPaid == 3
    fun showRefundOrigin(): Boolean = originCostFen > 0
    fun showRefundDispatch(): Boolean = dispatchCostFen > 0
    fun showCardTimes(): Boolean = favorableTypes.isNotEmpty()
    fun showHelmetInput(): Boolean = helmetPenaltyFen > 0
    fun payCostYuanMax(): Double = payCostFen / 100.0
    fun originCostYuanMax(): Double = originCostFen / 100.0
    fun dispatchCostYuanMax(): Double = dispatchCostFen / 100.0
    fun helmetYuanMax(): Double = helmetPenaltyFen / 100.0
}

fun ObjectionOrderDetail.processAmounts(): ObjectionProcessAmounts {
    val order = order
    return ObjectionProcessAmounts(
        izPaid = ticket.izPaid ?: order?.izPaid,
        payCostFen = order?.payCost ?: ticket.payCost ?: 0L,
        originCostFen = order?.originCost ?: ticket.originCost ?: 0L,
        dispatchCostFen = ticket.dispatchCost ?: order?.dispatchCost ?: 0L,
        helmetPenaltyFen = ticket.helmetPenalty ?: order?.helmetPenalty ?: 0L,
        favorableTypes = ticket.favorableTypes,
    )
}

fun ObjectionOrderItem.favorableTypeLabel(): String {
    val type = favorableTypes.firstOrNull() ?: return "无折扣"
    return when (type) {
        1 -> "骑行卡"
        2 -> "免单卡"
        3 -> "折扣卡"
        else -> "无折扣"
    }
}

fun ObjectionOrderItem.handleWayLabel(feeReasonable: Boolean, amounts: ObjectionProcessAmounts): String {
    if (feeReasonable) return "费用合理"
    val parts = mutableListOf<String>()
    if (amounts.showCardTimes()) parts += "退还次数"
    when (amounts.izPaid) {
        4 -> parts += "退回钱包"
        3 -> parts += "修改订单"
        else -> if (parts.isEmpty()) parts += "-"
    }
    return parts.joinToString(";")
}

data class ObjectionSendMode(
    val izSys: Boolean = false,
    val izSms: Boolean = false,
    val izApp: Boolean = false,
)

/** 对齐原版 InputFilterMinMax：0~max，最多 2 位小数，总长不超过 5；非法输入保留旧值。 */
fun filterObjectionAmountYuan(raw: String, maxYuan: Double, current: String): String {
    if (raw.isEmpty()) return ""
    if (raw == ".") return "0."
    if (!raw.matches(Regex("""^\d*\.?\d*$"""))) return current
    if (raw.length > 5) return current
    if (raw == "00") return current
    val dot = raw.indexOf('.')
    if (dot >= 0 && raw.length - dot - 1 > 2) return current
    val probe = if (raw.endsWith(".")) raw.dropLast(1) else raw
    if (probe.isEmpty()) return raw
    val v = probe.toDoubleOrNull() ?: return current
    val upper = maxYuan.coerceAtLeast(0.0)
    if (v < 0.0 || v > upper) return current
    return raw
}

fun filterObjectionCardTimes(raw: String, current: String): String {
    if (raw.isEmpty()) return ""
    if (raw.length > 1) return current
    return if (raw == "0" || raw == "1") raw else current
}

/** 元字符串 → 分；空返回 null。 */
fun objectionYuanToFen(raw: String): Long? {
    val text = raw.trim()
    if (text.isEmpty()) return null
    val neg = text.startsWith('-')
    val body = text.removePrefix("-").removePrefix("+")
    if (body.isEmpty() || body == ".") return null
    val parts = body.split('.')
    if (parts.size > 2) return null
    val whole = parts[0].ifEmpty { "0" }.toLongOrNull() ?: return null
    val frac = when {
        parts.size == 1 -> 0L
        else -> parts[1].padEnd(2, '0').take(2).toLongOrNull() ?: return null
    }
    val fen = whole * 100 + frac
    return if (neg) -fen else fen
}

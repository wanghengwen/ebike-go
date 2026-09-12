package com.luopingtech.ebike.ops.domain.model

import com.luopingtech.ebike.ops.core.i18n.Str

/** Sneak / violation report type from sneak_type_list/query (not repairConfig). */
data class SneakReportType(
    val id: String,
    val name: String,
    /** Legacy CommonRepairModel.type; 12 = "other". */
    val type: Int = 0,
) {
    val isOther: Boolean get() = type == 12
}

data class SneakReportRecord(
    val id: String,
    val carId: String = "",
    val reportedUserName: String = "",
    val reportedUserPhone: String = "",
    val reportedUserPin: String = "",
    val itinId: String = "",
    val typeLabels: List<String> = emptyList(),
    val otherType: String = "",
    val description: String = "",
    val photoUrls: List<String> = emptyList(),
    val createdAt: String = "",
    /** 0 pending … 5 cancelled. */
    val checkResult: Int = 0,
    val remark: String = "",
    val handleType: Int = 0,
    val serviceId: String = "",
) {
    val canCancel: Boolean get() = checkResult == 0
}

object SneakCheckResultLabels {
    fun strKey(checkResult: Int): Str = when (checkResult) {
        0 -> Str.SneakStatusPending
        1 -> Str.SneakStatusAutoReject
        2 -> Str.SneakStatusAutoPass
        3 -> Str.SneakStatusManualPass
        4 -> Str.SneakStatusManualReject
        5 -> Str.SneakStatusCancelled
        else -> Str.SneakStatusUnknown
    }
}

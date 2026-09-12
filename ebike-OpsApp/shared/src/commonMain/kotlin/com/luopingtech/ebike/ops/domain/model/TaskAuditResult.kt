package com.luopingtech.ebike.ops.domain.model

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings

/**
 * Task photo-audit result (legacy TaskAuditResultActivity / iOS TaskCheckResultViewController).
 *
 * checkResult: 0 pending / 1 auto reject / 2 auto pass / 3 manual pass / 4 manual reject.
 */
data class TaskAuditResult(
    val checkResult: Int? = null,
    val checkView: String = "",
    val remark: String = "",
    val photo: List<String> = emptyList(),
    val checkTime: String = "",
    val checkManName: String = "",
) {
    val statusLabel: String
        get() = when (checkResult) {
            1 -> Strings.t(Str.AuditResultAutoReject)
            2 -> Strings.t(Str.AuditResultAutoPass)
            3 -> Strings.t(Str.AuditResultManualPass)
            4 -> Strings.t(Str.AuditResultManualReject)
            else -> Strings.t(Str.AuditResultPending)
        }

    val isRejected: Boolean get() = checkResult == 1 || checkResult == 4
    val isPassed: Boolean get() = checkResult == 2 || checkResult == 3
}

/** Legacy showAuditResultView — finished and not auto-pass. */
fun OpsTask.canViewAuditResult(): Boolean =
    when (type) {
        OpsTaskType.Inspection,
        OpsTaskType.Repair,
        OpsTaskType.MoveCar,
        -> state == 2 && checkResult != 2
        else -> false
    }

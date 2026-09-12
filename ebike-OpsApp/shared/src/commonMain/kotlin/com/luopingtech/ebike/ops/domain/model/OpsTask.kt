package com.luopingtech.ebike.ops.domain.model

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings

/**
 * Field ops task types. Scaffold starts with change-battery; others follow the same shell.
 */
enum class OpsTaskType {
    ChangeBattery,
    MoveCar,
    Inspection,
    Repair,
}

/**
 * Change-battery task item aligned with Flutter `TaskModel` camelCase fields.
 *
 * state: 0 pending / 1 in progress / 2 finished / 3 closed
 */
data class OpsTask(
    val id: String,
    val type: OpsTaskType = OpsTaskType.ChangeBattery,
    val carId: String = "",
    val imei: String = "",
    val address: String = "",
    val areaId: String = "",
    val areaName: String = "",
    val state: Int = 0,
    val checkResult: Int? = null,
    val restBattery: Int = 0,
    val lat: Double = 0.0,
    val lng: Double = 0.0,
    val serviceId: String = "",
    val createdAt: String = "",
    /** 领取/开始时间（处理中时长用）。 */
    val startTime: String = "",
    val finishTime: String = "",
    /**
     * Repair drag-back: 1 = none, 2 = in progress, 3 = done (legacy dragState).
     * Null / 0 = not applicable (non-repair) or unknown.
     */
    val dragState: Int? = null,
    /** Task source: 2 = man-made (人工). See legacy TaskInfoModel.source. */
    val source: Int? = null,
    /** Move type: 1 = single, 2 = multi-vehicle batch parent. */
    val moveType: Int? = null,
    /** Parent batch record id when [isManMadeBatch]. */
    val recordId: String? = null,
    /** 维修停运标记（仅看停运车）。 */
    val izStop: Boolean = false,
) {
    val stateLabel: String
        get() = when (state) {
            1 -> Strings.t(Str.TaskStateInProgress)
            2 -> Strings.t(Str.TaskStateDone)
            3 -> Strings.t(Str.TaskStateClosed)
            else -> Strings.t(Str.TaskStatePending)
        }

    val isActionable: Boolean get() = state == 0 || state == 1

    val isDragBacking: Boolean get() = dragState == 2

    /** Legacy: source==2 && move_type==2 → open man_made_move_list children. */
    val isManMadeBatch: Boolean get() = source == 2 && moveType == 2

    val batchParentId: String
        get() = when {
            !isManMadeBatch -> id
            !recordId.isNullOrBlank() -> recordId
            else -> id
        }
}

/** Tenant rule for change-battery list filter (`task_rules/battery_range`). */
data class BatteryRange(
    val min: Int = 0,
    val max: Int = 30,
)

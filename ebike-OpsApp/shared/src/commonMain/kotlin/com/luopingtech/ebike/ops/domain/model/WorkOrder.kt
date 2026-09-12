package com.luopingtech.ebike.ops.domain.model

import com.luopingtech.ebike.ops.core.i18n.Str

/** Legacy work-order ledger (1208/1209), not task-center 123403/123404. */
enum class WorkOrderKind {
    Inspection,
    Repair,
}

/**
 * Shared FixState for alarm/fix tickets:
 * 0=new, 1=processing, 2=done, 3=dragged-back.
 */
data class WorkOrder(
    val id: String,
    val kind: WorkOrderKind,
    val carId: String,
    val imei: String = "",
    val state: Int = 0,
    /** AlarmState 1–14 for inspection; unused for repair. */
    val alarmType: Int? = null,
    /** Display name for repair parts (from nameExtraInfo). */
    val partNames: List<String> = emptyList(),
    val fixReason: String = "",
    val opManName: String = "",
    val opManPhone: String = "",
    val createdAt: String = "",
    val fixedTime: String = "",
    val carStopped: Boolean = false,
    val serviceId: String = "",
) {
    val canAccept: Boolean get() = state == 0
    val canFinish: Boolean get() = state == 1
}

object AlarmTypeLabels {
    /** Codes shown on inspection work-order filter chips (legacy AlarmState 1–14). */
    val ALL_CODES: List<Int> = (1..14).toList()

    fun strKey(code: Int): Str = when (code) {
        1 -> Str.AlarmTypeHighVoltage
        2 -> Str.AlarmTypeAbnormalMove
        3 -> Str.AlarmTypeOutOfArea
        4 -> Str.AlarmTypeNoParking
        5 -> Str.AlarmTypeNotAtStation
        6 -> Str.AlarmTypeBatteryRemoved
        7 -> Str.AlarmTypeOffline
        8 -> Str.AlarmTypeOrderNoTrip
        9 -> Str.AlarmTypeLost
        10 -> Str.AlarmTypeOrderTooLong
        11 -> Str.AlarmTypeShortOrder
        12 -> Str.AlarmTypeAbnormalUnlock
        13 -> Str.AlarmTypeHelmetLost
        14 -> Str.AlarmTypeHelmetFault
        else -> Str.AlarmTypeUnknown
    }
}

object WorkOrderStateLabels {
    fun strKey(state: Int): Str = when (state) {
        0 -> Str.WorkOrderStateNew
        1 -> Str.WorkOrderStateProcessing
        2 -> Str.WorkOrderStateDone
        3 -> Str.WorkOrderStateDragged
        else -> Str.WorkOrderStateUnknown
    }
}

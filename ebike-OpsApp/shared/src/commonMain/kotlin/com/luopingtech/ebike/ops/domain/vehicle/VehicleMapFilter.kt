package com.luopingtech.ebike.ops.domain.vehicle

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings

/**
 * Legacy [VehicleStateRing] / [StatisticsStateType] codes used by home map filters.
 * Keep numeric values identical to Merchant-Android.
 */
object VehicleRidingStates {
    const val RIDEABLE: Int = 1 // 待使用
    const val RIDING: Int = 2 // 骑行中
    const val TEMP_PARKING: Int = 3 // 临停
    const val RESERVE: Int = 4 // 预约中
    const val OPERATION: Int = 5 // 运维中
}

/** Legacy [VehicleStateOperation] subset used by home statistics / sold-out / badges. */
object VehicleOperationStates {
    const val OFF: Int = 1 // 下架（首页始终隐藏）
    const val MOVING_CAR: Int = 2 // 挪车中 / 调度
    const val CHANGING_BATTERY: Int = 3 // 换电中
    const val LOW_BATTERY: Int = 4 // 低电
    const val REPAIRING: Int = 5 // 报修
    const val DRAG_BACK: Int = 6 // 拖回
}

/**
 * Legacy [HomeFilterVehicleState] / [VehicleStateAlarm] codes for filter panel + badges.
 */
object VehicleAlarmStates {
    const val MOVE: Int = 2 // 异常移动
    const val OUT_GFENCE: Int = 3 // 出服务区
    const val NO_PARKING_ZONE: Int = 4 // 禁停区
    const val OUT_PARKING_ZONE: Int = 5 // 站点外
    const val POWER_CUT: Int = 6 // 电瓶移除
    const val OFFLINE: Int = 7 // 异常离线
    const val ORDER_WITHOUT_GPS: Int = 8 // 有单无程
    const val LOST: Int = 9 // 报失
    const val TOO_LONG_ORDER: Int = 10 // 订单超长
    const val TOO_SHORT_ORDER: Int = 11 // 短时订单
    const val UNLOCK_ABNORMAL: Int = 12 // 异常开锁
    const val HELMET_LOST: Int = 13 // 头盔丢失
    const val HELMET_FAULT: Int = 14 // 头盔故障
}

enum class VehicleAlarmFilter(val code: Int) {
    Offline(VehicleAlarmStates.OFFLINE),
    PowerCut(VehicleAlarmStates.POWER_CUT),
    OutGfence(VehicleAlarmStates.OUT_GFENCE),
    Move(VehicleAlarmStates.MOVE),
    Lost(VehicleAlarmStates.LOST),
    OrderWithoutGps(VehicleAlarmStates.ORDER_WITHOUT_GPS),
    OutParking(VehicleAlarmStates.OUT_PARKING_ZONE),
    UnlockAbnormal(VehicleAlarmStates.UNLOCK_ABNORMAL),
    NoParking(VehicleAlarmStates.NO_PARKING_ZONE),
    HelmetLost(VehicleAlarmStates.HELMET_LOST),
    HelmetFault(VehicleAlarmStates.HELMET_FAULT),
    ;

    val label: String
        get() = when (this) {
            Offline -> Strings.t(Str.Offline)
            PowerCut -> Strings.t(Str.AlarmPowerCut)
            OutGfence -> Strings.t(Str.AlarmOutGfence)
            Move -> Strings.t(Str.AlarmAbnormalMove)
            Lost -> Strings.t(Str.AlarmLost)
            OrderWithoutGps -> Strings.t(Str.AlarmOrderWithoutGps)
            OutParking -> Strings.t(Str.AlarmOutParking)
            UnlockAbnormal -> Strings.t(Str.AlarmUnlockAbnormal)
            NoParking -> Strings.t(Str.AlarmNoParking)
            HelmetLost -> Strings.t(Str.AlarmHelmetLost)
            HelmetFault -> Strings.t(Str.AlarmHelmetFault)
        }
}

object VehicleAlarmFilterLogic {
    /** Legacy always drops sold-out (operationState contains 1). */
    fun isSoldOut(operationStates: Collection<Int>): Boolean =
        operationStates.contains(VehicleOperationStates.OFF)

    /**
     * Empty [selected] = no alarm constraint.
     * Non-empty = vehicle must hit any selected code (OR), matching HomeFragmentV3.
     * Offline is alarmState code 7 only — do not treat heartbeat [isOnline]=false as a hit.
     */
    fun matchesAlarms(
        alarmStates: Collection<Int>,
        isOnline: Boolean,
        selected: Set<Int>,
    ): Boolean {
        if (selected.isEmpty()) return true
        return alarmStates.any { it in selected }
    }

    fun counts(
        vehicles: List<com.luopingtech.ebike.ops.domain.model.Vehicle>,
    ): Map<VehicleAlarmFilter, Int> =
        VehicleAlarmFilter.entries.associateWith { filter ->
            vehicles.count { v ->
                !isSoldOut(v.operationStates) &&
                    matchesAlarms(v.alarmStates, v.isOnline, setOf(filter.code))
            }
        }
}

enum class VehicleMapFilter {
    All,
    Warehouse,
    Ready,
    Booking,
    Riding,
    TempParking,
    LowBattery,
    Repairing,
    Moving,
    ;

    val label: String
        get() = when (this) {
            All -> Strings.t(Str.FilterAll)
            Warehouse -> Strings.t(Str.FilterWarehouse)
            Ready -> Strings.t(Str.FilterReady)
            Booking -> Strings.t(Str.FilterBooking)
            Riding -> Strings.t(Str.FilterRiding)
            TempParking -> Strings.t(Str.FilterTempParking)
            LowBattery -> Strings.t(Str.FilterLowBattery)
            Repairing -> Strings.t(Str.FilterRepairing)
            Moving -> Strings.t(Str.FilterMoving)
        }
}

object VehicleMapFilterLogic {
    fun matches(
        ridingState: Int?,
        operationStates: Collection<Int>,
        restBattery: Int,
        filter: VehicleMapFilter,
    ): Boolean = when (filter) {
        VehicleMapFilter.All -> true
        // Legacy HomeViewModelV3 warehouse branch is a stub (always false).
        VehicleMapFilter.Warehouse -> false
        VehicleMapFilter.Ready -> ridingState == VehicleRidingStates.RIDEABLE
        VehicleMapFilter.Booking -> ridingState == VehicleRidingStates.RESERVE
        VehicleMapFilter.Riding -> ridingState == VehicleRidingStates.RIDING
        VehicleMapFilter.TempParking -> ridingState == VehicleRidingStates.TEMP_PARKING
        VehicleMapFilter.LowBattery ->
            operationStates.contains(VehicleOperationStates.LOW_BATTERY)
        VehicleMapFilter.Repairing ->
            operationStates.contains(VehicleOperationStates.REPAIRING)
        VehicleMapFilter.Moving ->
            operationStates.contains(VehicleOperationStates.MOVING_CAR)
    }

    fun counts(
        vehicles: List<com.luopingtech.ebike.ops.domain.model.Vehicle>,
    ): Map<VehicleMapFilter, Int> {
        val visible = vehicles.filterNot { VehicleAlarmFilterLogic.isSoldOut(it.operationStates) }
        return VehicleMapFilter.entries.associateWith { filter ->
            if (filter == VehicleMapFilter.All) {
                visible.size
            } else {
                visible.count {
                    matches(it.ridingState, it.operationStates, it.restBattery, filter)
                }
            }
        }
    }
}

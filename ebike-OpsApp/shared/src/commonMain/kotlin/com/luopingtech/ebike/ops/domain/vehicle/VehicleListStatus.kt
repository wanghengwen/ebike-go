package com.luopingtech.ebike.ops.domain.vehicle

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.domain.model.Vehicle

enum class VehicleListStatusTone {
    Positive,
    Alert,
    Neutral,
}

data class VehicleListStatus(
    val label: String,
    val tone: VehicleListStatusTone,
)

/**
 * 工作台车辆列表「车辆状态」列：优先运维中 / 异常离线，其余回落到骑行态文案。
 * 「短时订单」遗留 carInfo 字段本接口没有，暂不捏造。
 */
fun Vehicle.listStatus(): VehicleListStatus {
    if (ridingState == VehicleRidingStates.OPERATION) {
        return VehicleListStatus(Strings.t(Str.RidingOperation), VehicleListStatusTone.Positive)
    }
    if (alarmStates.contains(VehicleAlarmStates.OFFLINE) || !isOnline) {
        return VehicleListStatus(Strings.t(Str.AlarmTypeOffline), VehicleListStatusTone.Alert)
    }
    if (operationStates.contains(VehicleOperationStates.MOVING_CAR)) {
        return VehicleListStatus(Strings.t(Str.FilterMoving), VehicleListStatusTone.Alert)
    }
    if (operationStates.contains(VehicleOperationStates.REPAIRING)) {
        return VehicleListStatus(Strings.t(Str.FilterRepairing), VehicleListStatusTone.Alert)
    }
    if (ridingState == VehicleRidingStates.RIDING || ridingState == VehicleRidingStates.TEMP_PARKING) {
        return VehicleListStatus(ridingLabel, VehicleListStatusTone.Alert)
    }
    return VehicleListStatus(ridingLabel, VehicleListStatusTone.Neutral)
}

package com.luopingtech.ebike.ops.domain.model

import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.core.i18n.Strings
import com.luopingtech.ebike.ops.domain.vehicle.VehicleAlarmFilter
import com.luopingtech.ebike.ops.domain.vehicle.VehicleAlarmStates
import com.luopingtech.ebike.ops.domain.vehicle.VehicleOperationStates
import com.luopingtech.ebike.ops.domain.vehicle.VehicleRidingStates

data class Vehicle(
    val carId: String,
    val imei: String = "",
    val lat: Double = 0.0,
    val lng: Double = 0.0,
    val restBattery: Int = 0,
    /** Legacy VehicleStateRing: 1待使用 2骑行 3临停 4预约 5运维. */
    val ridingState: Int? = null,
    /** Legacy operationState codes (low battery / repairing / moving…). */
    val operationStates: List<Int> = emptyList(),
    /** Legacy alarmState codes (offline / out of fence / …). */
    val alarmStates: List<Int> = emptyList(),
    val isOnline: Boolean = false,
    val serviceId: String = "",
    /** Voltage in millivolts from device detail/list. */
    val voltageMv: Int? = null,
    val batterySn: String = "",
    /** Legacy: 0 = open/unlocked, non-zero = locked. Null = unknown. */
    val batteryLock: Int? = null,
    val backWheelLock: Int? = null,
    val helmetLock: Int? = null,
    val helmetReact: Int? = null,
    val helmetState: Int? = null,
    val helmetMac: String = "",
    val forParkName: String = "",
    val noParkName: String = "",
    val isOutOfServiceArea: Boolean? = null,
    val isFenceEnable: Boolean? = null,
    val totalMiles: Double? = null,
    val serviceName: String = "",
    /** Vehicle model id used by repairConfig/list. */
    val model: String = "",
    /** Legacy izHaveOverload — saddle overload hardware present. */
    val izHaveOverload: Boolean = false,
    val acc: Int? = null,
    val defend: Int? = null,
    val timestamp: Long? = null,
    /** 关锁时间 epoch ms；闲置分档用。接口可能是数字或数字字符串。 */
    val lockTimeMs: Long = 0L,
    /** 开锁时间 epoch ms。 */
    val unlockTimeMs: Long = 0L,
    /** Firmware / device version from detail. */
    val version: String = "",
    /** Legacy gsmSignal, shown as `{n}dbm`. */
    val gsmSignal: Int? = null,
    /** >0 means inertial navigation on (legacy headingAngle). */
    val headingAngle: Int = -1,
    /** 0 = helmet not supported / unbound. */
    val helmetBind: Int? = null,
) {
    val batteryLabel: String get() = "$restBattery%"

    val ridingLabel: String
        get() = when (ridingState) {
            VehicleRidingStates.RIDEABLE -> Strings.t(Str.FilterReady)
            VehicleRidingStates.RIDING -> Strings.t(Str.FilterRiding)
            VehicleRidingStates.TEMP_PARKING -> Strings.t(Str.FilterTempParking)
            VehicleRidingStates.RESERVE -> Strings.t(Str.FilterBooking)
            VehicleRidingStates.OPERATION -> Strings.t(Str.RidingOperation)
            else -> Strings.t(Str.RidingUnknown)
        }

    val voltageLabel: String
        get() {
            val mv = voltageMv ?: return "-"
            if (mv <= 0) return "-"
            val whole = mv / 1000
            val frac = (mv % 1000) / 100
            return "${whole}.${frac}V"
        }

    /** Legacy VehicleDevInfoModel.showSingleText. */
    val signalLabel: String get() = "${gsmSignal ?: 0}dbm"

    val batteryLockLabel: String get() = lockLabel(batteryLock)

    val helmetLockLabel: String get() = lockLabel(helmetLock)

    val mileageLabel: String
        get() {
            val miles = totalMiles ?: return "-"
            return if (miles == miles.toLong().toDouble()) {
                "${miles.toLong()} km"
            } else {
                "$miles km"
            }
        }

    val operationLabels: String
        get() = operationStates.mapNotNull { code ->
            when (code) {
                VehicleOperationStates.OFF -> Strings.t(Str.OpsSoldOut)
                VehicleOperationStates.MOVING_CAR -> Strings.t(Str.FilterMoving)
                VehicleOperationStates.LOW_BATTERY -> Strings.t(Str.FilterLowBattery)
                VehicleOperationStates.REPAIRING -> Strings.t(Str.FilterRepairing)
                else -> null
            }
        }.distinct().joinToString(" · ").ifBlank { "-" }

    val alarmLabels: String
        get() = alarmStates.mapNotNull { code ->
            VehicleAlarmFilter.entries.firstOrNull { it.code == code }?.label
                ?: when (code) {
                    VehicleAlarmStates.OFFLINE -> Strings.t(Str.Offline)
                    else -> null
                }
        }.distinct().joinToString(" · ").ifBlank { "-" }

    val siteLabel: String
        get() = buildString {
            val park = forParkName.ifBlank { null }
            val noPark = noParkName.ifBlank { null }
            when {
                park != null && noPark != null -> append("$park / $noPark")
                park != null -> append(park)
                noPark != null -> append(noPark)
                else -> append("-")
            }
            if (isOutOfServiceArea == true) {
                append(" · ").append(Strings.t(Str.VehicleOutOfArea))
            }
        }

    private fun lockLabel(code: Int?): String = when (code) {
        null -> "-"
        0 -> Strings.t(Str.LockOpen)
        else -> Strings.t(Str.LockClosed)
    }
}

/**
 * Lightweight pin for host map rendering. Shared does not draw maps.
 */
enum class MapPinIcon {
    Default,
    /** 用车人起点（遗留 icon_user_start）。 */
    UserStart,
    /** 还车人终点（遗留 icon_user_end）。 */
    UserEnd,
    /** 车辆轨迹起点（遗留 btn_trajectory_origin）。 */
    TrackOrigin,
    /** 车辆轨迹终点（遗留 btn_trajectory_end）。 */
    TrackEnd,
    /** 电量绿标（遗留 ico_vehicle_normal，详情页常用）。 */
    VehicleNormal,
    /** 电量黄标（遗留 ico_vehicle_warning）。 */
    VehicleWarning,
    /** 电量红标（遗留 ico_vehicle_error）。 */
    VehicleError,
    /** 首页 V3 状态车标：可使用。 */
    VehicleReady,
    /** 骑行中。 */
    VehicleRiding,
    /** 预约中。 */
    VehicleBooking,
    /** 临停。 */
    VehicleTempParking,
    /** 低电。 */
    VehicleLowBattery,
    /** 挪车中。 */
    VehicleMoving,
    /** 报修。 */
    VehicleRepairing,
    /** 默认状态车标 icon_vehicle。 */
    VehicleHome,
    /** 普通停车区带杆图钉（遗留 icon_parking_normal_unselect）。 */
    Parking,
    /** 功能/正规停车区带杆图钉（遗留 icon_parking_function_unselect）。 */
    ParkingFunction,
    /** 停用停车区带杆图钉（遗留 icon_parking_hiden_unselect）。 */
    ParkingHidden,
    /** 禁停区带杆图钉（遗留 icon_no_parking_unselect）。 */
    NoParking,
}

fun MapPinIcon.legacyDrawableName(): String = when (this) {
    MapPinIcon.UserStart -> "icon_user_start"
    MapPinIcon.UserEnd -> "icon_user_end"
    MapPinIcon.TrackOrigin -> "btn_trajectory_origin"
    MapPinIcon.TrackEnd -> "btn_trajectory_end"
    MapPinIcon.VehicleWarning -> "ico_vehicle_warning"
    MapPinIcon.VehicleError -> "ico_vehicle_error"
    MapPinIcon.VehicleReady -> "icon_vehicle_ready"
    MapPinIcon.VehicleRiding -> "icon_vehicle_riding"
    MapPinIcon.VehicleBooking -> "icon_vehicle_booking"
    MapPinIcon.VehicleTempParking -> "icon_vehicle_temp_parking"
    MapPinIcon.VehicleLowBattery -> "icon_vehicle_low_battery"
    MapPinIcon.VehicleMoving -> "icon_vehicle_moving"
    MapPinIcon.VehicleRepairing -> "icon_vehicle_repairing"
    MapPinIcon.VehicleHome -> "icon_vehicle"
    MapPinIcon.Parking -> "icon_parking_normal_unselect"
    MapPinIcon.ParkingFunction -> "icon_parking_function_unselect"
    MapPinIcon.ParkingHidden -> "icon_parking_hiden_unselect"
    MapPinIcon.NoParking -> "icon_no_parking_unselect"
    MapPinIcon.VehicleNormal,
    MapPinIcon.Default,
    -> "ico_vehicle_normal"
}

data class MapPin(
    val id: String,
    val lat: Double,
    val lng: Double,
    val title: String,
    val subtitle: String = "",
    val restBattery: Int = 0,
    val ridingState: Int? = null,
    /** 1 = single vehicle; >1 = cluster marker. */
    val memberCount: Int = 1,
    val memberIds: List<String> = emptyList(),
    val icon: MapPinIcon = MapPinIcon.Default,
    /** Legacy tips_* 角标资源名；null 表示无角标。 */
    val badgeDrawableName: String? = null,
    /** 对齐 legacy Cluster.showCluster：聚合模式下即使只有 1 台车也画数字气泡。 */
    val showCluster: Boolean = false,
) {
    val isCluster: Boolean get() = memberCount > 1

    /**
     * 对齐 legacy DefaultOptionGenerator：`clusterItems.size > 1 || showCluster == true` 时画数字气泡，
     * 否则画车辆状态图标。注意点击语义仍看 [isCluster]（气泡里只有 1 台车时 tag 是那台车）。
     */
    val isClusterBubble: Boolean get() = memberCount > 1 || showCluster
}

/**
 * Legacy MapOptionProvide.getVehicleIcon(ridingState, operationState) — 首页 V3 单车点。
 * 运维态覆盖骑行态：报修 > 挪车 > 低电。
 */
fun Vehicle.homeMapPinIcon(): MapPinIcon {
    var icon = when (ridingState) {
        VehicleRidingStates.RIDEABLE -> MapPinIcon.VehicleReady
        VehicleRidingStates.RIDING -> MapPinIcon.VehicleRiding
        VehicleRidingStates.RESERVE -> MapPinIcon.VehicleBooking
        VehicleRidingStates.TEMP_PARKING -> MapPinIcon.VehicleTempParking
        else -> MapPinIcon.VehicleHome
    }
    if (operationStates.contains(VehicleOperationStates.LOW_BATTERY)) {
        icon = MapPinIcon.VehicleLowBattery
    }
    if (operationStates.contains(VehicleOperationStates.MOVING_CAR)) {
        icon = MapPinIcon.VehicleMoving
    }
    if (operationStates.contains(VehicleOperationStates.REPAIRING)) {
        icon = MapPinIcon.VehicleRepairing
    }
    return icon
}

/**
 * Legacy BaseStateModel.getBadgeIcon() → tips_* drawable name（无角标返回 null）。
 * 优先级：骑行态角标 → 运维态 → 告警态。
 */
fun Vehicle.homeMapBadgeDrawableName(): String? {
    when (ridingState) {
        VehicleRidingStates.TEMP_PARKING -> return "tips_parking"
        VehicleRidingStates.RESERVE -> return "tips_bebooked"
        VehicleRidingStates.RIDING -> return "tips_riding"
    }
    when {
        operationStates.contains(VehicleOperationStates.OFF) -> return "tips_off_shelf"
        operationStates.contains(VehicleOperationStates.CHANGING_BATTERY) -> return "tips_battery_changing"
        operationStates.contains(VehicleOperationStates.REPAIRING) -> return "tips_repair"
        operationStates.contains(VehicleOperationStates.DRAG_BACK) -> return "tips_dragback"
        operationStates.contains(VehicleOperationStates.LOW_BATTERY) -> return "tips_low_power"
        operationStates.contains(VehicleOperationStates.MOVING_CAR) -> return "tips_occupied"
    }
    val alarms = alarmStates.toSet() + if (!isOnline) setOf(VehicleAlarmStates.OFFLINE) else emptySet()
    return when {
        VehicleAlarmStates.LOST in alarms -> "tips_lost"
        VehicleAlarmStates.OFFLINE in alarms -> "tips_offline"
        VehicleAlarmStates.POWER_CUT in alarms -> "tips_battery_removal"
        VehicleAlarmStates.OUT_GFENCE in alarms -> "tips_out_service"
        VehicleAlarmStates.MOVE in alarms -> "tips_movement"
        VehicleAlarmStates.ORDER_WITHOUT_GPS in alarms -> "tips_bebooked"
        VehicleAlarmStates.TOO_LONG_ORDER in alarms -> "tips_order_timeout"
        VehicleAlarmStates.OUT_PARKING_ZONE in alarms -> "tips_offsite"
        VehicleAlarmStates.NO_PARKING_ZONE in alarms -> "tips_np_parking"
        VehicleAlarmStates.UNLOCK_ABNORMAL in alarms -> "tips_movement"
        VehicleAlarmStates.TOO_SHORT_ORDER in alarms -> "tips_short_term"
        else -> null
    }
}

/** Legacy MapOptionProvide.getVehicleIcon(restBattery) — 电量三色车标。 */
fun Vehicle.batteryMapPinIcon(): MapPinIcon = when {
    restBattery in 0..29 -> MapPinIcon.VehicleError
    restBattery in 30..59 -> MapPinIcon.VehicleWarning
    else -> MapPinIcon.VehicleNormal
}

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
) {
    val isCluster: Boolean get() = memberCount > 1
}

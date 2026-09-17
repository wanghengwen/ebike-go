package com.luopingtech.ebike.ops.data.vehicle

import com.luopingtech.ebike.ops.data.analysis.FlexibleBooleanSerializer
import com.luopingtech.ebike.ops.domain.model.MapPin
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.domain.model.homeMapPinIcon
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class VehicleDto(
    val carId: String = "",
    val imei: String = "",
    val lat: Double = 0.0,
    val lng: Double = 0.0,
    val restBattery: Int? = 0,
    val ridingState: Int? = null,
    val operationState: List<Int>? = null,
    val alarmState: List<Int>? = null,
    @SerialName("isOnline")
    val isOnline: Int? = 0,
    val serviceId: String? = "",
    val battery: String? = null,
    val voltage: Int? = null,
    val batterySn: String? = null,
    val batteryLock: Int? = null,
    val backWheelLock: Int? = null,
    val helmetLock: Int? = null,
    val helmetReact: Int? = null,
    val helmetState: Int? = null,
    val helmetMac: String? = null,
    val forParkName: String? = null,
    val noParkName: String? = null,
    /** Legacy VehicleModel uses Int 0/1; production also returns 0/1. */
    @Serializable(with = FlexibleBooleanSerializer::class)
    val isOutofServAera: Boolean? = null,
    @Serializable(with = FlexibleBooleanSerializer::class)
    val isFenceEnable: Boolean? = null,
    val totalMiles: Double? = null,
    val serviceName: String? = null,
    val model: String? = null,
    @Serializable(with = FlexibleBooleanSerializer::class)
    val izHaveOverload: Boolean? = null,
    val acc: Int? = null,
    val defend: Int? = null,
    @Serializable(with = FlexibleNullableEpochMsSerializer::class)
    val timestamp: Long? = null,
    @Serializable(with = FlexibleEpochMsSerializer::class)
    val lockTime: Long = 0L,
    @Serializable(with = FlexibleEpochMsSerializer::class)
    val unlockTime: Long = 0L,
    val version: String? = null,
    val gsmSignal: Int? = null,
    val headingAngle: Int? = null,
    val helmetBind: Int? = null,
) {
    fun toDomain(): Vehicle {
        val batteryFromRest = restBattery ?: 0
        val batteryFromText = battery?.filter { it.isDigit() }?.toIntOrNull()
        return Vehicle(
            carId = carId,
            imei = imei,
            lat = lat,
            lng = lng,
            restBattery = if (batteryFromRest > 0) batteryFromRest else (batteryFromText ?: 0),
            ridingState = ridingState,
            operationStates = operationState.orEmpty(),
            alarmStates = alarmState.orEmpty(),
            isOnline = isOnline == 1,
            serviceId = serviceId.orEmpty(),
            voltageMv = voltage,
            batterySn = batterySn.orEmpty(),
            batteryLock = batteryLock,
            backWheelLock = backWheelLock,
            helmetLock = helmetLock,
            helmetReact = helmetReact,
            helmetState = helmetState,
            helmetMac = helmetMac.orEmpty(),
            forParkName = forParkName.orEmpty(),
            noParkName = noParkName.orEmpty(),
            isOutOfServiceArea = isOutofServAera,
            isFenceEnable = isFenceEnable,
            totalMiles = totalMiles,
            serviceName = serviceName.orEmpty(),
            model = model.orEmpty(),
            izHaveOverload = izHaveOverload == true,
            acc = acc,
            defend = defend,
            timestamp = normalizeEpochMsOrNull(timestamp),
            lockTimeMs = normalizeEpochMs(lockTime).takeIf { lockTime > 0 } ?: lockTime,
            unlockTimeMs = normalizeEpochMs(unlockTime).takeIf { unlockTime > 0 } ?: unlockTime,
            version = version.orEmpty(),
            gsmSignal = gsmSignal,
            headingAngle = headingAngle ?: -1,
            helmetBind = helmetBind,
        )
    }
}

fun Vehicle.toMapPin(): MapPin = MapPin(
    id = carId,
    lat = lat,
    lng = lng,
    title = carId,
    subtitle = "$batteryLabel · $ridingLabel",
    restBattery = restBattery,
    ridingState = ridingState,
    memberCount = 1,
    memberIds = listOf(carId),
    icon = homeMapPinIcon(),
)

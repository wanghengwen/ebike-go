package com.luopingtech.ebike.ops.data.vehicle

import com.luopingtech.ebike.ops.data.analysis.FlexibleIntSerializer
import com.luopingtech.ebike.ops.data.analysis.FlexibleStringSerializer
import kotlinx.serialization.Serializable

@Serializable
data class VehicleChangeBatteryHistoryPageDto(
    val list: List<VehicleChangeBatteryHistoryDto> = emptyList(),
    val sum: Int? = null,
    val count: Int? = null,
)

@Serializable
data class VehicleMoveHistoryPageDto(
    val list: List<VehicleMoveHistoryDto> = emptyList(),
    val sum: Int? = null,
    val count: Int? = null,
)

@Serializable
data class VehicleScanLogDto(
    @Serializable(with = FlexibleStringSerializer::class)
    val carId: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val imei: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val phone: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val time: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val type: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val address: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val content: String = "",
    val latitude: Double? = null,
    val longitude: Double? = null,
)

@Serializable
data class VehicleScanLogContentDto(
    val userLat: Double? = null,
    val userLng: Double? = null,
)

@Serializable
data class VehicleSwitchLockDto(
    @Serializable(with = FlexibleStringSerializer::class)
    val pin: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val phone: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val name: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val imei: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val eventName: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val time: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val eventType: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val carId: String = "",
) {
    /** Legacy: 17514 = unlock, otherwise lock. */
    fun isLock(): Boolean = eventType != "17514"
}

@Serializable
data class VehicleChangeBatteryHistoryDto(
    @Serializable(with = FlexibleStringSerializer::class)
    val carId: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val changeBatteryLastTime: String = "",
    val restBatteryAfter: Int? = null,
    @Serializable(with = FlexibleStringSerializer::class)
    val openBatBoxTime: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val phone: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val opMan: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val closeBatBoxTime: String = "",
    @Serializable(with = FlexibleIntSerializer::class)
    val state: Int = 0,
    val checkResult: Int? = null,
    val restBatteryBefore: Int? = null,
)

@Serializable
data class VehicleMoveHistoryDto(
    @Serializable(with = FlexibleStringSerializer::class)
    val carId: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val operator: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val lastTime: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val phone: String = "",
    @Serializable(with = FlexibleStringSerializer::class)
    val endTime: String = "",
    val checkResult: Int? = null,
    @Serializable(with = FlexibleStringSerializer::class)
    val startTime: String = "",
)

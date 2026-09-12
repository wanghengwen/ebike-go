package com.luopingtech.ebike.ops.domain.model

/**
 * Row from tools/unlock_car_list (legacy VehicleUnLockedModel).
 */
data class UnlockedVehicle(
    val carId: String,
    val imei: String = "",
)

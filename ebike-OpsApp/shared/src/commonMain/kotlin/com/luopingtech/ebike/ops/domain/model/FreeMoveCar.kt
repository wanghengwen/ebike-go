package com.luopingtech.ebike.ops.domain.model

/**
 * In-progress free (self-service) move-car item — no task id.
 * Legacy: MoveVehicleBean / MultipleMoveVehicleModel on batch_list.
 */
data class FreeMoveCar(
    val carId: String,
    val imei: String = "",
    val restBattery: Int = 0,
    /** 0 开锁中 / 1 挪车中 等，遗留字段；列表展示用. */
    val state: Int = 1,
)

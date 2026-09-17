package com.luopingtech.ebike.ops.domain.model

/**
 * In-progress free (self-service) move-car item — no task id.
 * Legacy: MoveVehicleBean / MultipleMoveVehicleModel on batch_list.
 */
data class FreeMoveCar(
    val carId: String,
    val imei: String = "",
    val restBattery: Int = 0,
    /**
     * Riding / ops state. After unlock start, legacy forces [STATE_OPERATION]=5（运维中）.
     */
    val state: Int = 1,
    /** Legacy batch_list izFinish: false = 开锁列表，true = 已完成关锁列表. */
    val izFinish: Boolean = false,
) {
    companion object {
        const val STATE_OPERATION = 5
    }
}

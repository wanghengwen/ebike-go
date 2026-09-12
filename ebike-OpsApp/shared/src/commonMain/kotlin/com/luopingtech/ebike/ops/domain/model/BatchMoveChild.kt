package com.luopingtech.ebike.ops.domain.model

/**
 * Child row under a man-made multi-vehicle move parent (`man_made_move_list`).
 */
data class BatchMoveChild(
    val taskId: String,
    val carId: String,
    val imei: String = "",
    val restBattery: Int = 0,
    /** Child task state: 0 pending / 1 in progress / 2 finished. */
    val state: Int = 0,
    val izFinish: Boolean = false,
) {
    val isFinished: Boolean get() = izFinish || state == 2
}

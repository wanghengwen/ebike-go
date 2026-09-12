package com.luopingtech.ebike.ops.domain.model

/**
 * Legacy Flutter OverloadResult / paas SaddleOverloadContactCo.
 * Contact bit: 1 = pressed / active.
 */
data class SaddleOverloadContact(
    val frontSaddleContact: Int = 0,
    val centSaddleContact: Int = 0,
    val backSaddleContact: Int = 0,
) {
    val frontOn: Boolean get() = frontSaddleContact == 1
    val centerOn: Boolean get() = centSaddleContact == 1
    val backOn: Boolean get() = backSaddleContact == 1

    companion object {
        val Idle: SaddleOverloadContact = SaddleOverloadContact()
    }
}

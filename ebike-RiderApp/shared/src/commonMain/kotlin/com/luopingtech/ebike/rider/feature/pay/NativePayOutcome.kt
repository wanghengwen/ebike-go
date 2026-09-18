package com.luopingtech.ebike.rider.feature.pay

sealed interface NativePayOutcome {
    data object Paid : NativePayOutcome
    data object FallbackH5 : NativePayOutcome
    data object Cancelled : NativePayOutcome
    data class Failed(val message: String) : NativePayOutcome
}

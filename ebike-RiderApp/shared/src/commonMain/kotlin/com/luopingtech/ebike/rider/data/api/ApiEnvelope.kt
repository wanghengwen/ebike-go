package com.luopingtech.ebike.rider.data.api

import kotlinx.serialization.Serializable

@Serializable
data class ApiEnvelope<T>(
    val success: Boolean = false,
    val code: String = "",
    val msg: String? = null,
    val data: T? = null,
) {
    val isSuccessful: Boolean
        get() = success || code == "0" || code.equals("ok", ignoreCase = true) || code == "200"
}

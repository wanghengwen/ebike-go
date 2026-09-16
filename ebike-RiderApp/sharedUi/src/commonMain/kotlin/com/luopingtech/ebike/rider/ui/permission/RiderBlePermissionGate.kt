package com.luopingtech.ebike.rider.ui.permission

/** 宿主注入 BLE 运行时权限；null = 已授权可继续。 */
fun interface RiderBlePermissionGate {
    fun ensure(onResult: (String?) -> Unit)
}

val NoOpBlePermissionGate = RiderBlePermissionGate { onResult -> onResult(null) }

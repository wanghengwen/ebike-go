package com.luopingtech.ebike.rider.ui.permission

/** 宿主注入定位权限流程；null = 已授权可继续。 */
fun interface RiderLocationPermissionGate {
    fun ensure(onResult: (String?) -> Unit)
}

val NoOpLocationPermissionGate = RiderLocationPermissionGate { onResult -> onResult(null) }

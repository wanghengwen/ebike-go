package com.luopingtech.ebike.rider.ui.permission

import kotlinx.cinterop.ExperimentalForeignApi
import platform.CoreBluetooth.CBCentralManager
import platform.CoreBluetooth.CBManagerAuthorizationAllowedAlways
import platform.CoreBluetooth.CBManagerAuthorizationDenied
import platform.CoreBluetooth.CBManagerAuthorizationRestricted

/**
 * iOS 蓝牙许可。未决定时交给系统在首次 scan/connect 时弹窗
 *（文案来自 Info.plist `NSBluetoothAlwaysUsageDescription`）。
 */
@OptIn(ExperimentalForeignApi::class)
class IosBlePermissionGate(
    private val deniedMessage: String,
    private val allowWithoutPermission: Boolean = false,
) : RiderBlePermissionGate {
    override fun ensure(onResult: (String?) -> Unit) {
        if (allowWithoutPermission) {
            onResult(null)
            return
        }
        val auth = CBCentralManager.authorization
        when (auth) {
            CBManagerAuthorizationDenied,
            CBManagerAuthorizationRestricted,
            -> onResult(deniedMessage)
            CBManagerAuthorizationAllowedAlways -> onResult(null)
            else -> onResult(null)
        }
    }
}

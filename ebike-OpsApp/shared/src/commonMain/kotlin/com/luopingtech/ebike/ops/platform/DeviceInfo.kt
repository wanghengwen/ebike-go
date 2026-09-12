package com.luopingtech.ebike.ops.platform

/**
 * Device / app identity for analytics headers and diagnostics.
 */
interface DeviceInfo {
    val platform: String
    val osVersion: String
    val deviceModel: String
    val appVersion: String
}

expect fun createDeviceInfo(appVersion: String): DeviceInfo

package com.luopingtech.ebike.ops.platform

import android.os.Build

actual fun createDeviceInfo(appVersion: String): DeviceInfo = object : DeviceInfo {
    override val platform: String = "android"
    override val osVersion: String = Build.VERSION.RELEASE ?: ""
    override val deviceModel: String = listOf(Build.MANUFACTURER, Build.MODEL)
        .filter { it.isNotBlank() }
        .joinToString(" ")
    override val appVersion: String = appVersion
}

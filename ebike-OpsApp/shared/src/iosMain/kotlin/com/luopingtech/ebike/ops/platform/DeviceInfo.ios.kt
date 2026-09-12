package com.luopingtech.ebike.ops.platform

import platform.UIKit.UIDevice

actual fun createDeviceInfo(appVersion: String): DeviceInfo = object : DeviceInfo {
    override val platform: String = "ios"
    override val osVersion: String = UIDevice.currentDevice.systemVersion
    override val deviceModel: String = UIDevice.currentDevice.model
    override val appVersion: String = appVersion
}

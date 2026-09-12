package com.luopingtech.ebike.ops.domain.tools

data class OpsSettingConfig(
    val swapBatteryThreshold: Int = 20,
    val izAutoSwapBattery: Boolean = false,
)

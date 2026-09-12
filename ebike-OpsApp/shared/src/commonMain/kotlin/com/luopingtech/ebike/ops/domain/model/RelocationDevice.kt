package com.luopingtech.ebike.ops.domain.model

/**
 * Offline vehicle candidate for relocation (legacy Flutter DeviceInfoModel / deviceScan).
 */
data class RelocationDevice(
    val carId: String,
    val imei: String,
    val lat: Double = 0.0,
    val lng: Double = 0.0,
    val restBattery: Int = 0,
    /** 0 = offline (required by legacy RelocationBloc); non-zero = online. */
    val isOnline: Int = 0,
    val selected: Boolean = true,
) {
    val isOffline: Boolean get() = isOnline == 0
}

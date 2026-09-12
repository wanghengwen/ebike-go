package com.luopingtech.ebike.ops.platform

import com.luopingtech.ebike.ops.core.result.OpsResult

/** Reverse geocode lat/lng → human address (legacy SearchAddressHelp). */
interface ReverseGeocoder {
    suspend fun addressOf(lat: Double, lng: Double): OpsResult<String>
}

class UnsupportedReverseGeocoder : ReverseGeocoder {
    override suspend fun addressOf(lat: Double, lng: Double): OpsResult<String> =
        OpsResult.Err(com.luopingtech.ebike.ops.core.result.OpsError.unsupported("ReverseGeocoder"))
}

class DemoReverseGeocoder : ReverseGeocoder {
    override suspend fun addressOf(lat: Double, lng: Double): OpsResult<String> =
        OpsResult.Ok("Demo · ${lat.toFixed5()}, ${lng.toFixed5()}")
}

/** `String.format` 只有 JVM 有；经纬度定长五位小数自己拼。 */
private fun Double.toFixed5(): String {
    val scaled = kotlin.math.round(this * 100_000.0).toLong()
    val sign = if (scaled < 0) "-" else ""
    val abs = if (scaled < 0) -scaled else scaled
    return "$sign${abs / 100_000}.${(abs % 100_000).toString().padStart(5, '0')}"
}

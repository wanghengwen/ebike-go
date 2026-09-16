package com.luopingtech.ebike.rider.platform

import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult

interface ReverseGeocoder {
    suspend fun addressOf(lat: Double, lng: Double): RiderResult<String>
}

class UnsupportedReverseGeocoder : ReverseGeocoder {
    override suspend fun addressOf(lat: Double, lng: Double): RiderResult<String> =
        RiderResult.Err(RiderError.unsupported("ReverseGeocoder"))
}

class DemoReverseGeocoder : ReverseGeocoder {
    override suspend fun addressOf(lat: Double, lng: Double): RiderResult<String> =
        RiderResult.Ok("Demo · ${lat.toFixed5()}, ${lng.toFixed5()}")
}

private fun Double.toFixed5(): String {
    val scaled = kotlin.math.round(this * 100_000.0).toLong()
    val sign = if (scaled < 0) "-" else ""
    val abs = if (scaled < 0) -scaled else scaled
    return "$sign${abs / 100_000}.${(abs % 100_000).toString().padStart(5, '0')}"
}

package com.luopingtech.ebike.rider.platform

import android.content.Context
import android.location.Geocoder
import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import java.util.Locale
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

class AndroidReverseGeocoder(
    context: Context,
) : ReverseGeocoder {
    private val appContext = context.applicationContext

    override suspend fun addressOf(lat: Double, lng: Double): RiderResult<String> =
        withContext(Dispatchers.IO) {
            if (!Geocoder.isPresent()) {
                return@withContext RiderResult.Err(RiderError.unsupported("Geocoder"))
            }
            runCatching {
                @Suppress("DEPRECATION")
                val list = Geocoder(appContext, Locale.getDefault())
                    .getFromLocation(lat, lng, 1)
                val line = list?.firstOrNull()?.getAddressLine(0).orEmpty()
                if (line.isBlank()) {
                    RiderResult.Err(RiderError.business("GEO", "empty address"))
                } else {
                    RiderResult.Ok(line)
                }
            }.getOrElse {
                RiderResult.Err(RiderError.network(it.message ?: "geocode failed", it))
            }
        }
}

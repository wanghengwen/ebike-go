package com.luopingtech.ebike.ops.platform

import android.content.Context
import android.location.Geocoder
import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import java.util.Locale
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

class AndroidReverseGeocoder(
    context: Context,
) : ReverseGeocoder {
    private val appContext = context.applicationContext

    override suspend fun addressOf(lat: Double, lng: Double): OpsResult<String> =
        withContext(Dispatchers.IO) {
            if (!Geocoder.isPresent()) {
                return@withContext OpsResult.Err(OpsError.unsupported("Geocoder"))
            }
            runCatching {
                @Suppress("DEPRECATION")
                val list = Geocoder(appContext, Locale.getDefault())
                    .getFromLocation(lat, lng, 1)
                val line = list?.firstOrNull()?.getAddressLine(0).orEmpty()
                if (line.isBlank()) {
                    OpsResult.Err(OpsError.business("GEO", "empty address"))
                } else {
                    OpsResult.Ok(line)
                }
            }.getOrElse {
                OpsResult.Err(OpsError.network(it.message ?: "geocode failed", it))
            }
        }
}

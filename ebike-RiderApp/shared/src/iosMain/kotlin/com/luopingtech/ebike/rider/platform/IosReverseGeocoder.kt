package com.luopingtech.ebike.rider.platform

import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
import kotlin.coroutines.resume
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.coroutines.suspendCancellableCoroutine
import platform.CoreLocation.CLGeocoder
import platform.CoreLocation.CLLocation
import platform.CoreLocation.CLPlacemark

/**
 * `AndroidReverseGeocoder` 的 iOS 对应实现：CLGeocoder 反查地址。
 *
 * 只取一条结果的 `name`（Apple 给的单行门牌串），拼不出来时退回「行政区 + 街道」，
 * 与 Android 的 `getAddressLine(0)` 呈现粒度对齐。
 */
class IosReverseGeocoder : ReverseGeocoder {
    private val geocoder = CLGeocoder()

    @OptIn(ExperimentalForeignApi::class)
    override suspend fun addressOf(lat: Double, lng: Double): RiderResult<String> =
        suspendCancellableCoroutine { cont ->
            val location = CLLocation(latitude = lat, longitude = lng)
            geocoder.reverseGeocodeLocation(location) { placemarks, error ->
                if (!cont.isActive) return@reverseGeocodeLocation
                if (error != null) {
                    cont.resume(RiderResult.Err(RiderError.network(error.localizedDescription)))
                    return@reverseGeocodeLocation
                }
                val mark = placemarks?.firstOrNull() as? CLPlacemark
                val line = mark?.let(::describe).orEmpty()
                if (line.isBlank()) {
                    cont.resume(RiderResult.Err(RiderError.business("GEO", "empty address")))
                } else {
                    cont.resume(RiderResult.Ok(line))
                }
            }
            cont.invokeOnCancellation { geocoder.cancelGeocode() }
        }

    private fun describe(mark: CLPlacemark): String {
        mark.name?.takeIf { it.isNotBlank() }?.let { return it }
        return listOfNotNull(
            mark.locality,
            mark.subLocality,
            mark.thoroughfare,
            mark.subThoroughfare,
        ).filter { it.isNotBlank() }.joinToString("")
    }
}

package com.luopingtech.ebike.ops.platform

import com.luopingtech.ebike.ops.core.result.OpsError
import com.luopingtech.ebike.ops.core.result.OpsResult
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.cinterop.useContents
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.withTimeoutOrNull
import platform.CoreLocation.CLAuthorizationStatus
import platform.CoreLocation.CLLocation
import platform.CoreLocation.CLLocationManager
import platform.CoreLocation.CLLocationManagerDelegateProtocol
import platform.CoreLocation.kCLAuthorizationStatusAuthorizedAlways
import platform.CoreLocation.kCLAuthorizationStatusAuthorizedWhenInUse
import platform.CoreLocation.kCLAuthorizationStatusDenied
import platform.CoreLocation.kCLAuthorizationStatusNotDetermined
import platform.CoreLocation.kCLAuthorizationStatusRestricted
import platform.CoreLocation.kCLLocationAccuracyBest
import platform.Foundation.NSError
import platform.Foundation.NSThread
import platform.darwin.NSObject
import platform.darwin.dispatch_async
import platform.darwin.dispatch_get_main_queue

/**
 * CoreLocation 版定位，对应 Android 的 `LocationManager` 实现。
 *
 * 只申请 whenInUse：轨迹上报走前台 + `UIBackgroundModes.location`，
 * always 的系统弹窗对现场运维员劝退率太高，等真需要熄屏跑轨迹再升级。
 *
 * 授权与开关都切回主线程：CLLocationManager 的回调依赖创建它的 run loop，
 * 共享层的 feature 是在协程里调 [startTracking] 的，不切会静默不回调。
 */
@OptIn(ExperimentalForeignApi::class)
class CoreLocationTracker(
    private val distanceFilterMeters: Double = 10.0,
) : LocationTracker {

    private val updates = MutableSharedFlow<GeoPoint>(replay = 1, extraBufferCapacity = 64)

    private val locationDelegate = object : NSObject(), CLLocationManagerDelegateProtocol {
        override fun locationManager(manager: CLLocationManager, didUpdateLocations: List<*>) {
            val fix = didUpdateLocations.lastOrNull() as? CLLocation ?: return
            updates.tryEmit(fix.toGeoPoint())
        }

        override fun locationManager(manager: CLLocationManager, didFailWithError: NSError) {
            // 定位失败不做重试：上层拿不到点就不上报，下一个 fix 到了自然恢复。
        }
    }

    private val manager = CLLocationManager().apply {
        desiredAccuracy = kCLLocationAccuracyBest
        distanceFilter = distanceFilterMeters
        pausesLocationUpdatesAutomatically = false
        setDelegate(locationDelegate)
    }

    val authorizationStatus: CLAuthorizationStatus
        get() = CLLocationManager.authorizationStatus()

    val isAuthorized: Boolean
        get() = authorizationStatus == kCLAuthorizationStatusAuthorizedWhenInUse ||
            authorizationStatus == kCLAuthorizationStatusAuthorizedAlways

    fun requestAuthorization() = onMain { manager.requestWhenInUseAuthorization() }

    override suspend fun currentLocation(): OpsResult<GeoPoint> {
        manager.location?.let { return OpsResult.Ok(it.toGeoPoint()) }
        if (!CLLocationManager.locationServicesEnabled()) {
            return OpsResult.Err(OpsError.unsupported("Location services disabled"))
        }
        when (authorizationStatus) {
            kCLAuthorizationStatusDenied, kCLAuthorizationStatusRestricted ->
                return OpsResult.Err(OpsError.unsupported("Location permission denied"))
            kCLAuthorizationStatusNotDetermined -> requestAuthorization()
            else -> Unit
        }
        onMain { manager.requestLocation() }
        val fix = withTimeoutOrNull(REQUEST_TIMEOUT_MS) { updates.first() }
        return fix?.let { OpsResult.Ok(it) }
            ?: OpsResult.Err(OpsError.unsupported("Location unavailable"))
    }

    override fun track(): Flow<GeoPoint> = updates.asSharedFlow()

    override fun startTracking() = onMain {
        if (authorizationStatus == kCLAuthorizationStatusNotDetermined) {
            manager.requestWhenInUseAuthorization()
        }
        manager.startUpdatingLocation()
    }

    override fun stopTracking() = onMain { manager.stopUpdatingLocation() }

    private companion object {
        const val REQUEST_TIMEOUT_MS = 8_000L
    }
}

@OptIn(ExperimentalForeignApi::class)
private fun CLLocation.toGeoPoint(): GeoPoint =
    coordinate.useContents { GeoPoint(latitude = latitude, longitude = longitude) }

internal fun onMain(block: () -> Unit) {
    if (NSThread.isMainThread()) block() else dispatch_async(dispatch_get_main_queue(), block)
}

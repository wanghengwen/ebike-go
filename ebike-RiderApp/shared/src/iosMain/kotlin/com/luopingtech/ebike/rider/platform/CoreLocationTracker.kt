package com.luopingtech.ebike.rider.platform

import com.luopingtech.ebike.rider.core.result.RiderError
import com.luopingtech.ebike.rider.core.result.RiderResult
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
import platform.Foundation.NSBundle
import platform.Foundation.NSError
import platform.Foundation.NSThread
import platform.darwin.NSObject
import platform.darwin.dispatch_async
import platform.darwin.dispatch_get_main_queue

/**
 * CoreLocation 版定位，对应 Android 的 `LocationManager` 实现。
 *
 * 只申请 whenInUse：骑行轨迹靠 `UIBackgroundModes.location` +
 * [CLLocationManager.allowsBackgroundLocationUpdates] 拿后台更新（系统显示蓝色状态条），
 * 不需要 always —— always 的系统弹窗劝退率高得多，而且我们不需要熄屏后无限期定位。
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

    override suspend fun currentLocation(): RiderResult<GeoPoint> {
        manager.location?.let { return RiderResult.Ok(it.toGeoPoint()) }
        if (!CLLocationManager.locationServicesEnabled()) {
            return RiderResult.Err(RiderError.unsupported("Location services disabled"))
        }
        when (authorizationStatus) {
            kCLAuthorizationStatusDenied, kCLAuthorizationStatusRestricted ->
                return RiderResult.Err(RiderError.unsupported("Location permission denied"))
            kCLAuthorizationStatusNotDetermined -> requestAuthorization()
            else -> Unit
        }
        onMain { manager.requestLocation() }
        val fix = withTimeoutOrNull(REQUEST_TIMEOUT_MS) { updates.first() }
        return fix?.let { RiderResult.Ok(it) }
            ?: RiderResult.Err(RiderError.unsupported("Location unavailable"))
    }

    override fun track(): Flow<GeoPoint> = updates.asSharedFlow()

    override fun startTracking() = onMain {
        if (authorizationStatus == kCLAuthorizationStatusNotDetermined) {
            manager.requestWhenInUseAuthorization()
        }
        // 骑行中锁屏 / 切去导航时也要继续收点，否则计费时长和还车围栏判定会断档。
        // 必须在 startUpdatingLocation 之前置位，且只在 plist 声明了后台模式时置 ——
        // 没声明就置位会被 CoreLocation 直接抛异常。
        if (backgroundLocationDeclared) {
            manager.allowsBackgroundLocationUpdates = true
        }
        manager.startUpdatingLocation()
    }

    override fun stopTracking() = onMain {
        manager.stopUpdatingLocation()
        // 还车后立刻放掉后台定位权，别让蓝色状态条一直挂着。
        if (backgroundLocationDeclared) {
            manager.allowsBackgroundLocationUpdates = false
        }
    }

    private companion object {
        const val REQUEST_TIMEOUT_MS = 8_000L

        /** `UIBackgroundModes` 是否含 `location`，见 iosApp/project.yml。 */
        val backgroundLocationDeclared: Boolean by lazy {
            val modes = NSBundle.mainBundle.objectForInfoDictionaryKey("UIBackgroundModes")
            (modes as? List<*>)?.any { it as? String == "location" } == true
        }
    }
}

@OptIn(ExperimentalForeignApi::class)
private fun CLLocation.toGeoPoint(): GeoPoint =
    coordinate.useContents { GeoPoint(latitude = latitude, longitude = longitude) }

internal fun onMain(block: () -> Unit) {
    if (NSThread.isMainThread()) block() else dispatch_async(dispatch_get_main_queue(), block)
}

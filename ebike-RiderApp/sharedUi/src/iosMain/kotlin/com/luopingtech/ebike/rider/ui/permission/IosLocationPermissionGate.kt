package com.luopingtech.ebike.rider.ui.permission

import kotlinx.cinterop.ExperimentalForeignApi
import platform.CoreLocation.CLLocationManager
import platform.CoreLocation.CLLocationManagerDelegateProtocol
import platform.CoreLocation.kCLAuthorizationStatusAuthorizedAlways
import platform.CoreLocation.kCLAuthorizationStatusAuthorizedWhenInUse
import platform.CoreLocation.kCLAuthorizationStatusNotDetermined
import platform.darwin.NSObject

/**
 * iOS 的轨迹上报许可，对应 Android 的 `rememberTrackPermissionGate`。
 *
 * 两端的流程差别就在这里体现：Android 要分次弹、还要哄用户去开「始终允许」，
 * iOS 只有一次 whenInUse 弹窗，结果通过 `locationManagerDidChangeAuthorization` 回来。
 * 已授权时同步放行，不再弹窗。
 */
@OptIn(ExperimentalForeignApi::class)
class IosLocationPermissionGate(
    private val deniedMessage: String,
    private val allowWithoutPermission: Boolean = false,
) : RiderLocationPermissionGate {

    private var pending: ((String?) -> Unit)? = null

    private val authDelegate = object : NSObject(), CLLocationManagerDelegateProtocol {
        override fun locationManagerDidChangeAuthorization(manager: CLLocationManager) {
            val status = CLLocationManager.authorizationStatus()
            // 弹窗还在，用户没选：等下一次回调，不要提前判失败。
            if (status == kCLAuthorizationStatusNotDetermined) return
            val callback = pending ?: return
            pending = null
            callback(if (status.isAuthorized()) null else deniedMessage)
        }
    }

    private val manager = CLLocationManager().apply { setDelegate(authDelegate) }

    override fun ensure(onResult: (String?) -> Unit) {
        val status = CLLocationManager.authorizationStatus()
        when {
            status.isAuthorized() -> onResult(null)
            // demo 模式用的是模拟轨迹，没有系统定位也照样能演示上报。
            allowWithoutPermission -> onResult(null)
            status == kCLAuthorizationStatusNotDetermined -> {
                pending = onResult
                manager.requestWhenInUseAuthorization()
            }
            else -> onResult(deniedMessage)
        }
    }
}

private fun Int.isAuthorized(): Boolean =
    this == kCLAuthorizationStatusAuthorizedWhenInUse ||
        this == kCLAuthorizationStatusAuthorizedAlways

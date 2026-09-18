package com.luopingtech.ebike.ops

import android.app.ActivityManager
import android.os.Bundle
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Surface
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import com.luopingtech.ebike.ops.domain.model.Vehicle
import com.luopingtech.ebike.ops.domain.permission.OpsPermissions
import com.luopingtech.ebike.ops.platform.ActivityCodeScanner
import com.luopingtech.ebike.ops.platform.ActivityPhotoCapture
import com.luopingtech.ebike.ops.platform.AndroidLocationTracker
import com.luopingtech.ebike.ops.platform.TrackLocationService
import com.luopingtech.ebike.ops.platform.rememberTrackPermissionGate
import com.luopingtech.ebike.ops.scan.AndroidScanPreview
import com.luopingtech.ebike.ops.ui.feedback.LocalOpsToast
import com.luopingtech.ebike.ops.ui.map.LocalOpsMapRenderer
import com.luopingtech.ebike.ops.ui.map.opsMapRendererFor
import com.luopingtech.ebike.ops.ui.permission.LocalOpsTrackPermissionGate
import com.luopingtech.ebike.ops.ui.scan.LocalOpsScanPreview
import com.luopingtech.ebike.ops.ui.shell.OpsAppRoot
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import com.luopingtech.ebike.ops.ui.vehicle.LocalOpenVehicleDetail
import com.luopingtech.ebike.ops.ui.vehicle.OpenVehicleDetail
import com.luopingtech.ebike.ops.ui.vehicle.VehicleDetailScreen

/**
 * Android 宿主。界面一行都不在这里：登录、四个 Tab、扫码屏全在 :sharedUi 的 [OpsAppRoot]，
 * 这个 Activity 只负责三件事——绑定需要 Activity 的系统能力（扫码页、拍照）、
 * 把四种平台能力插进 CompositionLocal、然后让路。
 *
 * 保活 FGS 对齐原版 SwipeBackBaseActivity：退后台启、回前台停。
 *
 * iOS 宿主是同一份清单的另一种写法，见 iosApp/OpsAppHost。
 */
class MainActivity : ComponentActivity() {
    private lateinit var activityCodeScanner: ActivityCodeScanner
    private lateinit var activityPhotoCapture: ActivityPhotoCapture

    /** 对齐原版 isCurrentRunningForeground：记录上次是否在前台。 */
    private var wasRunningForeground = true

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val host = application as OpsApplication
        val app = host.opsApp
        activityCodeScanner = ActivityCodeScanner(this)
        host.codeScannerBridge.bind(activityCodeScanner)
        activityPhotoCapture = ActivityPhotoCapture(this)
        host.photoCaptureBridge.bind(activityPhotoCapture)
        setContent {
            OpsTheme(branding = app.config.branding) {
                val homeState by app.homeFeature.state.collectAsState()
                var detailVehicle by remember { mutableStateOf<Vehicle?>(null) }
                var detailServiceAreaId by remember { mutableStateOf<String?>(null) }
                val permissions = remember(homeState.session?.permissionCodes, app.isDemoMode) {
                    if (app.isDemoMode) OpsPermissions.demoFull()
                    else OpsPermissions.fromCodes(homeState.session?.permissionCodes.orEmpty())
                }
                CompositionLocalProvider(
                    LocalOpsMapRenderer provides opsMapRendererFor(app),
                    LocalOpsScanPreview provides AndroidScanPreview,
                    LocalOpsTrackPermissionGate provides rememberTrackPermissionGate(app),
                    LocalOpsToast provides { message: String ->
                        Toast.makeText(this, message, Toast.LENGTH_SHORT).show()
                    },
                    LocalOpenVehicleDetail provides OpenVehicleDetail { vehicle, serviceAreaId ->
                        detailVehicle = vehicle
                        detailServiceAreaId = serviceAreaId
                    },
                ) {
                    Box(modifier = Modifier.fillMaxSize()) {
                        Surface(modifier = Modifier.fillMaxSize()) {
                            OpsAppRoot(app)
                        }
                        detailVehicle?.let { vehicle ->
                            VehicleDetailScreen(
                                app = app,
                                vehicle = vehicle,
                                serviceAreaId = detailServiceAreaId,
                                permissions = permissions,
                                mapReady = homeState.mapReady,
                                mapProviderKind = homeState.mapProviderKind,
                                onClose = {
                                    detailVehicle = null
                                    detailServiceAreaId = null
                                },
                                modifier = Modifier.fillMaxSize(),
                            )
                        }
                    }
                }
            }
        }
    }

    // 对齐 SwipeBackBaseActivity：退后台启保活
    override fun onStop() {
        super.onStop()
        wasRunningForeground = isRunningForeground()
        if (!wasRunningForeground && shouldKeepAlive()) {
            TrackLocationService.start(this)
        }
    }

    // 对齐 SwipeBackBaseActivity：回前台停保活
    override fun onResume() {
        super.onResume()
        if (!wasRunningForeground) {
            TrackLocationService.stop(this)
        }
        wasRunningForeground = true
    }

    override fun onDestroy() {
        val host = application as OpsApplication
        host.codeScannerBridge.unbind(activityCodeScanner)
        host.photoCaptureBridge.unbind(activityPhotoCapture)
        super.onDestroy()
    }

    /** 原版只要进后台就启保活；我们仅在已开轨迹采点时启（location 型 FGS 更严）。 */
    private fun shouldKeepAlive(): Boolean {
        val host = application as? OpsApplication ?: return false
        if (!host.isOpsAppReady) return false
        if (host.opsApp.isDemoMode) return false
        val tracker = host.opsApp.locationTracker as? AndroidLocationTracker ?: return false
        return tracker.trackingActive && tracker.hasPermission()
    }

    private fun isRunningForeground(): Boolean {
        val activityManager = getSystemService(ACTIVITY_SERVICE) as ActivityManager
        val processes = activityManager.runningAppProcesses ?: return false
        val myName = applicationInfo.processName
        return processes.any {
            it.importance == ActivityManager.RunningAppProcessInfo.IMPORTANCE_FOREGROUND &&
                it.processName == myName
        }
    }
}

package com.luopingtech.ebike.rider.ui.shell

import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Surface
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.window.ComposeUIViewController
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.ui.feedback.IosToastHost
import com.luopingtech.ebike.rider.ui.map.LocalRiderMapRenderer
import com.luopingtech.ebike.rider.ui.map.riderMapRendererForIos
import com.luopingtech.ebike.rider.ui.permission.IosBlePermissionGate
import com.luopingtech.ebike.rider.ui.permission.IosLocationPermissionGate
import com.luopingtech.ebike.rider.ui.permission.LocalRiderBlePermissionGate
import com.luopingtech.ebike.rider.ui.permission.LocalRiderLocationPermissionGate
import com.luopingtech.ebike.rider.ui.scan.IosScanPreview
import com.luopingtech.ebike.rider.ui.scan.LocalRiderScanPreview
import com.luopingtech.ebike.rider.ui.theme.RiderTheme
import platform.UIKit.UIViewController

fun RiderAppViewController(app: RiderApp): UIViewController = ComposeUIViewController {
    RiderTheme(branding = app.config.branding) {
        val mapRenderer = remember(app) { riderMapRendererForIos(app) }
        val permissionGate = remember(app) {
            IosLocationPermissionGate(
                deniedMessage = app.i18n.t(Str.LocationPermissionRequired),
                allowWithoutPermission = app.isDemoMode,
            )
        }
        val blePermissionGate = remember(app) {
            IosBlePermissionGate(
                deniedMessage = app.i18n.t(Str.BlePermissionRequired),
                allowWithoutPermission = app.isDemoMode,
            )
        }
        CompositionLocalProvider(
            LocalRiderMapRenderer provides mapRenderer,
            LocalRiderScanPreview provides IosScanPreview,
            LocalRiderLocationPermissionGate provides permissionGate,
            LocalRiderBlePermissionGate provides blePermissionGate,
        ) {
            IosToastHost {
                Surface(modifier = Modifier.fillMaxSize()) {
                    RiderAppRoot(app)
                }
            }
        }
    }
}

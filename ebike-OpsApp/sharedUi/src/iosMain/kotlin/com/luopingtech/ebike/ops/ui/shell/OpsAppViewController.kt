package com.luopingtech.ebike.ops.ui.shell

import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Surface
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.window.ComposeUIViewController
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.i18n.Str
import com.luopingtech.ebike.ops.ui.feedback.IosToastHost
import com.luopingtech.ebike.ops.ui.map.LocalOpsMapRenderer
import com.luopingtech.ebike.ops.ui.map.opsMapRendererForIos
import com.luopingtech.ebike.ops.ui.permission.IosTrackPermissionGate
import com.luopingtech.ebike.ops.ui.permission.LocalOpsTrackPermissionGate
import com.luopingtech.ebike.ops.ui.scan.IosScanPreview
import com.luopingtech.ebike.ops.ui.scan.LocalOpsScanPreview
import com.luopingtech.ebike.ops.ui.theme.OpsTheme
import platform.UIKit.UIViewController

/**
 * iOS 宿主的入口，跟 Android 的 `MainActivity.setContent` 是同一件事：
 * 套上主题、把平台能力插进 CompositionLocal，然后交给共享层的 [OpsAppRoot]。
 *
 * 这份清单跟 `MainActivity` 一一对应：
 * 地图（腾讯就绪则用宿主注入，否则 MapKit）、取景框（AVFoundation）、
 * 定位许可（CoreLocation）、轻提示（Compose 浮层）。
 * 相机 / 相册 / Keychain / 定位这些不经过 Compose 的能力在 `createIosOpsApp` 里装。
 */
fun OpsAppViewController(app: OpsApp): UIViewController = ComposeUIViewController {
    OpsTheme(branding = app.config.branding) {
        val mapRenderer = remember(app) { opsMapRendererForIos(app) }
        val permissionGate = remember(app) {
            IosTrackPermissionGate(
                deniedMessage = app.i18n.t(Str.TrackNeedLocation),
                allowWithoutPermission = app.isDemoMode,
            )
        }
        CompositionLocalProvider(
            LocalOpsMapRenderer provides mapRenderer,
            LocalOpsScanPreview provides IosScanPreview,
            LocalOpsTrackPermissionGate provides permissionGate,
        ) {
            // Toast 的 provider 在 host 内部，所以它要包住真正的界面。
            IosToastHost {
                Surface(modifier = Modifier.fillMaxSize()) {
                    OpsAppRoot(app)
                }
            }
        }
    }
}

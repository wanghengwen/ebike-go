package com.luopingtech.ebike.ops

import android.os.Bundle
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Surface
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.ui.Modifier
import com.luopingtech.ebike.ops.platform.ActivityCodeScanner
import com.luopingtech.ebike.ops.platform.ActivityPhotoCapture
import com.luopingtech.ebike.ops.platform.rememberTrackPermissionGate
import com.luopingtech.ebike.ops.scan.AndroidScanPreview
import com.luopingtech.ebike.ops.ui.feedback.LocalOpsToast
import com.luopingtech.ebike.ops.ui.map.LocalOpsMapRenderer
import com.luopingtech.ebike.ops.ui.map.opsMapRendererFor
import com.luopingtech.ebike.ops.ui.permission.LocalOpsTrackPermissionGate
import com.luopingtech.ebike.ops.ui.scan.LocalOpsScanPreview
import com.luopingtech.ebike.ops.ui.shell.OpsAppRoot
import com.luopingtech.ebike.ops.ui.theme.OpsTheme

/**
 * Android 宿主。界面一行都不在这里：登录、四个 Tab、扫码屏全在 :sharedUi 的 [OpsAppRoot]，
 * 这个 Activity 只负责三件事——绑定需要 Activity 的系统能力（扫码页、拍照）、
 * 把四种平台能力插进 CompositionLocal、然后让路。
 *
 * iOS 宿主是同一份清单的另一种写法，见 iosApp/OpsAppHost。
 */
class MainActivity : ComponentActivity() {
    private lateinit var activityCodeScanner: ActivityCodeScanner
    private lateinit var activityPhotoCapture: ActivityPhotoCapture

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
                CompositionLocalProvider(
                    LocalOpsMapRenderer provides opsMapRendererFor(app),
                    LocalOpsScanPreview provides AndroidScanPreview,
                    LocalOpsTrackPermissionGate provides rememberTrackPermissionGate(app),
                    LocalOpsToast provides { message: String ->
                        Toast.makeText(this, message, Toast.LENGTH_SHORT).show()
                    },
                ) {
                    Surface(modifier = Modifier.fillMaxSize()) {
                        OpsAppRoot(app)
                    }
                }
            }
        }
    }

    override fun onDestroy() {
        val host = application as OpsApplication
        host.codeScannerBridge.unbind(activityCodeScanner)
        host.photoCaptureBridge.unbind(activityPhotoCapture)
        super.onDestroy()
    }
}

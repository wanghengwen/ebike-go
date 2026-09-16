package com.luopingtech.ebike.rider.ui.scan

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.UIKitView
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.i18n.Strings
import com.luopingtech.ebike.rider.platform.IosScanSession
import kotlinx.cinterop.BetaInteropApi
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.cinterop.readValue
import platform.AVFoundation.AVCaptureSession
import platform.AVFoundation.AVCaptureVideoPreviewLayer
import platform.AVFoundation.AVLayerVideoGravityResizeAspectFill
import platform.CoreGraphics.CGRectZero
import platform.UIKit.UIColor
import platform.UIKit.UIView

/**
 * iOS 的取景框，对应 Android 的 `AndroidScanPreview`（CameraX + ML Kit）。
 *
 * 相机权限在这里就地申请：扫码屏一露出就要出画面，把授权推给宿主反而会多一次黑屏。
 * 拒绝授权时明说原因，而不是留一块黑屏让人以为扫不出来。
 */
@OptIn(ExperimentalForeignApi::class)
object IosScanPreview : RiderScanPreview {

    @Composable
    override fun Preview(
        modifier: Modifier,
        torchOn: Boolean,
        enabled: Boolean,
        onCode: (String) -> Unit,
    ) {
        var granted by remember { mutableStateOf<Boolean?>(null) }
        LaunchedEffect(Unit) { granted = IosScanSession.requestCameraAccess() }

        if (granted != true) {
            Box(
                modifier = modifier.background(Color(0xFF1B1F2A)),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text = if (granted == null) {
                        Strings.t(Str.Loading)
                    } else {
                        Strings.t(Str.CameraPermissionRequired)
                    },
                    color = Color(0xFFCCCCCC),
                    style = MaterialTheme.typography.bodyMedium,
                    textAlign = TextAlign.Center,
                    modifier = Modifier.padding(24.dp),
                )
            }
            return
        }

        val scanSession = remember { IosScanSession() }
        val previewView = remember(scanSession) { CameraPreviewView(scanSession.session) }

        DisposableEffect(scanSession) {
            scanSession.start()
            onDispose {
                scanSession.onCode = null
                scanSession.stop()
            }
        }

        // enabled=false 时摘掉回调而不是停会话：画面留着，但不再往上吐码，
        // 否则手输框一弹相机重启，回来又是一次黑屏。
        DisposableEffect(enabled, onCode, scanSession) {
            scanSession.onCode = if (enabled) onCode else null
            onDispose { scanSession.onCode = null }
        }

        LaunchedEffect(torchOn, scanSession) { scanSession.setTorch(torchOn) }

        UIKitView(
            factory = { previewView },
            modifier = modifier,
        )
    }
}

@OptIn(ExperimentalForeignApi::class, BetaInteropApi::class)
private class CameraPreviewView(
    session: AVCaptureSession,
) : UIView(frame = CGRectZero.readValue()) {

    private val previewLayer = AVCaptureVideoPreviewLayer(session = session)

    init {
        setBackgroundColor(UIColor.blackColor)
        previewLayer.setVideoGravity(AVLayerVideoGravityResizeAspectFill)
        layer.addSublayer(previewLayer)
    }

    override fun layoutSubviews() {
        super.layoutSubviews()
        // 预览层不参与 Auto Layout，尺寸只能跟着宿主 view 的 bounds 走。
        previewLayer.setFrame(bounds)
    }
}

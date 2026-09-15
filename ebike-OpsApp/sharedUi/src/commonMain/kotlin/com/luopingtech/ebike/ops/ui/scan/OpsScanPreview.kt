package com.luopingtech.ebike.ops.ui.scan

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp

/**
 * 相机取景框。Android 是 CameraX + ML Kit，iOS 是 AVCaptureSession，
 * 共享界面只需要「一块会吐出码的画面」：给它尺寸和手电筒开关，它回码。
 *
 * [enabled] 为 false 时实现方要停掉解码（弹了手输框、正在提交），否则会一直重复回调。
 */
interface OpsScanPreview {
    @Composable
    fun Preview(
        modifier: Modifier,
        torchOn: Boolean,
        enabled: Boolean,
        onCode: (String) -> Unit,
    )
}

/** 宿主没接相机时的占位：明说这台设备扫不了，而不是黑屏假装在扫。 */
object UnavailableScanPreview : OpsScanPreview {
    @Composable
    override fun Preview(
        modifier: Modifier,
        torchOn: Boolean,
        enabled: Boolean,
        onCode: (String) -> Unit,
    ) {
        Box(
            modifier = modifier.background(Color(0xFF1B1F2A)),
            contentAlignment = Alignment.Center,
        ) {
            Text(
                text = "Camera scanner is not wired on this host yet.",
                color = Color(0xFFCCCCCC),
                style = MaterialTheme.typography.bodyMedium,
                textAlign = TextAlign.Center,
                modifier = Modifier.padding(24.dp),
            )
        }
    }
}

val LocalOpsScanPreview = staticCompositionLocalOf<OpsScanPreview> { UnavailableScanPreview }

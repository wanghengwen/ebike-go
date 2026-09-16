package com.luopingtech.ebike.rider.scan

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import com.luopingtech.ebike.rider.ui.scan.RiderScanPreview

/** 共享界面要的「会吐码的画面」，在 Android 上就是 CameraX + ML Kit 那块预览。 */
object AndroidScanPreview : RiderScanPreview {
    @Composable
    override fun Preview(
        modifier: Modifier,
        torchOn: Boolean,
        enabled: Boolean,
        onCode: (String) -> Unit,
    ) {
        RiderCameraScanPreview(
            onCode = onCode,
            modifier = modifier,
            torchOn = torchOn,
            enabled = enabled,
        )
    }
}

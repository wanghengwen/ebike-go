package com.luopingtech.ebike.rider.ui.platform

import androidx.compose.runtime.Composable

@Composable
actual fun RiderBackHandler(enabled: Boolean, onBack: () -> Unit) {
    // iOS 无系统返回键；侧滑由各屏自己处理（如 WKWebView）。
}

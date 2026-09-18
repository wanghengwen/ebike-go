package com.luopingtech.ebike.ops.ui.navigation

import androidx.compose.runtime.Composable

@Composable
actual fun OpsBackHandler(enabled: Boolean, onBack: () -> Unit) {
    // iOS 无 Activity 系统返回键；页面关闭走顶栏 onBack。
}

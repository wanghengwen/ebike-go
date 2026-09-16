package com.luopingtech.ebike.rider.ui.platform

import androidx.compose.runtime.Composable

/**
 * 拦截系统返回（Android 手势/导航键）。iOS 无系统返回键，actual 为空。
 */
@Composable
expect fun RiderBackHandler(enabled: Boolean = true, onBack: () -> Unit)

package com.luopingtech.ebike.ops.ui.navigation

import androidx.compose.runtime.Composable

/**
 * 拦截系统返回（对齐遗留 Activity.finish()：只关当前页，不结束宿主）。
 * Android 用 Activity BackHandler；iOS 无系统返回键，为空实现。
 */
@Composable
expect fun OpsBackHandler(enabled: Boolean = true, onBack: () -> Unit)

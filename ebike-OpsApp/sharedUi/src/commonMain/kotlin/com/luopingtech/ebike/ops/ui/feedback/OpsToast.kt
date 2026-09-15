package com.luopingtech.ebike.ops.ui.feedback

import androidx.compose.runtime.staticCompositionLocalOf

/**
 * 一次性轻提示。Android 宿主接 Toast，iOS 宿主接自己的 HUD / Snackbar。
 * 共享界面只管说"提示这句话"，不关心用什么控件弹。
 */
val LocalOpsToast = staticCompositionLocalOf<(String) -> Unit> { {} }

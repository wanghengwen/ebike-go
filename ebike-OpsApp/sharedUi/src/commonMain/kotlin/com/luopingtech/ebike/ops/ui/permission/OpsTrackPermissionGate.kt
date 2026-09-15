package com.luopingtech.ebike.ops.ui.permission

import androidx.compose.runtime.staticCompositionLocalOf

/**
 * 轨迹上报的权限闸门。
 *
 * 定位、通知、后台定位这三件事在两端的申请流程差得很远（Android 要分次弹、
 * 还要哄用户去开「始终允许」；iOS 是 always/whenInUse 两档 + 一次 plist 文案），
 * 所以共享界面只问一句「现在能开始上报吗」：
 * [onResult] 收到 null 就开上报，收到文案就把它显示成提示。
 */
fun interface OpsTrackPermissionGate {
    fun request(onResult: (String?) -> Unit)
}

/** 宿主没接权限流程时直接放行：定位拿不到自然不会有点上报，不需要在这里假装拦一下。 */
val LocalOpsTrackPermissionGate = staticCompositionLocalOf {
    OpsTrackPermissionGate { onResult -> onResult(null) }
}

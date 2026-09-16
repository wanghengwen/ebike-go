package com.luopingtech.ebike.rider.ui.shell

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.size
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.zIndex
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.config.H5ScreenKind
import com.luopingtech.ebike.rider.ui.auth.PhoneLoginScreen
import com.luopingtech.ebike.rider.ui.h5.H5HostScreen
import com.luopingtech.ebike.rider.ui.ride.RideFlowHost

private data class H5Nav(
    val kind: H5ScreenKind? = null,
    val hashRoute: String? = null,
)

/**
 * 整个 APP 的界面入口：未登录 → 手机号登录；已登录 → 骑行流程（`Idle` 相位就是首页地图）。
 *
 * H5 策略（省内存 + 回首页不闪）：
 * - 进首页不创建 WebView
 * - 第一次打开长尾页才挂 [H5HostScreen]
 * - 「关闭 / 回首页」只隐藏，不拆掉容器，首页一直在底下
 * - 再次打开同一容器，按新的 kind/hash 跳转
 */
@Composable
fun RiderAppRoot(app: RiderApp) {
    val authState by app.authFeature.state.collectAsState()
    var h5Nav by remember { mutableStateOf<H5Nav?>(null) }
    var h5Created by remember { mutableStateOf(false) }
    var h5Visible by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) {
        app.start(this)
    }

    // 登出后拆掉 H5，下次登录再懒创建
    LaunchedEffect(authState.session == null, authState.needLogin) {
        if (authState.needLogin || authState.session == null) {
            h5Created = false
            h5Visible = false
            h5Nav = null
        }
    }

    fun openH5(kind: H5ScreenKind?, hash: String?) {
        h5Nav = H5Nav(kind = kind, hashRoute = hash)
        h5Created = true
        h5Visible = true
    }

    fun hideH5() {
        h5Visible = false
    }

    Surface(modifier = Modifier.fillMaxSize()) {
        if (authState.needLogin || authState.session == null) {
            PhoneLoginScreen(app)
        } else {
            Box(modifier = Modifier.fillMaxSize()) {
                RideFlowHost(
                    app = app,
                    onOpenH5 = { kind, hash -> openH5(kind, hash) },
                )
                if (h5Created) {
                    val nav = h5Nav
                    // 可见时铺满拦截触摸；隐藏时 0×0 仍留在组合树里，WebView 不销毁
                    Surface(
                        modifier = Modifier
                            .zIndex(if (h5Visible) 2f else 0f)
                            .then(
                                if (h5Visible) Modifier.fillMaxSize()
                                else Modifier.size(0.dp),
                            ),
                    ) {
                        if (nav != null) {
                            H5HostScreen(
                                app = app,
                                kind = nav.kind,
                                hashRoute = nav.hashRoute,
                                onClose = { hideH5() },
                            )
                        }
                    }
                }
            }
        }
    }
}

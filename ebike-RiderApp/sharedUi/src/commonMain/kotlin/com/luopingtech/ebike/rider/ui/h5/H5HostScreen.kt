package com.luopingtech.ebike.rider.ui.h5

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.config.H5ScreenKind
import com.luopingtech.ebike.rider.core.config.H5ScreenUrls
import com.luopingtech.ebike.rider.core.h5.NativeHostBridge
import com.luopingtech.ebike.rider.core.h5.NativeHostHooks
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.ui.feedback.LocalRiderToast
import kotlinx.coroutines.launch

/**
 * 从首页 / Debug 打开的 H5 长尾容器。demo 未配 `h5.baseUrl` 时只显示说明，不加载 WebView。
 */
@Composable
fun H5HostScreen(
    app: RiderApp,
    kind: H5ScreenKind? = null,
    hashRoute: String? = null,
    onClose: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val language by app.i18n.languageFlow.collectAsState()
    val toast = LocalRiderToast.current
    val scope = rememberCoroutineScope()
    var overrideHash by remember { mutableStateOf<String?>(null) }

    // 父级再次 openH5（换 kind/hash）时清掉内部覆盖路由，按新入口加载
    LaunchedEffect(kind, hashRoute) {
        overrideHash = null
    }

    val url = remember(kind, hashRoute, overrideHash, language, app.config.h5.baseUrl) {
        val route = overrideHash ?: hashRoute
        when {
            route != null -> app.resolveH5Hash(route)
            kind != null -> app.resolveH5Url(kind)
            else -> null
        }
    }
    val defaultTitle = kind?.title(language)
        ?: hashRoute?.let { H5ScreenUrls.normalizeHash(it).removePrefix("#") }
        ?: app.i18n.t(Str.AppName)
    var title by remember(defaultTitle) { mutableStateOf(defaultTitle) }

    val hooks = remember {
        NativeHostHooks(
            close = onClose,
            setTitle = { title = it },
            toast = { toast(it) },
            navigate = { type, navUrl ->
                if (type == "back") {
                    onClose()
                } else if (!navUrl.isNullOrBlank() && !NativeHostBridge.isLoginNavigation(navUrl)) {
                    overrideHash = navUrl
                }
            },
            onLoginRequired = {
                scope.launch { app.authFeature.logout() }
                onClose()
            },
        )
    }
    hooks.close = onClose
    hooks.setTitle = { title = it }
    hooks.toast = { toast(it) }
    hooks.onLoginRequired = {
        scope.launch { app.authFeature.logout() }
        onClose()
    }

    val bridge = remember(app) { NativeHostBridge(app, hooks) }

    LaunchedEffect(language) {
        if (title == defaultTitle || title.isBlank()) {
            title = defaultTitle
        }
    }

    H5Screen(
        title = title,
        url = url,
        placeholder = app.i18n.t(Str.H5NotConfigured),
        failText = app.i18n.t(Str.H5LoadFailed),
        retryText = app.i18n.t(Str.Retry),
        closeText = app.i18n.t(Str.Close),
        backText = app.i18n.t(Str.GoBack),
        bridge = bridge,
        onClose = onClose,
        modifier = modifier,
    )
}

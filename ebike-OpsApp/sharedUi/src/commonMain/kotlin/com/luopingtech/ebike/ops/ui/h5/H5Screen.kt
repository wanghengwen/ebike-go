package com.luopingtech.ebike.ops.ui.h5

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.Stable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.ops.OpsApp
import com.luopingtech.ebike.ops.core.config.H5ScreenKind
import com.luopingtech.ebike.ops.core.i18n.Str

/**
 * WebView 的跨端状态。加载与失败由平台实现回写，翻页与重载由界面下发。
 *
 * 之所以不直接用第三方的 compose-webview-multiplatform：那个库最后一次发布停在
 * Compose 1.8 / Kotlin 2.1.20，落后我们三个 CMP 小版本，而管理类功能全走 H5，
 * 不能让它卡住升级。真要换回去，只需替换 [PlatformWebView] 的 iOS actual。
 */
@Stable
class WebViewHost {
    var loading: Boolean by mutableStateOf(true)
        internal set

    var failed: Boolean by mutableStateOf(false)
        internal set

    var canGoBack: Boolean by mutableStateOf(false)
        internal set

    /** 平台实现挂载时装配，卸载时置空。 */
    internal var controller: WebViewController? = null

    fun reload() {
        failed = false
        controller?.reload()
    }

    fun goBack() {
        controller?.goBack()
    }
}

internal interface WebViewController {
    fun reload()
    fun goBack()
}

/**
 * 平台 WebView 本体：Android 是 `android.webkit.WebView`，iOS 是 `WKWebView`。
 *
 * 系统返回键 / 侧滑返回的语义两端不同，交给各自实现；退无可退时调 [onNavigateOut]。
 */
@Composable
expect fun PlatformWebView(
    url: String,
    host: WebViewHost,
    onNavigateOut: () -> Unit,
    modifier: Modifier = Modifier,
)

/**
 * H5 屏外壳：标题栏、加载指示、失败重试。这部分两端一致，只有 WebView 本体分平台。
 */
@Composable
fun H5Screen(
    app: OpsApp,
    kind: H5ScreenKind,
    onClose: () -> Unit,
) {
    val language by app.i18n.languageFlow.collectAsState()
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)
    val title = t(kind.titleKey)
    val url = remember(kind, language) { app.resolveH5ScreenUrl(kind) }
    val host = remember { WebViewHost() }

    Column(modifier = Modifier.fillMaxSize()) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 8.dp, vertical = 4.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(title, style = MaterialTheme.typography.titleMedium)
            TextButton(onClick = onClose) { Text(t(Str.Close)) }
        }
        if (url.isNullOrBlank()) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Text(t(Str.H5LoadFailed), color = MaterialTheme.colorScheme.error)
            }
            return
        }
        Box(modifier = Modifier.fillMaxSize()) {
            PlatformWebView(
                url = url,
                host = host,
                onNavigateOut = onClose,
                modifier = Modifier.fillMaxSize(),
            )
            if (host.loading) {
                CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
            }
            if (host.failed) {
                Column(
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(16.dp),
                    horizontalAlignment = Alignment.CenterHorizontally,
                ) {
                    Text(t(Str.H5LoadFailed), color = MaterialTheme.colorScheme.error)
                    TextButton(onClick = { host.reload() }) { Text(t(Str.Retry)) }
                }
            }
        }
    }
}

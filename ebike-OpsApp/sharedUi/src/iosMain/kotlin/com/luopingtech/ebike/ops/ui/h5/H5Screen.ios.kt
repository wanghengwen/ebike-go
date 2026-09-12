package com.luopingtech.ebike.ops.ui.h5

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier

/**
 * TODO(iOS)：接 `WKWebView`。
 *
 * 计划是 `UIKitView` 包一个 WKWebView，用 `WKNavigationDelegate` 的
 * didStartProvisionalNavigation / didFinishNavigation / didFailProvisionalNavigation
 * 回写 [WebViewHost] 的 loading 与 failed，估计一百多行。
 *
 * 现在留占位而不是直接写，是因为手上没有 macOS，Kotlin/Native 的 ObjC 协议 interop
 * 写出来也编译不了、更跑不了 —— 未经验证却看着像对的代码，比一个明显的窟窿更危险。
 * 拿到 Mac 后第一件事就是补这里。备选是 io.github.kevinnzou:compose-webview-multiplatform，
 * 但它落后我们三个 CMP 小版本，只在自研受阻时才考虑。
 */
@Composable
actual fun PlatformWebView(
    url: String,
    host: WebViewHost,
    onNavigateOut: () -> Unit,
    modifier: Modifier,
) {
    host.loading = false
    host.failed = true
    Box(modifier = modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Text(
            text = "WebView is not implemented on iOS yet.",
            color = MaterialTheme.colorScheme.error,
            style = MaterialTheme.typography.bodyMedium,
        )
    }
}

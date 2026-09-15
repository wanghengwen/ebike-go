package com.luopingtech.ebike.ops.ui.h5

import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.viewinterop.UIKitView
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.cinterop.ObjCSignatureOverride
import kotlinx.cinterop.readValue
import platform.CoreGraphics.CGRectZero
import platform.Foundation.NSError
import platform.Foundation.NSURL
import platform.Foundation.NSURLRequest
import platform.WebKit.WKNavigation
import platform.WebKit.WKNavigationDelegateProtocol
import platform.WebKit.WKWebView
import platform.WebKit.WKWebViewConfiguration
import platform.darwin.NSObject

/**
 * iOS 的 H5 屏本体，对应 Android 的 `android.webkit.WebView`。
 *
 * 侧滑返回交给 `allowsBackForwardNavigationGestures`：管理类 H5 是多页流程，
 * 手势应该退 H5 的一页，跟 Android 那边拦系统返回键是同一个语义；
 * 退无可退时由外壳的关闭按钮走 [onNavigateOut]，不在这里劫持手势。
 */
@OptIn(ExperimentalForeignApi::class)
@Composable
actual fun PlatformWebView(
    url: String,
    host: WebViewHost,
    onNavigateOut: () -> Unit,
    modifier: Modifier,
) {
    val navigationDelegate = remember { WebNavigationDelegate(host) }
    val webView = remember {
        val configuration = WKWebViewConfiguration().apply {
            allowsInlineMediaPlayback = true
        }
        WKWebView(frame = CGRectZero.readValue(), configuration = configuration).apply {
            setNavigationDelegate(navigationDelegate)
            allowsBackForwardNavigationGestures = true
            setOpaque(false)
        }
    }

    DisposableEffect(host, webView) {
        host.controller = object : WebViewController {
            override fun reload() {
                webView.reload()
            }

            override fun goBack() {
                if (webView.canGoBack) webView.goBack() else onNavigateOut()
            }
        }
        onDispose {
            host.controller = null
            webView.setNavigationDelegate(null)
            webView.stopLoading()
        }
    }

    // 大屏 H5 是 SPA，内部跳转只改 hash。跟 Android 一样只比文档基址，
    // 否则每次 recomposition 都会把用户打回入口页。
    LaunchedEffect(url, webView) {
        val currentDoc = webView.URL?.absoluteString?.substringBefore('#')
        if (currentDoc.isNullOrBlank() || currentDoc != url.substringBefore('#')) {
            NSURL.URLWithString(url)?.let { webView.loadRequest(NSURLRequest.requestWithURL(it)) }
        }
    }

    UIKitView(
        factory = { webView },
        modifier = modifier,
    )
}

@OptIn(ExperimentalForeignApi::class)
private class WebNavigationDelegate(
    private val host: WebViewHost,
) : NSObject(), WKNavigationDelegateProtocol {

    @ObjCSignatureOverride
    override fun webView(webView: WKWebView, didStartProvisionalNavigation: WKNavigation?) {
        host.loading = true
        host.failed = false
    }

    @ObjCSignatureOverride
    override fun webView(webView: WKWebView, didFinishNavigation: WKNavigation?) {
        host.loading = false
        host.canGoBack = webView.canGoBack
    }

    @ObjCSignatureOverride
    override fun webView(
        webView: WKWebView,
        didFailProvisionalNavigation: WKNavigation?,
        withError: NSError,
    ) {
        host.loading = false
        host.failed = true
    }

    @ObjCSignatureOverride
    override fun webView(webView: WKWebView, didFailNavigation: WKNavigation?, withError: NSError) {
        host.loading = false
        host.failed = true
    }
}

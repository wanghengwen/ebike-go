package com.luopingtech.ebike.ops.ui.h5

import android.annotation.SuppressLint
import android.graphics.Bitmap
import android.view.ViewGroup
import android.webkit.WebChromeClient
import android.webkit.WebResourceError
import android.webkit.WebResourceRequest
import android.webkit.WebSettings
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.activity.compose.BackHandler
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.getValue
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.viewinterop.AndroidView

@SuppressLint("SetJavaScriptEnabled")
@Composable
actual fun PlatformWebView(
    url: String,
    host: WebViewHost,
    onNavigateOut: () -> Unit,
    modifier: Modifier,
) {
    var webView by remember { mutableStateOf<WebView?>(null) }

    DisposableEffect(host) {
        onDispose { host.controller = null }
    }

    // 管理类 H5 是多页流程（列表→详情、多步向导），系统返回键要退 H5 的一页，
    // 而不是把整个 WebView 关掉。
    BackHandler {
        val view = webView
        if (view != null && view.canGoBack()) view.goBack() else onNavigateOut()
    }

    AndroidView(
        factory = { context ->
            WebView(context).apply {
                layoutParams = ViewGroup.LayoutParams(
                    ViewGroup.LayoutParams.MATCH_PARENT,
                    ViewGroup.LayoutParams.MATCH_PARENT,
                )
                settings.javaScriptEnabled = true
                settings.domStorageEnabled = true
                settings.cacheMode = WebSettings.LOAD_DEFAULT
                settings.mixedContentMode = WebSettings.MIXED_CONTENT_COMPATIBILITY_MODE
                webChromeClient = WebChromeClient()
                webViewClient = object : WebViewClient() {
                    override fun onPageStarted(view: WebView?, url: String?, favicon: Bitmap?) {
                        host.loading = true
                        host.failed = false
                    }

                    override fun onPageFinished(view: WebView?, url: String?) {
                        host.loading = false
                        host.canGoBack = view?.canGoBack() == true
                    }

                    override fun onReceivedError(
                        view: WebView?,
                        request: WebResourceRequest?,
                        error: WebResourceError?,
                    ) {
                        if (request?.isForMainFrame == true) {
                            host.loading = false
                            host.failed = true
                        }
                    }
                }
                webView = this
                // 必须显式持有，否则匿名对象里的 reload() 解析成它自己而不是 WebView 的。
                val self = this
                host.controller = object : WebViewController {
                    // 失败后 url 仍是目标地址，reload 比重新 loadUrl 更贴近「重试」。
                    override fun reload() = self.reload()
                    override fun goBack() = self.goBack()
                }
                loadUrl(url)
            }
        },
        modifier = modifier,
        update = { view ->
            webView = view
            // SPA 会改 hash（大屏内跳转），view.url 随之变化。
            // 若拿完整 URL 比较，Compose recomposition 会把用户打回入口页。
            // 只在「文档基址」（# 之前）变化时才重新 load。
            val currentDoc = view.url?.substringBefore('#')
            val targetDoc = url.substringBefore('#')
            if (currentDoc.isNullOrBlank() || currentDoc != targetDoc) {
                view.loadUrl(url)
            }
        },
    )
}

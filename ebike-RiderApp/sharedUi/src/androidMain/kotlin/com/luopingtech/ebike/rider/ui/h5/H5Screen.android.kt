package com.luopingtech.ebike.rider.ui.h5

import android.annotation.SuppressLint
import android.content.Intent
import android.graphics.Bitmap
import android.net.Uri
import android.view.ViewGroup
import android.webkit.JavascriptInterface
import android.webkit.WebChromeClient
import android.webkit.WebResourceError
import android.webkit.WebResourceRequest
import android.webkit.WebSettings
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.activity.compose.BackHandler
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.viewinterop.AndroidView
import com.luopingtech.ebike.rider.core.h5.NativeHostBridge
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch

@SuppressLint("SetJavaScriptEnabled")
@Composable
actual fun PlatformWebView(
    url: String,
    host: WebViewHost,
    bridge: NativeHostBridge,
    onNavigateOut: () -> Unit,
    openEpoch: Int,
    backEnabled: Boolean,
    modifier: Modifier,
) {
    var webView by remember { mutableStateOf<WebView?>(null) }
    val scope = rememberCoroutineScope()
    // 每次从功能入口打开都记一次入口；栈底再 back 应关容器。
    val entryUrl = remember(openEpoch, url.substringBefore('#'), H5BackPolicy.hashPath(url)) { url }
    val session = remember {
        object {
            var appliedEpoch: Int = -1
            var clearHistoryAfterLoad: Boolean = false
        }
    }

    DisposableEffect(host) {
        onDispose { host.controller = null }
    }

    val webViewFallbackBack: (WebView) -> Unit = fallback@{ view ->
        // 桥脚本未就绪时的兜底：仍用 history，但退回本次入口不关容器
        if (H5BackPolicy.atEntry(view.url, entryUrl) || !view.canGoBack()) {
            onNavigateOut()
            return@fallback
        }
        view.goBack()
        view.post {
            if (H5BackPolicy.shouldCloseAfterBack(view.url, entryUrl)) {
                onNavigateOut()
            } else {
                host.canGoBack = view.canGoBack() && !H5BackPolicy.atEntry(view.url, entryUrl)
            }
        }
    }

    val handleBack: () -> Unit = {
        val view = webView
        if (view == null) {
            onNavigateOut()
        } else {
            view.evaluateJavascript(H5HostBackJs.SCRIPT) { raw ->
                when (H5HostBackJs.parseResult(raw)) {
                    "kept" -> {
                        // UniApp 已 navigateBack；栈深未知时保守认为还能再退
                        host.canGoBack = true
                    }
                    "close" -> onNavigateOut()
                    else -> webViewFallbackBack(view)
                }
            }
        }
    }

    BackHandler(enabled = backEnabled) { handleBack() }

    AndroidView(
        factory = { context ->
            WebView(context).apply {
                layoutParams = ViewGroup.LayoutParams(
                    ViewGroup.LayoutParams.MATCH_PARENT,
                    ViewGroup.LayoutParams.MATCH_PARENT,
                )
                settings.javaScriptEnabled = true
                settings.domStorageEnabled = true
                // H5 是持续迭代的外部产物；缓存住 index.html / chunk 会让线上更新长期不生效。
                settings.cacheMode = WebSettings.LOAD_NO_CACHE
                settings.mixedContentMode = WebSettings.MIXED_CONTENT_COMPATIBILITY_MODE
                clearCache(true)
                webChromeClient = WebChromeClient()
                webViewClient = object : WebViewClient() {
                    override fun onPageStarted(view: WebView?, pageUrl: String?, favicon: Bitmap?) {
                        // 仅首屏 / 显式 reload 打 loading；SPA 回退也会回调，再打会闪进度圈
                        if (!host.initialLoadDone) {
                            host.loading = true
                        }
                        host.failed = false
                    }

                    override fun onPageFinished(view: WebView?, pageUrl: String?) {
                        host.loading = false
                        host.initialLoadDone = true
                        if (session.clearHistoryAfterLoad) {
                            session.clearHistoryAfterLoad = false
                            view?.clearHistory()
                        }
                        val atEntry = H5BackPolicy.atEntry(pageUrl, entryUrl)
                        host.canGoBack = view?.canGoBack() == true && !atEntry
                        view?.evaluateJavascript(H5HostChromeJs.bootstrap("android"), null)
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
                setBackgroundColor(android.graphics.Color.WHITE)
                val self = this
                bridge.currentPageUrl = { self.url }
                bridge.evaluateJs = { js ->
                    self.post { self.evaluateJavascript(js, null) }
                }
                bridge.hooks.openNavigation = { lat, lng, name ->
                    openAndroidMaps(self.context, lat, lng, name)
                }
                addJavascriptInterface(
                    RiderNativeJsInterface(bridge, scope) { id, result ->
                        self.post {
                            self.evaluateJavascript(bridge.replyJs(id, result), null)
                        }
                    },
                    "__riderNative",
                )
                evaluateJavascript(H5HostChromeJs.bootstrap("android"), null)
                webView = this
                host.controller = object : WebViewController {
                    override fun reload() = self.reload()
                    override fun goBack() = handleBack()
                }
                session.appliedEpoch = openEpoch
                session.clearHistoryAfterLoad = true
                host.initialLoadDone = false
                host.loading = true
                loadUrl(url)
            }
        },
        modifier = modifier,
        update = { view ->
            webView = view
            bridge.currentPageUrl = { view.url }
            host.controller = object : WebViewController {
                override fun reload() = view.reload()
                override fun goBack() = handleBack()
            }
            // 每次从原生入口打开：整页 load 到入口并清 history，避免仍停在上次子页。
            if (openEpoch != session.appliedEpoch) {
                session.appliedEpoch = openEpoch
                session.clearHistoryAfterLoad = true
                host.initialLoadDone = false
                host.loading = true
                host.failed = false
                view.loadUrl(url)
                return@AndroidView
            }
            // SPA：同文档只改 hash 时也要跳（父级打开不同长尾页且 epoch 未变的兜底）
            val currentDoc = view.url?.substringBefore('#')
            val targetDoc = url.substringBefore('#')
            val currentHash = H5BackPolicy.hashPath(view.url)
            val targetHash = H5BackPolicy.hashPath(url)
            when {
                currentDoc.isNullOrBlank() || currentDoc != targetDoc -> view.loadUrl(url)
                currentHash != targetHash -> {
                    // 用 assign 进 history，便于容器「＜」按页回退
                    val escaped = url.replace("\\", "\\\\").replace("'", "\\'")
                    view.evaluateJavascript("window.location.assign('$escaped');", null)
                }
            }
        },
    )
}

private fun openAndroidMaps(context: android.content.Context, lat: Double, lng: Double, name: String) {
    val label = Uri.encode(name.ifBlank { "$lat,$lng" })
    val geo = Uri.parse("geo:$lat,$lng?q=$lat,$lng($label)")
    try {
        context.startActivity(
            Intent(Intent.ACTION_VIEW, geo).addFlags(Intent.FLAG_ACTIVITY_NEW_TASK),
        )
    } catch (_: Exception) {
        val web = Uri.parse("https://maps.google.com/?q=$lat,$lng")
        context.startActivity(
            Intent(Intent.ACTION_VIEW, web).addFlags(Intent.FLAG_ACTIVITY_NEW_TASK),
        )
    }
}

private class RiderNativeJsInterface(
    private val bridge: NativeHostBridge,
    private val scope: CoroutineScope,
    private val reply: (id: String, resultJson: String) -> Unit,
) {
    @JavascriptInterface
    fun invoke(payload: String) {
        // JS 接口回调不在主线程；拍照 / Activity Result 必须在 Main 启动。
        scope.launch(Dispatchers.Main) {
            val (id, result) = bridge.handlePayload(payload)
            reply(id, result)
        }
    }

    @JavascriptInterface
    fun platform(): String = bridge.platform()
}

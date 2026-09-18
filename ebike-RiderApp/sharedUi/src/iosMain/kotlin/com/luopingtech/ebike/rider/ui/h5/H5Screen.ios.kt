package com.luopingtech.ebike.rider.ui.h5

import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import androidx.compose.ui.viewinterop.UIKitView
import com.luopingtech.ebike.rider.core.h5.NativeHostBridge
import kotlinx.cinterop.ExperimentalForeignApi
import kotlinx.cinterop.ObjCSignatureOverride
import kotlinx.cinterop.readValue
import kotlinx.coroutines.launch
import platform.CoreGraphics.CGRectZero
import platform.Foundation.NSError
import platform.Foundation.NSURL
import platform.Foundation.NSURLRequest
import platform.UIKit.UIApplication
import platform.WebKit.WKNavigation
import platform.WebKit.WKNavigationDelegateProtocol
import platform.WebKit.WKScriptMessage
import platform.WebKit.WKScriptMessageHandlerProtocol
import platform.WebKit.WKUserContentController
import platform.WebKit.WKWebView
import platform.WebKit.WKWebViewConfiguration
import platform.darwin.NSObject

@OptIn(ExperimentalForeignApi::class)
@Composable
@Suppress("UNUSED_PARAMETER")
actual fun PlatformWebView(
    url: String,
    host: WebViewHost,
    bridge: NativeHostBridge,
    onNavigateOut: () -> Unit,
    openEpoch: Int,
    backEnabled: Boolean,
    modifier: Modifier,
) {
    val scope = rememberCoroutineScope()
    val navigationDelegate = remember { WebNavigationDelegate(host) }
    val messageHandler = remember {
        RiderNativeMessageHandler { payload ->
            scope.launch {
                val (id, result) = bridge.handlePayload(payload)
                bridge.evaluateJs(bridge.replyJs(id, result))
            }
        }
    }
    val webView = remember {
        val configuration = WKWebViewConfiguration().apply {
            allowsInlineMediaPlayback = true
            userContentController.addScriptMessageHandler(messageHandler, "riderNative")
        }
        WKWebView(frame = CGRectZero.readValue(), configuration = configuration).apply {
            setNavigationDelegate(navigationDelegate)
            allowsBackForwardNavigationGestures = true
            setOpaque(false)
        }
    }
    val entryUrl = remember(openEpoch, url.substringBefore('#'), H5BackPolicy.hashPath(url)) { url }

    DisposableEffect(host, webView, bridge) {
        bridge.currentPageUrl = { webView.URL?.absoluteString }
        bridge.evaluateJs = { js ->
            webView.evaluateJavaScript(js, null)
        }
        bridge.hooks.openNavigation = { lat, lng, name ->
            openIosMaps(lat, lng, name)
        }
        host.controller = object : WebViewController {
            override fun reload() {
                webView.reload()
            }

            override fun goBack() {
                webView.evaluateJavaScript(H5HostBackJs.SCRIPT) { raw, _ ->
                    when (H5HostBackJs.parseResult(raw as? String)) {
                        "kept" -> {
                            host.canGoBack = true
                        }
                        "close" -> onNavigateOut()
                        else -> {
                            val current = webView.URL?.absoluteString
                            if (H5BackPolicy.atEntry(current, entryUrl) || !webView.canGoBack) {
                                onNavigateOut()
                            } else {
                                webView.goBack()
                                scope.launch {
                                    kotlinx.coroutines.delay(50)
                                    val landed = webView.URL?.absoluteString
                                    if (H5BackPolicy.shouldCloseAfterBack(landed, entryUrl)) {
                                        onNavigateOut()
                                    } else {
                                        host.canGoBack = webView.canGoBack &&
                                            !H5BackPolicy.atEntry(landed, entryUrl)
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
        onDispose {
            host.controller = null
            bridge.evaluateJs = {}
            webView.configuration.userContentController.removeScriptMessageHandlerForName("riderNative")
            webView.setNavigationDelegate(null)
            webView.stopLoading()
        }
    }

    // openEpoch 变化：强制整页加载入口，丢掉上次子页；仅 URL 变化时再按文档/hash 跳转
    LaunchedEffect(openEpoch, url, webView) {
        val current = webView.URL?.absoluteString
        val currentDoc = current?.substringBefore('#')
        val targetDoc = url.substringBefore('#')
        val epochChanged = navigationDelegate.consumeOpenEpoch(openEpoch)
        when {
            epochChanged || currentDoc.isNullOrBlank() || currentDoc != targetDoc -> {
                host.initialLoadDone = false
                host.loading = true
                host.failed = false
                NSURL.URLWithString(url)?.let { webView.loadRequest(NSURLRequest.requestWithURL(it)) }
            }
            H5BackPolicy.hashPath(current) != H5BackPolicy.hashPath(url) -> {
                val escaped = url.replace("\\", "\\\\").replace("'", "\\'")
                webView.evaluateJavaScript("window.location.assign('$escaped');", null)
            }
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

    private var appliedOpenEpoch: Int = -1

    /** @return true 表示这是一次新的入口打开，需要整页加载。 */
    fun consumeOpenEpoch(epoch: Int): Boolean {
        if (epoch == appliedOpenEpoch) return false
        appliedOpenEpoch = epoch
        return true
    }

    @ObjCSignatureOverride
    override fun webView(webView: WKWebView, didStartProvisionalNavigation: WKNavigation?) {
        if (!host.initialLoadDone) {
            host.loading = true
        }
        host.failed = false
    }

    @ObjCSignatureOverride
    override fun webView(webView: WKWebView, didFinishNavigation: WKNavigation?) {
        host.loading = false
        host.initialLoadDone = true
        host.canGoBack = webView.canGoBack
        webView.evaluateJavaScript(H5HostChromeJs.bootstrap("ios"), null)
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

@OptIn(ExperimentalForeignApi::class)
private class RiderNativeMessageHandler(
    private val onMessage: (String) -> Unit,
) : NSObject(), WKScriptMessageHandlerProtocol {
    override fun userContentController(
        userContentController: WKUserContentController,
        didReceiveScriptMessage: WKScriptMessage,
    ) {
        val body = didReceiveScriptMessage.body
        onMessage(body as? String ?: body.toString())
    }
}

private fun openIosMaps(lat: Double, lng: Double, name: String) {
    val q = if (name.isBlank()) "$lat,$lng" else name
    val encoded = q.replace(" ", "+")
    val raw = "http://maps.apple.com/?ll=$lat,$lng&q=$encoded"
    val nsUrl = NSURL.URLWithString(raw) ?: return
    UIApplication.sharedApplication.openURL(nsUrl)
}

package com.luopingtech.ebike.rider.ui.h5

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.Stable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.rider.core.h5.NativeHostBridge

@Stable
class WebViewHost {
    var loading: Boolean by mutableStateOf(true)
        internal set

    var failed: Boolean by mutableStateOf(false)
        internal set

    var canGoBack: Boolean by mutableStateOf(false)
        internal set

    /** 首屏加载完后，SPA 内跳转不再打 loading，避免回退时闪进度圈。 */
    internal var initialLoadDone: Boolean = false

    internal var controller: WebViewController? = null

    fun reload() {
        failed = false
        loading = true
        initialLoadDone = false
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

@Composable
expect fun PlatformWebView(
    url: String,
    host: WebViewHost,
    bridge: NativeHostBridge,
    onNavigateOut: () -> Unit,
    openEpoch: Int = 0,
    /** 为 false 时不注册系统返回拦截（容器已隐藏但仍挂在组合树里）。 */
    backEnabled: Boolean = true,
    modifier: Modifier = Modifier,
)

/**
 * H5 外壳：左「＜」、中标题、右「关闭」。
 * 页内返回键由 H5 在原生宿主下自行隐藏，避免双层返回。
 */
@Composable
fun H5Screen(
    title: String,
    url: String?,
    placeholder: String,
    failText: String,
    retryText: String,
    closeText: String,
    backText: String,
    bridge: NativeHostBridge,
    onClose: () -> Unit,
    openEpoch: Int = 0,
    active: Boolean = true,
    modifier: Modifier = Modifier,
) {
    val host = remember { WebViewHost() }

    Column(
        modifier = modifier
            .fillMaxSize()
            .statusBarsPadding(),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 4.dp, vertical = 4.dp),
        ) {
            TextButton(
                onClick = { host.goBack() },
                modifier = Modifier.align(Alignment.CenterStart),
            ) { Text(backText) }
            Text(
                title,
                style = MaterialTheme.typography.titleMedium,
                modifier = Modifier.align(Alignment.Center),
            )
            TextButton(
                onClick = onClose,
                modifier = Modifier.align(Alignment.CenterEnd),
            ) { Text(closeText) }
        }
        if (url.isNullOrBlank()) {
            Box(modifier = Modifier.fillMaxSize().padding(24.dp), contentAlignment = Alignment.Center) {
                Text(placeholder, color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
            return
        }
        Box(modifier = Modifier.fillMaxSize()) {
            PlatformWebView(
                url = url,
                host = host,
                bridge = bridge,
                onNavigateOut = onClose,
                openEpoch = openEpoch,
                backEnabled = active,
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
                    Text(failText, color = MaterialTheme.colorScheme.error)
                    TextButton(onClick = { host.reload() }) { Text(retryText) }
                }
            }
        }
    }
}

package com.luopingtech.ebike.ops.ui.feedback

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.delay

/**
 * iOS 的轻提示，对应 Android 宿主注入的 `Toast.makeText`。
 *
 * 用 Compose 浮层而不是 UIKit 的 HUD：提示内容全部来自共享层，
 * 做成 Compose 就跟着主题和安全区一起走，也省掉一条 UIView 生命周期。
 */
@Composable
fun IosToastHost(content: @Composable () -> Unit) {
    var message by remember { mutableStateOf<String?>(null) }
    // 同一句提示连续弹两次也要重新计时，靠自增序号触发。
    var ticket by remember { mutableIntStateOf(0) }

    LaunchedEffect(ticket) {
        if (message != null) {
            delay(TOAST_DURATION_MS)
            message = null
        }
    }

    CompositionLocalProvider(
        LocalOpsToast provides { text: String ->
            if (text.isNotBlank()) {
                message = text
                ticket += 1
            }
        },
    ) {
        Box(modifier = Modifier.fillMaxSize()) {
            content()
            AnimatedVisibility(
                visible = message != null,
                enter = fadeIn(),
                exit = fadeOut(),
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .padding(bottom = 96.dp, start = 32.dp, end = 32.dp),
            ) {
                Surface(
                    shape = RoundedCornerShape(10.dp),
                    color = MaterialTheme.colorScheme.inverseSurface,
                    contentColor = MaterialTheme.colorScheme.inverseOnSurface,
                ) {
                    Text(
                        text = message.orEmpty(),
                        style = MaterialTheme.typography.bodyMedium,
                        textAlign = TextAlign.Center,
                        modifier = Modifier.padding(horizontal = 16.dp, vertical = 10.dp),
                    )
                }
            }
        }
    }
}

private const val TOAST_DURATION_MS = 2_000L

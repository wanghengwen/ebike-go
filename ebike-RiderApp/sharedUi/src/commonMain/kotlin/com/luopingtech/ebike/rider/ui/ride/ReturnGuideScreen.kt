package com.luopingtech.ebike.rider.ui.ride

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.domain.riding.ReturnGuidePageType
import com.luopingtech.ebike.rider.ui.theme.RiderTheme

/**
 * 定制还车引导（对应 UniApp `customizedReturn`）。
 *
 * 简化版：旧版每种 pageType 配一张示意图，这里只出文字步骤 + 「按好了，再试一次」。
 * 图片资源要等设计给多语言版本，先把流程打通。
 */
@Composable
fun ReturnGuideScreen(
    app: RiderApp,
    pageType: Int,
    onRetryReturn: () -> Unit,
    onApply: () -> Unit,
    onBack: () -> Unit,
    showApplyEntry: Boolean = true,
    modifier: Modifier = Modifier,
) {
    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)

    val instruction = when (pageType) {
        ReturnGuidePageType.DIRECTIONAL -> t(Str.ReturnGuideDirectional)
        ReturnGuidePageType.RFID -> t(Str.ReturnGuideRfid)
        ReturnGuidePageType.HELMET -> t(Str.ReturnGuideHelmet)
        ReturnGuidePageType.KICKSTAND -> t(Str.ReturnGuideKickstand)
        ReturnGuidePageType.CAMERA -> t(Str.ReturnGuideCamera)
        else -> t(Str.ReturnGuideDirectional)
    }

    RideScaffold(
        title = t(Str.ReturnGuideTitle),
        onBack = onBack,
        backLabel = t(Str.Back),
        modifier = modifier,
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(20.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            Surface(
                tonalElevation = 1.dp,
                shape = MaterialTheme.shapes.medium,
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text(
                    text = instruction,
                    style = MaterialTheme.typography.bodyLarge,
                    color = RiderTheme.colors.textPrimary,
                    modifier = Modifier.padding(20.dp),
                )
            }
            Button(onClick = onRetryReturn, modifier = Modifier.fillMaxWidth()) {
                Text(t(Str.RideReturn))
            }
            if (showApplyEntry) {
                TextButton(onClick = onApply, modifier = Modifier.fillMaxWidth()) {
                    Text(t(Str.ReturnGuideApply))
                }
            }
        }
    }
}

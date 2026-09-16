package com.luopingtech.ebike.rider.ui.ride

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.rider.RiderApp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.core.result.RiderResult
import com.luopingtech.ebike.rider.ui.feedback.LocalRiderToast
import com.luopingtech.ebike.rider.ui.theme.RiderTheme
import kotlinx.coroutines.launch

/**
 * 无法还车申诉（对应 UniApp `applyReturn`）。
 *
 * 进屏就问一次 `izCanCameraAudit`：这单可能已经申诉过或不在可申诉窗口内，
 * 让用户拍完照才被拒是旧版最常见的投诉点。
 *
 * 申诉成功后 [com.luopingtech.ebike.rider.feature.riding.RidingFeature] 会置 autoReturn 标记，
 * 回骑行页自动再跑一次还车。
 */
@Composable
fun ApplyReturnScreen(
    app: RiderApp,
    applyType: Int,
    onSubmitted: () -> Unit,
    onBack: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val scope = rememberCoroutineScope()
    val toast = LocalRiderToast.current
    var localPhotos by remember { mutableStateOf<List<String>>(emptyList()) }
    var reason by remember { mutableStateOf("") }
    var busy by remember { mutableStateOf(false) }
    var allowed by remember { mutableStateOf<Boolean?>(null) }

    fun t(key: Str, vararg args: Any?) = app.i18n.t(key, *args)

    LaunchedEffect(applyType) {
        allowed = app.ridingFeature.canApplyReturn(applyType)
    }

    RideScaffold(
        title = t(Str.ApplyReturnTitle),
        onBack = onBack,
        backLabel = t(Str.Back),
        busy = busy,
        busyLabel = t(Str.Loading),
        modifier = modifier,
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(20.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            when (allowed) {
                null -> Text(
                    text = t(Str.Loading),
                    style = MaterialTheme.typography.bodyMedium,
                    color = RiderTheme.colors.textSecondary,
                )

                false -> Text(
                    text = t(Str.ApplyReturnBlocked),
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.error,
                )

                true -> {
                    Text(
                        text = t(Str.ApplyReturnHint),
                        style = MaterialTheme.typography.bodyMedium,
                        color = RiderTheme.colors.textSecondary,
                    )

                    Row(
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                        modifier = Modifier.fillMaxWidth(),
                    ) {
                        OutlinedButton(
                            enabled = !busy && localPhotos.size < MAX_PHOTOS,
                            onClick = {
                                scope.launch {
                                    when (val shot = app.photoCapture.takePhoto("apply-return")) {
                                        is RiderResult.Ok -> localPhotos = localPhotos + shot.value
                                        is RiderResult.Err -> toast(shot.error.message)
                                    }
                                }
                            },
                        ) {
                            Text(t(Str.ApplyReturnAddPhoto))
                        }
                        Surface(tonalElevation = 1.dp, shape = MaterialTheme.shapes.small) {
                            Text(
                                text = "${localPhotos.size} / $MAX_PHOTOS",
                                style = MaterialTheme.typography.bodySmall,
                                color = RiderTheme.colors.textSecondary,
                                modifier = Modifier.padding(horizontal = 12.dp, vertical = 10.dp),
                            )
                        }
                    }

                    OutlinedTextField(
                        value = reason,
                        onValueChange = { reason = it.take(MAX_REASON_CHARS) },
                        label = { Text(t(Str.ApplyReturnDesc)) },
                        minLines = 3,
                        modifier = Modifier.fillMaxWidth(),
                    )

                    Button(
                        onClick = {
                            scope.launch {
                                busy = true
                                // 本地路径要先换成 CDN key，后端只收 URL。
                                when (val uploaded = app.mediaUploader.upload(localPhotos)) {
                                    is RiderResult.Err -> toast(uploaded.error.message)
                                    is RiderResult.Ok -> {
                                        val ok = app.ridingFeature.submitApplyReturn(
                                            photoUrls = uploaded.value,
                                            applyType = applyType,
                                            reason = reason.trim(),
                                        )
                                        if (ok) onSubmitted()
                                    }
                                }
                                busy = false
                            }
                        },
                        enabled = !busy && localPhotos.isNotEmpty() && reason.isNotBlank(),
                        modifier = Modifier.fillMaxWidth(),
                    ) {
                        Text(t(Str.ApplyReturnSubmit))
                    }
                }
            }
        }
    }
}

private const val MAX_PHOTOS = 2
private const val MAX_REASON_CHARS = 50

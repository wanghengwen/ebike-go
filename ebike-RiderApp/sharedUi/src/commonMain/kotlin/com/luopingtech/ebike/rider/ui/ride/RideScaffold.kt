package com.luopingtech.ebike.rider.ui.ride

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.luopingtech.ebike.rider.core.i18n.Str
import com.luopingtech.ebike.rider.ui.theme.RiderTheme

/**
 * 骑行域各屏共用的标题栏 + 忙碌遮罩。
 *
 * 遮罩要盖住整屏而不只是按钮：还车 / 开锁在途时任何一次误触都可能建出第二张单，
 * `RidingFeature` 的 mutex 只保证串行，不保证用户不排队。
 */
@Composable
internal fun RideScaffold(
    title: String,
    onBack: (() -> Unit)?,
    busy: Boolean = false,
    busyLabel: String = "",
    backLabel: String = "",
    modifier: Modifier = Modifier,
    trailing: @Composable () -> Unit = {},
    content: @Composable () -> Unit,
) {
    Box(modifier = modifier.fillMaxSize()) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .safeDrawingPadding(),
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 8.dp, vertical = 4.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                if (onBack != null) {
                    TextButton(onClick = onBack) { Text(backLabel) }
                }
                Text(
                    text = title,
                    style = MaterialTheme.typography.titleMedium,
                    color = RiderTheme.colors.textPrimary,
                    modifier = Modifier
                        .weight(1f)
                        .padding(horizontal = 8.dp),
                )
                trailing()
            }
            Box(modifier = Modifier.weight(1f)) { content() }
        }

        if (busy) {
            Surface(
                modifier = Modifier.fillMaxSize(),
                color = RiderTheme.colors.pageBackground.copy(alpha = BUSY_SCRIM_ALPHA),
            ) {
                Column(
                    modifier = Modifier.fillMaxSize(),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center,
                ) {
                    CircularProgressIndicator()
                    if (busyLabel.isNotBlank()) {
                        Text(
                            text = busyLabel,
                            style = MaterialTheme.typography.bodyMedium,
                            color = RiderTheme.colors.textSecondary,
                            textAlign = TextAlign.Center,
                            modifier = Modifier.padding(top = 12.dp),
                        )
                    }
                }
            }
        }
    }
}

/** 键值对一行，确认页 / 结费屏的清单都用它，省掉每处自己排版。 */
@Composable
internal fun RideDetailRow(
    label: String,
    value: String,
    emphasize: Boolean = false,
    modifier: Modifier = Modifier,
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Text(
            text = label,
            style = MaterialTheme.typography.bodyMedium,
            color = RiderTheme.colors.textSecondary,
        )
        Text(
            text = value,
            style = if (emphasize) {
                MaterialTheme.typography.titleMedium
            } else {
                MaterialTheme.typography.bodyMedium
            },
            color = if (emphasize) RiderTheme.colors.primary else RiderTheme.colors.textPrimary,
        )
    }
}

private const val BUSY_SCRIM_ALPHA = 0.86f
